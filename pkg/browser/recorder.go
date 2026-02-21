package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/google/uuid"
)

// RecorderConfig configures the browser recorder
type RecorderConfig struct {
	// Width is the browser viewport width
	Width int

	// Height is the browser viewport height
	Height int

	// OutputDir is where recordings are stored
	OutputDir string

	// FrameRate is the capture frame rate
	FrameRate int

	// Headless runs the browser without UI
	Headless bool

	// DefaultTimeout is the default step timeout in milliseconds
	DefaultTimeout int

	// CaptureEveryStep captures a screenshot after every step
	CaptureEveryStep bool

	// UserAgent overrides the browser user agent
	UserAgent string

	// DeviceScaleFactor sets the device pixel ratio
	DeviceScaleFactor float64
}

// DefaultRecorderConfig returns a default recorder configuration
func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		Width:             1920,
		Height:            1080,
		FrameRate:         30,
		Headless:          false,
		DefaultTimeout:    30000,
		CaptureEveryStep:  true,
		DeviceScaleFactor: 1,
	}
}

// Recorder handles browser automation and video capture
type Recorder struct {
	config    RecorderConfig
	browser   *rod.Browser
	page      *rod.Page
	capturer  *Capturer
	timing    *TimingData
	sessionID string
	results   []StepResult
}

// NewRecorder creates a new browser recorder
func NewRecorder(config RecorderConfig) (*Recorder, error) {
	if config.Width == 0 {
		config.Width = 1920
	}
	if config.Height == 0 {
		config.Height = 1080
	}
	if config.FrameRate == 0 {
		config.FrameRate = 30
	}
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30000
	}
	if config.DeviceScaleFactor == 0 {
		config.DeviceScaleFactor = 1
	}

	if config.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create capturer
	captureConfig := CaptureConfig{
		Mode:      CaptureModeScreenshot,
		OutputDir: config.OutputDir,
		Width:     config.Width,
		Height:    config.Height,
		FrameRate: config.FrameRate,
		Quality:   85,
		Format:    "mp4",
	}

	capturer, err := NewCapturer(captureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create capturer: %w", err)
	}

	sessionID := uuid.New().String()[:8]

	return &Recorder{
		config:    config,
		capturer:  capturer,
		sessionID: sessionID,
		results:   make([]StepResult, 0),
	}, nil
}

// Launch starts the browser
func (r *Recorder) Launch() error {
	// Configure launcher
	l := launcher.New().
		Headless(r.config.Headless).
		Set("window-size", fmt.Sprintf("%d,%d", r.config.Width, r.config.Height))

	// Add user agent if specified
	if r.config.UserAgent != "" {
		l = l.Set("user-agent", r.config.UserAgent)
	}

	u := l.MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	r.browser = browser

	return nil
}

// Navigate navigates to a URL and creates an initial page
func (r *Recorder) Navigate(url string) error {
	if r.browser == nil {
		return fmt.Errorf("browser not launched")
	}

	page := r.browser.MustPage(url)

	// Set viewport size
	page.MustSetViewport(r.config.Width, r.config.Height, r.config.DeviceScaleFactor, false)

	// Wait for page to load
	page.MustWaitLoad()

	r.page = page
	r.timing = NewTimingData(r.sessionID, url, r.config.FrameRate)

	return nil
}

// RecordSteps executes steps and captures video/screenshots
func (r *Recorder) RecordSteps(ctx context.Context, steps []Step) ([]StepResult, error) {
	if r.page == nil {
		return nil, fmt.Errorf("no page loaded - call Navigate first")
	}

	for i, step := range steps {
		result, err := r.executeStep(ctx, i, step)
		if err != nil {
			result.Error = err.Error()
			result.Success = false
		} else {
			result.Success = true
		}

		// Capture screenshot after step
		if r.config.CaptureEveryStep {
			startFrame := r.capturer.GetFrameCount()
			screenshotPath, captureErr := r.captureScreenshot()
			if captureErr == nil {
				result.Screenshot = screenshotPath
			}
			endFrame := r.capturer.GetFrameCount() - 1

			// Add to timing data
			r.timing.AddStepTiming(result, startFrame, endFrame)
		}

		r.results = append(r.results, *result)

		// Enforce minimum duration if specified
		if step.MinDuration > 0 {
			elapsed := result.Duration
			if elapsed < step.MinDuration {
				remaining := time.Duration(step.MinDuration-elapsed) * time.Millisecond
				select {
				case <-time.After(remaining):
				case <-ctx.Done():
					return r.results, ctx.Err()
				}
			}
		}

		// Stop on error unless context allows continuation
		if !result.Success {
			return r.results, fmt.Errorf("step %d failed: %s", i, result.Error)
		}
	}

	return r.results, nil
}

