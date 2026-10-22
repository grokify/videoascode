package htmldeck

import (
	"fmt"
	"regexp"
	"sort"
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
	seen := map[string]bool{}
	for _, s := range t.Slides {
		for l := range s.Transcripts {
			seen[l] = true
		}
	}
	var langs []string
	for l := range seen {
		complete := true
		for _, s := range t.Slides {
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

// ConfigOptions configures VideoConfig.
type ConfigOptions struct {
	// URL is the page to record, including any query (for example
	// "http://localhost:8000/index.html?capture=1"). Required.
	URL string
	// Name is the browser segment's name; default a slug of the title.
	Name string
	// NextKey advances one slide; default "ArrowRight".
	NextKey string
	// InitialWaitMs is the wait on the first slide before narration; default 800.
	InitialWaitMs int
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
	if opts.URL == "" {
		return nil, fmt.Errorf("a URL is required")
	}
	if t == nil || len(t.Slides) != len(d.Slides) {
		return nil, fmt.Errorf("transcript must have one entry per slide (%d)", len(d.Slides))
	}
	langs := CompleteLanguages(t)
	if len(langs) == 0 {
		return nil, fmt.Errorf("no language has narration on every slide (unnarrated slides: %v)", d.Unnarrated())
	}

	nextKey := firstNonEmpty(opts.NextKey, DefaultNextKey)
	wait := opts.InitialWaitMs
	if wait == 0 {
		wait = DefaultInitialWaitMs
	}
	ids := d.SlideIDs()
	steps := make([]browser.Step, len(d.Slides))
	for i := range d.Slides {
		if i == 0 {
			steps[i] = browser.Step{Action: browser.ActionWait, Duration: wait, Description: ids[i]}
		} else {
			steps[i] = browser.Step{Action: browser.ActionKeypress, Key: nextKey, Description: ids[i]}
		}
	}

	transcripts := make(map[string]config.SegmentTranscript, len(langs))
	for _, l := range langs {
		segs := make([]transcript.Segment, len(t.Slides))
		for i, s := range t.Slides {
			lc := s.Transcripts[l]
			segs[i] = transcript.Segment{Text: joinText(lc.Segments), Pause: totalPause(lc.Segments)}
		}
		transcripts[l] = config.SegmentTranscript{Segments: segs}
	}

	cfg := &config.VideoConfig{
		Version:         config.SchemaVersion,
		Title:           t.Metadata.Title,
		Description:     t.Metadata.Description,
		DefaultLanguage: t.Metadata.DefaultLanguage,
		DefaultVoice:    t.Metadata.DefaultVoice,
		Resolution:      config.Resolution{Width: orDefault(opts.Width, 1920), Height: orDefault(opts.Height, 1080)},
		FrameRate:       orDefault(opts.FrameRate, 30),
		Languages:       langs,
		Pronunciations:  opts.Pronunciations,
		Segments: []config.SegmentConfig{{
			Type:        config.SourceTypeBrowser,
			Name:        firstNonEmpty(opts.Name, slug(t.Metadata.Title), "deck"),
			URL:         opts.URL,
			Headless:    opts.Headless,
			Steps:       steps,
			Transcripts: transcripts,
		}},
	}
	if cfg.Title == "" {
		cfg.Title = "Deck"
	}
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = DefaultLanguage
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
