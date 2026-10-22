package htmldeck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/grokify/videoascode/pkg/config"
)

// ManifestFile is the lesson manifest's file name in a lesson output directory.
const ManifestFile = "lessons.json"

// Manifest lists a deck's lessons and their recorded videos, so a course
// platform can import them and later runs can tell which lessons changed.
type Manifest struct {
	// Title is the deck's title.
	Title string `json:"title"`
	// DefaultLanguage is the language of each lesson's primary video.
	DefaultLanguage string `json:"defaultLanguage,omitempty"`
	// Lessons are in deck order.
	Lessons []ManifestLesson `json:"lessons"`
}

// ManifestLesson is one lesson in a Manifest.
type ManifestLesson struct {
	// Number is the lesson's 1-based position in the deck.
	Number int `json:"number"`
	// ID identifies the lesson across runs (see Lesson.ID).
	ID string `json:"id"`
	// Title is the lesson's title.
	Title string `json:"title"`
	// Slides are the IDs of the lesson's slides.
	Slides []string `json:"slides"`
	// Videos maps a language to the lesson's video file, relative to the
	// manifest. Empty until the lesson is recorded.
	Videos map[string]string `json:"videos,omitempty"`
	// DurationMs is the length of the primary video.
	DurationMs int `json:"durationMs,omitempty"`
	// Fingerprint identifies the lesson content the videos were recorded
	// from (see LessonFingerprint).
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Lesson returns the entry with the given ID, or nil.
func (m *Manifest) Lesson(id string) *ManifestLesson {
	for i := range m.Lessons {
		if m.Lessons[i].ID == id {
			return &m.Lessons[i]
		}
	}
	return nil
}

// LoadManifest reads a manifest file.
func LoadManifest(path string) (*Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &m, nil
}

// Save writes the manifest as indented JSON.
func (m *Manifest) Save(path string) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

// LessonFingerprint identifies what a lesson's video is recorded from: the
// HTML of its slides and its config (narration in every language, voice,
// pronunciations, timing, and resolution), ignoring the URL, which changes
// between runs. extra adds settings from outside the config, such as a
// voice chosen on the command line. Changes to the deck's stylesheets,
// scripts, or images are not detected.
func LessonFingerprint(l Lesson, cfg *config.VideoConfig, extra ...string) string {
	c := *cfg
	c.Segments = append([]config.SegmentConfig(nil), cfg.Segments...)
	for i := range c.Segments {
		c.Segments[i].URL = ""
	}
	b, err := json.Marshal(&c)
	if err != nil {
		panic(fmt.Sprintf("marshal video config: %v", err)) // configs hold only plain data
	}
	h := sha256.New()
	h.Write(b)
	for _, s := range l.Slides {
		fmt.Fprintf(h, "\n%s", s.ContentHash)
	}
	for _, e := range extra {
		fmt.Fprintf(h, "\n%s", e)
	}
	return hex.EncodeToString(h.Sum(nil)[:8])
}
