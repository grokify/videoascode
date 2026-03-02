package tts

import (
	"fmt"

	"github.com/plexusone/omnivoice"
	_ "github.com/plexusone/omnivoice/providers/all" // Register all providers
)

// ProviderConfig holds configuration for creating TTS providers.
type ProviderConfig struct {
	// ElevenLabsAPIKey is the API key for ElevenLabs.
	ElevenLabsAPIKey string

	// DeepgramAPIKey is the API key for Deepgram.
	DeepgramAPIKey string
}

// Factory creates TTS providers based on configuration.
type Factory struct {
	config    ProviderConfig
	providers map[string]*Provider
	fallback  string
}

// NewFactory creates a new TTS provider factory.
func NewFactory(config ProviderConfig) *Factory {
	return &Factory{
		config:    config,
		providers: make(map[string]*Provider),
	}
}

// Get returns a provider by name, creating it if necessary.
// If name is empty, returns the fallback provider.
func (f *Factory) Get(name string) (*Provider, error) {
	if name == "" {
		name = f.fallback
	}
	if name == "" {
		// Default to elevenlabs if available
		if f.config.ElevenLabsAPIKey != "" {
			name = "elevenlabs"
		} else if f.config.DeepgramAPIKey != "" {
			name = "deepgram"
		} else {
			return nil, fmt.Errorf("no provider specified and no API keys configured")
		}
	}

	// Return cached provider if available
	if provider, ok := f.providers[name]; ok {
		return provider, nil
	}

	// Create provider
	provider, err := f.createProvider(name)
	if err != nil {
		return nil, err
	}

	f.providers[name] = provider
	if f.fallback == "" {
		f.fallback = name
	}

	return provider, nil
}

// SetFallback sets the default provider name.
func (f *Factory) SetFallback(name string) {
	f.fallback = name
}

// createProvider creates a new provider instance using the omnivoice registry.
func (f *Factory) createProvider(name string) (*Provider, error) {
	var apiKey string
	switch name {
	case "elevenlabs":
		if f.config.ElevenLabsAPIKey == "" {
			return nil, fmt.Errorf("ElevenLabs API key not configured")
		}
		apiKey = f.config.ElevenLabsAPIKey
	case "deepgram":
		if f.config.DeepgramAPIKey == "" {
			return nil, fmt.Errorf("Deepgram API key not configured")
		}
		apiKey = f.config.DeepgramAPIKey
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	// Use the omnivoice registry to create the provider
	provider, err := omnivoice.GetTTSProvider(name, omnivoice.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w", name, err)
	}

	return New(provider), nil
}

// Available returns a list of available provider names based on configured API keys.
func (f *Factory) Available() []string {
	var names []string
	if f.config.ElevenLabsAPIKey != "" {
		names = append(names, "elevenlabs")
	}
	if f.config.DeepgramAPIKey != "" {
		names = append(names, "deepgram")
	}
	return names
}
