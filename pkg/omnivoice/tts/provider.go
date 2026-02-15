// Package tts provides OmniVoice-based text-to-speech for marp2video.
package tts

import (
	"context"
	"fmt"
	"time"

	"github.com/grokify/marp2video/pkg/transcript"

	"github.com/agentplexus/omnivoice/tts"
)

// Provider wraps an OmniVoice TTS provider for use with marp2video.
type Provider struct {
	provider tts.Provider
	name     string
}

// New creates a new OmniVoice TTS provider wrapper.
func New(provider tts.Provider) *Provider {
	return &Provider{
		provider: provider,
		name:     provider.Name(),
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return p.name
}

// Synthesize converts text to speech using the OmniVoice provider.
func (p *Provider) Synthesize(ctx context.Context, text string, voice transcript.VoiceConfig) ([]byte, error) {
	config := VoiceConfigToSynthesisConfig(voice)

	result, err := p.provider.Synthesize(ctx, text, config)
	if err != nil {
		return nil, fmt.Errorf("%s tts failed: %w", p.name, err)
	}

	return result.Audio, nil
}

// SynthesizeWithDuration converts text to speech and returns audio with duration.
func (p *Provider) SynthesizeWithDuration(ctx context.Context, text string, voice transcript.VoiceConfig) ([]byte, time.Duration, error) {
	config := VoiceConfigToSynthesisConfig(voice)

	result, err := p.provider.Synthesize(ctx, text, config)
	if err != nil {
		return nil, 0, fmt.Errorf("%s tts failed: %w", p.name, err)
	}

	// Duration from result if available, otherwise caller should use ffprobe
	duration := time.Duration(result.DurationMs) * time.Millisecond

	return result.Audio, duration, nil
}

// VoiceConfigToSynthesisConfig converts marp2video VoiceConfig to OmniVoice SynthesisConfig.
func VoiceConfigToSynthesisConfig(voice transcript.VoiceConfig) tts.SynthesisConfig {
	config := tts.SynthesisConfig{
		VoiceID:         voice.VoiceID,
		Model:           voice.Model,
		OutputFormat:    voice.OutputFormat,
		SampleRate:      voice.SampleRate,
		Speed:           voice.Speed,
		Pitch:           voice.Pitch,
		Stability:       voice.Stability,
		SimilarityBoost: voice.SimilarityBoost,
	}

	// Style is ElevenLabs-specific and uses the Extensions mechanism
	if voice.Style > 0 {
		config.Extensions = map[string]any{
			"elevenlabs.style": voice.Style,
		}
	}

	return config
}

// UnderlyingProvider returns the wrapped OmniVoice provider for advanced operations.
func (p *Provider) UnderlyingProvider() tts.Provider {
	return p.provider
}
