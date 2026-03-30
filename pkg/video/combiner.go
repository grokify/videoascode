package video

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/grokify/ffutil"
	"github.com/grokify/mogo/log/slogutil"
)

// Combiner handles concatenation of video segments
type Combiner struct {
	outputDir string
}

// NewCombiner creates a new video combiner
func NewCombiner(outputDir string) *Combiner {
	return &Combiner{outputDir: outputDir}
}

// CombineVideos concatenates multiple video files into one
// Uses filter_complex concat to properly handle mixed audio formats (different sample rates)
func (c *Combiner) CombineVideos(ctx context.Context, videoPaths []string, outputPath string) error {
	_ = slogutil.LoggerFromContext(ctx, slogutil.Null())

	if len(videoPaths) == 0 {
		return fmt.Errorf("no video files to combine")
	}

	if len(videoPaths) == 1 {
		return copyFile(videoPaths[0], outputPath)
	}

	// Build the concat filter
	// [0:v][0:a][1:v][1:a]...[n:v][n:a]concat=n=N:v=1:a=1[outv][outa]
	var filterParts strings.Builder
	for i := range videoPaths {
		filterParts.WriteString(fmt.Sprintf("[%d:v][%d:a]", i, i))
	}
	filterParts.WriteString(fmt.Sprintf("concat=n=%d:v=1:a=1[outv][outa]", len(videoPaths)))

	// Get encoder settings
	encoderConfig := GetGlobalEncoderConfig()
	codec, codecArgs := GetVideoCodec(encoderConfig)

	// Build command using ffutil
	cmd := ffutil.New()
	for _, path := range videoPaths {
		cmd.Input(path)
	}
	cmd.FilterComplex(filterParts.String()).
		Args("-map", "[outv]").
		Args("-map", "[outa]").
		VideoCodec(codec).
		Args(codecArgs...).
		AudioCodec("aac").
		AudioBitrate("192k").
		AudioRate(44100).
		Output(outputPath)

	if err := cmd.Run(ctx); err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	return nil
}

// CombineVideosWithTransitions concatenates videos with crossfade transitions
func (c *Combiner) CombineVideosWithTransitions(ctx context.Context, videoPaths []string, outputPath string, transitionDuration float64) error {
	logger := slogutil.LoggerFromContext(ctx, slogutil.Null())

	if len(videoPaths) == 0 {
		return fmt.Errorf("no video files to combine")
	}

	if len(videoPaths) == 1 {
		// No transitions needed, just copy
		return copyFile(videoPaths[0], outputPath)
	}

	// Get durations of all videos
	durations := make([]float64, len(videoPaths))
	for i, path := range videoPaths {
		dur, err := GetVideoDuration(path)
		if err != nil {
			return fmt.Errorf("failed to get duration of video %d: %w", i, err)
		}
		durations[i] = dur
	}

	// Build the complex filter graph for video
	// Each xfade takes two inputs and produces one output
	// We chain them: [0][1]xfade[v01]; [v01][2]xfade[v012]; ...
	var videoFilter strings.Builder
	var audioFilter strings.Builder

	// Calculate cumulative offsets (accounting for transition overlap)
	offset := 0.0

	for i := 0; i < len(videoPaths)-1; i++ {
		inputA := fmt.Sprintf("[v%d]", i)
		inputB := fmt.Sprintf("[%d:v]", i+1)
		outputV := fmt.Sprintf("[v%d]", i+1)

		audioInputA := fmt.Sprintf("[a%d]", i)
		audioInputB := fmt.Sprintf("[%d:a]", i+1)
		audioOutputA := fmt.Sprintf("[a%d]", i+1)

		if i == 0 {
			inputA = "[0:v]"
			audioInputA = "[0:a]"
		}

		// Video xfade
		// offset is when the transition starts (end of current video minus transition duration)
		transitionOffset := offset + durations[i] - transitionDuration
		if transitionOffset < 0 {
			transitionOffset = 0
		}

		videoFilter.WriteString(fmt.Sprintf("%s%sxfade=transition=fade:duration=%.3f:offset=%.3f%s",
			inputA, inputB, transitionDuration, transitionOffset, outputV))

		// Audio crossfade using acrossfade
		audioFilter.WriteString(fmt.Sprintf("%s%sacrossfade=d=%.3f:c1=tri:c2=tri%s",
			audioInputA, audioInputB, transitionDuration, audioOutputA))

		if i < len(videoPaths)-2 {
			videoFilter.WriteString(";")
			audioFilter.WriteString(";")
		}

		// Update offset for next iteration (subtract overlap)
		offset = transitionOffset
	}

	// Final output labels
	finalVideoLabel := fmt.Sprintf("[v%d]", len(videoPaths)-1)
	finalAudioLabel := fmt.Sprintf("[a%d]", len(videoPaths)-1)

	// Complete filter
	filterComplex := videoFilter.String() + ";" + audioFilter.String()

	// Get encoder settings
	encConfig := GetGlobalEncoderConfig()
	encCodec, encCodecArgs := GetVideoCodec(encConfig)

	// Build command using ffutil
	cmd := ffutil.New()
	for _, path := range videoPaths {
		cmd.Input(path)
	}
	cmd.FilterComplex(filterComplex).
		Args("-map", finalVideoLabel).
		Args("-map", finalAudioLabel).
		VideoCodec(encCodec).
		Args(encCodecArgs...).
		AudioCodec("aac").
		AudioBitrate("192k").
		Output(outputPath)

	if err := cmd.Run(ctx); err != nil {
		// If xfade fails (e.g., older ffmpeg), fall back to simple concatenation
		logger.Warn("xfade transition failed, falling back to simple concatenation", "error", err)
		return c.CombineVideos(ctx, videoPaths, outputPath)
	}

	return nil
}

// GetVideoDuration gets the duration of a video file using ffprobe
func GetVideoDuration(videoPath string) (float64, error) {
	dur, err := ffutil.Duration(videoPath)
	if err != nil {
		return 0, err
	}
	return dur.Seconds(), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeFileSecure(dst, data)
}
