package review

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grokify/videoascode/pkg/htmldeck"
	"github.com/grokify/videoascode/pkg/transcript"
	"github.com/grokify/videoascode/pkg/tts"
)

// Navigation modes for the review page's deck frame.
const (
	// NavHash moves between slides by changing the deck's URL fragment. The
	// deck must follow hashchange events.
	NavHash = "hash"
	// NavReload reloads the deck at each slide's fragment. It works with any
	// deck that opens at the fragment, at the cost of a reload per move.
	NavReload = "reload"
)

// Input is what Build needs to describe a deck for review.
type Input struct {
	Deck       *htmldeck.Deck
	Transcript *transcript.Transcript
	// Pronunciations maps terms to spoken forms, per language.
	Pronunciations map[string]map[string]string
	Review         *File
	// Languages to review; default every language in the transcript, the
	// default language first.
	Languages []string
	Frame     Frame
	// StartFragment opens the deck at a slide: {id}, {n} (1-based) and {i}
	// (0-based) are replaced. Default "#{id}".
	StartFragment string
	// ReadOnly marks a snapshot whose decisions cannot be changed.
	ReadOnly bool
	// Generated is when the model was built.
	Generated time.Time
}

// Frame says how the review page loads the deck.
type Frame struct {
	// URL is the deck page, relative to the review page.
	URL string `json:"url"`
	// Query is added to the deck URL, e.g. "capture=1".
	Query string `json:"query,omitempty"`
	// Width and Height are the deck's design size; the frame is scaled to fit.
	Width  int `json:"width"`
	Height int `json:"height"`
	// Nav is NavHash or NavReload.
	Nav string `json:"nav"`
}

// Model describes a deck for review. It is the JSON the review page renders.
type Model struct {
	Title           string            `json:"title"`
	DefaultLanguage string            `json:"defaultLanguage"`
	Languages       []string          `json:"languages"`
	Frame           Frame             `json:"frame"`
	Lessons         []Lesson          `json:"lessons"`
	Slides          []Slide           `json:"slides"`
	Summary         map[string]Counts `json:"summary"`
	Voice           string            `json:"voice,omitempty"`
	ReadOnly        bool              `json:"readOnly"`
	Generated       time.Time         `json:"generated"`
	// Decks lists every deck under review when several are served
	// together; empty for a single deck.
	Decks []DeckLink `json:"decks,omitempty"`
}

// DeckLink points to one deck in a multi-deck review.
type DeckLink struct {
	Name string `json:"name"`
	// URL is relative to the deck's review page.
	URL     string `json:"url"`
	Current bool   `json:"current,omitempty"`
}

// Lesson is a run of slides recorded as one video.
type Lesson struct {
	Number int    `json:"number"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	// First and Last are 0-based slide indexes, inclusive.
	First int `json:"first"`
	Last  int `json:"last"`
}

// Slide is one slide's narration in every reviewed language.
type Slide struct {
	Index  int    `json:"index"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	Lesson int    `json:"lesson"`
	// Fragment opens the deck at this slide, e.g. "#intro".
	Fragment string               `json:"fragment"`
	Langs    map[string]Narration `json:"langs"`
}

// Narration is one slide's narration in one language.
type Narration struct {
	// Segments are the sentences as written, each with its pause.
	Segments []Segment `json:"segments"`
	// Text is the narration as written.
	Text string `json:"text"`
	// Spoken is the text TTS receives, after pronunciation substitutions.
	Spoken string `json:"spoken"`
	// Terms are the pronunciation entries applied to this text.
	Terms   []Term    `json:"terms,omitempty"`
	Words   int       `json:"words"`
	Seconds float64   `json:"seconds"`
	PauseMs int       `json:"pauseMs"`
	Hash    string    `json:"hash"`
	Missing bool      `json:"missing,omitempty"`
	Checks  []Finding `json:"checks,omitempty"`
	State   State     `json:"state"`
	// Entry is the latest decision, current or stale.
	Entry *Entry `json:"entry,omitempty"`
}

