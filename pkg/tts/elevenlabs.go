package tts

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/agentplexus/go-elevenlabs"
)

// Config holds ElevenLabs TTS configuration
type Config struct {
	APIKey    string
	VoiceID   string // Default: Adam voice
	OutputDir string
	Model     string // Default: eleven_multilingual_v2
}

// Generator handles text-to-speech generation
type Generator struct {
	config Config
}

// NewGenerator creates a new TTS generator
func NewGenerator(config Config) *Generator {
	if config.Model == "" {
		config.Model = elevenlabs.DefaultModelID
	}
	if config.VoiceID == "" {
		// Adam voice ID
		config.VoiceID = "pNInz6obpgDQGcFmaJgB"
	}
	return &Generator{config: config}
}

// AudioResult contains the generated audio file path and duration
type AudioResult struct {
	FilePath string
	Duration time.Duration
}

// GenerateAudio converts text to speech using ElevenLabs
func (g *Generator) GenerateAudio(ctx context.Context, text string, slideIndex int) (*AudioResult, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate output filename
	outputPath := filepath.Join(g.config.OutputDir, fmt.Sprintf("slide_%03d.mp3", slideIndex))

	// Create ElevenLabs client using functional options
	client, err := elevenlabs.NewClient(elevenlabs.WithAPIKey(g.config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create ElevenLabs client: %w", err)
	}

	// Create TTS request
	req := &elevenlabs.TTSRequest{
		VoiceID:       g.config.VoiceID,
		Text:          text,
		ModelID:       g.config.Model,
		VoiceSettings: elevenlabs.DefaultVoiceSettings(),
	}

	// Generate speech
	resp, err := client.TextToSpeech().Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate speech: %w", err)
	}

	// Read audio data from response
	audioData, err := io.ReadAll(resp.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	// Write audio to file
	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write audio file: %w", err)
	}

	// Get audio duration using ffprobe
	duration, err := getAudioDuration(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	return &AudioResult{
		FilePath: outputPath,
		Duration: duration,
	}, nil
}

// getAudioDuration uses ffprobe to get the duration of an audio file
func getAudioDuration(filePath string) (time.Duration, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
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
