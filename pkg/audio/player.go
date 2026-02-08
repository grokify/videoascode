package audio

import (
	"fmt"
	"os/exec"
	"time"
)

// Player handles audio playback
type Player struct{}

// NewPlayer creates a new audio player
func NewPlayer() *Player {
	return &Player{}
}

// GetDuration gets the duration of an audio file using ffprobe
func (p *Player) GetDuration(audioPath string) (time.Duration, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	var seconds float64
	_, err = fmt.Sscanf(string(output), "%f", &seconds)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return time.Duration(seconds * float64(time.Second)), nil
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