// Segment is one sentence and the pause after it.
type Segment struct {
	Text string `json:"text"`
	// Spoken is Text after pronunciation substitutions, for previewing
	// the segment with the browser's own speech.
	Spoken  string `json:"spoken,omitempty"`
	PauseMs int    `json:"pauseMs,omitempty"`
}

// Term is a pronunciation substitution: Written is said as Spoken.
type Term struct {
	Written string `json:"written"`
	Spoken  string `json:"spoken"`
}

// Counts summarizes review states for one language.
type Counts struct {
	Slides     int `json:"slides"`
	Approved   int `json:"approved"`
	Changes    int `json:"changes"`
	Stale      int `json:"stale"`
	Unreviewed int `json:"unreviewed"`
	Errors     int `json:"errors"`
	Warnings   int `json:"warnings"`
}

// Ready reports whether every slide is approved on its current text.
func (c Counts) Ready() bool { return c.Slides > 0 && c.Approved == c.Slides }

// Build describes a deck for review.
func Build(in Input) (*Model, error) {
	d, t := in.Deck, in.Transcript
	if d == nil || t == nil {
		return nil, fmt.Errorf("review: deck and transcript are required")
	}
	if len(t.Slides) != len(d.Slides) {
		return nil, fmt.Errorf("review: transcript has %d slides; deck has %d", len(t.Slides), len(d.Slides))
	}
	def := t.Metadata.DefaultLanguage
	if def == "" {
		def = htmldeck.DefaultLanguage
	}
	langs := in.Languages
	if len(langs) == 0 {
		langs = transcriptLanguages(t, def)
	}

	frame := in.Frame
	if frame.Width == 0 || frame.Height == 0 {
		frame.Width, frame.Height = 1920, 1080
	}
	if frame.Nav == "" {
		frame.Nav = NavHash
	}

	m := &Model{
		Title:           t.Metadata.Title,
		DefaultLanguage: def,
		Languages:       langs,
		Frame:           frame,
		Summary:         map[string]Counts{},
		Voice:           voiceLabel(t.Metadata.DefaultVoice),
		ReadOnly:        in.ReadOnly,
		Generated:       in.Generated,
	}
	ids := d.SlideIDs()
	lessonOf := make([]int, len(d.Slides))
	for _, l := range d.Lessons() {
		last := l.Slides[len(l.Slides)-1].Index
		m.Lessons = append(m.Lessons, Lesson{Number: l.Number, ID: l.ID, Title: l.Title, First: l.First(), Last: last})
		for _, s := range l.Slides {
			lessonOf[s.Index] = l.Number
		}
	}

	pron := tts.NewPronouncer(&tts.PronunciationDictionary{Terms: in.Pronunciations})
	for i, s := range d.Slides {
		title := s.Title
		if title == "" {
			title = s.ID
		}
		sv := Slide{
			Index:    i,
			ID:       ids[i],
			Title:    title,
			Lesson:   lessonOf[i],
			Fragment: Fragment(in.StartFragment, ids[i], i),
			Langs:    map[string]Narration{},
		}
		defText := joinSegments(t.Slides[i].Transcripts[def].Segments)
		for _, lang := range langs {
			lc, ok := t.Slides[i].Transcripts[lang]
			if !ok && defText == "" {
				// A silent slide is silent in every language.
				lc = t.Slides[i].Transcripts[def]
			}
			n := narration(lc.Segments, lang, pron, in.Pronunciations)
			n.Missing = lang != def && (!ok || n.Text == "") && defText != ""
			n.Checks = Check(CheckInput{
				Text: n.Text, PauseMs: n.PauseMs, Language: lang,
				Pronunciations: in.Pronunciations, Missing: n.Missing,
			})
			n.State = in.Review.StateOf(sv.ID, lang, n.Hash)
			if e, ok := in.Review.Get(sv.ID, lang); ok {
				n.Entry = &e
			}
			sv.Langs[lang] = n
			m.Summary[lang] = tally(m.Summary[lang], n)
		}
		m.Slides = append(m.Slides, sv)
	}
	return m, nil
}

