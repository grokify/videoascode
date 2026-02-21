package transcript

import (
	"encoding/json"
	"fmt"
	"os"
)

// Transcript represents the complete transcript for a presentation
type Transcript struct {
	Version  string   `json:"version"`
	Metadata Metadata `json:"metadata"`
	Slides   []Slide  `json:"slides"`
}

// Metadata contains presentation-level settings
type Metadata struct {
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	DefaultLanguage string            `json:"defaultLanguage"`        // BCP-47 code (e.g., "en-US", "en-GB", "fr-CA", "zh-Hans")
	DefaultVoice    VoiceConfig       `json:"defaultVoice"`           // Default voice settings
	DefaultVenue    string            `json:"defaultVenue,omitempty"` // udemy, youtube, coursera, etc.
	Tags            []string          `json:"tags,omitempty"`         // For organization/filtering
	Custom          map[string]string `json:"custom,omitempty"`       // User-defined metadata
}

// VoiceConfig specifies TTS voice settings (compatible with OmniVoice SynthesisConfig)
type VoiceConfig struct {
	Provider        string  `json:"provider,omitempty"`        // elevenlabs, deepgram, etc.
	VoiceID         string  `json:"voiceId"`                   // Provider-specific voice ID
	VoiceName       string  `json:"voiceName,omitempty"`       // Human-readable name
	Model           string  `json:"model,omitempty"`           // Provider-specific model
	OutputFormat    string  `json:"outputFormat,omitempty"`    // mp3, wav, pcm, opus
	SampleRate      int     `json:"sampleRate,omitempty"`      // 22050, 44100, etc.
	Speed           float64 `json:"speed,omitempty"`           // Speech speed multiplier (1.0 = normal)
	Pitch           float64 `json:"pitch,omitempty"`           // Pitch adjustment (-1.0 to 1.0)
	Stability       float64 `json:"stability,omitempty"`       // Voice consistency (0.0 to 1.0)
	SimilarityBoost float64 `json:"similarityBoost,omitempty"` // Voice similarity (0.0 to 1.0)
	Style           float64 `json:"style,omitempty"`           // Style exaggeration (0.0 to 1.0)
}

// SourceType identifies the segment content source
type SourceType string

const (
	// SourceTypeSlide indicates a Marp slide segment
	SourceTypeSlide SourceType = "slide"
	// SourceTypeBrowser indicates a browser-driven demo segment
	SourceTypeBrowser SourceType = "browser"
)

// Slide represents a single slide's transcript data
type Slide struct {
	Index       int                        `json:"index"`
	Title       string                     `json:"title,omitempty"`      // Optional slide title for reference
	SourceType  SourceType                 `json:"sourceType,omitempty"` // slide or browser (defaults to slide)
	Transcripts map[string]LanguageContent `json:"transcripts"`          // Keyed by language code
	Avatar      *AvatarConfig              `json:"avatar,omitempty"`     // Optional avatar/speaker config
	Notes       string                     `json:"notes,omitempty"`      // Internal notes (not spoken)

	// Browser-specific fields (only used when SourceType is "browser")
	BrowserURL   string        `json:"browserUrl,omitempty"`   // Starting URL for browser segment
	BrowserSteps []BrowserStep `json:"browserSteps,omitempty"` // Browser automation steps
}

// BrowserStep represents a single browser automation step in the transcript
type BrowserStep struct {
	// Action is the type of action (navigate, click, input, wait, etc.)
	Action string `json:"action"`

	// Selector is the CSS selector for element actions
	Selector string `json:"selector,omitempty"`

	// Value is used for input actions
	Value string `json:"value,omitempty"`

	// URL is used for navigate actions
	URL string `json:"url,omitempty"`

	// Duration is used for wait actions (milliseconds)
	Duration int `json:"duration,omitempty"`

	// Script is JavaScript code for evaluate actions
	Script string `json:"script,omitempty"`

	// Voiceover is the text to speak during this step
	Voiceover string `json:"voiceover,omitempty"`

	// Description provides context for the step
	Description string `json:"description,omitempty"`

	// ScrollX and ScrollY are pixel amounts for scroll actions
	ScrollX int `json:"scrollX,omitempty"`
	ScrollY int `json:"scrollY,omitempty"`

	// ScrollMode determines if scroll is relative (delta) or absolute (position)
	// Valid values: "relative" (default), "absolute"
	ScrollMode string `json:"scrollMode,omitempty"`

	// ScrollBehavior determines if scroll is instant or animated
	// Valid values: "auto" (instant, default), "smooth" (animated)
	ScrollBehavior string `json:"scrollBehavior,omitempty"`

	// Timing contains timing data after recording
	Timing *StepTimingInfo `json:"timing,omitempty"`
}

// StepTimingInfo contains timing data for a browser step
type StepTimingInfo struct {
	StartMs    int `json:"startMs"`    // Start time relative to segment start
	EndMs      int `json:"endMs"`      // End time relative to segment start
	DurationMs int `json:"durationMs"` // Actual step duration
}

// LanguageContent contains the transcript for one language
type LanguageContent struct {
	Voice    *VoiceConfig `json:"voice,omitempty"`  // Override default voice for this language
	Segments []Segment    `json:"segments"`         // Text segments with timing/effects
	Timing   *TimingInfo  `json:"timing,omitempty"` // Populated after TTS generation
}

