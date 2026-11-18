package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/videoascode/pkg/transcript"
	"github.com/grokify/videoascode/pkg/tts"
)

// retakeDeck writes a deck with lessons "one" (slides one, one-b, one-c)
// and "two" (slide two), and returns its plans.
func retakeDeck(t *testing.T) []lessonPlan {
	t.Helper()
	dir := t.TempDir()
	deckPath := filepath.Join(dir, "index.html")
	deckHTML := `<title>Deck</title>
<section class="slide" data-id="one" data-lesson="One" data-vo="First slide.">1</section>
<section class="slide" data-id="one-b" data-vo="Second slide.">2</section>
<section class="slide" data-id="one-c" data-vo="Third slide.">3</section>
<section class="slide" data-id="two" data-lesson="Two" data-vo="Second lesson.">4</section>`
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
	if len(plans) != 2 || len(plans[0].lesson.Slides) != 3 {
		t.Fatalf("unexpected lessons: %+v", plans)
	}
	return plans
}

func TestPlanRetakes(t *testing.T) {
	plans := retakeDeck(t)

	got, err := planRetakes(plans, []int{0, 1}, []string{"one-c", " two", "one", "one-c"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || len(got[0]) != 2 || got[0][0] != 2 || got[0][1] != 0 || len(got[1]) != 1 || got[1][0] != 0 {
		t.Errorf("planRetakes = %v, want map[0:[2 0] 1:[0]]", got)
	}
	if got, err := planRetakes(plans, []int{0}, nil); err != nil || len(got) != 0 {
		t.Errorf("no --retake: %v, %v", got, err)
	}

	for _, tc := range []struct {
		refs     []string
		selected []int
		want     string
	}{
		{[]string{"nope"}, []int{0, 1}, `no slide "nope"`},
		{[]string{"one-b", "two"}, []int{0}, `slide "two" is in lesson 2 (two)`},
	} {
		_, err := planRetakes(plans, tc.selected, tc.refs)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("planRetakes(%v): error %v, want one containing %q", tc.refs, err, tc.want)
		}
	}
}

func TestRetakeLessonClips(t *testing.T) {
	plans := retakeDeck(t)
	p := plans[0]
	audioDir := t.TempDir()
	voice := transcript.VoiceConfig{Provider: "fake", VoiceID: "v1"}
	gen := tts.NewSegmentTTSGenerator(nil, voice)
	langs := []string{"en-US"}

	seg, err := lessonSegment(p.cfg)
	if err != nil {
		t.Fatal(err)
	}
	langDir := filepath.Join(audioDir, "en-US")
	clipDir := filepath.Join(langDir, tts.ClipDir)
	if err := os.MkdirAll(clipDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var clips []string
	for _, vo := range seg.GetVoiceovers("en-US") {
		clips = append(clips, filepath.Join(clipDir, tts.ClipKey(vo.Text, "en-US", voice)+".mp3"))
	}
	if len(clips) != 3 {
		t.Fatalf("%d voiceovers, want 3", len(clips))
	}
	meta := filepath.Join(langDir, seg.GetID()+".json")
	for _, f := range append(append([]string{}, clips...), meta) {
		if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var out bytes.Buffer
	if err := retakeLessonClips(&out, gen, audioDir, p, []int{1}, langs); err != nil {
		t.Fatal(err)
	}
	for i, f := range clips {
		_, err := os.Stat(f)
		if kept := err == nil; kept != (i != 1) {
			t.Errorf("clip of slide %d kept = %v (stat: %v)", i, kept, err)
		}
	}
	if _, err := os.Stat(meta); !os.IsNotExist(err) {
		t.Errorf("lesson audio metadata not removed: %v", err)
	}
	if want := "re-take one-b (en-US): " + filepath.Base(clips[1]); !strings.Contains(out.String(), want) {
		t.Errorf("output %q missing %q", out.String(), want)
	}
}

func TestRunHTMLVideo_RetakeNeedsLessons(t *testing.T) {
	oldLessons, oldRetake := hvLessons, hvRetake
	t.Cleanup(func() { hvLessons, hvRetake = oldLessons, oldRetake })
	hvLessons, hvRetake = nil, []string{"one"}
	err := runHTMLVideo(htmlVideoCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--retake") {
		t.Errorf("runHTMLVideo: error %v, want one about --retake", err)
	}
}
