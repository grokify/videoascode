package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grokify/videoascode/pkg/browser"
)

func TestDefaultVideoConfig(t *testing.T) {
	cfg := DefaultVideoConfig()

	if cfg.Version != "1.0" {
		t.Errorf("Version = %s, want 1.0", cfg.Version)
	}
	if cfg.DefaultLanguage != "en-US" {
		t.Errorf("DefaultLanguage = %s, want en-US", cfg.DefaultLanguage)
	}
	if cfg.Resolution != Resolution1080p {
		t.Errorf("Resolution = %v, want %v", cfg.Resolution, Resolution1080p)
	}
	if cfg.FrameRate != 30 {
		t.Errorf("FrameRate = %d, want 30", cfg.FrameRate)
	}
	if cfg.Segments == nil {
		t.Error("Segments should not be nil")
	}
}

func TestVideoConfig_GetLanguages(t *testing.T) {
	tests := []struct {
		name     string
		cfg      VideoConfig
		expected []string
	}{
		{
			name:     "uses Languages if set",
			cfg:      VideoConfig{Languages: []string{"en-US", "es-ES"}, DefaultLanguage: "fr-FR"},
			expected: []string{"en-US", "es-ES"},
		},
		{
			name:     "falls back to DefaultLanguage",
			cfg:      VideoConfig{DefaultLanguage: "de-DE"},
			expected: []string{"de-DE"},
		},
		{
			name:     "empty Languages uses DefaultLanguage",
			cfg:      VideoConfig{Languages: []string{}, DefaultLanguage: "ja-JP"},
			expected: []string{"ja-JP"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.GetLanguages()
			if len(got) != len(tc.expected) {
				t.Errorf("GetLanguages() = %v, want %v", got, tc.expected)
				return
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("GetLanguages()[%d] = %s, want %s", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestVideoConfig_GetOutputDir(t *testing.T) {
	tests := []struct {
		name     string
		cfg      VideoConfig
		expected string
	}{
		{
			name:     "returns OutputDir if set",
			cfg:      VideoConfig{OutputDir: "/custom/output"},
			expected: "/custom/output",
		},
		{
			name:     "returns default if empty",
			cfg:      VideoConfig{},
			expected: "output",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.GetOutputDir()
			if got != tc.expected {
				t.Errorf("GetOutputDir() = %s, want %s", got, tc.expected)
			}
		})
	}
}

func TestSegmentConfig_IsSlideSegment(t *testing.T) {
	tests := []struct {
		segType   SourceType
		isSlide   bool
		isBrowser bool
	}{
		{SourceTypeSlide, true, false},
		{SourceTypeBrowser, false, true},
		{"unknown", false, false},
	}

	for _, tc := range tests {
		seg := SegmentConfig{Type: tc.segType}
		if got := seg.IsSlideSegment(); got != tc.isSlide {
			t.Errorf("IsSlideSegment() for %s = %v, want %v", tc.segType, got, tc.isSlide)
		}
		if got := seg.IsBrowserSegment(); got != tc.isBrowser {
			t.Errorf("IsBrowserSegment() for %s = %v, want %v", tc.segType, got, tc.isBrowser)
		}
	}
}

func TestSegmentConfig_GetVoiceover(t *testing.T) {
	seg := SegmentConfig{
		Transcripts: map[string]SegmentTranscript{
			"en-US": {Text: "Hello world"},
			"es-ES": {Text: ""},
		},
	}

	if got := seg.GetVoiceover("en-US"); got != "Hello world" {
		t.Errorf("GetVoiceover(en-US) = %s, want 'Hello world'", got)
	}
	if got := seg.GetVoiceover("es-ES"); got != "" {
		t.Errorf("GetVoiceover(es-ES) = %s, want ''", got)
	}
	if got := seg.GetVoiceover("fr-FR"); got != "" {
		t.Errorf("GetVoiceover(fr-FR) = %s, want ''", got)
	}
}

func TestSegmentConfig_GetStepVoiceovers(t *testing.T) {
	seg := SegmentConfig{
		Steps: []browser.Step{
			{Action: browser.ActionClick, Voiceover: "Click the button"},
			{Action: browser.ActionWait},
			{Action: browser.ActionInput, Voiceover: "Enter your name"},
		},
	}

	got := seg.GetStepVoiceovers()
	expected := []string{"Click the button", "Enter your name"}

	if len(got) != len(expected) {
		t.Errorf("GetStepVoiceovers() returned %d items, want %d", len(got), len(expected))
		return
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Errorf("GetStepVoiceovers()[%d] = %s, want %s", i, got[i], expected[i])
		}
	}
}

func TestResolutionConstants(t *testing.T) {
	if Resolution720p.Width != 1280 || Resolution720p.Height != 720 {
		t.Errorf("Resolution720p = %v, want 1280x720", Resolution720p)
	}
	if Resolution1080p.Width != 1920 || Resolution1080p.Height != 1080 {
		t.Errorf("Resolution1080p = %v, want 1920x1080", Resolution1080p)
	}
	if Resolution4K.Width != 3840 || Resolution4K.Height != 2160 {
		t.Errorf("Resolution4K = %v, want 3840x2160", Resolution4K)
	}
}

func TestVideoConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     VideoConfig
		wantErr bool
	}{
		{
			name:    "empty title",
			cfg:     VideoConfig{},
			wantErr: true,
		},
		{
			name:    "no segments",
			cfg:     VideoConfig{Title: "Test"},
			wantErr: true,
		},
		{
			name: "invalid resolution",
			cfg: VideoConfig{
				Title:      "Test",
				Resolution: Resolution{Width: 0, Height: 0},
				Segments:   []SegmentConfig{{Type: SourceTypeSlide, Source: "test.md"}},
			},
			wantErr: true,
		},
		{
			name: "invalid frame rate",
			cfg: VideoConfig{
				Title:      "Test",
				Resolution: Resolution1080p,
				FrameRate:  0,
				Segments:   []SegmentConfig{{Type: SourceTypeSlide, Source: "test.md"}},
			},
			wantErr: true,
		},
		{
			name: "valid config",
			cfg: VideoConfig{
				Title:      "Test",
				Resolution: Resolution1080p,
				FrameRate:  30,
				Segments:   []SegmentConfig{{Type: SourceTypeSlide, Source: "test.md"}},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestSegmentConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		seg     SegmentConfig
		wantErr bool
	}{
		{
			name:    "slide without source",
			seg:     SegmentConfig{Type: SourceTypeSlide},
			wantErr: true,
		},
		{
			name:    "slide with source",
			seg:     SegmentConfig{Type: SourceTypeSlide, Source: "slides.md"},
			wantErr: false,
		},
		{
			name:    "browser without URL",
			seg:     SegmentConfig{Type: SourceTypeBrowser},
			wantErr: true,
		},
		{
			name:    "browser with URL",
			seg:     SegmentConfig{Type: SourceTypeBrowser, URL: "https://example.com"},
			wantErr: false,
		},
		{
			name:    "unknown type",
			seg:     SegmentConfig{Type: "unknown"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.seg.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	cfg1 := &VideoConfig{
		Title:           "First",
		DefaultLanguage: "en-US",
		Resolution:      Resolution720p,
		FrameRate:       24,
	}
	cfg2 := &VideoConfig{
		Title:     "Second",
		FrameRate: 30,
		Segments:  []SegmentConfig{{Type: SourceTypeSlide, Source: "test.md"}},
	}

	result := MergeConfigs(cfg1, cfg2)

	if result.Title != "Second" {
		t.Errorf("Title = %s, want Second", result.Title)
	}
	if result.DefaultLanguage != "en-US" {
		t.Errorf("DefaultLanguage = %s, want en-US", result.DefaultLanguage)
	}
	if result.FrameRate != 30 {
		t.Errorf("FrameRate = %d, want 30", result.FrameRate)
	}
	if len(result.Segments) != 1 {
		t.Errorf("Segments count = %d, want 1", len(result.Segments))
	}
}

func TestMergeConfigs_Empty(t *testing.T) {
	result := MergeConfigs()
	if result == nil {
		t.Error("MergeConfigs() with no args should return empty config, not nil")
	}
}

func TestLoadFromFile_JSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	jsonContent := `{
		"title": "Test Video",
		"segments": [
			{"type": "slide", "source": "slides.md"}
		]
	}`

	path := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(path, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.Title != "Test Video" {
		t.Errorf("Title = %s, want 'Test Video'", cfg.Title)
	}
	if len(cfg.Segments) != 1 {
		t.Errorf("Segments count = %d, want 1", len(cfg.Segments))
	}
	// Check defaults were applied
	if cfg.Version != "1.0" {
		t.Errorf("Version = %s, want '1.0' (default)", cfg.Version)
	}
	if cfg.FrameRate != 30 {
		t.Errorf("FrameRate = %d, want 30 (default)", cfg.FrameRate)
	}
}

func TestLoadFromFile_YAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	yamlContent := `title: Test Video
segments:
  - type: slide
    source: slides.md
`

	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.Title != "Test Video" {
		t.Errorf("Title = %s, want 'Test Video'", cfg.Title)
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/config.json")
	if err == nil {
		t.Error("LoadFromFile() should return error for nonexistent file")
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfg := &VideoConfig{
		Title:     "Test",
		FrameRate: 30,
	}

	// Test JSON
	jsonPath := filepath.Join(tmpDir, "config.json")
	if err := cfg.SaveToFile(jsonPath); err != nil {
		t.Errorf("SaveToFile(json) error = %v", err)
	}
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("JSON file was not created")
	}

	// Test YAML
	yamlPath := filepath.Join(tmpDir, "config.yaml")
	if err := cfg.SaveToFile(yamlPath); err != nil {
		t.Errorf("SaveToFile(yaml) error = %v", err)
	}
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Error("YAML file was not created")
	}
}
