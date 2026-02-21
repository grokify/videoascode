package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	// Build ffmpeg command with filter_complex concat
	// This properly handles mixed audio formats by re-encoding
	args := []string{}

	// Add all input files
	for _, path := range videoPaths {
		args = append(args, "-i", path)
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

	args = append(args,
		"-filter_complex", filterParts.String(),
		"-map", "[outv]",
		"-map", "[outa]",
		"-c:v", codec,
	)
	args = append(args, codecArgs...)
	args = append(args,
		"-c:a", "aac",
		"-b:a", "192k",
		"-ar", "44100",
		"-y",
		outputPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w\nOutput: %s", err, string(output))
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

	// Build ffmpeg command with xfade filter
	args := []string{}

	// Add all input files
	for _, path := range videoPaths {
		args = append(args, "-i", path)
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

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", finalVideoLabel,
		"-map", finalAudioLabel,
		"-c:v", encCodec,
	)
	args = append(args, encCodecArgs...)
	args = append(args,
		"-c:a", "aac",
		"-b:a", "192k",
		"-y",
		outputPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If xfade fails (e.g., older ffmpeg), fall back to simple concatenation
		logger.Warn("xfade transition failed, falling back to simple concatenation", "error", err)
		return c.CombineVideos(ctx, videoPaths, outputPath)
	}
	_ = output

	return nil
}

// GetVideoDuration gets the duration of a video file using ffprobe
func GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	var duration float64
	_, err = fmt.Sscanf(string(output), "%f", &duration)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
