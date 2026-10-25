package htmldeck

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/grokify/videoascode/pkg/browser"
	"github.com/grokify/videoascode/pkg/config"
	"github.com/grokify/videoascode/pkg/transcript"
)

// TranscriptOptions configures Transcript.
type TranscriptOptions struct {
	// Title overrides the deck's <title>.
	Title string
	// Description is optional transcript metadata.
	Description string
	// Language is the language of the deck's narration; default "en-US".
	Language string
	// Voice is the default TTS voice.
	Voice transcript.VoiceConfig
	// Existing is a previously written transcript. Its other languages
	// (translations) are kept, matched to slides by position, and its
	// metadata fields fill any left empty here.
	Existing *transcript.Transcript
}

var sentenceEndRe = regexp.MustCompile(`([.?!])\s+`)

// Transcript returns the deck's narration as a videoascode transcript. The
// deck's language comes from the HTML; other languages come from
// opts.Existing, so regenerating never discards translations. Each slide's
// narration is split into sentences, and its pause follows the last one.
func (d *Deck) Transcript(opts TranscriptOptions) *transcript.Transcript {
	lang := opts.Language
	if lang == "" {
		lang = DefaultLanguage
	}
	t := &transcript.Transcript{
		Version: DefaultTranscriptVers,
		Metadata: transcript.Metadata{
			Title:           firstNonEmpty(opts.Title, d.Title),
			Description:     opts.Description,
			DefaultLanguage: lang,
			DefaultVoice:    opts.Voice,
		},
	}
	if ex := opts.Existing; ex != nil {
		m := ex.Metadata
		t.Metadata.Title = firstNonEmpty(t.Metadata.Title, m.Title)
		t.Metadata.Description = firstNonEmpty(t.Metadata.Description, m.Description)
		if t.Metadata.DefaultVoice.VoiceID == "" {
			t.Metadata.DefaultVoice = m.DefaultVoice
		}
		t.Metadata.DefaultVenue = m.DefaultVenue
		t.Metadata.Tags = m.Tags
		t.Metadata.Custom = m.Custom
	}

	for i, s := range d.Slides {
		lcs := map[string]transcript.LanguageContent{
			lang: {Segments: sentenceSegments(s.Narration, s.Pause)},
		}
		if opts.Existing != nil && i < len(opts.Existing.Slides) {
			for l, c := range opts.Existing.Slides[i].Transcripts {
				if l != lang {
					lcs[l] = c
				}
			}
		}
		title := s.Title
		if title == "" {
			title = s.ID
		}
		t.Slides = append(t.Slides, transcript.Slide{Index: i, Title: title, Transcripts: lcs})
	}
	return t
}

func sentenceSegments(text string, pause int) []transcript.Segment {
	if text == "" {
		return []transcript.Segment{{Text: "", Pause: pause}}
	}
	parts := strings.Split(sentenceEndRe.ReplaceAllString(text, "$1\n"), "\n")
	segs := make([]transcript.Segment, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			segs = append(segs, transcript.Segment{Text: p})
		}
	}
	segs[len(segs)-1].Pause = pause
	return segs
}

// CompleteLanguages returns, sorted, the languages with narration on every
// slide. Partly translated languages are left out.
func CompleteLanguages(t *transcript.Transcript) []string {
	return completeLanguages(t.Slides)
}

func completeLanguages(slides []transcript.Slide) []string {
	seen := map[string]bool{}
	for _, s := range slides {
		for l := range s.Transcripts {
			seen[l] = true
		}
	}
	var langs []string
	for l := range seen {
		complete := true
		for _, s := range slides {
			if joinText(s.Transcripts[l].Segments) == "" {
				complete = false
				break
			}
		}
		if complete {
			langs = append(langs, l)
		}
	}
	sort.Strings(langs)
	return langs
}

// DefaultStartFragment is the URL fragment that opens a deck at a lesson's
// first slide: "#" followed by the slide's ID.
const DefaultStartFragment = "#{id}"

// ConfigOptions configures VideoConfig and LessonVideoConfig.
type ConfigOptions struct {
	// URL is the page to record, including any query (for example
	// "http://localhost:8000/index.html?capture=1"). Required.
	URL string
	// Name is the browser segment's name; default a slug of the title, or
	// the lesson's Name for a lesson.
	Name string
	// NextKey advances one slide; default "ArrowRight".
	NextKey string
	// InitialWaitMs is the wait on the first slide before narration; default 800.
	InitialWaitMs int
	// StartFragment is appended to URL to open the deck at a lesson's first
	// slide. {id} is replaced by the slide's ID, {n} by its 1-based number,
	// and {i} by its 0-based index. Default "#{id}". It is not used for a
	// lesson that starts at the first slide.
	StartFragment string
	// Width and Height set the output resolution; default 1920x1080.
	Width, Height int
	// FrameRate defaults to 30.
	FrameRate int
	// Headless records without a visible browser window.
	Headless bool
	// Pronunciations maps terms to how TTS should say them, per language.
	Pronunciations map[string]map[string]string
}

