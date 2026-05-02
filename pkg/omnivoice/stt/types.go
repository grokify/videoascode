// Package stt provides OmniVoice-based speech-to-text for videoascode.
// Types are re-exported from github.com/plexusone/omnivoice-core/stt for convenience.
package stt

import (
	"github.com/plexusone/omnivoice-core/stt"
)

// Type aliases for omnivoice-core/stt types.
// This allows videoascode code to continue using omnistt.* imports.

// TranscriptionConfig configures a transcription request.
type TranscriptionConfig = stt.TranscriptionConfig

// TranscriptionResult contains the result of a transcription.
type TranscriptionResult = stt.TranscriptionResult

// Segment represents a segment of transcription.
type Segment = stt.Segment

// Word represents a single transcribed word with timing.
type Word = stt.Word

// Transcript is the canonical JSON transcript format.
type Transcript = stt.Transcript

// TranscriptSegment is a segment in the canonical transcript format.
type TranscriptSegment = stt.TranscriptSegment

// TranscriptWord is a word in the canonical transcript format.
type TranscriptWord = stt.TranscriptWord

// TranscriptMetadata contains provenance information.
type TranscriptMetadata = stt.TranscriptMetadata

// TranscriptOptions records transcription options used.
type TranscriptOptions = stt.TranscriptOptions

// NewTranscript creates a Transcript from a TranscriptionResult.
var NewTranscript = stt.NewTranscript

// LoadTranscript reads a transcript from a JSON file.
var LoadTranscript = stt.LoadTranscript
