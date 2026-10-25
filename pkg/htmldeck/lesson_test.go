package htmldeck

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/videoascode/pkg/browser"
	"github.com/grokify/videoascode/pkg/transcript"
)

const lessonDeck = `<html><head><title>Course Deck</title></head><body>
<section class="slide" data-id="title" data-vo="Welcome.">Title</section>
<section class="slide section" data-id="sec-a" data-title="Section: Basics" data-lesson="  The
  Basics " data-vo="Basics.">A</section>
<section class="slide" data-id="a2" data-vo="More basics.">A2</section>
<section class="slide section" data-id="sec-b" data-title="Section: Advanced" data-vo="Advanced. [PAUSE:400]">B</section>
<section class="slide" data-id="b2" data-vo="More advanced.">B2</section>
<section class="slide" data-id="c" data-lesson data-title="Wrap-up" data-vo="Goodbye.">C</section>
</body></html>`

func parseLessonDeck(t *testing.T, opts Options) *Deck {
	t.Helper()
	d, err := Parse(strings.NewReader(lessonDeck), opts)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func lessonSummary(ls []Lesson) []string {
	out := make([]string, len(ls))
	for i, l := range ls {
		ids := make([]string, len(l.Slides))
		for j, s := range l.Slides {
			ids[j] = s.ID
		}
		out[i] = l.ID + "|" + l.Title + "|" + strings.Join(ids, ",")
	}
	return out
}

func TestLessons(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want []string
	}{
		{
			name: "data-lesson attribute",
			want: []string{
				// The title slide joins the first lesson; the attribute's value
				// is the title, and an empty attribute falls back to data-title.
				"sec-a|The Basics|title,sec-a,a2,sec-b,b2",
				"c|Wrap-up|c",
			},
		},
		{
			name: "lesson class too",
			opts: Options{LessonClass: "section"},
			want: []string{
				"sec-a|The Basics|title,sec-a,a2",
				"sec-b|Section: Advanced|sec-b,b2",
				"c|Wrap-up|c",
			},
		},
		{
			name: "no lesson starts",
			opts: Options{LessonAttr: "data-none"},
			want: []string{"course-deck|Course Deck|title,sec-a,a2,sec-b,b2,c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lessonSummary(parseLessonDeck(t, tt.opts).Lessons())
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("lessons:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestLessonLookupAndName(t *testing.T) {
	d := parseLessonDeck(t, Options{LessonClass: "section"})
	for ref, wantID := range map[string]string{"2": "sec-b", "sec-b": "sec-b", "c": "c", "1": "sec-a"} {
		l, err := d.Lesson(ref)
		if err != nil {
			t.Fatalf("Lesson(%q): %v", ref, err)
		}
		if l.ID != wantID {
			t.Errorf("Lesson(%q) = %s, want %s", ref, l.ID, wantID)
		}
	}
	for _, ref := range []string{"0", "4", "nope"} {
		if _, err := d.Lesson(ref); err == nil {
			t.Errorf("Lesson(%q): expected an error", ref)
		}
	}

	l, err := d.Lesson("2")
	if err != nil {
		t.Fatal(err)
	}
	if got := l.Name(); got != "02-section-advanced" {
		t.Errorf("Name = %q", got)
	}
	if got := (Lesson{Number: 7}).Name(); got != "07" {
		t.Errorf("empty Name = %q", got)
	}
}

func TestLessonVideoConfig(t *testing.T) {
	d := parseLessonDeck(t, Options{LessonClass: "section"})
	tr := d.Transcript(TranscriptOptions{})
	lessons := d.Lessons()
	opts := ConfigOptions{URL: "http://localhost:8000/index.html?capture=1"}

	first, err := d.LessonVideoConfig(tr, lessons[0], opts)
	if err != nil {
		t.Fatal(err)
	}
	if u := first.Segments[0].URL; u != opts.URL {
		t.Errorf("first lesson URL = %s, want no fragment", u)
	}

	cfg, err := d.LessonVideoConfig(tr, lessons[1], opts)
	if err != nil {
		t.Fatal(err)
	}
	seg := cfg.Segments[0]
	if seg.URL != opts.URL+"#sec-b" {
		t.Errorf("URL = %s", seg.URL)
	}
	if cfg.Title != "Section: Advanced" || seg.Name != "02-section-advanced" {
		t.Errorf("title %q, name %q", cfg.Title, seg.Name)
	}
	if len(seg.Steps) != 2 || seg.Steps[0].Action != browser.ActionWait || seg.Steps[1].Description != "b2" {
		t.Errorf("steps = %+v", seg.Steps)
	}
	segs := seg.Transcripts["en-US"].Segments
	if len(segs) != 2 || segs[0].Text != "Advanced." || segs[0].Pause != 400 || segs[1].Text != "More advanced." {
		t.Errorf("transcript segments = %+v", segs)
	}

	opts.StartFragment = "#/{i}?n={n}"
	cfg, err = d.LessonVideoConfig(tr, lessons[1], opts)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Segments[0].URL; !strings.HasSuffix(got, "#/3?n=4") {
		t.Errorf("custom fragment URL = %s", got)
	}
}

func TestLessonVideoConfigLanguagesAndErrors(t *testing.T) {
	d := parseLessonDeck(t, Options{LessonClass: "section"})
	tr := d.Transcript(TranscriptOptions{})
	// French covers only the second lesson.
	for _, i := range []int{3, 4} {
		tr.Slides[i].Transcripts["fr-FR"] = transcript.LanguageContent{Segments: []transcript.Segment{{Text: "Bonjour."}}}
	}
	lessons := d.Lessons()
	opts := ConfigOptions{URL: "http://x/"}

	cfg, err := d.LessonVideoConfig(tr, lessons[1], opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(cfg.Languages, ",") != "en-US,fr-FR" {
		t.Errorf("lesson 2 languages = %v", cfg.Languages)
	}
	if cfg, err = d.LessonVideoConfig(tr, lessons[0], opts); err != nil {
		t.Fatal(err)
	} else if strings.Join(cfg.Languages, ",") != "en-US" {
		t.Errorf("lesson 1 languages = %v", cfg.Languages)
	}

	if _, err := d.LessonVideoConfig(tr, Lesson{Number: 9}, opts); err == nil {
		t.Error("expected an error for an empty lesson")
	}

	// A lesson whose first slide has no ID cannot be opened by {id}.
	noID, err := Parse(strings.NewReader(`<div class="slide" data-lesson="One" data-vo="One."></div><div class="slide" data-lesson="Two" data-vo="Two."></div>`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	ntr := noID.Transcript(TranscriptOptions{})
	if _, err := noID.LessonVideoConfig(ntr, noID.Lessons()[1], opts); err == nil || !strings.Contains(err.Error(), "needs an ID") {
		t.Errorf("err = %v", err)
	}
	opts.StartFragment = "#{n}"
	if cfg, err := noID.LessonVideoConfig(ntr, noID.Lessons()[1], opts); err != nil {
		t.Fatal(err)
	} else if cfg.Segments[0].URL != "http://x/#2" {
		t.Errorf("URL = %s", cfg.Segments[0].URL)
	}
}

func TestLessonFingerprint(t *testing.T) {
	fingerprint := func(src, u string) string {
		t.Helper()
		d, err := Parse(strings.NewReader(src), Options{LessonClass: "section"})
		if err != nil {
			t.Fatal(err)
		}
		l := d.Lessons()[1]
		cfg, err := d.LessonVideoConfig(d.Transcript(TranscriptOptions{}), l, ConfigOptions{URL: u})
		if err != nil {
			t.Fatal(err)
		}
		return LessonFingerprint(l, cfg)
	}
	base := fingerprint(lessonDeck, "http://127.0.0.1:5000/")
	if got := fingerprint(lessonDeck, "http://127.0.0.1:6000/"); got != base {
		t.Error("fingerprint depends on the URL")
	}
	for name, src := range map[string]string{
		"narration": strings.Replace(lessonDeck, "More advanced.", "Even more advanced.", 1),
		"slide":     strings.Replace(lessonDeck, ">B2<", ">B2, revised<", 1),
		"pause":     strings.Replace(lessonDeck, "[PAUSE:400]", "[PAUSE:500]", 1),
	} {
		if fingerprint(src, "http://127.0.0.1:5000/") == base {
			t.Errorf("%s change did not change the fingerprint", name)
		}
	}
	if fingerprint(strings.Replace(lessonDeck, ">A2<", ">A2, revised<", 1), "http://x/") != base {
		t.Error("a change to another lesson changed the fingerprint")
	}
}

func TestManifestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ManifestFile)
	m := &Manifest{Title: "Deck", DefaultLanguage: "en-US", Lessons: []ManifestLesson{
		{Number: 1, ID: "a", Title: "A", Slides: []string{"a"}, Videos: map[string]string{"en-US": "01-a.mp4"}, DurationMs: 1234, Fingerprint: "f"},
		{Number: 2, ID: "b", Title: "B", Slides: []string{"b", "b2"}},
	}}
	if err := m.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Lesson("a").Videos["en-US"] != "01-a.mp4" || got.Lesson("b").Slides[1] != "b2" || got.Lesson("z") != nil {
		t.Errorf("round trip = %+v", got)
	}
	if _, err := LoadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("expected an error for a missing manifest")
	}
}