// VideoConfig returns a browser video config that records the deck one slide
// per step, narrated from t (usually the result of Transcript). Every
// language with narration on all slides is included. Each step's narration
// is the slide's segments joined, with the slide's pause after it, so
// videoascode can time each slide to its own narration.
func (d *Deck) VideoConfig(t *transcript.Transcript, opts ConfigOptions) (*config.VideoConfig, error) {
	if err := d.checkTranscript(t); err != nil {
		return nil, err
	}
	name := firstNonEmpty(opts.Name, slug(t.Metadata.Title), "deck")
	return d.videoConfig(t, 0, len(d.Slides), t.Metadata.Title, name, opts.URL, opts)
}

// LessonVideoConfig returns a browser video config that records one lesson:
// it opens the deck at the lesson's first slide (see
// ConfigOptions.StartFragment) and steps through the lesson's slides. Every
// language with narration on all of the lesson's slides is included. The
// config's title is the lesson's title.
func (d *Deck) LessonVideoConfig(t *transcript.Transcript, l Lesson, opts ConfigOptions) (*config.VideoConfig, error) {
	if err := d.checkTranscript(t); err != nil {
		return nil, err
	}
	if len(l.Slides) == 0 || l.First()+len(l.Slides) > len(d.Slides) {
		return nil, fmt.Errorf("lesson %d does not belong to this deck", l.Number)
	}
	u := opts.URL
	if first := l.First(); first > 0 && u != "" {
		id := d.Slides[first].ID
		frag := firstNonEmpty(opts.StartFragment, DefaultStartFragment)
		if id == "" && strings.Contains(frag, "{id}") {
			return nil, fmt.Errorf("lesson %d: its first slide (%d) needs an ID to open the deck there; add data-id or set a start fragment using {n}", l.Number, first+1)
		}
		u += strings.NewReplacer("{id}", id, "{n}", strconv.Itoa(first+1), "{i}", strconv.Itoa(first)).Replace(frag)
	}
	return d.videoConfig(t, l.First(), l.First()+len(l.Slides), l.Title, firstNonEmpty(opts.Name, l.Name()), u, opts)
}

func (d *Deck) checkTranscript(t *transcript.Transcript) error {
	if t == nil || len(t.Slides) != len(d.Slides) {
		return fmt.Errorf("transcript must have one entry per slide (%d)", len(d.Slides))
	}
	return nil
}

// videoConfig builds a config recording slides [from, to).
func (d *Deck) videoConfig(t *transcript.Transcript, from, to int, title, name, u string, opts ConfigOptions) (*config.VideoConfig, error) {
	if u == "" {
		return nil, fmt.Errorf("a URL is required")
	}
	slides := t.Slides[from:to]
	langs := completeLanguages(slides)
	if len(langs) == 0 {
		var missing []int
		for _, n := range d.Unnarrated() {
			if n > from && n <= to {
				missing = append(missing, n)
			}
		}
		return nil, fmt.Errorf("no language has narration on every slide (unnarrated slides: %v)", missing)
	}

	nextKey := firstNonEmpty(opts.NextKey, DefaultNextKey)
	wait := opts.InitialWaitMs
	if wait == 0 {
		wait = DefaultInitialWaitMs
	}
	ids := d.SlideIDs()
	steps := make([]browser.Step, 0, to-from)
	for i := from; i < to; i++ {
		if i == from {
			steps = append(steps, browser.Step{Action: browser.ActionWait, Duration: wait, Description: ids[i]})
		} else {
			steps = append(steps, browser.Step{Action: browser.ActionKeypress, Key: nextKey, Description: ids[i]})
		}
	}

	transcripts := make(map[string]config.SegmentTranscript, len(langs))
	for _, l := range langs {
		segs := make([]transcript.Segment, len(slides))
		for i, s := range slides {
			lc := s.Transcripts[l]
			segs[i] = transcript.Segment{Text: joinText(lc.Segments), Pause: totalPause(lc.Segments)}
		}
		transcripts[l] = config.SegmentTranscript{Segments: segs}
	}

	cfg := &config.VideoConfig{
		Version:         config.SchemaVersion,
		Title:           firstNonEmpty(title, "Deck"),
		Description:     t.Metadata.Description,
		DefaultLanguage: firstNonEmpty(t.Metadata.DefaultLanguage, DefaultLanguage),
		DefaultVoice:    t.Metadata.DefaultVoice,
		Resolution:      config.Resolution{Width: orDefault(opts.Width, 1920), Height: orDefault(opts.Height, 1080)},
		FrameRate:       orDefault(opts.FrameRate, 30),
		Languages:       langs,
		Pronunciations:  opts.Pronunciations,
		Segments: []config.SegmentConfig{{
			Type:        config.SourceTypeBrowser,
			Name:        name,
			URL:         u,
			Headless:    opts.Headless,
			Steps:       steps,
			Transcripts: transcripts,
		}},
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("generated config is invalid: %w", err)
	}
	return cfg, nil
}

func joinText(segs []transcript.Segment) string {
	parts := make([]string, 0, len(segs))
	for _, s := range segs {
		if t := strings.TrimSpace(s.Text); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

func totalPause(segs []transcript.Segment) int {
	total := 0
	for _, s := range segs {
		total += s.Pause
	}
	return total
}

var nonSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

// slug returns a lowercase, hyphenated form of s, at most 48 characters.
func slug(s string) string {
	out := strings.Trim(nonSlugRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if len(out) > 48 {
		out = strings.TrimRight(out[:48], "-")
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func orDefault(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
