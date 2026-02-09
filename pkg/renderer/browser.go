package renderer

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
)

// BrowserController manages the browser for displaying presentations
type BrowserController struct {
	browser *rod.Browser
	page    *rod.Page
	width   int
	height  int
}

// NewBrowserController creates a new browser controller
func NewBrowserController(width, height int) (*BrowserController, error) {
	// Launch browser with custom size
	u := launcher.New().
		Headless(false). // Set to true for headless mode
		Set("window-size", fmt.Sprintf("%d,%d", width, height)).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	return &BrowserController{
		browser: browser,
		width:   width,
		height:  height,
	}, nil
}

// LoadPresentation opens the HTML presentation in the browser
func (bc *BrowserController) LoadPresentation(htmlPath string) error {
	// Convert to absolute path for file:// URL
	absPath, err := filepath.Abs(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	page := bc.browser.MustPage(fmt.Sprintf("file://%s", absPath))

	// Set viewport size
	page.MustSetViewport(bc.width, bc.height, 1, false)

	// Wait for page to load
	page.MustWaitLoad()

	bc.page = page
	return nil
}

// NavigateToSlide navigates to a specific slide using keyboard navigation
func (bc *BrowserController) NavigateToSlide(slideIndex int) error {
	if bc.page == nil {
		return fmt.Errorf("no page loaded")
	}

	// Use page with timeout for operations
	page := bc.page.Timeout(5 * time.Second)

	fmt.Printf("[NAV] Clicking body...\n")
	// Click on the page to ensure it has focus
	body, err := page.Element("body")
	if err != nil {
		return fmt.Errorf("failed to find body element: %w", err)
	}
	if err := body.Click("left", 1); err != nil {
		return fmt.Errorf("failed to click body: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("[NAV] Pressing Home...\n")
	// Navigate to first slide by pressing Home key
	if err := page.Keyboard.Type(input.Home); err != nil {
		return fmt.Errorf("failed to press Home: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Navigate forward to target slide using arrow keys
	fmt.Printf("[NAV] Pressing ArrowRight %d times...\n", slideIndex)
	for i := 0; i < slideIndex; i++ {
		fmt.Printf("[NAV] ArrowRight %d/%d\n", i+1, slideIndex)
		if err := page.Keyboard.Type(input.ArrowRight); err != nil {
			return fmt.Errorf("failed to press ArrowRight: %w", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("[NAV] Navigation done\n")
	return nil
}

// GetWindowInfo returns the browser window position and size for recording
func (bc *BrowserController) GetWindowInfo() (x, y, width, height int, err error) {
	// For now, return configured size
	// TODO: Get actual window position using browser APIs
	return 0, 0, bc.width, bc.height, nil
}

// Close closes the browser
func (bc *BrowserController) Close() error {
	if bc.browser != nil {
		return bc.browser.Close()
	}
	return nil
}

// WaitForDuration waits for a specific duration
func (bc *BrowserController) WaitForDuration(ctx context.Context, duration time.Duration) {
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}
}
