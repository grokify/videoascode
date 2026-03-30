package transcript

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLanguageContent_GetFullText(t *testing.T) {
	lc := LanguageContent{
		Segments: []Segment{
			{Text: "Hello"},
			{Text: "World"},
			{Text: "Test"},
		},
	}

	expected := "Hello World Test"
	if got := lc.GetFullText(); got != expected {
		t.Errorf("GetFullText() = %s, want %s", got, expected)
	}
}

func TestLanguageContent_GetFullText_Empty(t *testing.T) {
	lc := LanguageContent{}
	if got := lc.GetFullText(); got != "" {
		t.Errorf("GetFullText() on empty = %s, want ''", got)
	}
}

func TestLanguageContent_GetTotalPauseDuration(t *testing.T) {
	lc := LanguageContent{
		Segments: []Segment{
			{Text: "First", Pause: 500},
			{Text: "Second", Pause: 1000},
			{Text: "Third"},
		},
	}

	if got := lc.GetTotalPauseDuration(); got != 1500 {
		t.Errorf("GetTotalPauseDuration() = %d, want 1500", got)
	}
}

func TestSlide_GetEffectiveSourceType(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceType
		expected   SourceType
	}{
		{"empty defaults to slide", "", SourceTypeSlide},
		{"slide stays slide", SourceTypeSlide, SourceTypeSlide},
		{"browser stays browser", SourceTypeBrowser, SourceTypeBrowser},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := Slide{SourceType: tc.sourceType}
			if got := s.GetEffectiveSourceType(); got != tc.expected {
				t.Errorf("GetEffectiveSourceType() = %s, want %s", got, tc.expected)
			}
		})
	}
}

func TestSlide_IsSlideSegment(t *testing.T) {
	slide := Slide{SourceType: SourceTypeSlide}
	if !slide.IsSlideSegment() {
		t.Error("IsSlideSegment() = false, want true")
	}
	if slide.IsBrowserSegment() {
		t.Error("IsBrowserSegment() = true, want false")
	}

	browser := Slide{SourceType: SourceTypeBrowser}
	if browser.IsSlideSegment() {
		t.Error("IsSlideSegment() for browser = true, want false")
	}
	if !browser.IsBrowserSegment() {
		t.Error("IsBrowserSegment() for browser = false, want true")
	}
}

func TestSlide_GetBrowserVoiceovers(t *testing.T) {
	slide := Slide{
		SourceType: SourceTypeBrowser,
		BrowserSteps: []BrowserStep{
			{Action: "click", Voiceover: "Click the button"},
			{Action: "wait"},
			{Action: "input", Voiceover: "Enter your name"},
		},
	}

	voiceovers := slide.GetBrowserVoiceovers()
	expected := []string{"Click the button", "Enter your name"}
	if len(voiceovers) != len(expected) {
		t.Fatalf("GetBrowserVoiceovers() returned %d items, want %d", len(voiceovers), len(expected))
	}
	for i, v := range voiceovers {
		if v != expected[i] {
			t.Errorf("GetBrowserVoiceovers()[%d] = %s, want %s", i, v, expected[i])
		}
	}

	// Non-browser slide should return nil
	slideOnly := Slide{SourceType: SourceTypeSlide}
	if got := slideOnly.GetBrowserVoiceovers(); got != nil {
		t.Errorf("GetBrowserVoiceovers() for slide = %v, want nil", got)
	}
}

func TestTranscript_GetSlideTranscript(t *testing.T) {
	tr := &Transcript{
		Metadata: Metadata{DefaultLanguage: "en-US"},
		Slides: []Slide{
			{
				Index: 0,
				Transcripts: map[string]LanguageContent{
					"en-US": {Segments: []Segment{{Text: "English"}}},
					"es-ES": {Segments: []Segment{{Text: "Spanish"}}},
				},
			},
		},
	}

	// Get existing language
	content, err := tr.GetSlideTranscript(0, "es-ES")
	if err != nil {
		t.Fatalf("GetSlideTranscript() error = %v", err)
	}
	if content.GetFullText() != "Spanish" {
		t.Errorf("GetSlideTranscript(es-ES) = %s, want 'Spanish'", content.GetFullText())
	}

	// Fall back to default language
	content, err = tr.GetSlideTranscript(0, "fr-FR")
	if err != nil {
		t.Fatalf("GetSlideTranscript(fr-FR) fallback error = %v", err)
	}
	if content.GetFullText() != "English" {
		t.Errorf("GetSlideTranscript(fr-FR) fallback = %s, want 'English'", content.GetFullText())
	}

	// Out of range index
	_, err = tr.GetSlideTranscript(10, "en-US")
	if err == nil {
		t.Error("GetSlideTranscript() with out of range index should return error")
	}
}

