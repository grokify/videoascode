package stt

import (
	"context"
	"errors"
	"testing"

	"github.com/plexusone/omnivoice-core/stt"
	"github.com/plexusone/omnivoice-core/stt/providertest"
)

func TestNew(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	if provider == nil {
		t.Fatal("New() returned nil")
	}
	if provider.Name() != "mock" {
		t.Errorf("Name() = %s, want mock", provider.Name())
	}
}

func TestProvider_Name(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	if got := provider.Name(); got != "mock" {
		t.Errorf("Name() = %s, want mock", got)
	}
}

func TestProvider_Transcribe(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	config := TranscriptionConfig{
		Language:             "en-US",
		EnablePunctuation:    true,
		EnableWordTimestamps: true,
	}

	result, err := provider.Transcribe(context.Background(), []byte("fake-audio"), config)
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	if result.Text == "" {
		t.Error("Transcribe() returned empty text")
	}

	if result.Language != "en-US" {
		t.Errorf("Language = %s, want en-US", result.Language)
	}

	if len(result.Segments) == 0 {
		t.Error("Transcribe() returned no segments")
	}

	// Check word timestamps are present
	if len(result.Segments) > 0 && len(result.Segments[0].Words) == 0 {
		t.Error("Transcribe() returned no words in segment")
	}
}

func TestProvider_Transcribe_ContextCancellation(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := TranscriptionConfig{Language: "en-US"}

	_, err := provider.Transcribe(ctx, []byte("audio"), config)
	if err == nil {
		t.Error("Transcribe() should return error on cancelled context")
	}
}

func TestProvider_Transcribe_Error(t *testing.T) {
	testErr := errors.New("test transcription error")

	mock := providertest.NewMockProvider()
	mock.TranscribeFunc = func(ctx context.Context, audio []byte, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
		return nil, testErr
	}
	provider := New(mock)

	config := TranscriptionConfig{Language: "en-US"}

	_, err := provider.Transcribe(context.Background(), []byte("audio"), config)
	if err == nil {
		t.Fatal("Transcribe() expected error, got nil")
	}

	// Error should be wrapped with provider name
	if !errors.Is(err, testErr) {
		t.Errorf("Expected test error in chain, got: %v", err)
	}
}

func TestProvider_TranscribeFile(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	config := TranscriptionConfig{
		Language:          "en-US",
		EnablePunctuation: true,
	}

	result, err := provider.TranscribeFile(context.Background(), "/path/to/audio.mp3", config)
	if err != nil {
		t.Fatalf("TranscribeFile() error = %v", err)
	}

	if result.Text == "" {
		t.Error("TranscribeFile() returned empty text")
	}
}

func TestProvider_TranscribeFile_ContextCancellation(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := TranscriptionConfig{Language: "en-US"}

	_, err := provider.TranscribeFile(ctx, "/path/to/audio.mp3", config)
	if err == nil {
		t.Error("TranscribeFile() should return error on cancelled context")
	}
}

func TestProvider_TranscribeURL(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	config := TranscriptionConfig{
		Language:          "en-US",
		EnablePunctuation: true,
	}

	result, err := provider.TranscribeURL(context.Background(), "https://example.com/audio.mp3", config)
	if err != nil {
		t.Fatalf("TranscribeURL() error = %v", err)
	}

	if result.Text == "" {
		t.Error("TranscribeURL() returned empty text")
	}
}

func TestProvider_TranscribeURL_ContextCancellation(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := TranscriptionConfig{Language: "en-US"}

	_, err := provider.TranscribeURL(ctx, "https://example.com/audio.mp3", config)
	if err == nil {
		t.Error("TranscribeURL() should return error on cancelled context")
	}
}

func TestProvider_UnderlyingProvider(t *testing.T) {
	mock := providertest.NewMockProvider()
	provider := New(mock)

	underlying := provider.UnderlyingProvider()
	if underlying == nil {
		t.Fatal("UnderlyingProvider() returned nil")
	}

	if underlying.Name() != "mock" {
		t.Errorf("UnderlyingProvider().Name() = %s, want mock", underlying.Name())
	}
}