// Segment represents a portion of speech with optional effects
type Segment struct {
	Text     string       `json:"text"`               // Text to speak
	Pause    int          `json:"pause,omitempty"`    // Pause after segment (milliseconds)
	Emphasis string       `json:"emphasis,omitempty"` // none, moderate, strong
	Rate     string       `json:"rate,omitempty"`     // slow, medium, fast
	Pitch    string       `json:"pitch,omitempty"`    // low, medium, high, +Xst, -Xst
	Voice    *VoiceConfig `json:"voice,omitempty"`    // Override voice for this segment
	SSML     *SSMLHints   `json:"ssml,omitempty"`     // Additional SSML hints
}

// SSMLHints provides SSML-compatible markup hints
type SSMLHints struct {
	Breaks   []string `json:"breaks,omitempty"`   // e.g., ["400ms", "1s"]
	Emphasis []string `json:"emphasis,omitempty"` // Words to emphasize
	Prosody  string   `json:"prosody,omitempty"`  // Custom prosody settings
	SayAs    string   `json:"sayAs,omitempty"`    // date, time, telephone, etc.
	Phoneme  string   `json:"phoneme,omitempty"`  // IPA pronunciation
	SubAlias string   `json:"subAlias,omitempty"` // Substitution text
}

// TimingInfo contains timing data (populated after TTS generation)
type TimingInfo struct {
	AudioDuration int `json:"audioDuration"` // Audio duration in milliseconds
	PauseDuration int `json:"pauseDuration"` // Total pause duration in milliseconds
	TotalDuration int `json:"totalDuration"` // Total slide duration in milliseconds
}

// AvatarConfig specifies virtual avatar/speaker settings
type AvatarConfig struct {
	Provider string            `json:"provider"`           // heygen, synthesia, d-id, etc.
	AvatarID string            `json:"avatarId"`           // Provider-specific avatar ID
	Position string            `json:"position,omitempty"` // bottom-right, bottom-left, full, pip
	Size     string            `json:"size,omitempty"`     // small, medium, large
	Style    string            `json:"style,omitempty"`    // casual, professional, etc.
	Custom   map[string]string `json:"custom,omitempty"`   // Provider-specific settings
}

// LoadFromFile loads a transcript from a JSON file
func LoadFromFile(path string) (*Transcript, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript file: %w", err)
	}

	var t Transcript
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("failed to parse transcript JSON: %w", err)
	}

	return &t, nil
}

// SaveToFile saves the transcript to a JSON file
func (t *Transcript) SaveToFile(path string) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transcript: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write transcript file: %w", err)
	}

	return nil
}

// GetSlideTranscript returns the transcript for a slide in the specified language
// Falls back to default language if the requested language is not available
func (t *Transcript) GetSlideTranscript(slideIndex int, language string) (*LanguageContent, error) {
	if slideIndex < 0 || slideIndex >= len(t.Slides) {
		return nil, fmt.Errorf("slide index %d out of range", slideIndex)
	}

	slide := t.Slides[slideIndex]

	// Try requested language first
	if content, ok := slide.Transcripts[language]; ok {
		return &content, nil
	}

	// Fall back to default language
	if content, ok := slide.Transcripts[t.Metadata.DefaultLanguage]; ok {
		return &content, nil
	}

	return nil, fmt.Errorf("no transcript found for slide %d in language %s or default %s",
		slideIndex, language, t.Metadata.DefaultLanguage)
}

// GetFullText returns the complete text for a language content (for TTS)
func (lc *LanguageContent) GetFullText() string {
	var text string
	for i, seg := range lc.Segments {
		if i > 0 {
			text += " "
		}
		text += seg.Text
	}
	return text
}

// GetTotalPauseDuration returns the total pause duration in milliseconds
func (lc *LanguageContent) GetTotalPauseDuration() int {
	var total int
	for _, seg := range lc.Segments {
		total += seg.Pause
	}
	return total
}

// GetEffectiveSourceType returns the source type, defaulting to slide for backward compatibility
func (s *Slide) GetEffectiveSourceType() SourceType {
	if s.SourceType == "" {
		return SourceTypeSlide
	}
	return s.SourceType
}

// IsSlideSegment returns true if this is a slide segment
func (s *Slide) IsSlideSegment() bool {
	return s.GetEffectiveSourceType() == SourceTypeSlide
}

// IsBrowserSegment returns true if this is a browser segment
func (s *Slide) IsBrowserSegment() bool {
	return s.GetEffectiveSourceType() == SourceTypeBrowser
}

// GetBrowserVoiceovers returns all voiceover texts from browser steps
func (s *Slide) GetBrowserVoiceovers() []string {
	if !s.IsBrowserSegment() {
		return nil
	}
	voiceovers := make([]string, 0, len(s.BrowserSteps))
	for _, step := range s.BrowserSteps {
		if step.Voiceover != "" {
			voiceovers = append(voiceovers, step.Voiceover)
		}
	}
	return voiceovers
}

// GetBrowserSlides returns only browser-type slides from the transcript
func (t *Transcript) GetBrowserSlides() []Slide {
	slides := make([]Slide, 0)
	for _, slide := range t.Slides {
		if slide.IsBrowserSegment() {
			slides = append(slides, slide)
		}
	}
	return slides
}

// GetSlideSlides returns only slide-type slides from the transcript
func (t *Transcript) GetSlideSlides() []Slide {
	slides := make([]Slide, 0)
	for _, slide := range t.Slides {
		if slide.IsSlideSegment() {
			slides = append(slides, slide)
		}
	}
	return slides
}
