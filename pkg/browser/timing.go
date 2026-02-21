package browser

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TimingData contains timing information for a recorded browser session
type TimingData struct {
	// SessionID uniquely identifies this recording session
	SessionID string `json:"sessionId"`

	// URL is the starting URL
	URL string `json:"url"`

	// StartTime is when the session began
	StartTime time.Time `json:"startTime"`

	// EndTime is when the session completed
	EndTime time.Time `json:"endTime"`

	// TotalDuration is the session duration in milliseconds
	TotalDuration int `json:"totalDurationMs"`

	// Steps contains timing for each step
	Steps []StepTiming `json:"steps"`

	// FrameCount is the total number of frames captured
	FrameCount int `json:"frameCount"`

	// FrameRate is the capture frame rate
	FrameRate int `json:"frameRate"`
}

// StepTiming contains timing information for a single step
type StepTiming struct {
	// Index is the step's position in the sequence
	Index int `json:"index"`

	// Action is the step action type
	Action StepAction `json:"action"`

	// Description provides context for the step
	Description string `json:"description,omitempty"`

	// Voiceover is the text spoken during this step
	Voiceover string `json:"voiceover,omitempty"`

	// StartMs is the start time relative to session start (milliseconds)
	StartMs int `json:"startMs"`

	// EndMs is the end time relative to session start (milliseconds)
	EndMs int `json:"endMs"`

	// DurationMs is the step duration (milliseconds)
	DurationMs int `json:"durationMs"`

	// StartFrame is the first frame index for this step
	StartFrame int `json:"startFrame"`

	// EndFrame is the last frame index for this step
	EndFrame int `json:"endFrame"`

	// Screenshot is the path to the step's screenshot (if captured)
	Screenshot string `json:"screenshot,omitempty"`
}

// NewTimingData creates a new timing data container
func NewTimingData(sessionID, url string, frameRate int) *TimingData {
	return &TimingData{
		SessionID: sessionID,
		URL:       url,
		StartTime: time.Now(),
		FrameRate: frameRate,
		Steps:     make([]StepTiming, 0),
	}
}

// AddStepTiming adds timing data for a step
func (td *TimingData) AddStepTiming(result *StepResult, startFrame, endFrame int) {
	sessionStart := td.StartTime

	timing := StepTiming{
		Index:       result.Index,
		Action:      result.Step.Action,
		Description: result.Step.Description,
		Voiceover:   result.Step.Voiceover,
		StartMs:     int(result.StartTime.Sub(sessionStart).Milliseconds()),
		EndMs:       int(result.EndTime.Sub(sessionStart).Milliseconds()),
		DurationMs:  result.Duration,
		StartFrame:  startFrame,
		EndFrame:    endFrame,
		Screenshot:  result.Screenshot,
	}

	td.Steps = append(td.Steps, timing)
}

// Finalize completes the timing data
func (td *TimingData) Finalize(frameCount int) {
	td.EndTime = time.Now()
	td.TotalDuration = int(td.EndTime.Sub(td.StartTime).Milliseconds())
	td.FrameCount = frameCount
}

// SaveToFile saves timing data to a JSON file
func (td *TimingData) SaveToFile(path string) error {
	data, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal timing data: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write timing data file: %w", err)
	}

	return nil
}

// LoadTimingData loads timing data from a JSON file
func LoadTimingData(path string) (*TimingData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read timing data file: %w", err)
	}

	var td TimingData
	if err := json.Unmarshal(data, &td); err != nil {
		return nil, fmt.Errorf("failed to parse timing data JSON: %w", err)
	}

	return &td, nil
}

// GetStepAtTime returns the step active at a given time (milliseconds from start)
func (td *TimingData) GetStepAtTime(ms int) *StepTiming {
	for i := range td.Steps {
		if ms >= td.Steps[i].StartMs && ms < td.Steps[i].EndMs {
			return &td.Steps[i]
		}
	}
	return nil
}

// GetStepAtFrame returns the step active at a given frame
func (td *TimingData) GetStepAtFrame(frame int) *StepTiming {
	for i := range td.Steps {
		if frame >= td.Steps[i].StartFrame && frame <= td.Steps[i].EndFrame {
			return &td.Steps[i]
		}
	}
	return nil
}

// ToTranscriptSegments converts timing data to transcript-compatible segments
// This is useful for generating voiceover timing from browser recording
func (td *TimingData) ToTranscriptSegments() []TranscriptSegment {
	segments := make([]TranscriptSegment, 0, len(td.Steps))

	for _, step := range td.Steps {
		if step.Voiceover == "" {
			continue
		}

		segments = append(segments, TranscriptSegment{
			Text:    step.Voiceover,
			StartMs: step.StartMs,
			EndMs:   step.EndMs,
		})
	}

	return segments
}

// TranscriptSegment represents a segment for transcript generation
type TranscriptSegment struct {
	Text    string `json:"text"`
	StartMs int    `json:"startMs"`
	EndMs   int    `json:"endMs"`
}

// AdjustForTTSDuration adjusts step timing to accommodate TTS audio duration
// If the TTS audio is longer than the step duration, extend the step
func (td *TimingData) AdjustForTTSDuration(stepIndex int, ttsDurationMs int) {
	if stepIndex < 0 || stepIndex >= len(td.Steps) {
		return
	}

	step := &td.Steps[stepIndex]
	currentDuration := step.DurationMs

	if ttsDurationMs > currentDuration {
		// Extend this step
		extension := ttsDurationMs - currentDuration
		step.EndMs += extension
		step.DurationMs = ttsDurationMs

		// Shift all subsequent steps
		for i := stepIndex + 1; i < len(td.Steps); i++ {
			td.Steps[i].StartMs += extension
			td.Steps[i].EndMs += extension
		}

		// Update total duration
		td.TotalDuration += extension
	}
}
