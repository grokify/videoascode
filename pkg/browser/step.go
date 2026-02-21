package browser

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StepAction defines the type of browser action
type StepAction string

const (
	// ActionNavigate navigates to a URL
	ActionNavigate StepAction = "navigate"
	// ActionClick clicks an element
	ActionClick StepAction = "click"
	// ActionInput enters text into an element
	ActionInput StepAction = "input"
	// ActionScroll scrolls the page or element
	ActionScroll StepAction = "scroll"
	// ActionWait waits for a duration
	ActionWait StepAction = "wait"
	// ActionWaitFor waits for an element to appear
	ActionWaitFor StepAction = "waitFor"
	// ActionScreenshot captures a screenshot
	ActionScreenshot StepAction = "screenshot"
	// ActionEvaluate executes JavaScript
	ActionEvaluate StepAction = "evaluate"
	// ActionHover hovers over an element
	ActionHover StepAction = "hover"
	// ActionSelect selects an option from a dropdown
	ActionSelect StepAction = "select"
	// ActionKeypress sends keyboard input
	ActionKeypress StepAction = "keypress"
)

// ScrollMode defines how scroll coordinates are interpreted
type ScrollMode string

const (
	// ScrollModeRelative scrolls by the specified delta (default)
	ScrollModeRelative ScrollMode = "relative"
	// ScrollModeAbsolute scrolls to the specified position
	ScrollModeAbsolute ScrollMode = "absolute"
)

// ScrollBehavior defines how scrolling is animated
type ScrollBehavior string

const (
	// ScrollBehaviorAuto uses instant scrolling (default)
	ScrollBehaviorAuto ScrollBehavior = "auto"
	// ScrollBehaviorSmooth uses smooth animated scrolling
	ScrollBehaviorSmooth ScrollBehavior = "smooth"
)

// TextMatch defines how text content is matched
type TextMatch string

const (
	// TextMatchContains matches if element text contains the search text
	TextMatchContains TextMatch = "contains"
	// TextMatchExact matches if element text equals the search text exactly
	TextMatchExact TextMatch = "exact"
	// TextMatchRegex matches if element text matches the regex pattern
	TextMatchRegex TextMatch = "regex"
)

