package audio

import (
	"fmt"
	"os/exec"
	"path/filepath"
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

	// If audio is already long enough, just convert to m4a (for consistency)
	if currentDuration >= targetDuration {
		cmd := exec.Command("ffmpeg",
			"-i", audioPath,
			"-c:a", "aac",
			"-b:a", "192k",
			"-y",
			outputPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
		}
		return outputPath, nil
	}

	// Pad with silence using the apad filter
	// The apad filter extends the audio with silence to reach whole_dur seconds
	targetSeconds := targetDuration.Seconds()
	apadFilter := fmt.Sprintf("apad=whole_dur=%.3f", targetSeconds)

	cmd := exec.Command("ffmpeg",
		"-i", audioPath,
		"-af", apadFilter,
		"-c:a", "aac",
		"-b:a", "192k",
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg padding failed: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}
