package stt

import "time"

// TranscriptionConfig configures a transcription request.
type TranscriptionConfig struct {
	// Language is the BCP-47 language code (e.g., "en-US").
	Language string

	// Model is the provider-specific model identifier.
	Model string

	// EnablePunctuation adds punctuation to transcription.
	EnablePunctuation bool

	// EnableWordTimestamps includes word-level timestamps.
	EnableWordTimestamps bool

	// EnableSpeakerDiarization identifies different speakers.
	EnableSpeakerDiarization bool

	// MaxSpeakers is the maximum number of speakers to detect.
	MaxSpeakers int
}

// TranscriptionResult contains the result of a transcription.
type TranscriptionResult struct {
	// Text is the full transcription text.
	Text string

	// Segments contains segment-level details with timing.
	Segments []Segment

	// Language is the detected language.
	Language string

	// Duration is the audio duration.
	Duration time.Duration
}

// Segment represents a segment of transcription.
type Segment struct {
	// Text is the transcribed text for this segment.
	Text string

	// StartTime is when the segment starts.
	StartTime time.Duration

	// EndTime is when the segment ends.
	EndTime time.Duration

	// Confidence is the average confidence for this segment.
	Confidence float64

	// Speaker is the speaker identifier (if diarization enabled).
	Speaker string

	// Words contains word-level details.
	Words []Word
}

// Word represents a single transcribed word with timing.
type Word struct {
	// Text is the transcribed word.
	Text string

	// StartTime is when the word starts.
	StartTime time.Duration

	// EndTime is when the word ends.
	EndTime time.Duration

	// Confidence is the recognition confidence (0.0 to 1.0).
	Confidence float64

	// Speaker is the speaker identifier.
	Speaker string
}
