// Package media provides utilities for working with audio and video files.
package media

import (
	"time"

	"github.com/grokify/ffutil"
)

// GetAudioDuration returns the duration of an audio file using ffprobe.
// This consolidates the duplicate implementations that were in:
// - pkg/tts/audio.go
// - pkg/audio/player.go
// - pkg/video/image_video.go
func GetAudioDuration(path string) (time.Duration, error) {
	return ffutil.Duration(path)
}

// GetVideoDuration returns the duration of a video file using ffprobe.
func GetVideoDuration(path string) (time.Duration, error) {
	return ffutil.Duration(path)
}

// GetDuration returns the duration of a media file (audio or video).
func GetDuration(path string) (time.Duration, error) {
	return ffutil.Duration(path)
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
	if !ffutil.FFprobeAvailable() {
		return ffutil.Available() // Returns error with details
	}
	return nil
}

// CheckFFmpeg verifies that ffmpeg is installed and accessible.
func CheckFFmpeg() error {
	if !ffutil.FFmpegAvailable() {
		return ffutil.Available() // Returns error with details
	}
	return nil
}
