// Package segment provides abstractions for content units that can be
// rendered to video. Both Marp slides and browser demos implement the
// Segment interface, allowing them to share the same pipeline.
package segment

import (
	"github.com/grokify/marp2video/pkg/transcript"
)

// SourceType identifies the content source type
type SourceType string

const (
	// SourceTypeSlide represents a static slide (e.g., Marp presentation)
	SourceTypeSlide SourceType = "slide"
	// SourceTypeBrowser represents a browser-driven demo
	SourceTypeBrowser SourceType = "browser"
)

// Segment represents a unit of content that can be rendered to video.
// Both slides and browser demos implement this interface.
type Segment interface {
	// GetID returns a unique identifier for this segment (e.g., "segment_001")
	GetID() string

	// GetIndex returns the position of this segment in the sequence (0-based)
	GetIndex() int

	// GetSourceType returns whether this is a slide or browser segment
	GetSourceType() SourceType

	// GetTitle returns an optional human-readable title
	GetTitle() string

	// GetVoiceovers returns the voiceover content for a specific language.
	// Slides typically have one voiceover; browser segments may have multiple
	// (one per step with voiceover text).
	GetVoiceovers(language string) []Voiceover

	// GetLanguages returns all available language codes for this segment
	GetLanguages() []string

	// GetTranscripts returns the raw transcript data for all languages
	GetTranscripts() map[string]transcript.LanguageContent
}

// Voiceover represents a piece of text to be spoken
type Voiceover struct {
	// Index is the order within the segment (0 for slides, 0-N for browser steps)
	Index int

	// Text is the content to speak
	Text string

	// Language is the BCP-47 language code
	Language string

	// Voice overrides the default voice configuration
	Voice *transcript.VoiceConfig

	// Pause is the duration to pause after speaking (milliseconds)
	Pause int

	// MinDuration ensures the voiceover takes at least this long (milliseconds)
	// Useful for browser steps where the action must complete before moving on
	MinDuration int

	// StepIndex links back to the browser step index (only for browser segments)
	StepIndex int
}

// AudioResult contains the result of TTS generation for a segment
type AudioResult struct {
	// SegmentID matches the segment's GetID()
	SegmentID string

	// AudioFiles maps language code to audio file path
	AudioFiles map[string]string

	// Durations maps language code to duration in milliseconds
	Durations map[string]int

	// MaxDuration is the maximum duration across all languages
	MaxDuration int

	// VoiceoverCount is how many voiceovers were generated
	VoiceoverCount int
}

// VideoResult contains the result of video generation for a segment
type VideoResult struct {
	// SegmentID matches the segment's GetID()
	SegmentID string

	// VideoPath is the path to the generated video file
	VideoPath string

	// Duration is the video duration in milliseconds
	Duration int

	// FrameCount is the number of frames in the video
	FrameCount int
}

// TimingInfo contains timing data for synchronization
type TimingInfo struct {
	// StartMs is the start time relative to the full video (milliseconds)
	StartMs int

	// EndMs is the end time relative to the full video (milliseconds)
	EndMs int

	// Duration is the segment duration (milliseconds)
	Duration int

	// VoiceoverTimings contains per-voiceover timing
	VoiceoverTimings []VoiceoverTiming
}

// VoiceoverTiming contains timing for a single voiceover
type VoiceoverTiming struct {
	// Index matches Voiceover.Index
	Index int

	// StartMs relative to segment start
	StartMs int

	// EndMs relative to segment start
	EndMs int

	// Text is the voiceover text
	Text string
}
