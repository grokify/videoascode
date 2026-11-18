package review

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grokify/videoascode/pkg/htmldeck"
	"github.com/grokify/videoascode/pkg/transcript"
)

const testDeck = `<!doctype html>
<html><head><title>Test Deck</title></head><body>
<section class="slide" data-id="intro" data-lesson="Welcome" data-vo="Welcome to PBHQ. [PAUSE:500]"></section>
<section class="slide" data-id="pricing" data-vo="It costs $5 a month, see example.com & more."></section>
<section class="slide" data-id="quiet" data-vo="[PAUSE:1500]"></section>
<section class="slide" data-id="next" data-lesson="Next Steps" data-vo="PBHQ Cloud is next."></section>
</body></html>`

var testPron = map[string]map[string]string{
	"PBHQ":       {"en-US": "P B H Q"},
	"PBHQ Cloud": {"en-US": "P B H Q Cloud"},
}

func testInput(t *testing.T, rf *File) Input {
	t.Helper()
	d, err := htmldeck.Parse(strings.NewReader(testDeck), htmldeck.Options{})
	if err != nil {
		t.Fatal(err)
	}
	tr := d.Transcript(htmldeck.TranscriptOptions{})
	tr.Slides[0].Transcripts["fr-FR"] = transcript.LanguageContent{
		Segments: []transcript.Segment{{Text: "Bienvenue sur PBHQ.", Pause: 500}},
	}
	return Input{Deck: d, Transcript: tr, Pronunciations: testPron, Review: rf}
}

func build(t *testing.T, rf *File) *Model {
	t.Helper()
	m, err := Build(testInput(t, rf))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestBuild(t *testing.T) {
	m := build(t, NewFile())
	if m.Title != "Test Deck" || len(m.Slides) != 4 || len(m.Lessons) != 2 {
		t.Fatalf("title %q, %d slides, %d lessons", m.Title, len(m.Slides), len(m.Lessons))
	}
	if got := strings.Join(m.Languages, ","); got != "en-US,fr-FR" {
		t.Errorf("languages = %s", got)
	}
	if l := m.Lessons[1]; l.First != 3 || l.Last != 3 || l.Title != "Next Steps" {
		t.Errorf("lesson 2 = %+v", l)
	}
	if f := m.Slides[1].Fragment; f != "#pricing" {
		t.Errorf("fragment = %q", f)
	}

	intro := m.Slides[0].Langs["en-US"]
	if intro.Spoken != "Welcome to P B H Q." || intro.PauseMs != 500 {
		t.Errorf("intro spoken %q pause %d", intro.Spoken, intro.PauseMs)
	}
	if intro.State != StateUnreviewed {
		t.Errorf("state = %s", intro.State)
	}
	// Each segment carries its spoken form, for the browser preview.
	if len(intro.Segments) != 1 || intro.Segments[0].Spoken != "Welcome to P B H Q." || intro.Segments[0].PauseMs != 500 {
		t.Errorf("intro segments = %+v", intro.Segments)
	}

	// The longer term wins; PBHQ inside PBHQ Cloud is not listed separately.
	next := m.Slides[3].Langs["en-US"]
	if len(next.Terms) != 1 || next.Terms[0].Written != "PBHQ Cloud" {
		t.Errorf("terms = %+v", next.Terms)
	}

	// fr-FR exists only on slide 1: missing elsewhere, except the silent slide.
	for i, want := range []bool{false, true, false, true} {
		if got := m.Slides[i].Langs["fr-FR"].Missing; got != want {
			t.Errorf("slide %d fr-FR missing = %v, want %v", i+1, got, want)
		}
	}
	if c := m.Summary["fr-FR"]; c.Errors != 2 {
		t.Errorf("fr-FR errors = %d, want 2", c.Errors)
	}
}

func TestBuildFragmentTemplate(t *testing.T) {
	in := testInput(t, NewFile())
	in.StartFragment = "#/{n}"
	m, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if f := m.Slides[2].Fragment; f != "#/3" {
		t.Errorf("fragment = %q", f)
	}
}

func TestBuildMismatch(t *testing.T) {
	in := testInput(t, nil)
	in.Transcript.Slides = in.Transcript.Slides[:2]
	if _, err := Build(in); err == nil {
		t.Error("want error for a transcript shorter than the deck")
	}
}

func TestReviewStates(t *testing.T) {
	rf := NewFile()
	m := build(t, rf)
	intro := m.Slides[0].Langs["en-US"]
	rf.Set("intro", "en-US", Entry{Status: StatusApproved, TextHash: intro.Hash})
	rf.Set("pricing", "en-US", Entry{Status: StatusChanges, TextHash: m.Slides[1].Langs["en-US"].Hash, Note: "say five dollars"})
	rf.Set("next", "en-US", Entry{Status: StatusApproved, TextHash: "old"})

	m = build(t, rf)
	want := []State{StateApproved, StateChanges, StateUnreviewed, StateStale}
	for i, w := range want {
		if got := m.Slides[i].Langs["en-US"].State; got != w {
			t.Errorf("slide %d state = %s, want %s", i+1, got, w)
		}
	}
	c := m.Summary["en-US"]
	if c.Approved != 1 || c.Changes != 1 || c.Stale != 1 || c.Unreviewed != 1 || c.Ready() {
		t.Errorf("counts = %+v", c)
	}
	if got := m.Unready([]string{"en-US"})["en-US"]; len(got) != 3 || got[0] != 2 {
		t.Errorf("unready = %v", got)
	}

	// Changing a pronunciation changes what TTS hears, so approval goes stale.
	in := testInput(t, rf)
	in.Pronunciations = map[string]map[string]string{"PBHQ": {"en-US": "pee bee aitch queue"}}
	m2, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := m2.Slides[0].Langs["en-US"].State; got != StateStale {
		t.Errorf("after pronunciation change: state = %s", got)
	}
}

func TestFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.json")
	f, err := LoadFile(path)
	if err != nil {
		t.Fatalf("missing file: %v", err)
	}
	f.Set("intro", "en-US", Entry{Status: StatusApproved, TextHash: "abc", UpdatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)})
	if err := f.Save(path); err != nil {
		t.Fatal(err)
	}
	g, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if g.StateOf("intro", "en-US", "abc") != StateApproved {
		t.Errorf("round trip lost the decision: %+v", g.Slides)
	}
	g.Clear("intro", "en-US")
	if len(g.Slides) != 0 {
		t.Errorf("clear left %+v", g.Slides)
	}
}

