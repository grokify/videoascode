package htmldeck

import (
	"strings"
	"testing"

	"github.com/grokify/videoascode/pkg/browser"
	"github.com/grokify/videoascode/pkg/transcript"
)

const sampleDeck = `<!doctype html>
<html><head><title>  Sample
  Deck </title></head>
<body>
  <div id="deck">
    <section class="slide title active" data-id="intro" data-title="Intro &amp; Overview"
      data-vo="Welcome to the deck. It has three slides! Ready?">
      <h1>Intro</h1>
      <div class="slide-inner">nested elements are part of this slide</div>
    </section>
    <section data-vo="Second slide. [PAUSE:500]Then a beat. [PAUSE:250]" class="slide" id="two">
      <p>Attribute order does not matter.</p>
    </section>
    <section class="slide" data-id="three">
      <p>Narration from speaker notes.</p>
      <aside class="notes">
        Notes are   read as plain text,
        <em>markup</em> removed.
      </aside>
    </section>
    <section class="slideshow">not a slide: class must match exactly</section>
  </div>
</body></html>`

func parseSample(t *testing.T) *Deck {
	t.Helper()
	d, err := Parse(strings.NewReader(sampleDeck), Options{})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestParse(t *testing.T) {
	d := parseSample(t)
	if d.Title != "Sample Deck" {
		t.Errorf("Title = %q", d.Title)
	}
	if len(d.Slides) != 3 {
		t.Fatalf("got %d slides, want 3", len(d.Slides))
	}

	tests := []struct {
		id, title, narration string
		pause                int
	}{
		{"intro", "Intro & Overview", "Welcome to the deck. It has three slides! Ready?", 0},
		{"two", "", "Second slide. Then a beat.", 750},
		{"three", "", "Notes are read as plain text, markup removed.", 0},
	}
	for i, tt := range tests {
		s := d.Slides[i]
		if s.Index != i || s.ID != tt.id || s.Title != tt.title || s.Narration != tt.narration || s.Pause != tt.pause {
			t.Errorf("slide %d = %+v, want id=%q title=%q narration=%q pause=%d", i, s, tt.id, tt.title, tt.narration, tt.pause)
		}
	}
	if got := d.Unnarrated(); len(got) != 0 {
		t.Errorf("Unnarrated() = %v", got)
	}
}

func TestParseOptionsAndErrors(t *testing.T) {
	src := `<div class="page" data-say="Hello.">A</div><div class="page">B</div>`
	d, err := Parse(strings.NewReader(src), Options{SlideClass: "page", NarrationAttr: "data-say"})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Slides) != 2 || d.Slides[0].Narration != "Hello." {
		t.Fatalf("slides = %+v", d.Slides)
	}
	if got := d.Unnarrated(); len(got) != 1 || got[0] != 2 {
		t.Errorf("Unnarrated() = %v, want [2]", got)
	}
	if got := d.SlideIDs(); got[0] != "slide-1" || got[1] != "slide-2" {
		t.Errorf("SlideIDs() = %v", got)
	}

	if _, err := Parse(strings.NewReader(`<p>no slides</p>`), Options{}); err == nil {
		t.Error("Parse with no slides: want error")
	}
	if _, err := Parse(strings.NewReader(`<div class="slide" data-vo="[PAUSE:99999999999999999999]"></div>`), Options{}); err == nil {
		t.Error("Parse with overflowing pause: want error")
	}
}

