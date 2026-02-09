package tts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManifest(t *testing.T) {
	m := NewManifest("en-US")

	if m.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", m.Version)
	}
	if m.Language != "en-US" {
		t.Errorf("expected language en-US, got %s", m.Language)
	}
	if m.GeneratedAt.IsZero() {
		t.Error("expected non-zero GeneratedAt")
	}
	if len(m.Slides) != 0 {
		t.Errorf("expected empty slides, got %d", len(m.Slides))
	}
}

func TestManifest_AddSlide(t *testing.T) {
	m := NewManifest("en-US")

	m.AddSlide(SlideAudio{
		Index:         0,
		Title:         "Title Slide",
		AudioFile:     "slide_000.mp3",
		AudioDuration: 5000,
		PauseDuration: 500,
		TotalDuration: 5500,
	})

	if len(m.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(m.Slides))
	}

	slide := m.Slides[0]
	if slide.Index != 0 {
		t.Errorf("expected index 0, got %d", slide.Index)
	}
	if slide.TotalDuration != 5500 {
		t.Errorf("expected total duration 5500, got %d", slide.TotalDuration)
	}
}

func TestManifest_GetTotalDuration(t *testing.T) {
	m := NewManifest("en-US")

	m.AddSlide(SlideAudio{Index: 0, TotalDuration: 5000})
	m.AddSlide(SlideAudio{Index: 1, TotalDuration: 3000})
	m.AddSlide(SlideAudio{Index: 2, TotalDuration: 4500})

	total := m.GetTotalDuration()
	if total != 12500 {
		t.Errorf("expected total 12500, got %d", total)
	}
}

func TestManifest_GetSlide(t *testing.T) {
	m := NewManifest("en-US")

	m.AddSlide(SlideAudio{Index: 0, Title: "First"})
	m.AddSlide(SlideAudio{Index: 2, Title: "Third"}) // Skipping index 1

	// Find existing slide
	slide, err := m.GetSlide(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slide.Title != "Third" {
		t.Errorf("expected title 'Third', got '%s'", slide.Title)
	}

	// Try to find non-existent slide
	_, err = m.GetSlide(1)
	if err == nil {
		t.Error("expected error for missing slide")
	}
}

func TestManifest_SaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "manifest_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create and save manifest
	m := NewManifest("es-ES")
	m.GeneratedAt = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	m.AddSlide(SlideAudio{
		Index:         0,
		Title:         "Título",
		AudioFile:     "slide_000.mp3",
		AudioDuration: 5000,
		PauseDuration: 500,
		TotalDuration: 5500,
	})

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if err := m.SaveToFile(manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Load manifest
	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	// Verify loaded data
	if loaded.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", loaded.Version)
	}
	if loaded.Language != "es-ES" {
		t.Errorf("expected language es-ES, got %s", loaded.Language)
	}
	if len(loaded.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(loaded.Slides))
	}
	if loaded.Slides[0].Title != "Título" {
		t.Errorf("expected title 'Título', got '%s'", loaded.Slides[0].Title)
	}
}

func TestLoadManifest_NotExists(t *testing.T) {
	_, err := LoadManifest("/nonexistent/path/manifest.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "manifest_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write invalid JSON
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not json"), 0600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = LoadManifest(invalidPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
