package tts

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/grokify/videoascode/pkg/segment"
	"github.com/grokify/videoascode/pkg/transcript"
)

func TestSegmentMetadataSaveLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "segment_metadata_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	metadataPath := filepath.Join(tmpDir, "segment_000.json")

	// Create test metadata with per-voiceover durations
	original := segmentMetadata{
		SegmentID:     "segment_000",
		TotalDuration: 15000, // 15 seconds total
		VoiceoverDurations: map[int]int{
			0: 3000, // voiceover 0: 3 seconds
			1: 5000, // voiceover 1: 5 seconds
			2: 4000, // voiceover 2: 4 seconds
			3: 3000, // voiceover 3: 3 seconds
		},
		VoiceoverFiles: map[int]string{
			0: filepath.Join(tmpDir, "segment_000", "voiceover_000.mp3"),
			1: filepath.Join(tmpDir, "segment_000", "voiceover_001.mp3"),
			2: filepath.Join(tmpDir, "segment_000", "voiceover_002.mp3"),
			3: filepath.Join(tmpDir, "segment_000", "voiceover_003.mp3"),
		},
	}

	// Save metadata
	if err := saveSegmentMetadata(metadataPath, original); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Fatal("metadata file was not created")
	}

	// Load metadata
	loaded, err := loadSegmentMetadata(metadataPath)
	if err != nil {
		t.Fatalf("failed to load metadata: %v", err)
	}

	// Verify loaded data matches original
	if loaded.SegmentID != original.SegmentID {
		t.Errorf("SegmentID mismatch: got %s, want %s", loaded.SegmentID, original.SegmentID)
	}

	if loaded.TotalDuration != original.TotalDuration {
		t.Errorf("TotalDuration mismatch: got %d, want %d", loaded.TotalDuration, original.TotalDuration)
	}

	if len(loaded.VoiceoverDurations) != len(original.VoiceoverDurations) {
		t.Errorf("VoiceoverDurations length mismatch: got %d, want %d",
			len(loaded.VoiceoverDurations), len(original.VoiceoverDurations))
	}

	for idx, duration := range original.VoiceoverDurations {
		if loaded.VoiceoverDurations[idx] != duration {
			t.Errorf("VoiceoverDurations[%d] mismatch: got %d, want %d",
				idx, loaded.VoiceoverDurations[idx], duration)
		}
	}
}

func TestSegmentMetadataLoadNonExistent(t *testing.T) {
	_, err := loadSegmentMetadata("/nonexistent/path/metadata.json")
	if err == nil {
		t.Error("expected error when loading non-existent metadata file")
	}
}

func TestCalculateMaxVoiceoverDurations(t *testing.T) {
	// Simulate audio results for multiple languages
	audioResults := map[string]map[string]*SegmentAudioResult{
		"en-US": {
			"segment_000": {
				SegmentID: "segment_000",
				Duration:  10000,
				VoiceoverDurations: map[int]int{
					0: 2000, // English voiceover 0: 2s
					1: 3000, // English voiceover 1: 3s
					2: 5000, // English voiceover 2: 5s
				},
			},
		},
		"fr-FR": {
			"segment_000": {
				SegmentID: "segment_000",
				Duration:  12000, // French is longer
				VoiceoverDurations: map[int]int{
					0: 2500, // French voiceover 0: 2.5s (longer than English)
					1: 4000, // French voiceover 1: 4s (longer than English)
					2: 5500, // French voiceover 2: 5.5s (longer than English)
				},
			},
		},
		"zh-Hans": {
			"segment_000": {
				SegmentID: "segment_000",
				Duration:  8000, // Chinese is shorter
				VoiceoverDurations: map[int]int{
					0: 1500, // Chinese voiceover 0: 1.5s
					1: 2500, // Chinese voiceover 1: 2.5s
					2: 4000, // Chinese voiceover 2: 4s
				},
			},
		},
	}

	// Calculate max durations
	maxDurations := CalculateMaxVoiceoverDurations(audioResults, "segment_000")

	// Verify we get the maximum for each voiceover
	expectedMax := map[int]int{
		0: 2500, // French is longest for voiceover 0
		1: 4000, // French is longest for voiceover 1
		2: 5500, // French is longest for voiceover 2
	}

	if len(maxDurations) != len(expectedMax) {
		t.Errorf("maxDurations length mismatch: got %d, want %d", len(maxDurations), len(expectedMax))
	}

	for idx, expected := range expectedMax {
		if maxDurations[idx] != expected {
			t.Errorf("maxDurations[%d] mismatch: got %d, want %d", idx, maxDurations[idx], expected)
		}
	}
}

func TestCalculateMaxVoiceoverDurationsEmpty(t *testing.T) {
	// Test with empty results
	audioResults := map[string]map[string]*SegmentAudioResult{}
	maxDurations := CalculateMaxVoiceoverDurations(audioResults, "segment_000")

	if len(maxDurations) != 0 {
		t.Errorf("expected empty map for empty results, got %d entries", len(maxDurations))
	}
}

func TestCalculateMaxVoiceoverDurationsNoVoiceovers(t *testing.T) {
	// Test with results that have no voiceover durations (e.g., from old cache)
	audioResults := map[string]map[string]*SegmentAudioResult{
		"en-US": {
			"segment_000": {
				SegmentID:          "segment_000",
				Duration:           10000,
				VoiceoverDurations: map[int]int{}, // Empty - no per-voiceover data
			},
		},
	}

	maxDurations := CalculateMaxVoiceoverDurations(audioResults, "segment_000")

	if len(maxDurations) != 0 {
		t.Errorf("expected empty map when no voiceover durations, got %d entries", len(maxDurations))
	}
}

