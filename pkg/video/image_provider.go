package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/grokify/marp2video/pkg/media"
	"github.com/grokify/marp2video/pkg/segment"
)

// ImageVideoProvider creates videos from static images (for slide segments).
type ImageVideoProvider struct {
	width     int
	height    int
	frameRate int
}

// NewImageVideoProvider creates a new image video provider.
func NewImageVideoProvider(width, height, frameRate int) *ImageVideoProvider {
	return &ImageVideoProvider{
		width:     width,
		height:    height,
		frameRate: frameRate,
	}
}

// SupportsSegmentType returns true for slide segments.
func (p *ImageVideoProvider) SupportsSegmentType(sourceType segment.SourceType) bool {
	return sourceType == segment.SourceTypeSlide
}

// CreateVideo generates a video from a slide's image and audio.
func (p *ImageVideoProvider) CreateVideo(ctx context.Context, seg segment.Segment, audioPath string, outputPath string) (int, error) {
	// Get the slide segment to access the image path
	slideSeg, ok := seg.(*segment.SlideSegment)
	if !ok {
		return 0, fmt.Errorf("expected SlideSegment, got %T", seg)
	}

	imagePath := slideSeg.GetImagePath()
	if imagePath == "" {
		return 0, fmt.Errorf("slide segment %s has no image path", seg.GetID())
	}

	// Get audio duration
	audioDuration, err := media.GetAudioDuration(audioPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return 0, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build ffmpeg command
	// Scale image to target resolution while maintaining aspect ratio, then pad
	scaleFilter := fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black",
		p.width, p.height, p.width, p.height,
	)

	durationSec := audioDuration.Seconds()

	// Get encoder settings
	encoderConfig := GetGlobalEncoderConfig()
	codec, codecArgs := GetVideoCodec(encoderConfig)

	args := []string{
		"-y",
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", codec,
	}
	args = append(args, codecArgs...)
	// Add stillimage tune only for libx264
	if codec == "libx264" {
		args = append(args, "-tune", "stillimage")
	}
	args = append(args,
		"-c:a", "aac",
		"-b:a", "192k",
		"-pix_fmt", "yuv420p",
		"-vf", scaleFilter,
		"-t", fmt.Sprintf("%.3f", durationSec),
		"-movflags", "+faststart",
		outputPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return int(audioDuration.Milliseconds()), nil
}

// CreateVideoWithDuration generates a video with a specific duration (padding audio if needed).
func (p *ImageVideoProvider) CreateVideoWithDuration(ctx context.Context, seg segment.Segment, audioPath string, outputPath string, durationMs int) (int, error) {
	slideSeg, ok := seg.(*segment.SlideSegment)
	if !ok {
		return 0, fmt.Errorf("expected SlideSegment, got %T", seg)
	}

	imagePath := slideSeg.GetImagePath()
	if imagePath == "" {
		return 0, fmt.Errorf("slide segment %s has no image path", seg.GetID())
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return 0, fmt.Errorf("failed to create output directory: %w", err)
	}

	scaleFilter := fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black",
		p.width, p.height, p.width, p.height,
	)

	durationSec := float64(durationMs) / 1000.0

	// Get encoder settings
	encConfig := GetGlobalEncoderConfig()
	encCodec, encCodecArgs := GetVideoCodec(encConfig)

	// Use apad filter to extend audio to match video duration
	args := []string{
		"-y",
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", encCodec,
	}
	args = append(args, encCodecArgs...)
	// Add stillimage tune only for libx264
	if encCodec == "libx264" {
		args = append(args, "-tune", "stillimage")
	}
	args = append(args,
		"-c:a", "aac",
		"-b:a", "192k",
		"-pix_fmt", "yuv420p",
		"-vf", scaleFilter,
		"-af", fmt.Sprintf("apad=whole_dur=%.3f", durationSec),
		"-t", fmt.Sprintf("%.3f", durationSec),
		"-movflags", "+faststart",
		outputPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return durationMs, nil
}
