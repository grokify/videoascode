package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grokify/videoascode/pkg/browser"
	"github.com/grokify/videoascode/pkg/media"
	"github.com/grokify/videoascode/pkg/segment"
)

// writeFileSecure validates that path contains no ".." traversal sequences,
// then writes data to the cleaned path with mode 0600.
func writeFileSecure(path string, data []byte) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid path: contains '..' traversal sequence: %s", path)
	}
	cleanPath := filepath.Clean(path)
	return os.WriteFile(cleanPath, data, 0600) //nolint:gosec // G703: Path validated above - no '..' allowed
}

// BrowserVideoProvider creates videos by recording browser sessions.
type BrowserVideoProvider struct {
	width     int
	height    int
	frameRate int
	headless  bool
	workDir   string
}

// NewBrowserVideoProvider creates a new browser video provider.
func NewBrowserVideoProvider(width, height, frameRate int, headless bool, workDir string) *BrowserVideoProvider {
	return &BrowserVideoProvider{
		width:     width,
		height:    height,
		frameRate: frameRate,
		headless:  headless,
		workDir:   workDir,
	}
}

// SupportsSegmentType returns true for browser segments.
func (p *BrowserVideoProvider) SupportsSegmentType(sourceType segment.SourceType) bool {
	return sourceType == segment.SourceTypeBrowser
}

// CreateVideo records a browser session and combines it with audio.
func (p *BrowserVideoProvider) CreateVideo(ctx context.Context, seg segment.Segment, audioPath string, outputPath string) (int, error) {
	browserSeg, ok := seg.(*segment.BrowserSegment)
	if !ok {
		return 0, fmt.Errorf("expected BrowserSegment, got %T", seg)
	}

	// Check if video was already recorded
	if existingVideo := browserSeg.GetVideoPath(); existingVideo != "" {
		// Just combine with audio
		return p.combineWithAudio(ctx, existingVideo, audioPath, outputPath)
	}

	// Create working directory for this segment
	segWorkDir := filepath.Join(p.workDir, seg.GetID())
	if err := os.MkdirAll(segWorkDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Create recorder
	recorderCfg := browser.RecorderConfig{
		Width:            p.width,
		Height:           p.height,
		OutputDir:        segWorkDir,
		FrameRate:        p.frameRate,
		Headless:         p.headless,
		DefaultTimeout:   30000,
		CaptureEveryStep: true,
	}

	recorder, err := browser.NewRecorder(recorderCfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create recorder: %w", err)
	}
	defer recorder.Close()

	// Launch browser
	if err := recorder.Launch(); err != nil {
		return 0, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Navigate to URL
	url := browserSeg.GetURL()
	if url == "" {
		return 0, fmt.Errorf("browser segment %s has no URL", seg.GetID())
	}

	if err := recorder.Navigate(url); err != nil {
		return 0, fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Execute steps
	steps := browserSeg.GetSteps()
	if _, err := recorder.RecordSteps(ctx, steps); err != nil {
		return 0, fmt.Errorf("failed to record steps: %w", err)
	}

	// Generate video from screenshots
	videoPath, err := recorder.GenerateVideo(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to generate video: %w", err)
	}

	// Combine with audio
	return p.combineWithAudio(ctx, videoPath, audioPath, outputPath)
}

// combineWithAudio merges video and audio tracks.
func (p *BrowserVideoProvider) combineWithAudio(ctx context.Context, videoPath, audioPath, outputPath string) (int, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return 0, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get durations
	videoDuration, err := media.GetVideoDuration(videoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	audioDuration, err := media.GetAudioDuration(audioPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	// Use the longer duration
	finalDuration := max(videoDuration, audioDuration)

	// Get encoder settings
	encoderConfig := GetGlobalEncoderConfig()
	codec, codecArgs := GetVideoCodec(encoderConfig)

	// Build ffmpeg command
	args := []string{
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", codec,
	}
	args = append(args, codecArgs...)
	args = append(args,
		"-c:a", "aac",
		"-b:a", "192k",
	)

	// If audio is longer than video, extend video with last frame
	if audioDuration > videoDuration {
		args = append(args, "-vf", fmt.Sprintf("tpad=stop_mode=clone:stop_duration=%.3f",
			(audioDuration-videoDuration).Seconds()))
	}

	// If video is longer than audio, pad audio with silence
	if videoDuration > audioDuration {
		args = append(args, "-af", fmt.Sprintf("apad=whole_dur=%.3f", videoDuration.Seconds()))
	}

	args = append(args,
		"-shortest",
		"-movflags", "+faststart",
		outputPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return int(finalDuration.Milliseconds()), nil
}

// CreateVideoWithTiming records a browser session using timing from TTS durations.
// This ensures steps are paced to match voiceover length.
func (p *BrowserVideoProvider) CreateVideoWithTiming(ctx context.Context, seg segment.Segment, stepDurations map[int]int, outputPath string) (int, error) {
	browserSeg, ok := seg.(*segment.BrowserSegment)
	if !ok {
		return 0, fmt.Errorf("expected BrowserSegment, got %T", seg)
	}

	// Update step minDurations based on TTS timing
	browserSeg.UpdateStepMinDurations(stepDurations)

	// Create working directory
	segWorkDir := filepath.Join(p.workDir, seg.GetID())
	if err := os.MkdirAll(segWorkDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Create recorder
	recorderCfg := browser.RecorderConfig{
		Width:            p.width,
		Height:           p.height,
		OutputDir:        segWorkDir,
		FrameRate:        p.frameRate,
		Headless:         p.headless,
		DefaultTimeout:   30000,
		CaptureEveryStep: true,
	}

	recorder, err := browser.NewRecorder(recorderCfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create recorder: %w", err)
	}
	defer recorder.Close()

	// Launch and navigate
	if err := recorder.Launch(); err != nil {
		return 0, fmt.Errorf("failed to launch browser: %w", err)
	}

	url := browserSeg.GetURL()
	if err := recorder.Navigate(url); err != nil {
		return 0, fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Execute steps (with updated minDurations)
	steps := browserSeg.GetSteps()
	if _, err := recorder.RecordSteps(ctx, steps); err != nil {
		return 0, fmt.Errorf("failed to record steps: %w", err)
	}

	// Generate video
	videoPath, err := recorder.GenerateVideo(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to generate video: %w", err)
	}

	// Copy to output path
	data, err := os.ReadFile(videoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read video: %w", err)
	}
	if err := writeFileSecure(outputPath, data); err != nil {
		return 0, fmt.Errorf("failed to write video: %w", err)
	}

	// Get final duration
	duration, err := media.GetVideoDurationMs(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	// Store video path on segment for reuse
	browserSeg.SetVideoPath(outputPath)

	return duration, nil
}