func TestCheck(t *testing.T) {
	codes := func(fs []Finding) string {
		var out []string
		for _, f := range fs {
			out = append(out, string(f.Level)+":"+f.Code)
		}
		return strings.Join(out, " ")
	}
	long := strings.Repeat("word ", LongSentenceWords+1) + "end."
	tests := []struct {
		name string
		in   CheckInput
		want string
	}{
		{"clean", CheckInput{Text: "A plain sentence. Another one."}, ""},
		{"empty", CheckInput{}, "error:no-narration"},
		{"silent", CheckInput{PauseMs: 800}, "info:silent"},
		{"missing", CheckInput{Missing: true, Language: "fr-FR"}, "error:missing-translation"},
		{"url and symbol", CheckInput{Text: "Visit example.com & sign up."}, "warn:url warn:symbol"},
		{"directive", CheckInput{Text: "Wait [BREAK:2] here."}, "warn:unknown-directive"},
		{"numbers", CheckInput{Text: "It grew 80% to 1M users."}, "info:numbers"},
		{"acronym known", CheckInput{Text: "Use PBHQ.", Language: "en-US", Pronunciations: testPron}, ""},
		{"terms are not rechecked", CheckInput{Text: "Use PBHQ Cloud, B2B and v0.", Language: "en-US",
			Pronunciations: map[string]map[string]string{"PBHQ Cloud": {"en-US": "x"}, "B2B": {"en-US": "x"}, "v0": {"en-US": "x"}}}, ""},
		{"digit inside a word is not a number", CheckInput{Text: "Sell B2B.", Language: "en-US"}, "info:acronym"},
		{"acronym unknown", CheckInput{Text: "Add UTM tags.", Language: "en-US", Pronunciations: testPron}, "info:acronym"},
		{"long sentence", CheckInput{Text: long}, "warn:long-sentence"},
	}
	for _, tt := range tests {
		if got := codes(Check(tt.in)); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestHandler(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "deck.html"), []byte(testDeck), 0o600); err != nil {
		t.Fatal(err)
	}
	reviewPath := filepath.Join(dir, "review.json")
	load := func() (*Model, error) {
		rf, err := LoadFile(reviewPath)
		if err != nil {
			return nil, err
		}
		return Build(testInput(t, rf))
	}
	now := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	srv := httptest.NewServer(NewHandler(HandlerOptions{
		Load: load, ReviewPath: reviewPath, DeckDir: dir, Reviewer: "tester",
		Now: func() time.Time { return now },
	}))
	defer srv.Close()

	get := func(path string) (int, string) {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var b bytes.Buffer
		if _, err := b.ReadFrom(resp.Body); err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode, b.String()
	}
	post := func(d Decision) *http.Response {
		b, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.Post(srv.URL+"/api/review", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}

	if code, body := get("/"); code != 200 || !strings.Contains(body, `<script type="application/json" id="model">null</script>`) {
		t.Errorf("page: %d, live model placeholder missing", code)
	}
	if code, body := get("/deck/deck.html"); code != 200 || !strings.Contains(body, "Test Deck") {
		t.Errorf("deck: %d", code)
	}
	if code, _ := get("/api/model"); code != 200 {
		t.Errorf("model: %d", code)
	}

	resp := post(Decision{SlideID: "intro", Language: "en-US", Status: StatusApproved, Note: "  fine  "})
	var m Model
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	n := m.Slides[0].Langs["en-US"]
	if n.State != StateApproved || n.Entry == nil || n.Entry.Note != "fine" || n.Entry.Reviewer != "tester" || !n.Entry.UpdatedAt.Equal(now) {
		t.Errorf("after approve: %+v %+v", n.State, n.Entry)
	}

	for _, bad := range []Decision{
		{SlideID: "intro", Language: "en-US", Status: "maybe"},
		{SlideID: "nope", Language: "en-US", Status: StatusApproved},
		{SlideID: "intro", Language: "de-DE", Status: StatusApproved},
	} {
		if r := post(bad); r.StatusCode < 400 {
			t.Errorf("%+v: status %d, want an error", bad, r.StatusCode)
		} else {
			r.Body.Close()
		}
	}

	r := post(Decision{SlideID: "intro", Language: "en-US"})
	r.Body.Close()
	rf, err := LoadFile(reviewPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := rf.Get("intro", "en-US"); ok {
		t.Error("clearing left the decision in place")
	}
}

func TestExport(t *testing.T) {
	m := build(t, NewFile())
	m.Title = "Deck </script><b>"
	var buf bytes.Buffer
	if err := Export(&buf, m); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	// The speech library, the model and the page script.
	if strings.Count(page, "</script>") != 3 {
		t.Errorf("embedded model closed its script element early")
	}
	if !strings.Contains(page, "PlexusSpeech") {
		t.Error("speech library not embedded")
	}
	start := strings.Index(page, `id="model">`) + len(`id="model">`)
	end := strings.Index(page[start:], "</script>")
	var got Model
	if err := json.Unmarshal([]byte(page[start:start+end]), &got); err != nil {
		t.Fatalf("embedded model: %v", err)
	}
	if !got.ReadOnly || len(got.Slides) != 4 {
		t.Errorf("readOnly %v, %d slides", got.ReadOnly, len(got.Slides))
	}
	if m.ReadOnly {
		t.Error("Export changed the caller's model")
	}
	if !strings.Contains(page, "<title>Deck &lt;/script&gt;&lt;b&gt;") {
		t.Error("title not escaped")
	}
}

func TestMultiHandler(t *testing.T) {
	deck := func(name string) DeckOptions {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "deck.html"), []byte(testDeck), 0o600); err != nil {
			t.Fatal(err)
		}
		reviewPath := filepath.Join(dir, "review.json")
		return DeckOptions{Slug: name, Name: "Deck " + name, HandlerOptions: HandlerOptions{
			Load: func() (*Model, error) {
				rf, err := LoadFile(reviewPath)
				if err != nil {
					return nil, err
				}
				return Build(testInput(t, rf))
			},
			ReviewPath: reviewPath,
			DeckDir:    dir,
		}}
	}
	h, err := NewMultiHandler([]DeckOptions{deck("one"), deck("two")})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	get := func(path string) (int, string) {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var b bytes.Buffer
		if _, err := b.ReadFrom(resp.Body); err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode, b.String()
	}
	if code, body := get("/"); code != 200 || !strings.Contains(body, `href="d/two/"`) || !strings.Contains(body, "Deck one") {
		t.Errorf("index: %d\n%s", code, body)
	}
	code, body := get("/d/two/api/model")
	if code != 200 {
		t.Fatalf("model: %d %s", code, body)
	}
	var m Model
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Decks) != 2 || m.Decks[0].URL != "../one/" || !m.Decks[1].Current || m.Decks[0].Name != "Deck one" {
		t.Errorf("decks = %+v", m.Decks)
	}
	if code, _ := get("/d/one/deck/deck.html"); code != 200 {
		t.Errorf("deck file: %d", code)
	}

	for _, bad := range [][]DeckOptions{nil, {deck("Bad Slug")}, {deck("x"), deck("x")}} {
		if _, err := NewMultiHandler(bad); err == nil {
			t.Errorf("want error for %d decks", len(bad))
		}
	}
}
