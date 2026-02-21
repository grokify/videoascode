// Package media provides utilities for working with audio and video files.
package media

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetAudioDuration returns the duration of an audio file using ffprobe.
// This consolidates the duplicate implementations that were in:
// - pkg/tts/audio.go
// - pkg/audio/player.go
// - pkg/video/image_video.go
func GetAudioDuration(path string) (time.Duration, error) {
	return getDuration(path, "a")
}

// GetVideoDuration returns the duration of a video file using ffprobe.
func GetVideoDuration(path string) (time.Duration, error) {
	return getDuration(path, "v")
}

// GetDuration returns the duration of a media file (audio or video).
func GetDuration(path string) (time.Duration, error) {
	return getDuration(path, "")
}

// getDuration uses ffprobe to get the duration of a media file.
// streamType can be "a" for audio, "v" for video, or "" for any.
func getDuration(path, streamType string) (time.Duration, error) {
	args := []string{
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
	}

	if streamType != "" {
		args = append(args, "-select_streams", streamType)
	}

	args = append(args, path)

	cmd := exec.Command("ffprobe", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" || output == "N/A" {
		return 0, fmt.Errorf("ffprobe returned no duration for %s", path)
	}

	seconds, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration '%s': %w", output, err)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// GetDurationMs returns the duration in milliseconds.
func GetDurationMs(path string) (int, error) {
	d, err := GetDuration(path)
	if err != nil {
		return 0, err
	}
	return int(d.Milliseconds()), nil
}

// GetAudioDurationMs returns the audio duration in milliseconds.
func GetAudioDurationMs(path string) (int, error) {
	d, err := GetAudioDuration(path)
	if err != nil {
		return 0, err
	}
	return int(d.Milliseconds()), nil
}

// GetVideoDurationMs returns the video duration in milliseconds.
func GetVideoDurationMs(path string) (int, error) {
	d, err := GetVideoDuration(path)
	if err != nil {
		return 0, err
	}
	return int(d.Milliseconds()), nil
}

// CheckFFprobe verifies that ffprobe is installed and accessible.
func CheckFFprobe() error {
	cmd := exec.Command("ffprobe", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffprobe not found: %w", err)
	}
	return nil
}

// CheckFFmpeg verifies that ffmpeg is installed and accessible.
func CheckFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}
	return nil
}
