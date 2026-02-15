package tts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	omnitts "github.com/grokify/marp2video/pkg/omnivoice/tts"
	"github.com/grokify/marp2video/pkg/transcript"
	"github.com/grokify/mogo/log/slogutil"
)

// ProgressFunc is called during generation with current progress
type ProgressFunc func(current, total int, slideName string)

// TranscriptGeneratorConfig holds configuration for transcript-based TTS generation
type TranscriptGeneratorConfig struct {
	omnitts.ProviderConfig        // Embedded provider config (ElevenLabsAPIKey, DeepgramAPIKey)
	DefaultProvider        string // "elevenlabs" or "deepgram" - overrides voice config if set
	OutputDir              string
	Force                  bool         // If true, regenerate audio even if files exist
	ProgressFunc           ProgressFunc // Optional callback for progress updates
}

// TranscriptGenerator generates audio from transcript files
type TranscriptGenerator struct {
	config   TranscriptGeneratorConfig
	factory  *omnitts.Factory
	provider *omnitts.Provider
}

// NewTranscriptGenerator creates a new transcript-based TTS generator
func NewTranscriptGenerator(config TranscriptGeneratorConfig) (*TranscriptGenerator, error) {
	factory := omnitts.NewFactory(config.ProviderConfig)

	if config.DefaultProvider != "" {
		factory.SetFallback(config.DefaultProvider)
	}

	// Get the default provider to validate configuration
	provider, err := factory.Get("")
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS provider: %w", err)
	}

	return &TranscriptGenerator{
		config:   config,
		factory:  factory,
		provider: provider,
	}, nil
}

// GenerateFromTranscript generates audio files for all slides in a transcript
// Returns a manifest with timing information for use by the video recorder
func (g *TranscriptGenerator) GenerateFromTranscript(ctx context.Context, t *transcript.Transcript, language string) (*Manifest, error) {
	logger := slogutil.LoggerFromContext(ctx, slogutil.Null())

	// Create output directory
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Try to load existing manifest for resuming interrupted runs
	manifestPath := filepath.Join(g.config.OutputDir, "manifest.json")
	existingManifest, _ := LoadManifest(manifestPath) // Ignore error, may not exist

	// Create new manifest (will be populated with existing + new entries)
	manifest := NewManifest(language)
	numSlides := len(t.Slides)

	// Process each slide
	for i, slide := range t.Slides {
		// Report progress
		if g.config.ProgressFunc != nil {
			g.config.ProgressFunc(i+1, numSlides, slide.Title)
		}

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

		// Check if audio file already exists
		audioPath := filepath.Join(g.config.OutputDir, fmt.Sprintf("slide_%03d.mp3", slide.Index))

		// Skip if file exists and we're not forcing regeneration
		if !g.config.Force {
			if existingEntry, err := g.getExistingSlideAudio(audioPath, existingManifest, slide.Index); err == nil {
				logger.Info("skipping existing audio",
					"slide", slide.Index,
					"file", audioPath)
				manifest.AddSlide(*existingEntry)
				continue
			}
		}

		// Determine voice configuration
		voiceConfig := g.resolveVoiceConfig(t.Metadata.DefaultVoice, content.Voice)

		// If provider is being overridden via --provider flag, clear provider-specific settings
		// (e.g., ElevenLabs model "eleven_multilingual_v2" is not valid for Deepgram)
		if g.config.DefaultProvider != "" && voiceConfig.Provider != "" && voiceConfig.Provider != g.config.DefaultProvider {
			voiceConfig.Model = ""    // Clear model - let provider use its default
			voiceConfig.VoiceID = ""  // Clear voice ID - let provider use its default
			voiceConfig.Provider = "" // Clear so getProviderForVoice uses default
		}

		// Get provider for this voice (may be different per language/voice)
		provider, err := g.getProviderForVoice(voiceConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider for slide %d: %w", slide.Index, err)
		}

		logger.Info("generating audio",
			"slide", slide.Index,
			"provider", provider.Name(),
			"voice", voiceConfig.VoiceID,
			"textLength", len(text))

		audioDuration, err := g.generateSlideAudio(ctx, provider, text, audioPath, voiceConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate audio for slide %d: %w", slide.Index, err)
		}

		// Calculate pause duration
		pauseDuration := content.GetTotalPauseDuration()

		// Add to manifest
		slideAudio := SlideAudio{
			Index:         slide.Index,
			Title:         slide.Title,
			AudioFile:     filepath.Base(audioPath),
			AudioDuration: int(audioDuration.Milliseconds()),
			PauseDuration: pauseDuration,
			TotalDuration: int(audioDuration.Milliseconds()) + pauseDuration,
		}
		manifest.AddSlide(slideAudio)

		// Save manifest after each slide (for resume support)
		if err := manifest.SaveToFile(manifestPath); err != nil {
			logger.Warn("failed to save manifest", "error", err)
		}

		logger.Info("generated audio",
			"slide", slide.Index,
			"audioDurationMs", audioDuration.Milliseconds(),
			"pauseDurationMs", pauseDuration)
	}

	return manifest, nil
}