func TestTranscript_GetBrowserSlides(t *testing.T) {
	tr := &Transcript{
		Slides: []Slide{
			{Index: 0, SourceType: SourceTypeSlide},
			{Index: 1, SourceType: SourceTypeBrowser},
			{Index: 2, SourceType: SourceTypeSlide},
			{Index: 3, SourceType: SourceTypeBrowser},
		},
	}

	browserSlides := tr.GetBrowserSlides()
	if len(browserSlides) != 2 {
		t.Errorf("GetBrowserSlides() returned %d items, want 2", len(browserSlides))
	}
	if browserSlides[0].Index != 1 {
		t.Errorf("First browser slide index = %d, want 1", browserSlides[0].Index)
	}
}

func TestTranscript_GetSlideSlides(t *testing.T) {
	tr := &Transcript{
		Slides: []Slide{
			{Index: 0, SourceType: SourceTypeSlide},
			{Index: 1, SourceType: SourceTypeBrowser},
			{Index: 2}, // Empty defaults to slide
		},
	}

	slideSlides := tr.GetSlideSlides()
	if len(slideSlides) != 2 {
		t.Errorf("GetSlideSlides() returned %d items, want 2", len(slideSlides))
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "transcript_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	jsonContent := `{
		"version": "1.0",
		"metadata": {
			"title": "Test",
			"defaultLanguage": "en-US"
		},
		"slides": [
			{
				"index": 0,
				"transcripts": {
					"en-US": {
						"segments": [{"text": "Hello"}]
					}
				}
			}
		]
	}`

	path := filepath.Join(tmpDir, "transcript.json")
	if err := os.WriteFile(path, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tr, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if tr.Version != "1.0" {
		t.Errorf("Version = %s, want '1.0'", tr.Version)
	}
	if tr.Metadata.Title != "Test" {
		t.Errorf("Title = %s, want 'Test'", tr.Metadata.Title)
	}
	if len(tr.Slides) != 1 {
		t.Errorf("Slides count = %d, want 1", len(tr.Slides))
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/transcript.json")
	if err == nil {
		t.Error("LoadFromFile() should return error for nonexistent file")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "transcript_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	path := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(path, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadFromFile(path)
	if err == nil {
		t.Error("LoadFromFile() should return error for invalid JSON")
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "transcript_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tr := &Transcript{
		Version: "1.0",
		Metadata: Metadata{
			Title:           "Test",
			DefaultLanguage: "en-US",
		},
		Slides: []Slide{
			{Index: 0},
		},
	}

	path := filepath.Join(tmpDir, "output.json")
	if err := tr.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Verify file exists and can be loaded back
	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("Failed to load saved file: %v", err)
	}
	if loaded.Version != "1.0" {
		t.Errorf("Loaded version = %s, want '1.0'", loaded.Version)
	}
}

func TestVoiceConfig_Fields(t *testing.T) {
	vc := VoiceConfig{
		Provider:        "elevenlabs",
		VoiceID:         "abc123",
		VoiceName:       "Test Voice",
		Model:           "eleven_turbo_v2",
		OutputFormat:    "mp3",
		SampleRate:      44100,
		Speed:           1.0,
		Pitch:           0.0,
		Stability:       0.5,
		SimilarityBoost: 0.75,
		Style:           0.3,
	}

	if vc.Provider != "elevenlabs" {
		t.Errorf("Provider = %s, want 'elevenlabs'", vc.Provider)
	}
	if vc.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", vc.SampleRate)
	}
}

func TestSourceTypeConstants(t *testing.T) {
	if SourceTypeSlide != "slide" {
		t.Errorf("SourceTypeSlide = %s, want 'slide'", SourceTypeSlide)
	}
	if SourceTypeBrowser != "browser" {
		t.Errorf("SourceTypeBrowser = %s, want 'browser'", SourceTypeBrowser)
	}
}

func TestBrowserStep_Fields(t *testing.T) {
	step := BrowserStep{
		Action:         "click",
		Selector:       "#button",
		Value:          "test",
		URL:            "https://example.com",
		Duration:       1000,
		Script:         "console.log('test')",
		Voiceover:      "Click the button",
		Description:    "Test step",
		ScrollX:        100,
		ScrollY:        200,
		ScrollMode:     "absolute",
		ScrollBehavior: "smooth",
	}

	if step.Action != "click" {
		t.Errorf("Action = %s, want 'click'", step.Action)
	}
	if step.ScrollY != 200 {
		t.Errorf("ScrollY = %d, want 200", step.ScrollY)
	}
}

func TestSSMLHints_Fields(t *testing.T) {
	hints := SSMLHints{
		Breaks:   []string{"500ms", "1s"},
		Emphasis: []string{"important", "key"},
		Prosody:  "rate='slow'",
		SayAs:    "date",
		Phoneme:  "ˈtɛst",
		SubAlias: "test alias",
	}

	if len(hints.Breaks) != 2 {
		t.Errorf("Breaks count = %d, want 2", len(hints.Breaks))
	}
	if hints.SayAs != "date" {
		t.Errorf("SayAs = %s, want 'date'", hints.SayAs)
	}
}
