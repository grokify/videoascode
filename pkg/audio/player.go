package audio

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/grokify/ffutil"
)

// Player handles audio playback
type Player struct{}

// NewPlayer creates a new audio player
func NewPlayer() *Player {
	return &Player{}
}

// GetDuration gets the duration of an audio file using ffprobe
func (p *Player) GetDuration(audioPath string) (time.Duration, error) {
	return ffutil.Duration(audioPath)
}

// Play plays an audio file (blocking)
func (p *Player) Play(audioPath string) error {
	// Use ffplay (part of ffmpeg) for audio playback
	cmd := exec.Command("ffplay",
		"-nodisp",
		"-autoexit",
		audioPath,
	)

	return cmd.Run()
}

// PadToLength pads an audio file with silence to reach the target duration.
// If the audio is already longer than or equal to the target duration, it is copied unchanged.
// Returns the path to the padded audio file in the output directory.
func (p *Player) PadToLength(audioPath string, targetDuration time.Duration, outputDir string) (string, error) {
	// Get current duration
	currentDuration, err := p.GetDuration(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to get audio duration: %w", err)
	}

	// Generate output path
	baseName := filepath.Base(audioPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]
	outputPath := filepath.Join(outputDir, nameWithoutExt+"_padded.m4a")

	ctx := context.Background()

	// If audio is already long enough, just convert to m4a (for consistency)
	if currentDuration >= targetDuration {
		err := ffutil.New().
			Input(audioPath).
			AudioCodec("aac").
			AudioBitrate("192k").
			Output(outputPath).
			Run(ctx)
		if err != nil {
			return "", fmt.Errorf("ffmpeg conversion failed: %w", err)
		}
		return outputPath, nil
	}

	// Pad with silence using the apad filter
	// The apad filter extends the audio with silence to reach whole_dur seconds
	targetSeconds := targetDuration.Seconds()
	apadFilter := fmt.Sprintf("apad=whole_dur=%.3f", targetSeconds)

	err = ffutil.New().
		Input(audioPath).
		AudioFilter(apadFilter).
		AudioCodec("aac").
		AudioBitrate("192k").
		Output(outputPath).
		Run(ctx)
	if err != nil {
		return "", fmt.Errorf("ffmpeg padding failed: %w", err)
	}

	return outputPath, nil
}