// executeStep executes a single step
func (r *Recorder) executeStep(ctx context.Context, index int, step Step) (*StepResult, error) {
	result := &StepResult{
		Step:      step,
		Index:     index,
		StartTime: time.Now(),
	}

	timeout := step.GetEffectiveTimeout(r.config.DefaultTimeout)
	page := r.page.Timeout(timeout)

	var err error

	switch step.Action {
	case ActionNavigate:
		err = r.executeNavigate(page, step)
	case ActionClick:
		err = r.executeClick(page, step)
	case ActionInput:
		err = r.executeInput(page, step)
	case ActionScroll:
		err = r.executeScroll(page, step)
	case ActionWait:
		err = r.executeWait(ctx, step)
	case ActionWaitFor:
		err = r.executeWaitFor(page, step)
	case ActionScreenshot:
		result.Screenshot, err = r.captureScreenshot()
	case ActionEvaluate:
		err = r.executeEvaluate(page, step)
	case ActionHover:
		err = r.executeHover(page, step)
	case ActionSelect:
		err = r.executeSelect(page, step)
	case ActionKeypress:
		err = r.executeKeypress(page, step)
	default:
		err = fmt.Errorf("unknown action: %s", step.Action)
	}

	result.EndTime = time.Now()
	result.Duration = int(result.EndTime.Sub(result.StartTime).Milliseconds())

	return result, err
}

func (r *Recorder) executeNavigate(page *rod.Page, step Step) error {
	if err := page.Navigate(step.URL); err != nil {
		return fmt.Errorf("navigate failed: %w", err)
	}
	page.MustWaitLoad()
	return nil
}

func (r *Recorder) executeClick(page *rod.Page, step Step) error {
	var el *rod.Element
	var err error

	if step.Text != "" {
		// Click by text content
		el, err = r.findElementByText(page, step)
		if err != nil {
			return fmt.Errorf("element with text %q not found: %w", step.Text, err)
		}
	} else {
		// Click by selector
		el, err = page.Element(step.Selector)
		if err != nil {
			return fmt.Errorf("element not found: %w", err)
		}
	}

	// Wait for element to be visible and clickable
	if err := el.WaitVisible(); err != nil {
		return fmt.Errorf("element not visible: %w", err)
	}

	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click failed: %w", err)
	}

	return nil
}

// findElementByText finds an element by its text content
func (r *Recorder) findElementByText(page *rod.Page, step Step) (*rod.Element, error) {
	textMatch := step.TextMatch
	if textMatch == "" {
		textMatch = TextMatchContains
	}

	switch textMatch {
	case TextMatchExact:
		return r.findElementByExactText(page, step)
	case TextMatchRegex:
		return r.findElementByRegexText(page, step)
	default: // TextMatchContains
		return r.findElementByContainsText(page, step)
	}
}

// findElementByContainsText finds element containing the specified text
func (r *Recorder) findElementByContainsText(page *rod.Page, step Step) (*rod.Element, error) {
	if step.TextScope != "" {
		// Search within scoped elements using Evaluate to get object reference
		script := fmt.Sprintf(`
			(() => {
				const scope = document.querySelectorAll(%q);
				for (const container of scope) {
					const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, null, false);
					while (walker.nextNode()) {
						if (walker.currentNode.textContent.includes(%q)) {
							let el = walker.currentNode.parentElement;
							// Find clickable parent if current element isn't clickable
							while (el && !['A', 'BUTTON', 'INPUT'].includes(el.tagName) &&
								   !el.onclick && el.getAttribute('role') !== 'button') {
								el = el.parentElement;
							}
							return el || walker.currentNode.parentElement;
						}
					}
				}
				return null;
			})()
		`, step.TextScope, step.Text)

		// Use Evaluate with ByValue=false to get object reference
		opts := &rod.EvalOptions{
			JS:      script,
			ByValue: false,
		}
		result, err := page.Evaluate(opts)
		if err != nil {
			return nil, fmt.Errorf("text search failed: %w", err)
		}

		if result.ObjectID == "" {
			return nil, fmt.Errorf("no element found containing text %q in scope %q", step.Text, step.TextScope)
		}

		return page.ElementFromObject(result)
	}

	// Use Rod's built-in ElementR for simple contains matching
	// ElementR finds elements matching selector with text matching regex
	// We use a regex that matches the literal text
	return page.ElementR("*", step.Text)
}

