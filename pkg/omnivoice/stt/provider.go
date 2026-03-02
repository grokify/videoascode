// Package stt provides OmniVoice-based speech-to-text for marp2video.
package stt

import (
	"context"
	"fmt"

	"github.com/plexusone/omnivoice"
)

// Provider wraps an OmniVoice STT provider for use with marp2video.
type Provider struct {
	provider omnivoice.STTProvider
	name     string
}

// New creates a new OmniVoice STT provider wrapper.
func New(provider omnivoice.STTProvider) *Provider {
	return &Provider{
		provider: provider,
		name:     provider.Name(),
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return p.name
}

// TranscribeFile transcribes an audio file and returns the result.
func (p *Provider) TranscribeFile(ctx context.Context, filePath string, config TranscriptionConfig) (*TranscriptionResult, error) {
	omniConfig := configToOmniVoice(config)

	result, err := p.provider.TranscribeFile(ctx, filePath, omniConfig)
	if err != nil {
		return nil, fmt.Errorf("%s stt failed: %w", p.name, err)
	}

	return resultFromOmniVoice(result), nil
}

// TranscribeURL transcribes audio from a URL and returns the result.
func (p *Provider) TranscribeURL(ctx context.Context, url string, config TranscriptionConfig) (*TranscriptionResult, error) {
	omniConfig := configToOmniVoice(config)

	result, err := p.provider.TranscribeURL(ctx, url, omniConfig)
	if err != nil {
		return nil, fmt.Errorf("%s stt failed: %w", p.name, err)
	}

	return resultFromOmniVoice(result), nil
}

// Transcribe transcribes audio bytes and returns the result.
func (p *Provider) Transcribe(ctx context.Context, audio []byte, config TranscriptionConfig) (*TranscriptionResult, error) {
	omniConfig := configToOmniVoice(config)

	result, err := p.provider.Transcribe(ctx, audio, omniConfig)
	if err != nil {
		return nil, fmt.Errorf("%s stt failed: %w", p.name, err)
	}

	return resultFromOmniVoice(result), nil
}

// UnderlyingProvider returns the wrapped OmniVoice provider for advanced operations.
func (p *Provider) UnderlyingProvider() omnivoice.STTProvider {
	return p.provider
}

// configToOmniVoice converts marp2video config to OmniVoice config.
func configToOmniVoice(config TranscriptionConfig) omnivoice.TranscriptionConfig {
	return omnivoice.TranscriptionConfig{
		Language:                 config.Language,
		Model:                    config.Model,
		EnablePunctuation:        config.EnablePunctuation,
		EnableWordTimestamps:     config.EnableWordTimestamps,
		EnableSpeakerDiarization: config.EnableSpeakerDiarization,
		MaxSpeakers:              config.MaxSpeakers,
	}
}

// resultFromOmniVoice converts OmniVoice result to marp2video result.
func resultFromOmniVoice(result *omnivoice.TranscriptionResult) *TranscriptionResult {
	r := &TranscriptionResult{
		Text:     result.Text,
		Language: result.Language,
		Duration: result.Duration,
	}

	for _, seg := range result.Segments {
		segment := Segment{
			Text:       seg.Text,
			StartTime:  seg.StartTime,
			EndTime:    seg.EndTime,
			Confidence: seg.Confidence,
			Speaker:    seg.Speaker,
		}

		for _, w := range seg.Words {
			segment.Words = append(segment.Words, Word{
				Text:       w.Text,
				StartTime:  w.StartTime,
				EndTime:    w.EndTime,
				Confidence: w.Confidence,
				Speaker:    w.Speaker,
			})
		}

		r.Segments = append(r.Segments, segment)
	}

	return r
}