// Step represents a single browser automation step
type Step struct {
	// Action is the type of action to perform
	Action StepAction `json:"action" yaml:"action"`

	// Selector is the CSS selector or XPath for element targeting
	Selector string `json:"selector,omitempty" yaml:"selector,omitempty"`

	// Value is used for input actions (text to type) or select actions (option value)
	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	// URL is used for navigate actions
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Duration is used for wait actions (milliseconds)
	Duration int `json:"duration,omitempty" yaml:"duration,omitempty"`

	// Script is JavaScript code for evaluate actions
	Script string `json:"script,omitempty" yaml:"script,omitempty"`

	// Key is for keypress actions (e.g., "Enter", "Tab", "Escape")
	Key string `json:"key,omitempty" yaml:"key,omitempty"`

	// Voiceover is the text to speak during this step
	Voiceover string `json:"voiceover,omitempty" yaml:"voiceover,omitempty"`

	// MinDuration ensures the step takes at least this long (milliseconds)
	// Useful for ensuring voiceover completes before next step
	MinDuration int `json:"minDuration,omitempty" yaml:"minDuration,omitempty"`

	// Description is a human-readable description of the step
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// ScrollX and ScrollY are pixel amounts for scroll actions
	ScrollX int `json:"scrollX,omitempty" yaml:"scrollX,omitempty"`
	ScrollY int `json:"scrollY,omitempty" yaml:"scrollY,omitempty"`

	// ScrollMode determines if scroll is relative (delta) or absolute (position)
	// Default is "relative" for backward compatibility
	ScrollMode ScrollMode `json:"scrollMode,omitempty" yaml:"scrollMode,omitempty"`

	// ScrollBehavior determines if scroll is instant ("auto") or animated ("smooth")
	// When "smooth", the step automatically waits for the scroll animation to complete
	ScrollBehavior ScrollBehavior `json:"scrollBehavior,omitempty" yaml:"scrollBehavior,omitempty"`

	// Text enables click-by-text-content instead of selector
	// When set, finds elements containing this text and clicks them
	Text string `json:"text,omitempty" yaml:"text,omitempty"`

	// TextScope restricts text search to elements matching this selector
	// Example: ".sidebar" to only search within sidebar elements
	TextScope string `json:"textScope,omitempty" yaml:"textScope,omitempty"`

	// TextMatch determines how text matching is performed
	// "contains" (default), "exact", or "regex"
	TextMatch TextMatch `json:"textMatch,omitempty" yaml:"textMatch,omitempty"`

	// Timeout overrides the default step timeout (milliseconds)
	Timeout int `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// StepResult contains execution results for a step
type StepResult struct {
	// Step is the original step definition
	Step Step `json:"step"`

	// Index is the step's position in the sequence
	Index int `json:"index"`

	// StartTime is when the step began execution
	StartTime time.Time `json:"startTime"`

	// EndTime is when the step completed
	EndTime time.Time `json:"endTime"`

	// Duration is the actual execution time in milliseconds
	Duration int `json:"durationMs"`

	// Screenshot is the path to the screenshot taken after this step
	Screenshot string `json:"screenshot,omitempty"`

	// Error contains any error message from execution
	Error string `json:"error,omitempty"`

	// Success indicates whether the step completed successfully
	Success bool `json:"success"`
}

// StepSequence represents a sequence of steps with metadata
type StepSequence struct {
	// Name identifies this step sequence
	Name string `json:"name,omitempty"`

	// URL is the starting URL for the sequence
	URL string `json:"url"`

	// Steps is the ordered list of steps to execute
	Steps []Step `json:"steps"`

	// DefaultTimeout is the default timeout for steps (milliseconds)
	DefaultTimeout int `json:"defaultTimeout,omitempty"`
}

// Validate checks if the step is valid
func (s *Step) Validate() error {
	switch s.Action {
	case ActionNavigate:
		if s.URL == "" {
			return fmt.Errorf("navigate action requires URL")
		}
	case ActionClick:
		// Click supports either selector or text-based targeting
		if s.Selector == "" && s.Text == "" {
			return fmt.Errorf("click action requires selector or text")
		}
	case ActionInput, ActionHover, ActionWaitFor, ActionSelect:
		if s.Selector == "" {
			return fmt.Errorf("%s action requires selector", s.Action)
		}
		if s.Action == ActionInput && s.Value == "" {
			return fmt.Errorf("input action requires value")
		}
	case ActionWait:
		if s.Duration <= 0 {
			return fmt.Errorf("wait action requires positive duration")
		}
	case ActionEvaluate:
		if s.Script == "" {
			return fmt.Errorf("evaluate action requires script")
		}
	case ActionKeypress:
		if s.Key == "" {
			return fmt.Errorf("keypress action requires key")
		}
	case ActionScroll:
		// Scroll can have either scrollX, scrollY, or both
		// Validate scrollMode if set
		if s.ScrollMode != "" && s.ScrollMode != ScrollModeRelative && s.ScrollMode != ScrollModeAbsolute {
			return fmt.Errorf("scrollMode must be 'relative' or 'absolute'")
		}
		// Validate scrollBehavior if set
		if s.ScrollBehavior != "" && s.ScrollBehavior != ScrollBehaviorAuto && s.ScrollBehavior != ScrollBehaviorSmooth {
			return fmt.Errorf("scrollBehavior must be 'auto' or 'smooth'")
		}
	case ActionScreenshot:
		// No additional validation needed
	default:
		return fmt.Errorf("unknown action: %s", s.Action)
	}
	return nil
}

// GetEffectiveTimeout returns the timeout to use for this step
func (s *Step) GetEffectiveTimeout(defaultTimeout int) time.Duration {
	if s.Timeout > 0 {
		return time.Duration(s.Timeout) * time.Millisecond
	}
	if defaultTimeout > 0 {
		return time.Duration(defaultTimeout) * time.Millisecond
	}
	return 30 * time.Second // Default 30s timeout
}

// Validate checks if the step sequence is valid
func (seq *StepSequence) Validate() error {
	if seq.URL == "" {
		return fmt.Errorf("step sequence requires URL")
	}
	for i, step := range seq.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step %d: %w", i, err)
		}
	}
	return nil
}

// LoadStepSequence loads a step sequence from a JSON file
func LoadStepSequence(path string) (*StepSequence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read step sequence file: %w", err)
	}

	var seq StepSequence
	if err := json.Unmarshal(data, &seq); err != nil {
		return nil, fmt.Errorf("failed to parse step sequence JSON: %w", err)
	}

	if err := seq.Validate(); err != nil {
		return nil, fmt.Errorf("invalid step sequence: %w", err)
	}

	return &seq, nil
}

// SaveStepSequence saves a step sequence to a JSON file
func (seq *StepSequence) SaveToFile(path string) error {
	data, err := json.MarshalIndent(seq, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal step sequence: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write step sequence file: %w", err)
	}

	return nil
}
