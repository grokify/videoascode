package browser

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CaptureMode defines how video is captured
type CaptureMode string

const (
	// CaptureModeScreenshot captures screenshots and stitches them into video
	CaptureModeScreenshot CaptureMode = "screenshot"
	// CaptureModeVideo captures continuous video (requires headed browser)
	CaptureModeVideo CaptureMode = "video"
)

// CaptureConfig configures video capture settings
type CaptureConfig struct {
	// Mode determines capture strategy
	Mode CaptureMode

	// OutputDir is where captured frames/video are stored
	OutputDir string

	// Width is the capture width in pixels
	Width int

	// Height is the capture height in pixels
	Height int

	// FrameRate is frames per second
	FrameRate int

	// Quality is the output quality (0-100, higher is better)
	Quality int

	// Format is the output video format (mp4, webm)
	Format string
}

// DefaultCaptureConfig returns a default capture configuration
func DefaultCaptureConfig() CaptureConfig {
	return CaptureConfig{
		Mode:      CaptureModeScreenshot,
		Width:     1920,
		Height:    1080,
		FrameRate: 30,
		Quality:   85,
		Format:    "mp4",
	}
}

// Capturer handles screenshot capture and video generation
type Capturer struct {
	config     CaptureConfig
	frameCount int
	framesDir  string
}

// NewCapturer creates a new capturer
func NewCapturer(config CaptureConfig) (*Capturer, error) {
	if config.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	framesDir := filepath.Join(config.OutputDir, "frames")
	if err := os.MkdirAll(framesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create frames directory: %w", err)
	}

	return &Capturer{
		config:    config,
		framesDir: framesDir,
	}, nil
}

// SaveFrame saves a screenshot frame
func (c *Capturer) SaveFrame(data []byte) (string, error) {
	framePath := filepath.Join(c.framesDir, fmt.Sprintf("frame_%06d.png", c.frameCount))
	if err := os.WriteFile(framePath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write frame %d: %w", c.frameCount, err)
	}
	c.frameCount++
	return framePath, nil
}

// GetFrameCount returns the current frame count
func (c *Capturer) GetFrameCount() int {
	return c.frameCount
}

// GetFramePath returns the path for a specific frame
func (c *Capturer) GetFramePath(index int) string {
	return filepath.Join(c.framesDir, fmt.Sprintf("frame_%06d.png", index))
}

// GenerateVideo stitches captured frames into a video file
func (c *Capturer) GenerateVideo(ctx context.Context, outputPath string) error {
	if c.frameCount == 0 {
		return fmt.Errorf("no frames captured")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	inputPattern := filepath.Join(c.framesDir, "frame_%06d.png")

	args := []string{
		"-y", // Overwrite output
		"-framerate", fmt.Sprintf("%d", c.config.FrameRate),
		"-i", inputPattern,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", fmt.Sprintf("%d", qualityToCRF(c.config.Quality)),
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
	}

	// Add resolution scaling if needed
	args = append(args,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
			c.config.Width, c.config.Height, c.config.Width, c.config.Height),
	)

	args = append(args, outputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GenerateVideoWithDurations generates video with per-frame durations
// durations is a map of frame index to duration in milliseconds
// This is useful when steps have minimum durations for voiceover sync
func (c *Capturer) GenerateVideoWithDurations(ctx context.Context, outputPath string, durations map[int]int) error {
	if c.frameCount == 0 {
		return fmt.Errorf("no frames captured")
	}

	// Create a concat demuxer file for variable frame durations
	concatPath := filepath.Join(c.config.OutputDir, "concat.txt")
	if err := c.writeConcatFile(concatPath, durations); err != nil {
		return fmt.Errorf("failed to write concat file: %w", err)
	}

	// Build FFmpeg command using concat demuxer
	args := []string{
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", concatPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", fmt.Sprintf("%d", qualityToCRF(c.config.Quality)),
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
			c.config.Width, c.config.Height, c.config.Width, c.config.Height),
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// writeConcatFile writes an FFmpeg concat demuxer file
func (c *Capturer) writeConcatFile(path string, durations map[int]int) error {
	var sb strings.Builder

	defaultDuration := 1000 / c.config.FrameRate // Default frame duration in ms

	for i := 0; i < c.frameCount; i++ {
		framePath := c.GetFramePath(i)

		// Get duration for this frame (or use default)
		durationMs := defaultDuration
		if d, ok := durations[i]; ok {
			durationMs = d
		}

		// Convert to seconds for FFmpeg
		durationSec := float64(durationMs) / 1000.0

		fmt.Fprintf(&sb, "file '%s'\n", framePath)
		fmt.Fprintf(&sb, "duration %.3f\n", durationSec)
	}

	// FFmpeg concat demuxer requires the last file to be listed twice
	// to avoid cutting off the last frame
	if c.frameCount > 0 {
		lastFrame := c.GetFramePath(c.frameCount - 1)
		fmt.Fprintf(&sb, "file '%s'\n", lastFrame)
	}

	return os.WriteFile(path, []byte(sb.String()), 0600)
}

// Cleanup removes temporary frame files
func (c *Capturer) Cleanup() error {
	return os.RemoveAll(c.framesDir)
}

// qualityToCRF converts quality percentage to FFmpeg CRF value
// CRF ranges from 0 (lossless) to 51 (worst)
// Quality 100 -> CRF 18 (high quality)
// Quality 50 -> CRF 28 (medium quality)
// Quality 0 -> CRF 38 (low quality)
func qualityToCRF(quality int) int {
	if quality < 0 {
		quality = 0
	}
	if quality > 100 {
		quality = 100
	}
	// Map 0-100 to 38-18 (inverted, lower CRF = better quality)
	return 38 - (quality * 20 / 100)
}

// AddAudioToVideo combines video with audio track
func AddAudioToVideo(ctx context.Context, videoPath, audioPath, outputPath string) error {
	args := []string{
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-shortest", // End when shortest input ends
		"-movflags", "+faststart",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// PadVideoToDuration extends video to a specific duration using the last frame
func PadVideoToDuration(ctx context.Context, videoPath string, targetDurationMs int, outputPath string) error {
	// Use tpad filter to extend video
	durationSec := float64(targetDurationMs) / 1000.0

	args := []string{
		"-y",
		"-i", videoPath,
		"-vf", fmt.Sprintf("tpad=stop_mode=clone:stop_duration=%.3f", durationSec),
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
