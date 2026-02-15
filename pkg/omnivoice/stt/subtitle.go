// Package stt provides OmniVoice-based speech-to-text for marp2video.
// This file re-exports the omnivoice/subtitle package for convenience.
package stt

import (
	"github.com/agentplexus/omnivoice/stt"
	"github.com/agentplexus/omnivoice/subtitle"
)

// Re-export types and functions from omnivoice/subtitle for convenience.
// This allows marp2video users to access subtitle functionality through this package.

// SubtitleOptions is an alias for subtitle.Options.
type SubtitleOptions = subtitle.Options

// SubtitleFormat is an alias for subtitle.Format.
type SubtitleFormat = subtitle.Format

// Format constants.
const (
	FormatSRT = subtitle.FormatSRT
	FormatVTT = subtitle.FormatVTT
)

// DefaultSubtitleOptions returns sensible defaults for subtitle generation.
func DefaultSubtitleOptions() SubtitleOptions {
	return subtitle.DefaultOptions()
}

// GenerateSRTFromResult generates SRT from a local TranscriptionResult.
// For direct omnivoice results, use subtitle.GenerateSRT instead.
func GenerateSRTFromResult(result *TranscriptionResult, opts SubtitleOptions) string {
	// Convert local result to omnivoice result
	omniResult := resultToOmniVoice(result)
	return subtitle.GenerateSRT(omniResult, opts)
}

// GenerateVTTFromResult generates VTT from a local TranscriptionResult.
// For direct omnivoice results, use subtitle.GenerateVTT instead.
func GenerateVTTFromResult(result *TranscriptionResult, opts SubtitleOptions) string {
	omniResult := resultToOmniVoice(result)
	return subtitle.GenerateVTT(omniResult, opts)
}

// SaveSRTFromResult saves SRT from a local TranscriptionResult.
func SaveSRTFromResult(result *TranscriptionResult, filePath string, opts SubtitleOptions) error {
	omniResult := resultToOmniVoice(result)
	return subtitle.SaveSRT(omniResult, filePath, opts)
}

// SaveVTTFromResult saves VTT from a local TranscriptionResult.
func SaveVTTFromResult(result *TranscriptionResult, filePath string, opts SubtitleOptions) error {
	omniResult := resultToOmniVoice(result)
	return subtitle.SaveVTT(omniResult, filePath, opts)
}

// resultToOmniVoice converts local TranscriptionResult to omnivoice stt.TranscriptionResult.
func resultToOmniVoice(result *TranscriptionResult) *stt.TranscriptionResult {
	r := &stt.TranscriptionResult{
		Text:     result.Text,
		Language: result.Language,
		Duration: result.Duration,
	}

	for _, seg := range result.Segments {
		omniSeg := stt.Segment{
			Text:       seg.Text,
			StartTime:  seg.StartTime,
			EndTime:    seg.EndTime,
			Confidence: seg.Confidence,
			Speaker:    seg.Speaker,
		}

		for _, w := range seg.Words {
			omniSeg.Words = append(omniSeg.Words, stt.Word{
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
