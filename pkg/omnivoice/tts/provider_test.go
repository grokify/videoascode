package tts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grokify/videoascode/pkg/transcript"
	"github.com/plexusone/omnivoice-core/tts/providertest"
)

func TestNew(t *testing.T) {
	mock := providertest.NewElevenLabsMock()
	provider := New(mock)

	if provider == nil {
		t.Fatal("New() returned nil")
	}
	if provider.Name() != "elevenlabs" {
		t.Errorf("Name() = %s, want elevenlabs", provider.Name())
	}
}

func TestProvider_Name(t *testing.T) {
	tests := []struct {
		name     string
		mock     *providertest.MockProvider
		expected string
	}{
		{"elevenlabs", providertest.NewElevenLabsMock(), "elevenlabs"},
		{"deepgram", providertest.NewDeepgramMock(), "deepgram"},
		{"openai", providertest.NewOpenAIMock(), "openai"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := New(tc.mock)
			if got := provider.Name(); got != tc.expected {
				t.Errorf("Name() = %s, want %s", got, tc.expected)
			}
		})
	}
}

func TestProvider_Synthesize(t *testing.T) {
	mock := providertest.NewMockProviderWithOptions(
		providertest.WithName("test-provider"),
		providertest.WithFixedDuration(1000),
	)
	provider := New(mock)

	voice := transcript.VoiceConfig{
		VoiceID: "test-voice",
		Model:   "test-model",
	}

	audio, err := provider.Synthesize(context.Background(), "Hello world", voice)
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}

	if len(audio) == 0 {
		t.Error("Synthesize() returned empty audio")
	}

	// Verify it's a valid WAV (starts with RIFF)
	if len(audio) >= 4 && string(audio[0:4]) != "RIFF" {
		t.Error("Expected WAV audio format")
	}
}

func TestProvider_Synthesize_EmptyText(t *testing.T) {
	mock := providertest.NewElevenLabsMock()
	provider := New(mock)

	voice := transcript.VoiceConfig{VoiceID: "test"}

	// Empty text should still work (provider may return minimal audio)
	audio, err := provider.Synthesize(context.Background(), "", voice)
	if err != nil {
		t.Fatalf("Synthesize() with empty text error = %v", err)
	}

	// Should return some audio (even if minimal)
	if len(audio) == 0 {
		t.Error("Synthesize() with empty text returned no audio")
	}
}

func TestProvider_Synthesize_Error(t *testing.T) {
	mock := providertest.NewMockProviderWithOptions(
		providertest.WithName("failing-provider"),
		providertest.WithError(providertest.ErrMockRateLimit),
	)
	provider := New(mock)

	voice := transcript.VoiceConfig{VoiceID: "test"}

	_, err := provider.Synthesize(context.Background(), "test", voice)
	if err == nil {
		t.Fatal("Synthesize() expected error, got nil")
	}

	// Error should be wrapped with provider name
	if !errors.Is(err, providertest.ErrMockRateLimit) {
		t.Errorf("Expected ErrMockRateLimit in error chain, got: %v", err)
	}
}

func TestProvider_Synthesize_ContextCancellation(t *testing.T) {
	mock := providertest.NewMockProviderWithOptions(
		providertest.WithLatency(5 * time.Second),
	)
	provider := New(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	voice := transcript.VoiceConfig{VoiceID: "test"}

	start := time.Now()
	_, err := provider.Synthesize(ctx, "test", voice)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Synthesize() should return error on context cancellation")
	}

	if elapsed > 500*time.Millisecond {
		t.Errorf("Should have cancelled quickly, took %v", elapsed)
	}
}

func TestProvider_SynthesizeWithDuration(t *testing.T) {
	mock := providertest.NewMockProviderWithOptions(
		providertest.WithFixedDuration(2500), // 2.5 seconds
	)
	provider := New(mock)

	voice := transcript.VoiceConfig{VoiceID: "test"}

	audio, duration, err := provider.SynthesizeWithDuration(context.Background(), "Hello world", voice)
	if err != nil {
		t.Fatalf("SynthesizeWithDuration() error = %v", err)
	}

	if len(audio) == 0 {
		t.Error("SynthesizeWithDuration() returned empty audio")
	}

	expectedDuration := 2500 * time.Millisecond
	if duration != expectedDuration {
		t.Errorf("Duration = %v, want %v", duration, expectedDuration)
	}
}