// findElementByExactText finds element with exactly matching text
func (r *Recorder) findElementByExactText(page *rod.Page, step Step) (*rod.Element, error) {
	scopeSelector := "*"
	if step.TextScope != "" {
		scopeSelector = step.TextScope
	}

	script := fmt.Sprintf(`
		(() => {
			const elements = document.querySelectorAll(%q);
			for (const el of elements) {
				if (el.textContent.trim() === %q) {
					return el;
				}
			}
			return null;
		})()
	`, scopeSelector, step.Text)

	// Use Evaluate with ByValue=false to get object reference
	opts := &rod.EvalOptions{
		JS:      script,
		ByValue: false,
	}
	result, err := page.Evaluate(opts)
	if err != nil {
		return nil, fmt.Errorf("text search failed: %w", err)
	}

	if result.ObjectID == "" {
		return nil, fmt.Errorf("no element found with exact text %q", step.Text)
	}

	return page.ElementFromObject(result)
}

// findElementByRegexText finds element with text matching regex pattern
func (r *Recorder) findElementByRegexText(page *rod.Page, step Step) (*rod.Element, error) {
	scopeSelector := "*"
	if step.TextScope != "" {
		scopeSelector = step.TextScope
	}

	// Use Rod's ElementR which supports regex matching
	return page.ElementR(scopeSelector, step.Text)
}

func (r *Recorder) executeInput(page *rod.Page, step Step) error {
	el, err := page.Element(step.Selector)
	if err != nil {
		return fmt.Errorf("element not found: %w", err)
	}

	// Clear existing content first
	if err := el.SelectAllText(); err == nil {
		el.MustInput("")
	}

	// Type the new value
	if err := el.Input(step.Value); err != nil {
		return fmt.Errorf("input failed: %w", err)
	}

	return nil
}

func (r *Recorder) executeScroll(page *rod.Page, step Step) error {
	fmt.Fprintf(os.Stderr, "executeScroll: scrollX=%d, scrollY=%d, mode=%s, behavior=%s, selector=%q\n",
		step.ScrollX, step.ScrollY, step.ScrollMode, step.ScrollBehavior, step.Selector)

	if step.Selector != "" {
		// Scroll element into view
		el, err := page.Element(step.Selector)
		if err != nil {
			return fmt.Errorf("element not found: %w", err)
		}
		if err := el.ScrollIntoView(); err != nil {
			return fmt.Errorf("scroll into view failed: %w", err)
		}
		return nil
	}

	// Determine scroll mode (default to relative for backward compatibility)
	scrollMode := step.ScrollMode
	if scrollMode == "" {
		scrollMode = ScrollModeRelative
	}

	// Determine scroll behavior (default to auto/instant)
	scrollBehavior := step.ScrollBehavior
	if scrollBehavior == "" {
		scrollBehavior = ScrollBehaviorAuto
	}

	// Get scroll position before
	beforeResult, err := page.Eval(`() => ({x: window.scrollX, y: window.scrollY})`)
	if err == nil && beforeResult != nil {
		fmt.Fprintf(os.Stderr, "executeScroll: before position x=%v, y=%v\n", beforeResult.Value.Get("x"), beforeResult.Value.Get("y"))
	}

	// For React/SPA apps, we need to find the actual scrollable container
	// Try multiple approaches: window, document.documentElement, document.body, and main content divs
	scrollScript := fmt.Sprintf(`() => {
		const scrollY = %d;
		const scrollX = %d;
		const behavior = '%s';
		const isAbsolute = %v;

		// Helper to scroll an element
		const doScroll = (el, name) => {
			if (!el) return false;
			const before = el.scrollTop;
			if (isAbsolute) {
				el.scrollTo({top: scrollY, left: scrollX, behavior: behavior});
			} else {
				el.scrollBy({top: scrollY, left: scrollX, behavior: behavior});
			}
			// Check if scroll happened
			const after = el.scrollTop;
			console.log(name + ': before=' + before + ', after=' + after);
			return after !== before || (isAbsolute && after === scrollY);
		};

		// Try window first
		const windowBefore = window.scrollY;
		if (isAbsolute) {
			window.scrollTo({top: scrollY, left: scrollX, behavior: behavior});
		} else {
			window.scrollBy({top: scrollY, left: scrollX, behavior: behavior});
		}
		if (window.scrollY !== windowBefore || (isAbsolute && window.scrollY === scrollY)) {
			return {scrolled: true, element: 'window', position: window.scrollY};
		}

		// Try document.documentElement
		if (doScroll(document.documentElement, 'documentElement')) {
			return {scrolled: true, element: 'documentElement', position: document.documentElement.scrollTop};
		}

		// Try document.body
		if (doScroll(document.body, 'body')) {
			return {scrolled: true, element: 'body', position: document.body.scrollTop};
		}

		// Try common scrollable container selectors used in React apps
		const selectors = ['main', '[role="main"]', '.main-content', '#root > div', '#app > div', '.content', '.page-content'];
		for (const sel of selectors) {
			const el = document.querySelector(sel);
			if (el && el.scrollHeight > el.clientHeight) {
				if (doScroll(el, sel)) {
					return {scrolled: true, element: sel, position: el.scrollTop};
				}
			}
		}

		return {scrolled: false, element: 'none', position: 0};
	}`, step.ScrollY, step.ScrollX, func() string {
		if scrollBehavior == ScrollBehaviorSmooth {
			return "smooth"
		}
		return "auto"
	}(), scrollMode == ScrollModeAbsolute)

	fmt.Fprintf(os.Stderr, "executeScroll: attempting scroll scrollY=%d, scrollX=%d, mode=%s\n",
		step.ScrollY, step.ScrollX, scrollMode)

	result, err := page.Eval(scrollScript)
	if err != nil {
		return fmt.Errorf("scroll failed: %w", err)
	}

	if result != nil {
		scrolled := result.Value.Get("scrolled").Bool()
		element := result.Value.Get("element").Str()
		position := result.Value.Get("position").Int()
		fmt.Fprintf(os.Stderr, "executeScroll: result scrolled=%v, element=%s, position=%d\n",
			scrolled, element, position)
	}

	// Wait for smooth scroll to complete if applicable
	if scrollBehavior == ScrollBehaviorSmooth {
		if err := r.waitForScrollComplete(page); err != nil {
			return fmt.Errorf("scroll wait failed: %w", err)
		}
	}

	// Small delay to ensure scroll is rendered before screenshot
	time.Sleep(100 * time.Millisecond)

	return nil
}

