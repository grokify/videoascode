// Package review supports a human review of a deck's narration before it is
// sent to a TTS provider.
//
// A reviewer reads each slide's narration, per language, next to the slide
// itself, and marks it approved or in need of changes. Decisions are kept in a
// review file (voiceover/review.json next to the deck, by default), keyed by
// slide ID and language. Each decision records a hash of the text it was made
// on, so editing a slide's narration, pause or pronunciations makes its
// decision stale and the slide needs review again.
//
// The package builds a view model of a deck for review (Build), runs checks
// that catch text TTS tends to misread (Check), and serves an interactive
// review page (NewHandler) or writes a read-only snapshot of one (Export).
package review

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"
)

// FileVersion is the version of the review file format.
const FileVersion = "1.0"

// DefaultFile is the review file's name, kept next to the transcript.
const DefaultFile = "review.json"

// Status is a reviewer's decision on one slide in one language.
type Status string

const (
	// StatusApproved means the narration is ready for TTS.
	StatusApproved Status = "approved"
	// StatusChanges means the narration needs changes; the note says which.
	StatusChanges Status = "changes"
)

// State is a slide's review state in one language, as shown to a reviewer.
type State string

const (
	// StateUnreviewed means no decision has been made.
	StateUnreviewed State = "unreviewed"
	// StateApproved means the current text is approved.
	StateApproved State = "approved"
	// StateChanges means changes were requested on the current text.
	StateChanges State = "changes"
	// StateStale means a decision was made on text that has since changed.
	StateStale State = "stale"
)

// File holds review decisions for one deck.
type File struct {
	Version string `json:"version"`
	// Slides maps a slide ID to its decisions, keyed by language.
	Slides map[string]map[string]Entry `json:"slides"`
}

// Entry is one review decision.
type Entry struct {
	Status Status `json:"status"`
	// Note explains a request for changes, or anything else worth keeping.
	Note string `json:"note,omitempty"`
	// TextHash identifies the text the decision was made on (see TextHash).
	TextHash string `json:"textHash"`
	// Reviewer optionally names who made the decision.
	Reviewer  string    `json:"reviewer,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewFile returns an empty review file.
func NewFile() *File {
	return &File{Version: FileVersion, Slides: map[string]map[string]Entry{}}
}

// LoadFile reads a review file. A missing file is an empty review.
func LoadFile(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return NewFile(), nil
	}
	if err != nil {
		return nil, err
	}
	f := NewFile()
	if err := json.Unmarshal(b, f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if f.Slides == nil {
		f.Slides = map[string]map[string]Entry{}
	}
	return f, nil
}

// Save writes the review file.
func (f *File) Save(path string) error {
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

// Get returns the decision for a slide and language, if any.
func (f *File) Get(slideID, lang string) (Entry, bool) {
	if f == nil {
		return Entry{}, false
	}
	e, ok := f.Slides[slideID][lang]
	return e, ok
}

// Set records a decision for a slide and language.
func (f *File) Set(slideID, lang string, e Entry) {
	if f.Slides == nil {
		f.Slides = map[string]map[string]Entry{}
	}
	if f.Slides[slideID] == nil {
		f.Slides[slideID] = map[string]Entry{}
	}
	f.Slides[slideID][lang] = e
}

// Clear removes the decision for a slide and language.
func (f *File) Clear(slideID, lang string) {
	delete(f.Slides[slideID], lang)
	if len(f.Slides[slideID]) == 0 {
		delete(f.Slides, slideID)
	}
}

// StateOf returns the review state of text with the given hash.
func (f *File) StateOf(slideID, lang, textHash string) State {
	e, ok := f.Get(slideID, lang)
	switch {
	case !ok:
		return StateUnreviewed
	case e.TextHash != textHash:
		return StateStale
	case e.Status == StatusApproved:
		return StateApproved
	default:
		return StateChanges
	}
}

// TextHash identifies narration as it will be synthesized: the text TTS
// receives, after pronunciation substitutions, and the pause that follows it.
func TextHash(spoken string, pauseMs int) string {
	sum := sha256.Sum256([]byte(spoken + "\x00" + strconv.Itoa(pauseMs)))
	return hex.EncodeToString(sum[:6])
}
