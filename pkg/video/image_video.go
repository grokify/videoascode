package video

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grokify/ffutil"
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
	// Ensure output directory exists
	if err := os.MkdirAll(c.config.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get actual audio duration using ffprobe for precise timing
	audioDuration, err := getAudioDurationSeconds(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to get audio duration: %w", err)
	}

	outputPath := filepath.Join(c.config.OutputDir, fmt.Sprintf("slide_%03d.mp4", slideIndex))

	// Get encoder settings
	encoderConfig := GetGlobalEncoderConfig()
	codec, codecArgs := GetVideoCodec(encoderConfig)

	// Build ffmpeg command to create video from static image with audio
	// -loop 1: loop the image
	// -t: explicit duration matching audio length
	cmd := ffutil.New().
		InputImage(imagePath, c.config.FrameRate).
		Input(audioPath).
		VideoCodec(codec).
		Args(codecArgs...).
		AudioCodec("aac").
		AudioBitrate("192k").
		PixelFormat("yuv420p").
		Duration(audioDuration).
		Output(outputPath)

	// Add stillimage tune only for libx264
	if codec == "libx264" {
		cmd.Args("-tune", "stillimage")
	}

	if err := cmd.Run(ctx); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	return outputPath, nil
}

// CreateSlideVideoWithSize creates a video with specific dimensions
func (c *ImageVideoConverter) CreateSlideVideoWithSize(ctx context.Context, slideIndex int, imagePath, audioPath string, duration time.Duration, width, height int) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(c.config.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get actual audio duration using ffprobe for precise timing
	audioDuration, err := getAudioDurationSeconds(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to get audio duration: %w", err)
	}

	outputPath := filepath.Join(c.config.OutputDir, fmt.Sprintf("slide_%03d.mp4", slideIndex))

	// Scale filter to resize image to target dimensions
	scaleFilter := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
		width, height, width, height)

	// Get encoder settings
	encConfig := GetGlobalEncoderConfig()
	encCodec, encCodecArgs := GetVideoCodec(encConfig)

	// Build ffmpeg command using ffutil
	cmd := ffutil.New().
		InputImage(imagePath, c.config.FrameRate).
		Input(audioPath).
		VideoFilter(scaleFilter).
		VideoCodec(encCodec).
		Args(encCodecArgs...).
		AudioCodec("aac").
		AudioBitrate("192k").
		PixelFormat("yuv420p").
		Duration(audioDuration).
		Output(outputPath)

	// Add stillimage tune only for libx264
	if encCodec == "libx264" {
		cmd.Args("-tune", "stillimage")
	}

	if err := cmd.Run(ctx); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	return outputPath, nil
}

// getAudioDurationSeconds uses ffprobe to get the exact duration of an audio file in seconds
func getAudioDurationSeconds(audioPath string) (float64, error) {
	dur, err := ffutil.Duration(audioPath)
	if err != nil {
		return 0, err
	}
	return dur.Seconds(), nil
}
