package tts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	omnitts "github.com/grokify/marp2video/pkg/omnivoice/tts"
	"github.com/grokify/marp2video/pkg/transcript"
)

// Config holds TTS generator configuration.
type Config struct {
	omnitts.ProviderConfig        // Embedded provider config (ElevenLabsAPIKey, DeepgramAPIKey)
	DefaultProvider        string // "elevenlabs" or "deepgram"
	VoiceID                string
	Model                  string
	OutputDir              string
}

// Generator handles text-to-speech generation using OmniVoice providers.
type Generator struct {
	config   Config
	provider *omnitts.Provider
}

// NewGenerator creates a new TTS generator.
func NewGenerator(config Config) (*Generator, error) {
	factory := omnitts.NewFactory(config.ProviderConfig)

	if config.DefaultProvider != "" {
		factory.SetFallback(config.DefaultProvider)
	}

	provider, err := factory.Get("")
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS provider: %w", err)
	}

	return &Generator{
		config:   config,
		provider: provider,
	}, nil
}

// AudioResult contains the generated audio file path and duration.
type AudioResult struct {
	FilePath string
	Duration time.Duration
}

// GenerateAudio converts text to speech using OmniVoice.
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

	// Create voice config
	voiceConfig := transcript.VoiceConfig{
		VoiceID: g.config.VoiceID,
		Model:   g.config.Model,
	}

	// Set defaults if not specified
	if voiceConfig.VoiceID == "" {
		voiceConfig.VoiceID = "pNInz6obpgDQGcFmaJgB" // Adam voice
	}
	if voiceConfig.Model == "" {
		voiceConfig.Model = "eleven_multilingual_v2"
	}

	// Generate speech
	audioData, err := g.provider.Synthesize(ctx, text, voiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate speech: %w", err)
	}

	// Write audio to file
	if err := os.WriteFile(outputPath, audioData, 0600); err != nil {
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
