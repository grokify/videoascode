package video

import (
	"context"
	"fmt"
	"os/exec"
)

// ReplaceAudio replaces the audio track in a video file.
func ReplaceAudio(ctx context.Context, videoPath, audioPath, outputPath string) error {
	args := []string{
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-map", "0:v:0",
		"-map", "1:a:0",
		"-shortest",
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

// AddAudioToVideo combines a video (possibly silent) with an audio track.
func AddAudioToVideo(ctx context.Context, videoPath, audioPath, outputPath string) error {
	args := []string{
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-shortest",
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
