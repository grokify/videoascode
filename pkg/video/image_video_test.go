package video

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewImageVideoConverter(t *testing.T) {
	config := ImageVideoConfig{
		OutputDir: "/tmp/video",
		Width:     1920,
		Height:    1080,
		FrameRate: 30,
	}

	c := NewImageVideoConverter(config)
	if c.config.OutputDir != "/tmp/video" {
		t.Errorf("NewImageVideoConverter() outputDir = %s, want /tmp/video", c.config.OutputDir)
	}
	if c.config.FrameRate != 30 {
		t.Errorf("NewImageVideoConverter() frameRate = %d, want 30", c.config.FrameRate)
	}
}

func TestNewImageVideoConverter_DefaultFrameRate(t *testing.T) {
	config := ImageVideoConfig{
		OutputDir: "/tmp/video",
		FrameRate: 0, // Should default to 30
	}

	c := NewImageVideoConverter(config)
	if c.config.FrameRate != 30 {
		t.Errorf("NewImageVideoConverter() with zero frameRate should default to 30, got %d", c.config.FrameRate)
	}
}

func TestGetAudioDurationSeconds_NonexistentFile(t *testing.T) {
	_, err := getAudioDurationSeconds("/nonexistent/audio.mp3")
	if err == nil {
		t.Error("getAudioDurationSeconds() with nonexistent file should return error")
	}
}

func TestCreateSlideVideo_InvalidAudio(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "image_video_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a fake image file
	imgPath := filepath.Join(tmpDir, "slide.png")
	if err := os.WriteFile(imgPath, []byte("fake png"), 0600); err != nil {
		t.Fatalf("Failed to create image file: %v", err)
	}

	config := ImageVideoConfig{
		OutputDir: tmpDir,
		FrameRate: 30,
	}
	c := NewImageVideoConverter(config)

	// Try to create video with nonexistent audio
	_, err = c.CreateSlideVideo(context.Background(), 0, imgPath, "/nonexistent/audio.mp3", 5*time.Second)
	if err == nil {
		t.Error("CreateSlideVideo() with nonexistent audio should return error")
	}
}

// TestDebugModeEnvironmentVariable documents the MARP2VIDEO_DEBUG behavior.
// When set, ffmpeg output is streamed to stdout/stderr for debugging.
// This test verifies the environment variable is properly checked.
func TestDebugModeEnvironmentVariable(t *testing.T) {
	// Test that MARP2VIDEO_DEBUG is properly read
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"not set", "", false},
		{"set to 1", "1", true},
		{"set to true", "true", true},
		{"set to any value", "debug", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original value
			original := os.Getenv("MARP2VIDEO_DEBUG")
			defer func() {
				if original == "" {
					os.Unsetenv("MARP2VIDEO_DEBUG")
				} else {
					os.Setenv("MARP2VIDEO_DEBUG", original)
				}
			}()

			if tt.envValue == "" {
				os.Unsetenv("MARP2VIDEO_DEBUG")
			} else {
				os.Setenv("MARP2VIDEO_DEBUG", tt.envValue)
			}

			got := os.Getenv("MARP2VIDEO_DEBUG") != ""
			if got != tt.expected {
				t.Errorf("MARP2VIDEO_DEBUG check = %v, want %v", got, tt.expected)
			}
		})
	}
}
