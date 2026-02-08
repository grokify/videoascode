package video

import (
	"testing"
	"time"
)

func TestNewRecorder(t *testing.T) {
	tests := []struct {
		name           string
		config         RecorderConfig
		wantFrameRate  int
	}{
		{
			name:          "default frame rate",
			config:        RecorderConfig{OutputDir: "/tmp"},
			wantFrameRate: 30,
		},
		{
			name:          "custom frame rate",
			config:        RecorderConfig{OutputDir: "/tmp", FrameRate: 60},
			wantFrameRate: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRecorder(tt.config)
			if r.config.FrameRate != tt.wantFrameRate {
				t.Errorf("NewRecorder() FrameRate = %d, want %d", r.config.FrameRate, tt.wantFrameRate)
			}
		})
	}
}

func TestBuildMacOSCommand(t *testing.T) {
	r := NewRecorder(RecorderConfig{
		OutputDir:    "/tmp/video",
		Width:        1920,
		Height:       1080,
		FrameRate:    30,
		ScreenDevice: "2:none",
	})

	args := r.buildMacOSCommand("/tmp/output.mp4", "/tmp/audio.mp3", 5*time.Second)

	// Check essential arguments are present
	expectedArgs := map[string]string{
		"-f":               "avfoundation",
		"-framerate":       "30",
		"-i":               "2:none", // First -i is screen device
		"-t":               "5.00",
		"-vcodec":          "libx264",
		"-acodec":          "aac",
	}

	argMap := make(map[string]string)
	for i := 0; i < len(args)-1; i++ {
		if args[i][0] == '-' {
			argMap[args[i]] = args[i+1]
		}
	}

	for key, want := range expectedArgs {
		if got, ok := argMap[key]; !ok {
			t.Errorf("buildMacOSCommand() missing arg %s", key)
		} else if key != "-i" && got != want {
			t.Errorf("buildMacOSCommand() %s = %s, want %s", key, got, want)
		}
	}

	// Check output path is last argument
	if args[len(args)-1] != "/tmp/output.mp4" {
		t.Errorf("buildMacOSCommand() output path = %s, want /tmp/output.mp4", args[len(args)-1])
	}
}

func TestBuildLinuxCommand(t *testing.T) {
	r := NewRecorder(RecorderConfig{
		OutputDir: "/tmp/video",
		Width:     1280,
		Height:    720,
		FrameRate: 24,
	})

	args := r.buildLinuxCommand("/tmp/output.mp4", "/tmp/audio.mp3", 10*time.Second)

	// Verify essential x11grab arguments
	foundX11grab := false
	foundVideoSize := false
	for i, arg := range args {
		if arg == "-f" && i+1 < len(args) && args[i+1] == "x11grab" {
			foundX11grab = true
		}
		if arg == "-video_size" && i+1 < len(args) && args[i+1] == "1280x720" {
			foundVideoSize = true
		}
	}

	if !foundX11grab {
		t.Error("buildLinuxCommand() should use x11grab")
	}
	if !foundVideoSize {
		t.Error("buildLinuxCommand() should set video_size to 1280x720")
	}
}

func TestBuildWindowsCommand(t *testing.T) {
	r := NewRecorder(RecorderConfig{
		OutputDir: "/tmp/video",
		Width:     1920,
		Height:    1080,
		FrameRate: 30,
	})

	args := r.buildWindowsCommand("/tmp/output.mp4", "/tmp/audio.mp3", 5*time.Second)

	// Verify gdigrab is used
	foundGdigrab := false
	foundDesktop := false
	for i, arg := range args {
		if arg == "-f" && i+1 < len(args) && args[i+1] == "gdigrab" {
			foundGdigrab = true
		}
		if arg == "-i" && i+1 < len(args) && args[i+1] == "desktop" {
			foundDesktop = true
		}
	}

	if !foundGdigrab {
		t.Error("buildWindowsCommand() should use gdigrab")
	}
	if !foundDesktop {
		t.Error("buildWindowsCommand() should capture 'desktop'")
	}
}

func TestCheckFFmpeg(t *testing.T) {
	// This test checks that CheckFFmpeg doesn't panic
	// In CI without ffmpeg, it will return an error which is fine
	err := CheckFFmpeg()
	// We just verify the function runs without panicking
	_ = err
}