// getProviderForVoice returns the appropriate provider for a voice configuration
func (g *TranscriptGenerator) getProviderForVoice(voice transcript.VoiceConfig) (*omnitts.Provider, error) {
	// If --provider was explicitly set, use it (overrides voice config)
	if g.config.DefaultProvider != "" {
		return g.provider, nil
	}
	// Otherwise, use voice config's provider if specified
	if voice.Provider != "" {
		return g.factory.Get(voice.Provider)
	}
	return g.provider, nil
}

// getExistingSlideAudio checks if audio file exists and returns manifest entry if available
func (g *TranscriptGenerator) getExistingSlideAudio(audioPath string, existingManifest *Manifest, slideIndex int) (*SlideAudio, error) {
	// Check if file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file does not exist")
	}

	// Try to get duration from existing manifest
	if existingManifest != nil {
		if entry, err := existingManifest.GetSlide(slideIndex); err == nil {
			return entry, nil
		}
	}

	// File exists but no manifest entry - get duration from file
	duration, err := getAudioDuration(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	return &SlideAudio{
		Index:         slideIndex,
		AudioFile:     filepath.Base(audioPath),
		AudioDuration: int(duration.Milliseconds()),
		TotalDuration: int(duration.Milliseconds()),
	}, nil
}

// resolveVoiceConfig merges default voice with language-specific override
func (g *TranscriptGenerator) resolveVoiceConfig(defaultVoice transcript.VoiceConfig, override *transcript.VoiceConfig) transcript.VoiceConfig {
	if override == nil {
		return defaultVoice
	}

	// Start with default, override non-zero values
	result := defaultVoice

	if override.Provider != "" {
		result.Provider = override.Provider
	}
	if override.VoiceID != "" {
		result.VoiceID = override.VoiceID
	}
	if override.VoiceName != "" {
		result.VoiceName = override.VoiceName
	}
	if override.Model != "" {
		result.Model = override.Model
	}
	if override.OutputFormat != "" {
		result.OutputFormat = override.OutputFormat
	}
	if override.SampleRate != 0 {
		result.SampleRate = override.SampleRate
	}
	if override.Speed != 0 {
		result.Speed = override.Speed
	}
	if override.Pitch != 0 {
		result.Pitch = override.Pitch
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
func (g *TranscriptGenerator) generateSlideAudio(ctx context.Context, provider *omnitts.Provider, text, outputPath string, voice transcript.VoiceConfig) (time.Duration, error) {
	// Generate speech using OmniVoice provider
	audioData, err := provider.Synthesize(ctx, text, voice)
	if err != nil {
		return 0, fmt.Errorf("TTS synthesis failed: %w", err)
	}

	// Write audio to file
	if err := os.WriteFile(outputPath, audioData, 0600); err != nil {
		return 0, fmt.Errorf("failed to write audio file: %w", err)
	}

	// Get duration using ffprobe
	duration, err := getAudioDuration(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	return duration, nil
}