func narration(segs []transcript.Segment, lang string, pron *tts.Pronouncer, terms map[string]map[string]string) Narration {
	n := Narration{}
	for _, s := range segs {
		if s.Text != "" || s.Pause > 0 {
			n.Segments = append(n.Segments, Segment{Text: s.Text, Spoken: pron.Apply(s.Text, lang), PauseMs: s.Pause})
		}
		n.PauseMs += s.Pause
	}
	n.Text = joinSegments(segs)
	n.Spoken = pron.Apply(n.Text, lang)
	n.Words = Words(n.Text)
	n.Seconds = Seconds(n.Text, n.PauseMs)
	n.Hash = TextHash(n.Spoken, n.PauseMs)
	n.Terms = appliedTerms(n.Text, lang, terms)
	return n
}

// appliedTerms lists the pronunciation entries whose written form appears
// in text, longest first. A term found only inside a longer applied term
// (PBHQ inside PBHQ Cloud, say) is left out.
func appliedTerms(text, lang string, terms map[string]map[string]string) []Term {
	var cands []Term
	for written, byLang := range terms {
		if spoken, ok := byLang[lang]; ok && written != "" && strings.Contains(text, written) {
			cands = append(cands, Term{Written: written, Spoken: spoken})
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		if len(cands[i].Written) != len(cands[j].Written) {
			return len(cands[i].Written) > len(cands[j].Written)
		}
		return cands[i].Written < cands[j].Written
	})
	var out []Term
	rest := text
	for _, t := range cands {
		if strings.Contains(rest, t.Written) {
			out = append(out, t)
			rest = strings.ReplaceAll(rest, t.Written, "\x00")
		}
	}
	return out
}

func tally(c Counts, n Narration) Counts {
	c.Slides++
	switch n.State {
	case StateApproved:
		c.Approved++
	case StateChanges:
		c.Changes++
	case StateStale:
		c.Stale++
	default:
		c.Unreviewed++
	}
	for _, f := range n.Checks {
		switch f.Level {
		case LevelError:
			c.Errors++
		case LevelWarn:
			c.Warnings++
		}
	}
	return c
}

func joinSegments(segs []transcript.Segment) string {
	parts := make([]string, 0, len(segs))
	for _, s := range segs {
		if t := strings.TrimSpace(s.Text); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

// transcriptLanguages returns every language in t, def first, the rest sorted.
func transcriptLanguages(t *transcript.Transcript, def string) []string {
	var rest []string
	for _, s := range t.Slides {
		for l := range s.Transcripts {
			if l != def && !slices.Contains(rest, l) {
				rest = append(rest, l)
			}
		}
	}
	sort.Strings(rest)
	return append([]string{def}, rest...)
}

func voiceLabel(v transcript.VoiceConfig) string {
	name := v.VoiceName
	if name == "" {
		name = v.VoiceID
	}
	switch {
	case v.Provider != "" && name != "":
		return v.Provider + " · " + name
	case v.Provider != "":
		return v.Provider + " · no voice set"
	case name != "":
		return name
	}
	return ""
}

// Fragment expands a start-fragment template ({id}, {n} 1-based, {i}
// 0-based) for one slide.
func Fragment(template, id string, index int) string {
	if template == "" {
		template = htmldeck.DefaultStartFragment
	}
	return strings.NewReplacer("{id}", id, "{n}", strconv.Itoa(index+1), "{i}", strconv.Itoa(index)).Replace(template)
}

// Unready lists, for each language, the 1-based numbers of slides that are
// not approved on their current text.
func (m *Model) Unready(langs []string) map[string][]int {
	out := map[string][]int{}
	for _, lang := range langs {
		for _, s := range m.Slides {
			if n, ok := s.Langs[lang]; !ok || n.State != StateApproved {
				out[lang] = append(out[lang], s.Index+1)
			}
		}
	}
	return out
}
