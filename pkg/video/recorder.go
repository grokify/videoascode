package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/grokify/mogo/log/slogutil"
)

// RecorderConfig holds video recording configuration
type RecorderConfig struct {
	OutputDir    string
	Width        int
	Height       int
	FrameRate    int
	AudioInput   string // Path to audio file to play during recording
	ScreenDevice string // Screen capture device (macOS: auto-detected if empty)
}

// Recorder handles screen recording with audio
type Recorder struct {
	config RecorderConfig
}

// NewRecorder creates a new video recorder
func NewRecorder(config RecorderConfig) *Recorder {
	if config.FrameRate == 0 {
		config.FrameRate = 30
	}
	// Auto-detect screen device on macOS if not specified
	if runtime.GOOS == "darwin" && config.ScreenDevice == "" {
		config.ScreenDevice = detectMacOSScreenDevice()
	}
	return &Recorder{config: config}
}

// detectMacOSScreenDevice finds the first screen capture device
func detectMacOSScreenDevice() string {
	cmd := exec.Command("ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	output, _ := cmd.CombinedOutput()

	// Parse output to find "Capture screen X" device
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Look for lines like "[4] Capture screen 0"
		if strings.Contains(line, "Capture screen") {
			// Extract the device number in brackets
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start != -1 && end > start {
				deviceNum := line[start+1 : end]
				return deviceNum + ":none"
			}
		}
	}
	// Fallback to device 1 if detection fails
	return "1:none"
}

// RecordSlide records a single slide with audio playback
func (r *Recorder) RecordSlide(ctx context.Context, slideIndex int, audioPath string, duration time.Duration) (string, error) {
	logger := slogutil.LoggerFromContext(ctx, nil)

	// Ensure output directory exists
	if err := os.MkdirAll(r.config.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(r.config.OutputDir, fmt.Sprintf("slide_%03d.mp4", slideIndex))

	// Build ffmpeg command based on platform
	cmd, err := r.buildRecordCommand(outputPath, audioPath, duration)
	if err != nil {
		return "", err
	}

	// Start recording
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start recording: %w", err)
	}

	// Wait for duration or context cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(duration):
		// Stop recording gracefully by sending interrupt to ffmpeg
		if cmd.Process != nil {
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				logger.Warn("failed to send interrupt signal", "error", err)
			}
		}
		<-done // Wait for process to finish
	case <-ctx.Done():
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				logger.Warn("failed to kill process", "error", err)
			}
		}
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("recording failed: %w", err)
		}
	}

	return outputPath, nil
}

// buildRecordCommand creates the ffmpeg command for screen recording
func (r *Recorder) buildRecordCommand(outputPath, audioPath string, duration time.Duration) (*exec.Cmd, error) {
	var args []string

	switch runtime.GOOS {
	case "darwin": // macOS
		args = r.buildMacOSCommand(outputPath, audioPath, duration)
	case "linux":
		args = r.buildLinuxCommand(outputPath, audioPath, duration)
	case "windows":
		args = r.buildWindowsCommand(outputPath, audioPath, duration)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	cmd := exec.Command("ffmpeg", args...)
	return cmd, nil
}

// buildMacOSCommand builds ffmpeg command for macOS using avfoundation
func (r *Recorder) buildMacOSCommand(outputPath, audioPath string, duration time.Duration) []string {
	// Using screen capture with audio overlay
	// Output format optimized for YouTube/Udemy upload
	screenDevice := r.config.ScreenDevice
	if screenDevice == "" {
		screenDevice = "1:none" // Fallback
	}
	return []string{
		"-f", "avfoundation",
		"-capture_cursor", "1",
		"-framerate", fmt.Sprintf("%d", r.config.FrameRate),
		"-i", screenDevice, // Auto-detected screen capture device
		"-i", audioPath, // Audio input
		"-t", fmt.Sprintf("%.2f", duration.Seconds()),
		"-map", "0:v", // Map video from screen capture
		"-map", "1:a", // Map audio from audio file
		"-vcodec", "libx264",
		"-preset", "medium", // Better quality for uploads
		"-crf", "23",        // Constant rate factor (18-28 is good, lower = better)
		"-pix_fmt", "yuv420p",
		"-acodec", "aac",    // AAC audio for compatibility
		"-b:a", "192k",      // Audio bitrate
		"-y",
		outputPath,
	}
}

// buildLinuxCommand builds ffmpeg command for Linux using x11grab
func (r *Recorder) buildLinuxCommand(outputPath, audioPath string, duration time.Duration) []string {
	// Output format optimized for YouTube/Udemy upload
	return []string{
		"-f", "x11grab",
		"-framerate", fmt.Sprintf("%d", r.config.FrameRate),
		"-video_size", fmt.Sprintf("%dx%d", r.config.Width, r.config.Height),
		"-i", ":0.0",
		"-i", audioPath,
		"-t", fmt.Sprintf("%.2f", duration.Seconds()),
		"-map", "0:v",
		"-map", "1:a",
		"-vcodec", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		"-acodec", "aac",
		"-b:a", "192k",
		"-y",
		outputPath,
	}
}

// buildWindowsCommand builds ffmpeg command for Windows using gdigrab
func (r *Recorder) buildWindowsCommand(outputPath, audioPath string, duration time.Duration) []string {
	// Output format optimized for YouTube/Udemy upload
	return []string{
		"-f", "gdigrab",
		"-framerate", fmt.Sprintf("%d", r.config.FrameRate),
		"-i", "desktop",
		"-i", audioPath,
		"-t", fmt.Sprintf("%.2f", duration.Seconds()),
		"-map", "0:v",
		"-map", "1:a",
		"-vcodec", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		"-acodec", "aac",
		"-b:a", "192k",
		"-y",
		outputPath,
	}
}

// CheckFFmpeg verifies that ffmpeg is installed
func CheckFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found. Please install ffmpeg")
	}
	return nil
}
