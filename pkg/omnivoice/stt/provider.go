// Package stt provides OmniVoice-based speech-to-text for videoascode.
package stt

import (
	"context"
	"fmt"

	"github.com/plexusone/omnivoice"
	"github.com/plexusone/omnivoice-core/stt"
)

// Provider wraps an OmniVoice STT provider for use with videoascode.
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
func (p *Provider) TranscribeFile(ctx context.Context, filePath string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	result, err := p.provider.TranscribeFile(ctx, filePath, config)
	if err != nil {
		return nil, fmt.Errorf("%s stt failed: %w", p.name, err)
	}
	return result, nil
}

// TranscribeURL transcribes audio from a URL and returns the result.
func (p *Provider) TranscribeURL(ctx context.Context, url string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	result, err := p.provider.TranscribeURL(ctx, url, config)
	if err != nil {
		return nil, fmt.Errorf("%s stt failed: %w", p.name, err)
	}
	return result, nil
}

// Transcribe transcribes audio bytes and returns the result.
func (p *Provider) Transcribe(ctx context.Context, audio []byte, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	result, err := p.provider.Transcribe(ctx, audio, config)
	if err != nil {
		return nil, fmt.Errorf("%s stt failed: %w", p.name, err)
	}
	return result, nil
}

// UnderlyingProvider returns the wrapped OmniVoice provider for advanced operations.
func (p *Provider) UnderlyingProvider() omnivoice.STTProvider {
	return p.provider
}
