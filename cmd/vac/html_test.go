package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/grokify/videoascode/pkg/htmldeck"
)

func TestServeDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<p>deck</p>"), 0o600); err != nil {
		t.Fatal(err)
	}

	base, stop, err := serveDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	if !strings.HasPrefix(base, "http://127.0.0.1:") {
		t.Errorf("base URL %q is not on loopback", base)
	}
	resp, err := http.Get(base + "/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || string(body) != "<p>deck</p>" {
		t.Errorf("GET index.html = %d %q", resp.StatusCode, body)
	}
}

func TestWithQuery(t *testing.T) {
	saved := hdQuery
	t.Cleanup(func() { hdQuery = saved })

	for _, tt := range []struct{ query, url, want string }{
		{"capture=1", "http://h/index.html", "http://h/index.html?capture=1"},
		{"capture=1", "http://h/index.html?x=1", "http://h/index.html?x=1&capture=1"},
		{"", "http://h/index.html", "http://h/index.html"},
	} {
		hdQuery = tt.query
		if got := withQuery(tt.url); got != tt.want {
			t.Errorf("withQuery(%q) with query %q = %q, want %q", tt.url, tt.query, got, tt.want)
		}
	}
}

func TestPlanAndSelectLessons(t *testing.T) {
	dir := t.TempDir()
	deckPath := filepath.Join(dir, "index.html")
	deckHTML := `<title>Deck</title>
<section class="slide" data-id="one" data-lesson="One" data-vo="First lesson.">1</section>
<section class="slide" data-id="two" data-lesson="Two" data-vo="Second lesson.">2</section>
<section class="slide" data-id="three" data-lesson="Three" data-vo="Third lesson.">3</section>`
	if err := os.WriteFile(deckPath, []byte(deckHTML), 0o600); err != nil {
		t.Fatal(err)
	}
	oldInput, oldOut := hdInput, hdOutputDir
	t.Cleanup(func() { hdInput, hdOutputDir = oldInput, oldOut })
	hdInput, hdOutputDir = deckPath, ""

	in, err := loadHTMLDeck()
	if err != nil {
		t.Fatal(err)
	}
	_, plans, err := planLessons(in, "http://x/index.html")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range plans {
		if p.status != statusNotRecorded {
			t.Fatalf("lesson %d: status %q before any recording", p.lesson.Number, p.status)
		}
	}

	// Lesson 1 recorded from the current deck, lesson 2 from an older one,
	// lesson 3 listed but its video deleted.
	out := filepath.Join(dir, "lessons")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"01-one.mp4", "02-two.mp4"} {
		if err := os.WriteFile(filepath.Join(out, f), []byte("video"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	m := &htmldeck.Manifest{Title: "Deck", Lessons: []htmldeck.ManifestLesson{
		{Number: 1, ID: "one", Videos: map[string]string{"en-US": "01-one.mp4"}, DurationMs: 61000, Fingerprint: plans[0].fingerprint},
		{Number: 2, ID: "two", Videos: map[string]string{"en-US": "02-two.mp4"}, Fingerprint: "older"},
		{Number: 3, ID: "three", Videos: map[string]string{"en-US": "03-three.mp4"}, Fingerprint: plans[2].fingerprint},
	}}
	if err := m.Save(filepath.Join(out, htmldeck.ManifestFile)); err != nil {
		t.Fatal(err)
	}
	_, plans, err = planLessons(in, "http://y/index.html")
	if err != nil {
		t.Fatal(err)
	}
	wantStatus := []string{statusRecorded, statusChanged, statusNotRecorded}
	for i, p := range plans {
		if p.status != wantStatus[i] {
			t.Errorf("lesson %d: status %q, want %q", i+1, p.status, wantStatus[i])
		}
	}

	for refs, want := range map[string]string{
		"changed":   "1,2",
		"all":       "0,1,2",
		"3,one":     "0,2",
		"two,2,two": "1",
	} {
		got, err := selectLessons(in.deck, plans, strings.Split(refs, ","))
		if err != nil {
			t.Fatalf("%s: %v", refs, err)
		}
		gotS := make([]string, len(got))
		for i, n := range got {
			gotS[i] = strconv.Itoa(n)
		}
		if strings.Join(gotS, ",") != want {
			t.Errorf("selectLessons(%s) = %v, want %s", refs, got, want)
		}
	}
	if _, err := selectLessons(in.deck, plans, []string{"nope"}); err == nil {
		t.Error("expected an error for an unknown lesson")
	}

	var buf bytes.Buffer
	htmlLessonsCmd.SetOut(&buf)
	t.Cleanup(func() { htmlLessonsCmd.SetOut(nil) })
	if err := runHTMLLessons(htmlLessonsCmd, nil); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"3 lessons", "recorded (1:01)", "changed", "not recorded"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("lessons output missing %q:\n%s", want, buf.String())
		}
	}
}

func TestLanguageVideoPath(t *testing.T) {
	if got := languageVideoPath("out/01-a.mp4", "fr-FR"); got != "out/01-a_fr-FR.mp4" {
		t.Errorf("got %s", got)
	}
}
