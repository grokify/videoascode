package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ImageVideoConfig holds configuration for image-to-video conversion
type ImageVideoConfig struct {
	OutputDir string
	Width     int
	Height    int
	FrameRate int
}

// ImageVideoConverter creates videos from static images with audio
type ImageVideoConverter struct {
	config ImageVideoConfig
}

// NewImageVideoConverter creates a new image video converter
func NewImageVideoConverter(config ImageVideoConfig) *ImageVideoConverter {
	if config.FrameRate == 0 {
		config.FrameRate = 30
	}
	return &ImageVideoConverter{config: config}
}

// CreateSlideVideo creates a video from a static image and audio file
func (c *ImageVideoConverter) CreateSlideVideo(ctx context.Context, slideIndex int, imagePath, audioPath string, duration time.Duration) (string, error) {
	_ = ctx // reserved for future use

	// Ensure output directory exists
	if err := os.MkdirAll(c.config.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(c.config.OutputDir, fmt.Sprintf("slide_%03d.mp4", slideIndex))

	// Build ffmpeg command to create video from static image with audio
	// -loop 1: loop the image
	// -t: duration (use audio duration or specified duration)
	// -tune stillimage: optimize encoding for still images
	// -shortest: stop when the shortest input ends (the audio)
	args := []string{
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-c:a", "aac",
		"-b:a", "192k",
		"-pix_fmt", "yuv420p",
		"-t", fmt.Sprintf("%.2f", duration.Seconds()),
		"-y",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)

	// Show debug output if enabled
	if os.Getenv("MARP2VIDEO_DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}

// CreateSlideVideoWithSize creates a video with specific dimensions
func (c *ImageVideoConverter) CreateSlideVideoWithSize(ctx context.Context, slideIndex int, imagePath, audioPath string, duration time.Duration, width, height int) (string, error) {
	_ = ctx // reserved for future use

	// Ensure output directory exists
	if err := os.MkdirAll(c.config.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(c.config.OutputDir, fmt.Sprintf("slide_%03d.mp4", slideIndex))

	// Scale filter to resize image to target dimensions
	scaleFilter := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
		width, height, width, height)

	args := []string{
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-vf", scaleFilter,
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-c:a", "aac",
		"-b:a", "192k",
		"-pix_fmt", "yuv420p",
		"-t", fmt.Sprintf("%.2f", duration.Seconds()),
		"-y",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)

	if os.Getenv("MARP2VIDEO_DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}