func TestTranscript(t *testing.T) {
	d := parseSample(t)
	existing := &transcript.Transcript{
		Metadata: transcript.Metadata{
			Title:        "Old title",
			DefaultVoice: transcript.VoiceConfig{Provider: "deepgram", VoiceID: "aura-asteria-en"},
			DefaultVenue: "youtube",
		},
		Slides: []transcript.Slide{
			{Index: 0, Transcripts: map[string]transcript.LanguageContent{
				"en-US": {Segments: []transcript.Segment{{Text: "stale English"}}},
				"fr-FR": {Segments: []transcript.Segment{{Text: "Bienvenue."}}},
			}},
		},
	}

	tr := d.Transcript(TranscriptOptions{Existing: existing})
	if tr.Version != DefaultTranscriptVers || tr.Metadata.Title != "Sample Deck" || tr.Metadata.DefaultLanguage != "en-US" {
		t.Errorf("metadata = %+v", tr.Metadata)
	}
	if tr.Metadata.DefaultVoice.VoiceID != "aura-asteria-en" || tr.Metadata.DefaultVenue != "youtube" {
		t.Errorf("existing voice and venue not kept: %+v", tr.Metadata)
	}
	if len(tr.Slides) != 3 {
		t.Fatalf("got %d slides", len(tr.Slides))
	}

	en := tr.Slides[0].Transcripts["en-US"].Segments
	if len(en) != 3 || en[0].Text != "Welcome to the deck." || en[2].Text != "Ready?" {
		t.Errorf("sentences = %+v", en)
	}
	if fr := tr.Slides[0].Transcripts["fr-FR"].Segments; len(fr) != 1 || fr[0].Text != "Bienvenue." {
		t.Errorf("translation not kept: %+v", tr.Slides[0].Transcripts)
	}
	if two := tr.Slides[1].Transcripts["en-US"].Segments; two[len(two)-1].Pause != 750 {
		t.Errorf("pause not on last segment: %+v", two)
	}
	if tr.Slides[1].Title != "two" {
		t.Errorf("untitled slide should fall back to its ID, got %q", tr.Slides[1].Title)
	}

	if got := CompleteLanguages(tr); len(got) != 1 || got[0] != "en-US" {
		t.Errorf("CompleteLanguages = %v (fr-FR covers only slide 1)", got)
	}
}

func TestVideoConfig(t *testing.T) {
	d := parseSample(t)
	tr := d.Transcript(TranscriptOptions{Voice: transcript.VoiceConfig{Provider: "elevenlabs", VoiceID: "v1"}})
	for i := range tr.Slides {
		tr.Slides[i].Transcripts["de-DE"] = transcript.LanguageContent{Segments: []transcript.Segment{{Text: "Hallo."}}}
	}

	cfg, err := d.VideoConfig(tr, ConfigOptions{URL: "http://localhost:8000/index.html?capture=1", Headless: true})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Title != "Sample Deck" || cfg.DefaultVoice.VoiceID != "v1" || cfg.Resolution.Width != 1920 || cfg.FrameRate != 30 {
		t.Errorf("config = %+v", cfg)
	}
	if got := strings.Join(cfg.Languages, ","); got != "de-DE,en-US" {
		t.Errorf("Languages = %s", got)
	}

	seg := cfg.Segments[0]
	if seg.Name != "sample-deck" || !seg.Headless || len(seg.Steps) != 3 {
		t.Fatalf("segment = %+v", seg)
	}
	if s := seg.Steps[0]; s.Action != browser.ActionWait || s.Duration != DefaultInitialWaitMs || s.Description != "intro" {
		t.Errorf("first step = %+v", s)
	}
	if s := seg.Steps[2]; s.Action != browser.ActionKeypress || s.Key != DefaultNextKey || s.Description != "three" {
		t.Errorf("last step = %+v", s)
	}

	en := seg.Transcripts["en-US"].Segments
	if len(en) != 3 {
		t.Fatalf("en-US has %d segments, want one per step", len(en))
	}
	if en[0].Text != "Welcome to the deck. It has three slides! Ready?" || en[1].Pause != 750 {
		t.Errorf("en-US segments = %+v", en)
	}
}

func TestVideoConfigErrors(t *testing.T) {
	d := parseSample(t)
	tr := d.Transcript(TranscriptOptions{})
	if _, err := d.VideoConfig(tr, ConfigOptions{}); err == nil {
		t.Error("no URL: want error")
	}
	if _, err := d.VideoConfig(&transcript.Transcript{}, ConfigOptions{URL: "http://x"}); err == nil {
		t.Error("transcript of the wrong length: want error")
	}

	silent, err := Parse(strings.NewReader(`<div class="slide"></div>`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := silent.VideoConfig(silent.Transcript(TranscriptOptions{}), ConfigOptions{URL: "http://x"}); err == nil {
		t.Error("no narrated language: want error")
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Web APIs — A Field Guide": "web-apis-a-field-guide",
		"  ":                       "",
		strings.Repeat("ab ", 30):  "ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab",
	} {
		if got := slug(in); got != want {
			t.Errorf("slug(%q) = %q, want %q", in, got, want)
		}
	}
}
