// Package stt provides OmniVoice-based speech-to-text for marp2video.
// This file re-exports the omnivoice subtitle package for convenience.
package stt

import (
	"github.com/plexusone/omnivoice"
)

// Re-export types and functions from omnivoice subtitle for convenience.
// This allows marp2video users to access subtitle functionality through this package.

// SubtitleOptions is an alias for omnivoice.SubtitleOptions.
type SubtitleOptions = omnivoice.SubtitleOptions

// SubtitleFormat is an alias for omnivoice.SubtitleFormat.
type SubtitleFormat = omnivoice.SubtitleFormat

// Format constants.
const (
	FormatSRT = omnivoice.SubtitleFormatSRT
	FormatVTT = omnivoice.SubtitleFormatVTT
)

// DefaultSubtitleOptions returns sensible defaults for subtitle generation.
func DefaultSubtitleOptions() SubtitleOptions {
	return omnivoice.DefaultSubtitleOptions()
}

// GenerateSRTFromResult generates SRT from a local TranscriptionResult.
// For direct omnivoice results, use omnivoice.GenerateSRT instead.
func GenerateSRTFromResult(result *TranscriptionResult, opts SubtitleOptions) string {
	// Convert local result to omnivoice result
	omniResult := resultToOmniVoice(result)
	return omnivoice.GenerateSRT(omniResult, opts)
}

// GenerateVTTFromResult generates VTT from a local TranscriptionResult.
// For direct omnivoice results, use omnivoice.GenerateVTT instead.
func GenerateVTTFromResult(result *TranscriptionResult, opts SubtitleOptions) string {
	omniResult := resultToOmniVoice(result)
	return omnivoice.GenerateVTT(omniResult, opts)
}

// SaveSRTFromResult saves SRT from a local TranscriptionResult.
func SaveSRTFromResult(result *TranscriptionResult, filePath string, opts SubtitleOptions) error {
	omniResult := resultToOmniVoice(result)
	return omnivoice.SaveSRT(omniResult, filePath, opts)
}

// SaveVTTFromResult saves VTT from a local TranscriptionResult.
func SaveVTTFromResult(result *TranscriptionResult, filePath string, opts SubtitleOptions) error {
	omniResult := resultToOmniVoice(result)
	return omnivoice.SaveVTT(omniResult, filePath, opts)
}

// resultToOmniVoice converts local TranscriptionResult to omnivoice TranscriptionResult.
func resultToOmniVoice(result *TranscriptionResult) *omnivoice.TranscriptionResult {
	r := &omnivoice.TranscriptionResult{
		Text:     result.Text,
		Language: result.Language,
		Duration: result.Duration,
	}

	for _, seg := range result.Segments {
		omniSeg := omnivoice.Segment{
			Text:       seg.Text,
			StartTime:  seg.StartTime,
			EndTime:    seg.EndTime,
			Confidence: seg.Confidence,
			Speaker:    seg.Speaker,
		}

		for _, w := range seg.Words {
			omniSeg.Words = append(omniSeg.Words, omnivoice.Word{
				Text:       w.Text,
				StartTime:  w.StartTime,
				EndTime:    w.EndTime,
				Confidence: w.Confidence,
				Speaker:    w.Speaker,
			})
		}

		r.Segments = append(r.Segments, omniSeg)
	}

	return r
}
