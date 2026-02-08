package tts

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Manifest contains metadata about generated audio files for use by the video recorder
type Manifest struct {
	Version     string       `json:"version"`
	Language    string       `json:"language"`
	GeneratedAt time.Time    `json:"generatedAt"`
	Slides      []SlideAudio `json:"slides"`
}

// SlideAudio contains audio information for a single slide
type SlideAudio struct {
	Index         int    `json:"index"`
	Title         string `json:"title,omitempty"`
	AudioFile     string `json:"audioFile"`
	AudioDuration int    `json:"audioDurationMs"` // Audio duration in milliseconds
	PauseDuration int    `json:"pauseDurationMs"` // Total pause duration in milliseconds
	TotalDuration int    `json:"totalDurationMs"` // Total slide duration (audio + pauses)
}

// NewManifest creates a new manifest
func NewManifest(language string) *Manifest {
	return &Manifest{
		Version:     "1.0",
		Language:    language,
		GeneratedAt: time.Now().UTC(),
		Slides:      []SlideAudio{},
	}
}

// AddSlide adds a slide audio entry to the manifest
func (m *Manifest) AddSlide(slide SlideAudio) {
	m.Slides = append(m.Slides, slide)
}

// SaveToFile saves the manifest to a JSON file
func (m *Manifest) SaveToFile(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// LoadManifest loads a manifest from a JSON file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	return &m, nil
}

// GetTotalDuration returns the total duration of all slides in milliseconds
func (m *Manifest) GetTotalDuration() int {
	var total int
	for _, slide := range m.Slides {
		total += slide.TotalDuration
	}
	return total
}

// GetSlide returns the audio info for a specific slide index
func (m *Manifest) GetSlide(index int) (*SlideAudio, error) {
	for _, slide := range m.Slides {
		if slide.Index == index {
			return &slide, nil
		}
	}
	return nil, fmt.Errorf("slide %d not found in manifest", index)
}
