package tts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/agentplexus/go-elevenlabs"
	"github.com/grokify/marp2video/pkg/transcript"
	"github.com/grokify/mogo/log/slogutil"
)

// TranscriptGeneratorConfig holds configuration for transcript-based TTS generation
type TranscriptGeneratorConfig struct {
	APIKey    string
	OutputDir string
}

// TranscriptGenerator generates audio from transcript files
type TranscriptGenerator struct {
	config TranscriptGeneratorConfig
	client *elevenlabs.Client
}

// NewTranscriptGenerator creates a new transcript-based TTS generator
func NewTranscriptGenerator(config TranscriptGeneratorConfig) (*TranscriptGenerator, error) {
	client, err := elevenlabs.NewClient(elevenlabs.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create ElevenLabs client: %w", err)
	}

	return &TranscriptGenerator{
		config: config,
		client: client,
	}, nil
}

// GenerateFromTranscript generates audio files for all slides in a transcript
// Returns a manifest with timing information for use by the video recorder
func (g *TranscriptGenerator) GenerateFromTranscript(ctx context.Context, t *transcript.Transcript, language string) (*Manifest, error) {
	logger := slogutil.LoggerFromContext(ctx, nil)

	// Create output directory
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create manifest
	manifest := NewManifest(language)

	// Process each slide
	for _, slide := range t.Slides {
		// Get transcript for this language
		content, err := t.GetSlideTranscript(slide.Index, language)
		if err != nil {
			logger.Warn("skipping slide without transcript",
				"index", slide.Index,
				"language", language,
				"error", err)
			continue
		}

		// Get full text for TTS
		text := content.GetFullText()
		if text == "" {
			logger.Warn("skipping slide with empty text", "index", slide.Index)
			continue
		}

		// Determine voice configuration
		voiceConfig := g.resolveVoiceConfig(t.Metadata.DefaultVoice, content.Voice)

		// Generate audio
		audioPath := filepath.Join(g.config.OutputDir, fmt.Sprintf("slide_%03d.mp3", slide.Index))

		logger.Info("generating audio",
			"slide", slide.Index,
			"voice", voiceConfig.VoiceID,
			"textLength", len(text))

		audioDuration, err := g.generateSlideAudio(ctx, text, audioPath, voiceConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate audio for slide %d: %w", slide.Index, err)
		}

		// Calculate pause duration
		pauseDuration := content.GetTotalPauseDuration()

		// Add to manifest
		manifest.AddSlide(SlideAudio{
			Index:         slide.Index,
			Title:         slide.Title,
			AudioFile:     filepath.Base(audioPath),
			AudioDuration: int(audioDuration.Milliseconds()),
			PauseDuration: pauseDuration,
			TotalDuration: int(audioDuration.Milliseconds()) + pauseDuration,
		})

		logger.Info("generated audio",
			"slide", slide.Index,
			"audioDurationMs", audioDuration.Milliseconds(),
			"pauseDurationMs", pauseDuration)
	}

	return manifest, nil
}

// resolveVoiceConfig merges default voice with language-specific override
func (g *TranscriptGenerator) resolveVoiceConfig(defaultVoice transcript.VoiceConfig, override *transcript.VoiceConfig) transcript.VoiceConfig {
	if override == nil {
		return defaultVoice
	}

	// Start with default, override non-zero values
	result := defaultVoice

	if override.VoiceID != "" {
		result.VoiceID = override.VoiceID
	}
	if override.VoiceName != "" {
		result.VoiceName = override.VoiceName
	}
	if override.Model != "" {
		result.Model = override.Model
	}
	if override.Stability != 0 {
		result.Stability = override.Stability
	}
	if override.SimilarityBoost != 0 {
		result.SimilarityBoost = override.SimilarityBoost
	}
	if override.Style != 0 {
		result.Style = override.Style
	}

	return result
}

// generateSlideAudio generates a single audio file and returns its duration
func (g *TranscriptGenerator) generateSlideAudio(ctx context.Context, text, outputPath string, voice transcript.VoiceConfig) (time.Duration, error) {
	// Build voice settings
	voiceSettings := elevenlabs.DefaultVoiceSettings()
	if voice.Stability != 0 {
		voiceSettings.Stability = voice.Stability
	}
	if voice.SimilarityBoost != 0 {
		voiceSettings.SimilarityBoost = voice.SimilarityBoost
	}
	if voice.Style != 0 {
		voiceSettings.Style = voice.Style
	}

	// Determine model
	model := voice.Model
	if model == "" {
		model = elevenlabs.DefaultModelID
	}

	// Create TTS request
	req := &elevenlabs.TTSRequest{
		VoiceID:       voice.VoiceID,
		Text:          text,
		ModelID:       model,
		VoiceSettings: voiceSettings,
	}

	// Generate speech
	resp, err := g.client.TextToSpeech().Generate(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("ElevenLabs TTS failed: %w", err)
	}

	// Read and save audio
	audioData, err := readAllAndClose(resp.Audio)
	if err != nil {
		return 0, fmt.Errorf("failed to read audio: %w", err)
	}

	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return 0, fmt.Errorf("failed to write audio file: %w", err)
	}

	// Get duration
	duration, err := getAudioDuration(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	return duration, nil
}

// readAllAndClose reads all data from an io.ReadCloser and closes it
func readAllAndClose(rc interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var data []byte
	buf := make([]byte, 32*1024)
	for {
		n, err := rc.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return data, nil
}
