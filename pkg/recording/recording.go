// Package recording writes a lesson's synthesized narration into a
// course-spec deck's recordings/ directory: the lesson's audio, WebVTT
// captions, and its entry in recordings.json, with each slide's start time
// and narration fingerprint. Course platforms such as AcademyOS play the
// audio with the slides, and use the fingerprints to tell when a slide's
// narration changed after recording.
package recording

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/academyosio/course-spec/deck"
)

// Slide is one slide's narration in a lesson's audio.
type Slide struct {
	// ID is the slide's data-id.
	ID string
	// Title is the slide's title, for chapters.
	Title string
	// Sentences are the slide's narration, for captions.
	Sentences []string
	// SpeechMs is the length of the slide's clip; PauseMs the silence after it.
	SpeechMs int
	PauseMs  int
}

// Timing is a slide's place in the lesson's audio.
type Timing struct {
	StartMs    int
	DurationMs int
}

// Timings returns each slide's start and length in the lesson's audio,
// whose clips are joined in order, each followed by its slide's pause, and
// the audio's total length.
func Timings(slides []Slide) ([]Timing, int) {
	ts := make([]Timing, len(slides))
	at := 0
	for i, s := range slides {
		d := max(0, s.SpeechMs) + max(0, s.PauseMs)
		ts[i] = Timing{StartMs: at, DurationMs: d}
		at += d
	}
	return ts, at
}

// Captions returns WebVTT captions with a cue per sentence. Each sentence's
// time is estimated from its share of its slide's characters, as the clips
// carry no word timings.
func Captions(slides []Slide) string {
	var b strings.Builder
	b.WriteString("WEBVTT\n")
	ts, _ := Timings(slides)
	n := 0
	for i, s := range slides {
		total := 0
		for _, t := range s.Sentences {
			total += len([]rune(strings.TrimSpace(t)))
		}
		at := float64(ts[i].StartMs)
		for _, t := range s.Sentences {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			share := float64(len([]rune(t))) / float64(max(1, total))
			end := at + share*float64(max(0, s.SpeechMs))
			n++
			fmt.Fprintf(&b, "\n%d\n%s --> %s\n%s\n", n, vttTime(int(at+0.5)), vttTime(int(end+0.5)), t)
			at = end
		}
	}
	return b.String()
}

func vttTime(ms int) string {
	return fmt.Sprintf("%02d:%02d:%02d.%03d", ms/3600000, ms/60000%60, ms/1000%60, ms%1000)
}

// YouTube's chapter rules: the first at 0:00, at least three, each at
// least ten seconds long.
const (
	minChapters  = 3
	minChapterMs = 10000
)

// Chapters returns YouTube chapter lines, "m:ss Title", one per slide, with
// slides shorter than ten seconds joined to the chapter before. offsetMs is
// when the first slide's narration starts in the video. It returns nil when
// the lesson can't have three chapters.
func Chapters(slides []Slide, offsetMs int) []string {
	ts, total := Timings(slides)
	type chapter struct {
		startMs int
		title   string
	}
	var cs []chapter
	for i, s := range slides {
		title := firstNonEmpty(s.Title, s.ID)
		start := ts[i].StartMs + offsetMs
		if i == 0 {
			start = 0
		}
		// A short slide, or one too soon after the last chapter, is part of it.
		if len(cs) > 0 && (ts[i].DurationMs < minChapterMs || start-cs[len(cs)-1].startMs < minChapterMs) {
			continue
		}
		cs = append(cs, chapter{start, title})
	}
	// The last chapter must also last ten seconds.
	for len(cs) > 1 && total+offsetMs-cs[len(cs)-1].startMs < minChapterMs {
		cs = cs[:len(cs)-1]
	}
	if len(cs) < minChapters {
		return nil
	}
	lines := make([]string, len(cs))
	for i, c := range cs {
		s := c.startMs / 1000
		if s >= 3600 {
			lines[i] = fmt.Sprintf("%d:%02d:%02d %s", s/3600, s/60%60, s%60, c.title)
		} else {
			lines[i] = fmt.Sprintf("%d:%02d %s", s/60, s%60, c.title)
		}
	}
	return lines
}

// Lesson is one lesson's recording in one language.
type Lesson struct {
	// DeckPath is the deck's HTML file.
	DeckPath string
	// SectionID is the deck's section ID, the lesson ID course-spec gives a
	// deck without lesson starts; empty when unknown.
	SectionID string
	// Language is the narration's language; DefaultLanguage the deck's own.
	Language        string
	DefaultLanguage string
	// AudioFile is the lesson's joined audio, an MP3.
	AudioFile string
	// Slides are the lesson's slides in order.
	Slides []Slide
	Voice  deck.RecordingVoice
}

