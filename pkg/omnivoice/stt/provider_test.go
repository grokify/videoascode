package stt

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestConfigToOmniVoice(t *testing.T) {
	config := TranscriptionConfig{
		Language:                 "es-ES",
		Model:                    "nova-2",
		EnablePunctuation:        true,
		EnableWordTimestamps:     true,
		EnableSpeakerDiarization: true,
		MaxSpeakers:              3,
	}

	omniConfig := configToOmniVoice(config)

	if omniConfig.Language != "es-ES" {
		t.Errorf("Language = %s, want es-ES", omniConfig.Language)
	}
	if omniConfig.Model != "nova-2" {
		t.Errorf("Model = %s, want nova-2", omniConfig.Model)
	}
	if !omniConfig.EnablePunctuation {
		t.Error("EnablePunctuation should be true")
	}
	if !omniConfig.EnableWordTimestamps {
		t.Error("EnableWordTimestamps should be true")
	}
	if !omniConfig.EnableSpeakerDiarization {
		t.Error("EnableSpeakerDiarization should be true")
	}
	if omniConfig.MaxSpeakers != 3 {
		t.Errorf("MaxSpeakers = %d, want 3", omniConfig.MaxSpeakers)
	}
}

func TestResultFromOmniVoice(t *testing.T) {
	omniResult := &stt.TranscriptionResult{
		Text:     "Hello world",
		Language: "en-US",
		Duration: 2 * time.Second,
		Segments: []stt.Segment{
			{
				Text:       "Hello world",
				StartTime:  0,
				EndTime:    2 * time.Second,
				Confidence: 0.95,
				Speaker:    "speaker1",
				Words: []stt.Word{
					{
						Text:       "Hello",
						StartTime:  0,
						EndTime:    500 * time.Millisecond,
						Confidence: 0.98,
						Speaker:    "speaker1",
					},
					{
						Text:       "world",
						StartTime:  600 * time.Millisecond,
						EndTime:    2 * time.Second,
						Confidence: 0.92,
						Speaker:    "speaker1",
					},
				},
			},
		},
	}

	result := resultFromOmniVoice(omniResult)

	if result.Text != "Hello world" {
		t.Errorf("Text = %s, want 'Hello world'", result.Text)
	}
	if result.Language != "en-US" {
		t.Errorf("Language = %s, want en-US", result.Language)
	}
	if result.Duration != 2*time.Second {
		t.Errorf("Duration = %v, want 2s", result.Duration)
	}
	if len(result.Segments) != 1 {
		t.Fatalf("Segments count = %d, want 1", len(result.Segments))
	}

	seg := result.Segments[0]
	if seg.Text != "Hello world" {
		t.Errorf("Segment.Text = %s, want 'Hello world'", seg.Text)
	}
	if seg.Confidence != 0.95 {
		t.Errorf("Segment.Confidence = %f, want 0.95", seg.Confidence)
	}
	if seg.Speaker != "speaker1" {
		t.Errorf("Segment.Speaker = %s, want speaker1", seg.Speaker)
	}
	if len(seg.Words) != 2 {
		t.Fatalf("Words count = %d, want 2", len(seg.Words))
	}

	word := seg.Words[0]
	if word.Text != "Hello" {
		t.Errorf("Word.Text = %s, want 'Hello'", word.Text)
	}
	if word.Confidence != 0.98 {
		t.Errorf("Word.Confidence = %f, want 0.98", word.Confidence)
	}
}

func TestResultFromOmniVoice_EmptySegments(t *testing.T) {
	omniResult := &stt.TranscriptionResult{
		Text:     "Hello",
		Language: "en-US",
		Duration: time.Second,
		Segments: nil,
	}

	result := resultFromOmniVoice(omniResult)

	if result.Text != "Hello" {
		t.Errorf("Text = %s, want 'Hello'", result.Text)
	}
	if result.Segments != nil {
		t.Errorf("Segments should be nil, got %v", result.Segments)
	}
}
