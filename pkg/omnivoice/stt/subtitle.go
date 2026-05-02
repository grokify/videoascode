// Package stt provides OmniVoice-based speech-to-text for videoascode.
// This file re-exports the omnivoice subtitle package for convenience.
package stt

import (
	"github.com/plexusone/omnivoice"
	"github.com/plexusone/omnivoice-core/stt"
)

// Re-export types and functions from omnivoice subtitle for convenience.
// This allows videoascode users to access subtitle functionality through this package.

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

// GenerateSRTFromResult generates SRT from a TranscriptionResult.
func GenerateSRTFromResult(result *stt.TranscriptionResult, opts SubtitleOptions) string {
	return omnivoice.GenerateSRT(result, opts)
}

// GenerateVTTFromResult generates VTT from a TranscriptionResult.
func GenerateVTTFromResult(result *stt.TranscriptionResult, opts SubtitleOptions) string {
	return omnivoice.GenerateVTT(result, opts)
}

// SaveSRTFromResult saves SRT from a TranscriptionResult.
func SaveSRTFromResult(result *stt.TranscriptionResult, filePath string, opts SubtitleOptions) error {
	return omnivoice.SaveSRT(result, filePath, opts)
}

// SaveVTTFromResult saves VTT from a TranscriptionResult.
func SaveVTTFromResult(result *stt.TranscriptionResult, filePath string, opts SubtitleOptions) error {
	return omnivoice.SaveVTT(result, filePath, opts)
}
