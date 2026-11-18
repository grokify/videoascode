package recording

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/academyosio/course-spec/deck"
)

const testDeck = `<html><body>
<section class="slide" data-id="intro" data-vo="Welcome. [PAUSE:500]"></section>
<section class="slide" data-id="tractors" data-lesson="Tractors" data-title="Tractors" data-vo="He built tractors. From war surplus."></section>
<section class="slide" data-id="cars" data-vo="Then cars."></section>
<section class="slide" data-id="later" data-lesson="Later" data-vo="Later."></section>
</body></html>`

func writeDeck(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "sections", "01-career")
	if err := os.MkdirAll(filepath.Join(dir, "deck", "voiceover"), 0o750); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"section.json":                   `{"id": "career", "title": "Career", "source": {"type": "htmlDeck", "path": "deck/index.html"}}`,
		"deck/index.html":                testDeck,
		"deck/voiceover/transcript.json": `{"slides": [{"id": "intro", "index": 0, "transcripts": {"it-IT": {"segments": [{"text": "Benvenuti."}]}}}, {"id": "tractors", "index": 1, "transcripts": {"it-IT": {"segments": [{"text": "Trattori."}]}}}, {"id": "cars", "index": 2, "transcripts": {"it-IT": {"segments": [{"text": "Poi auto."}]}}}]}`,
		"audio.mp3":                      "ID3 audio",
	} {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "deck", "index.html")
}

var firstLesson = []Slide{
	{ID: "intro", Title: "Intro", Sentences: []string{"Welcome."}, SpeechMs: 1000, PauseMs: 500},
	{ID: "tractors", Title: "Tractors", Sentences: []string{"He built tractors.", "From war surplus."}, SpeechMs: 3500},
	{ID: "cars", Title: "Cars", Sentences: []string{"Then cars."}, SpeechMs: 1200},
}

func TestTimings(t *testing.T) {
	ts, total := Timings(firstLesson)
	want := []Timing{{0, 1500}, {1500, 3500}, {5000, 1200}}
	if !reflect.DeepEqual(ts, want) || total != 6200 {
		t.Errorf("timings = %v, %d", ts, total)
	}
}

func TestCaptions(t *testing.T) {
	got := Captions(firstLesson)
	for _, want := range []string{
		"WEBVTT\n",
		"\n1\n00:00:00.000 --> 00:00:01.000\nWelcome.\n",
		// 18 and 17 characters share 3.5 s.
		"\n2\n00:00:01.500 --> 00:00:03.300\nHe built tractors.\n",
		"\n3\n00:00:03.300 --> 00:00:05.000\nFrom war surplus.\n",
		"\n4\n00:00:05.000 --> 00:00:06.200\nThen cars.\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("captions lack %q:\n%s", want, got)
		}
	}
	if vttTime(3723456) != "01:02:03.456" {
		t.Errorf("vttTime = %s", vttTime(3723456))
	}
}

func TestChapters(t *testing.T) {
	slides := []Slide{
		{ID: "a", Title: "Start", SpeechMs: 12000},
		{ID: "b", Title: "Short", SpeechMs: 4000}, // joins Start
		{ID: "c", Title: "Middle", SpeechMs: 15000},
		{ID: "d", Title: "End", SpeechMs: 70000},
	}
	got := Chapters(slides, 1000)
	want := []string{"0:00 Start", "0:17 Middle", "0:32 End"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("chapters = %v", got)
	}
	if got := Chapters(slides[:2], 0); got != nil {
		t.Errorf("too few chapters = %v", got)
	}
}

func TestWrite(t *testing.T) {
	deckPath := writeDeck(t)
	sec, err := SectionID(deckPath)
	if err != nil || sec != "career" {
		t.Fatalf("section = %q, %v", sec, err)
	}
	audio := filepath.Join(filepath.Dir(filepath.Dir(deckPath)), "audio.mp3")
	l := Lesson{DeckPath: deckPath, SectionID: sec, Language: "en-US", DefaultLanguage: "en-US", AudioFile: audio,
		Slides: firstLesson, Voice: deck.RecordingVoice{Provider: "elevenlabs", VoiceID: "v1", Name: "Marco"}}

	if st, err := Status(deckPath, sec, "intro", "en-US", "en-US"); err != nil || st != deck.RecordingMissing {
		t.Fatalf("status before = %s, %v", st, err)
	}
	rec, err := Write(l)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Lesson != "tractors" || rec.Audio != "en-US/tractors.mp3" || rec.DurationMs != 6200 || len(rec.Slides) != 3 || rec.Slides[1].StartMs != 1500 {
		t.Errorf("recording = %+v", rec)
	}
	recDir := filepath.Join(filepath.Dir(deckPath), "recordings")
	if data, err := os.ReadFile(filepath.Join(recDir, "en-US", "tractors.mp3")); err != nil || string(data) != "ID3 audio" { //nolint:gosec // test file
		t.Errorf("audio = %q, %v", data, err)
	}
	if data, err := os.ReadFile(filepath.Join(recDir, "en-US", "tractors.vtt")); err != nil || !strings.HasPrefix(string(data), "WEBVTT") { //nolint:gosec // test file
		t.Errorf("captions = %q, %v", data, err)
	}
	if st, err := Status(deckPath, sec, "intro", "en-US", "en-US"); err != nil || st != deck.RecordingCurrent {
		t.Errorf("status after = %s, %v", st, err)
	}

	// An Italian recording uses the transcript, and both are kept.
	it := l
	it.Language = "it-IT"
	if _, err := Write(it); err != nil {
		t.Fatal(err)
	}
	all, err := deck.ReadRecordings(filepath.Join(recDir, "recordings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Lessons) != 2 || all.Find("tractors", "it-IT") == nil {
		t.Errorf("recordings = %+v", all.Lessons)
	}
	if st, err := Status(deckPath, sec, "intro", "it-IT", "en-US"); err != nil || st != deck.RecordingCurrent {
		t.Errorf("Italian status = %s, %v", st, err)
	}

	// Editing a slide's narration makes the recording stale.
	html, err := os.ReadFile(deckPath) //nolint:gosec // test file
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deckPath, []byte(strings.Replace(string(html), "Then cars.", "Then fast cars.", 1)), 0o600); err != nil { //nolint:gosec // a test file in a temporary directory
		t.Fatal(err)
	}
	if st, err := Status(deckPath, sec, "intro", "en-US", "en-US"); err != nil || st != deck.RecordingStale {
		t.Errorf("status after an edit = %s, %v", st, err)
	}
}

func TestWriteChecksSlides(t *testing.T) {
	deckPath := writeDeck(t)
	audio := filepath.Join(filepath.Dir(filepath.Dir(deckPath)), "audio.mp3")
	l := Lesson{DeckPath: deckPath, SectionID: "career", Language: "en-US", DefaultLanguage: "en-US", AudioFile: audio, Slides: firstLesson[:2]}
	if _, err := Write(l); err == nil || !strings.Contains(err.Error(), "3 slides") {
		t.Errorf("a partial lesson = %v", err)
	}
	l.Slides = []Slide{{ID: "nope"}}
	if _, err := Write(l); err == nil {
		t.Error("an unknown slide was accepted")
	}
}