// fakeSegment is a minimal segment.Segment with fixed voiceovers.
type fakeSegment struct {
	id  string
	vos []segment.Voiceover
}

func (f *fakeSegment) GetID() string                     { return f.id }
func (f *fakeSegment) GetIndex() int                     { return 0 }
func (f *fakeSegment) GetSourceType() segment.SourceType { return segment.SourceTypeBrowser }
func (f *fakeSegment) GetTitle() string                  { return f.id }
func (f *fakeSegment) GetVoiceovers(string) []segment.Voiceover {
	return f.vos
}
func (f *fakeSegment) GetLanguages() []string { return []string{"en-US"} }
func (f *fakeSegment) GetTranscripts() map[string]transcript.LanguageContent {
	return nil
}

// countingGenerator returns a generator whose provider writes a short silent
// MP3 per call, and the number of calls made so far.
func countingGenerator(t *testing.T) (*SegmentTTSGenerator, *[]string) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	silence := filepath.Join(t.TempDir(), "silence.mp3")
	if err := generateSilence(context.Background(), 300, silence); err != nil {
		t.Fatal(err)
	}
	audio, err := os.ReadFile(silence)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	g := NewSegmentTTSGenerator(nil, transcript.VoiceConfig{Provider: "fake", VoiceID: "v1"})
	g.synth = func(_ context.Context, text string, _ transcript.VoiceConfig) ([]byte, error) {
		calls = append(calls, text)
		return audio, nil
	}
	return g, &calls
}

func TestGenerateForSegment_ReusesClips(t *testing.T) {
	g, calls := countingGenerator(t)
	dir := t.TempDir()
	ctx := context.Background()
	seg := &fakeSegment{id: "lesson", vos: []segment.Voiceover{
		{Index: 0, Text: "First slide."},
		{Index: 1, Text: "Second slide.", Pause: 200},
	}}

	res, err := g.GenerateForSegment(ctx, seg, "en-US", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 {
		t.Fatalf("first run: %d TTS calls, want 2", len(*calls))
	}
	if res.AudioFile != filepath.Join(dir, "lesson.mp3") {
		t.Errorf("AudioFile = %s", res.AudioFile)
	}

	// Unchanged: no new TTS calls.
	if _, err := g.GenerateForSegment(ctx, seg, "en-US", dir); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 {
		t.Fatalf("unchanged run: %d TTS calls, want 2", len(*calls))
	}

	// One slide edited: only that clip is synthesized.
	seg.vos[1].Text = "Second slide, revised."
	res, err = g.GenerateForSegment(ctx, seg, "en-US", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 3 || (*calls)[2] != "Second slide, revised." {
		t.Fatalf("edited run: calls = %q", *calls)
	}
	if res.VoiceoverAudioFiles[0] == res.VoiceoverAudioFiles[1] {
		t.Error("distinct voiceovers share a clip")
	}

	// A pause change reassembles the segment without new TTS calls.
	seg.vos[1].Pause = 900
	res2, err := g.GenerateForSegment(ctx, seg, "en-US", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 3 {
		t.Fatalf("pause change: %d TTS calls, want 3", len(*calls))
	}
	if res2.Duration-res.Duration != 700 {
		t.Errorf("pause change: duration %d -> %d, want +700", res.Duration, res2.Duration)
	}
}

func TestGenerateForSegment_VoiceAndPronunciationChangeClip(t *testing.T) {
	g, calls := countingGenerator(t)
	dir := t.TempDir()
	ctx := context.Background()
	seg := &fakeSegment{id: "s", vos: []segment.Voiceover{{Index: 0, Text: "Run vac."}}}

	if _, err := g.GenerateForSegment(ctx, seg, "en-US", dir); err != nil {
		t.Fatal(err)
	}
	g.SetPronunciations(&PronunciationDictionary{Terms: map[string]map[string]string{"vac": {"en-US": "V A C"}}})
	if _, err := g.GenerateForSegment(ctx, seg, "en-US", dir); err != nil {
		t.Fatal(err)
	}
	g.defaultVoice.VoiceID = "v2"
	if _, err := g.GenerateForSegment(ctx, seg, "en-US", dir); err != nil {
		t.Fatal(err)
	}
	want := []string{"Run vac.", "Run V A C.", "Run V A C."}
	if len(*calls) != len(want) {
		t.Fatalf("calls = %q, want %q", *calls, want)
	}
	for i := range want {
		if (*calls)[i] != want[i] {
			t.Errorf("call %d = %q, want %q", i, (*calls)[i], want[i])
		}
	}
}

func TestClipKey(t *testing.T) {
	v := transcript.VoiceConfig{Provider: "deepgram", VoiceID: "aura"}
	k := ClipKey("Hello.", "en-US", v)
	if len(k) != 16 {
		t.Errorf("key %q: want 16 hex chars", k)
	}
	if k != ClipKey("Hello.", "en-US", v) {
		t.Error("key is not stable")
	}
	v2 := v
	v2.Speed = 1.1
	for _, other := range []string{
		ClipKey("Hello!", "en-US", v),
		ClipKey("Hello.", "fr-FR", v),
		ClipKey("Hello.", "en-US", v2),
	} {
		if other == k {
			t.Error("different inputs share a key")
		}
	}
}