// waitForScrollComplete waits for a smooth scroll animation to finish
func (r *Recorder) waitForScrollComplete(page *rod.Page) error {
	// Use a polling approach: check if scroll position is stable
	// The scrollend event is not universally supported, so we poll instead
	script := `
		new Promise((resolve) => {
			let lastY = window.scrollY;
			let lastX = window.scrollX;
			let stableCount = 0;
			const checkInterval = 50; // ms
			const stableThreshold = 3; // consecutive stable checks

			const check = () => {
				const currentY = window.scrollY;
				const currentX = window.scrollX;

				if (currentY === lastY && currentX === lastX) {
					stableCount++;
					if (stableCount >= stableThreshold) {
						resolve(true);
						return;
					}
				} else {
					stableCount = 0;
					lastY = currentY;
					lastX = currentX;
				}

				setTimeout(check, checkInterval);
			};

			// Start checking after a short delay to let scroll begin
			setTimeout(check, 50);
		})
	`

	_, err := page.Eval(script)
	return err
}

func (r *Recorder) executeWait(ctx context.Context, step Step) error {
	select {
	case <-time.After(time.Duration(step.Duration) * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Recorder) executeWaitFor(page *rod.Page, step Step) error {
	timeout := step.GetEffectiveTimeout(r.config.DefaultTimeout)
	_, err := page.Timeout(timeout).Element(step.Selector)
	if err != nil {
		return fmt.Errorf("element not found within timeout: %w", err)
	}
	return nil
}

func (r *Recorder) executeEvaluate(page *rod.Page, step Step) error {
	_, err := page.Eval(step.Script)
	if err != nil {
		return fmt.Errorf("script evaluation failed: %w", err)
	}
	return nil
}

func (r *Recorder) executeHover(page *rod.Page, step Step) error {
	el, err := page.Element(step.Selector)
	if err != nil {
		return fmt.Errorf("element not found: %w", err)
	}

	if err := el.Hover(); err != nil {
		return fmt.Errorf("hover failed: %w", err)
	}

	return nil
}

func (r *Recorder) executeSelect(page *rod.Page, step Step) error {
	el, err := page.Element(step.Selector)
	if err != nil {
		return fmt.Errorf("element not found: %w", err)
	}

	if err := el.Select([]string{step.Value}, true, rod.SelectorTypeText); err != nil {
		return fmt.Errorf("select failed: %w", err)
	}

	return nil
}

func (r *Recorder) executeKeypress(page *rod.Page, step Step) error {
	key := keyFromString(step.Key)
	if err := page.Keyboard.Press(key); err != nil {
		return fmt.Errorf("keypress failed: %w", err)
	}
	return nil
}

// captureScreenshot captures the current page as a screenshot
func (r *Recorder) captureScreenshot() (string, error) {
	if r.page == nil {
		return "", fmt.Errorf("no page loaded")
	}

	// Use fullPage=false to capture only the visible viewport (respects scroll position)
	// fullPage=true would capture entire page regardless of scroll
	// Don't use Clip - it uses absolute page coordinates which ignores scroll position
	data, err := r.page.Screenshot(false, &proto.PageCaptureScreenshot{
		Format:      proto.PageCaptureScreenshotFormatPng,
		Quality:     nil, // PNG doesn't use quality
		FromSurface: true,
	})
	if err != nil {
		return "", fmt.Errorf("screenshot failed: %w", err)
	}

	return r.capturer.SaveFrame(data)
}

// GenerateVideo creates the final video from captured frames
func (r *Recorder) GenerateVideo(ctx context.Context) (string, error) {
	r.timing.Finalize(r.capturer.GetFrameCount())

	videoPath := filepath.Join(r.config.OutputDir, fmt.Sprintf("recording_%s.mp4", r.sessionID))

	// Build frame durations from step timing and MinDurations
	frameDurations := r.buildFrameDurations()

	// Use duration-aware video generation if we have timing data
	if len(frameDurations) > 0 {
		if err := r.capturer.GenerateVideoWithDurations(ctx, videoPath, frameDurations); err != nil {
			return "", fmt.Errorf("failed to generate video with durations: %w", err)
		}
	} else {
		if err := r.capturer.GenerateVideo(ctx, videoPath); err != nil {
			return "", fmt.Errorf("failed to generate video: %w", err)
		}
	}

	return videoPath, nil
}

// buildFrameDurations creates a map of frame index to duration in milliseconds
// based on step MinDuration values set from TTS timing
func (r *Recorder) buildFrameDurations() map[int]int {
	durations := make(map[int]int)

	for i, result := range r.results {
		step := result.Step

		// Only set duration if step has a MinDuration (from TTS timing)
		if step.MinDuration > 0 {
			// Find the frame(s) for this step from timing data
			if i < len(r.timing.Steps) {
				timing := r.timing.Steps[i]
				// Assign MinDuration to each frame in this step's range
				// Usually StartFrame == EndFrame (1 frame per step)
				for frame := timing.StartFrame; frame <= timing.EndFrame; frame++ {
					// Divide duration across frames if multiple
					frameCount := timing.EndFrame - timing.StartFrame + 1
					durations[frame] = step.MinDuration / frameCount
				}
			}
		}
	}

	return durations
}

// GetTimingData returns the timing data for the recording
func (r *Recorder) GetTimingData() *TimingData {
	return r.timing
}

// GetResults returns all step results
func (r *Recorder) GetResults() []StepResult {
	return r.results
}

// GetVideoPath returns the path to the generated video
func (r *Recorder) GetVideoPath() string {
	return filepath.Join(r.config.OutputDir, fmt.Sprintf("recording_%s.mp4", r.sessionID))
}

// SaveTimingData saves timing data to a JSON file
func (r *Recorder) SaveTimingData() error {
	timingPath := filepath.Join(r.config.OutputDir, fmt.Sprintf("timing_%s.json", r.sessionID))
	return r.timing.SaveToFile(timingPath)
}

// Close releases browser resources
func (r *Recorder) Close() error {
	if r.browser != nil {
		return r.browser.Close()
	}
	return nil
}

// Cleanup removes temporary files but keeps the final video
func (r *Recorder) Cleanup() error {
	return r.capturer.Cleanup()
}

// keyFromString converts a key name to a Rod input key
func keyFromString(key string) input.Key {
	switch key {
	case "Enter":
		return input.Enter
	case "Tab":
		return input.Tab
	case "Escape":
		return input.Escape
	case "Backspace":
		return input.Backspace
	case "Delete":
		return input.Delete
	case "ArrowUp":
		return input.ArrowUp
	case "ArrowDown":
		return input.ArrowDown
	case "ArrowLeft":
		return input.ArrowLeft
	case "ArrowRight":
		return input.ArrowRight
	case "Home":
		return input.Home
	case "End":
		return input.End
	case "PageUp":
		return input.PageUp
	case "PageDown":
		return input.PageDown
	case "Space":
		return input.Space
	default:
		// For single character keys, use the first rune
		if len(key) == 1 {
			return input.Key(key[0])
		}
		return input.Enter // Default fallback
	}
}