// Write copies the lesson's audio into the deck's recordings/ directory as
// <language>/<lesson>.mp3, writes its captions beside it, and sets its entry
// in recordings.json. The lesson is the course-spec lesson holding the
// first slide; its slides must be the ones recorded. It returns the entry.
func Write(l Lesson) (*deck.LessonRecording, error) {
	if len(l.Slides) == 0 {
		return nil, errors.New("the lesson has no slides")
	}
	d, err := deck.ParseFile(l.DeckPath)
	if err != nil {
		return nil, err
	}
	cl, err := courseLesson(d, l.SectionID, l.Slides)
	if err != nil {
		return nil, err
	}
	var translations deck.Narration
	if l.Language != l.DefaultLanguage {
		tp := filepath.Join(filepath.Dir(l.DeckPath), "voiceover", "transcript.json")
		t, err := deck.ReadTranscript(tp)
		if err != nil {
			return nil, fmt.Errorf("read the translations: %w", err)
		}
		translations, _, _ = t.Match(d)
	}
	fps := cl.SlideNarration(l.Language, l.DefaultLanguage, translations)

	ts, total := Timings(l.Slides)
	rec := deck.LessonRecording{
		Lesson:     cl.ID,
		Language:   l.Language,
		Audio:      l.Language + "/" + cl.ID + ".mp3",
		Captions:   l.Language + "/" + cl.ID + ".vtt",
		DurationMs: total,
	}
	if l.Voice != (deck.RecordingVoice{}) {
		v := l.Voice
		rec.Voice = &v
	}
	for i, s := range l.Slides {
		rec.Slides = append(rec.Slides, deck.RecordedSlide{ID: s.ID, StartMs: ts[i].StartMs, DurationMs: ts[i].DurationMs, Narration: fps[i]})
	}
	if err := rec.CheckPaths(); err != nil {
		return nil, err
	}

	deckDir := filepath.Dir(l.DeckPath)
	dir := filepath.Join(deckDir, deck.RecordingsDir)
	if err := os.MkdirAll(filepath.Join(dir, l.Language), 0o755); err != nil { //nolint:gosec // course files are meant to be read
		return nil, err
	}
	if err := copyFile(l.AudioFile, filepath.Join(dir, filepath.FromSlash(rec.Audio))); err != nil {
		return nil, fmt.Errorf("copy the audio: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rec.Captions)), []byte(Captions(l.Slides)), 0o644); err != nil { //nolint:gosec // course files are meant to be read
		return nil, err
	}

	index := filepath.Join(deckDir, filepath.FromSlash(deck.RecordingsFile))
	all, err := deck.ReadRecordings(index)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		all = &deck.Recordings{}
	case err != nil:
		return nil, err
	}
	all.Set(rec)
	if err := all.Write(index); err != nil {
		return nil, err
	}
	return &rec, nil
}

// courseLesson returns the course-spec lesson holding the slides, checking
// they are exactly its slides.
func courseLesson(d *deck.Deck, sectionID string, slides []Slide) (deck.Lesson, error) {
	fallback := firstNonEmpty(sectionID, slides[0].ID)
	for _, l := range d.Lessons(fallback) {
		if len(l.Slides) == 0 || !containsSlide(l, slides[0].ID) {
			continue
		}
		if len(l.Slides) != len(slides) {
			return deck.Lesson{}, fmt.Errorf("lesson %s has %d slides in the deck, but %d were recorded", l.ID, len(l.Slides), len(slides))
		}
		for i, s := range l.Slides {
			if s.ID != slides[i].ID {
				return deck.Lesson{}, fmt.Errorf("lesson %s: slide %d is %q in the deck, but %q was recorded", l.ID, i+1, s.ID, slides[i].ID)
			}
		}
		return l, nil
	}
	return deck.Lesson{}, fmt.Errorf("slide %q is not in the deck", slides[0].ID)
}

func containsSlide(l deck.Lesson, id string) bool {
	for _, s := range l.Slides {
		if s.ID == id {
			return true
		}
	}
	return false
}

// Status returns the recording status of the lesson holding firstSlide, in
// a language: current, stale or missing.
func Status(deckPath, sectionID, firstSlide, lang, defaultLang string) (string, error) {
	d, err := deck.ParseFile(deckPath)
	if err != nil {
		return "", err
	}
	r, err := deck.ReadRecordings(filepath.Join(filepath.Dir(deckPath), filepath.FromSlash(deck.RecordingsFile)))
	if errors.Is(err, fs.ErrNotExist) {
		return deck.RecordingMissing, nil
	}
	if err != nil {
		return "", err
	}
	for _, l := range d.Lessons(firstNonEmpty(sectionID, firstSlide)) {
		if !containsSlide(l, firstSlide) {
			continue
		}
		var translations deck.Narration
		if lang != defaultLang {
			t, err := deck.ReadTranscript(filepath.Join(filepath.Dir(deckPath), "voiceover", "transcript.json"))
			if err != nil {
				return "", err
			}
			translations, _, _ = t.Match(d)
		}
		return l.CheckRecording(r.Find(l.ID, lang), defaultLang, translations).Status, nil
	}
	return deck.RecordingMissing, nil
}

// SectionID reads the ID of the course-spec section holding the deck, from
// the nearest section.json above it; "" when there is none.
func SectionID(deckPath string) (string, error) {
	dir := filepath.Dir(deckPath)
	for range 3 {
		p := filepath.Join(dir, "section.json")
		data, err := os.ReadFile(p) //nolint:gosec // beside the deck the user named
		switch {
		case err == nil:
			var m struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(data, &m); err != nil {
				return "", fmt.Errorf("%s: %w", p, err)
			}
			return m.ID, nil
		case !errors.Is(err, fs.ErrNotExist):
			return "", err
		}
		dir = filepath.Dir(dir)
	}
	return "", nil
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src) //nolint:gosec // a clip this run wrote
	if err != nil {
		return err
	}
	defer func() {
		if cerr := in.Close(); err == nil {
			err = cerr
		}
	}()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644) //nolint:gosec // course files are meant to be read
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		if cerr := out.Close(); cerr != nil {
			return errors.Join(err, cerr)
		}
		return err
	}
	return out.Close()
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