func TestProvider_SynthesizeWithDuration_Error(t *testing.T) {
	mock := providertest.NewMockProviderWithOptions(
		providertest.WithError(providertest.ErrMockNetworkError),
	)
	provider := New(mock)

	voice := transcript.VoiceConfig{VoiceID: "test"}

	_, _, err := provider.SynthesizeWithDuration(context.Background(), "test", voice)
	if err == nil {
		t.Fatal("SynthesizeWithDuration() expected error, got nil")
	}
}

func TestProvider_UnderlyingProvider(t *testing.T) {
	mock := providertest.NewElevenLabsMock()
	provider := New(mock)

	underlying := provider.UnderlyingProvider()
	if underlying == nil {
		t.Fatal("UnderlyingProvider() returned nil")
	}

	if underlying.Name() != "elevenlabs" {
		t.Errorf("UnderlyingProvider().Name() = %s, want elevenlabs", underlying.Name())
	}
}

func TestVoiceConfigToSynthesisConfig(t *testing.T) {
	voice := transcript.VoiceConfig{
		VoiceID:         "voice-123",
		Model:           "eleven_turbo_v2",
		OutputFormat:    "mp3",
		SampleRate:      44100,
		Speed:           1.2,
		Pitch:           0.5,
		Stability:       0.7,
		SimilarityBoost: 0.8,
		Style:           0.3,
	}

	config := VoiceConfigToSynthesisConfig(voice)

	if config.VoiceID != "voice-123" {
		t.Errorf("VoiceID = %s, want voice-123", config.VoiceID)
	}
	if config.Model != "eleven_turbo_v2" {
		t.Errorf("Model = %s, want eleven_turbo_v2", config.Model)
	}
	if config.OutputFormat != "mp3" {
		t.Errorf("OutputFormat = %s, want mp3", config.OutputFormat)
	}
	if config.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", config.SampleRate)
	}
	if config.Speed != 1.2 {
		t.Errorf("Speed = %f, want 1.2", config.Speed)
	}
	if config.Pitch != 0.5 {
		t.Errorf("Pitch = %f, want 0.5", config.Pitch)
	}
	if config.Stability != 0.7 {
		t.Errorf("Stability = %f, want 0.7", config.Stability)
	}
	if config.SimilarityBoost != 0.8 {
		t.Errorf("SimilarityBoost = %f, want 0.8", config.SimilarityBoost)
	}

	// Check Style is passed via Extensions
	if config.Extensions == nil {
		t.Fatal("Extensions should not be nil when Style > 0")
	}
	if style, ok := config.Extensions["elevenlabs.style"]; !ok || style != 0.3 {
		t.Errorf("Extensions[elevenlabs.style] = %v, want 0.3", style)
	}
}

func TestVoiceConfigToSynthesisConfig_NoStyle(t *testing.T) {
	voice := transcript.VoiceConfig{
		VoiceID: "voice-123",
		Style:   0, // No style
	}

	config := VoiceConfigToSynthesisConfig(voice)

	// Extensions should be nil when Style is 0
	if config.Extensions != nil {
		t.Errorf("Extensions should be nil when Style is 0, got %v", config.Extensions)
	}
}

func TestProvider_WithRealisticTiming(t *testing.T) {
	mock := providertest.NewMockProviderWithOptions(
		providertest.WithRealisticTiming(),
	)
	provider := New(mock)

	voice := transcript.VoiceConfig{VoiceID: "test"}

	// Short text
	_, duration1, _ := provider.SynthesizeWithDuration(context.Background(), "Hi", voice)

	// Longer text
	_, duration2, _ := provider.SynthesizeWithDuration(context.Background(), "This is a much longer piece of text for testing", voice)

	if duration2 <= duration1 {
		t.Errorf("Longer text should have longer duration: %v vs %v", duration2, duration1)
	}
}

func TestProvider_WithFailAfterN(t *testing.T) {
	mock := providertest.NewMockProviderWithOptions(
		providertest.WithFailAfterN(2, providertest.ErrMockQuotaExceeded),
	)
	provider := New(mock)

	voice := transcript.VoiceConfig{VoiceID: "test"}

	// First two should succeed
	_, err1 := provider.Synthesize(context.Background(), "test1", voice)
	_, err2 := provider.Synthesize(context.Background(), "test2", voice)

	if err1 != nil || err2 != nil {
		t.Errorf("First two calls should succeed: err1=%v, err2=%v", err1, err2)
	}

	// Third should fail
	_, err3 := provider.Synthesize(context.Background(), "test3", voice)
	if err3 == nil {
		t.Error("Third call should fail")
	}
	if !errors.Is(err3, providertest.ErrMockQuotaExceeded) {
		t.Errorf("Expected ErrMockQuotaExceeded, got: %v", err3)
	}
}
