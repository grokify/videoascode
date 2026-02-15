package video

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// bcp47ToISO639Map maps BCP-47 language codes to ISO 639-2 (3-letter) codes.
// ffmpeg uses ISO 639-2 for subtitle track metadata.
var bcp47ToISO639Map = map[string]string{
	// English variants
	"en":    "eng",
	"en-US": "eng",
	"en-GB": "eng",
	"en-AU": "eng",

	// French variants
	"fr":    "fra",
	"fr-FR": "fra",
	"fr-CA": "fra",

	// German variants
	"de":    "deu",
	"de-DE": "deu",
	"de-AT": "deu",

	// Spanish variants
	"es":    "spa",
	"es-ES": "spa",
	"es-MX": "spa",

	// Chinese variants
	"zh":      "zho",
	"zh-Hans": "zho",
	"zh-Hant": "zho",
	"zh-CN":   "zho",
	"zh-TW":   "zho",

	// Japanese
	"ja":    "jpn",
	"ja-JP": "jpn",

	// Korean
	"ko":    "kor",
	"ko-KR": "kor",

	// Portuguese variants
	"pt":    "por",
	"pt-BR": "por",
	"pt-PT": "por",

	// Italian
	"it":    "ita",
	"it-IT": "ita",

	// Russian
	"ru":    "rus",
	"ru-RU": "rus",

	// Arabic
	"ar": "ara",

	// Hindi
	"hi":    "hin",
	"hi-IN": "hin",
}

// BCP47ToISO639 converts a BCP-47 language code to ISO 639-2 (3-letter) code.
// If the code is not found, it returns the first 3 characters or the original code.
func BCP47ToISO639(bcp47 string) string {
	// Direct lookup
	if iso, ok := bcp47ToISO639Map[bcp47]; ok {
		return iso
	}

	// Try without region (e.g., "en-US" -> "en")
	if idx := strings.Index(bcp47, "-"); idx > 0 {
		base := bcp47[:idx]
		if iso, ok := bcp47ToISO639Map[base]; ok {
			return iso
		}
		// Return the base language code (before hyphen) if not in map
		if len(base) >= 2 && len(base) <= 3 {
			return strings.ToLower(base)
		}
	}

	// Fallback: return first 3 chars if it looks like ISO 639-2 already
	if len(bcp47) == 3 {
		return strings.ToLower(bcp47)
	}

	// Last resort: return first 2-3 chars if no hyphen
	if len(bcp47) >= 2 {
		return strings.ToLower(bcp47[:min(3, len(bcp47))])
	}

	return "und" // undefined
}

// DetectLanguageFromSubtitlePath extracts the language code from a subtitle filename.
// Supports patterns like "en-US.srt", "subtitles/fr-FR.vtt", "slide_001.en-US.srt"
func DetectLanguageFromSubtitlePath(path string) string {
	base := filepath.Base(path)

	// Remove extension
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Check if the whole name is a locale (e.g., "en-US.srt")
	if isValidLocale(name) {
		return name
	}

	// Check if it ends with a locale (e.g., "slide_001.en-US.srt")
	parts := strings.Split(name, ".")
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if isValidLocale(lastPart) {
			return lastPart
		}
	}

	return ""
}

// isValidLocale checks if a string looks like a valid BCP-47 locale code.
// Supports: language-region (en-US) or language-script (zh-Hans)
func isValidLocale(s string) bool {
	parts := strings.Split(s, "-")
	// Only support 2-part locales (language-region or language-script)
	if len(parts) != 2 {
		return false
	}

	// First part should be 2-3 lowercase letters (language)
	lang := parts[0]
	if len(lang) < 2 || len(lang) > 3 {
		return false
	}
	for _, c := range lang {
		if c < 'a' || c > 'z' {
			return false
		}
	}

	// Second part can be region (2 uppercase) or script (4 mixed case)
	second := parts[1]
	if len(second) == 2 {
		// Region code (e.g., "US", "FR")
		for _, c := range second {
			if c < 'A' || c > 'Z' {
				return false
			}
		}
	} else if len(second) == 4 {
		// Script code (e.g., "Hans", "Hant")
		if second[0] < 'A' || second[0] > 'Z' {
			return false
		}
		for _, c := range second[1:] {
			if c < 'a' || c > 'z' {
				return false
			}
		}
	} else {
		return false
	}

	return true
}

// EmbedSubtitles embeds a subtitle file into a video as a soft subtitle track.
// The subtitle track can be toggled on/off by the viewer.
// Supports SRT and VTT formats.
func EmbedSubtitles(videoPath, subtitlePath, language, outputPath string) error {
	// Validate subtitle format
	ext := strings.ToLower(filepath.Ext(subtitlePath))
	if ext != ".srt" && ext != ".vtt" {
		return fmt.Errorf("unsupported subtitle format: %s (use .srt or .vtt)", ext)
	}

	// Build ffmpeg command
	// -c:v copy - copy video stream without re-encoding
	// -c:a copy - copy audio stream without re-encoding
	// -c:s mov_text - encode subtitles for MP4 container
	// -metadata:s:s:0 language=XXX - set subtitle track language
	//nolint:gosec // G204: arguments are internal file paths, not user input
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-i", subtitlePath,
		"-c:v", "copy",
		"-c:a", "copy",
		"-c:s", "mov_text",
		"-metadata:s:s:0", fmt.Sprintf("language=%s", language),
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg subtitle embedding failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// BurnSubtitles burns subtitles directly into the video frames.
// The subtitles become permanent and cannot be toggled off.
// This is useful for social media or when subtitle track support is limited.
func BurnSubtitles(videoPath, subtitlePath, outputPath string) error {
	// Validate subtitle format
	ext := strings.ToLower(filepath.Ext(subtitlePath))
	if ext != ".srt" && ext != ".vtt" && ext != ".ass" {
		return fmt.Errorf("unsupported subtitle format: %s (use .srt, .vtt, or .ass)", ext)
	}

	// Build ffmpeg command
	// -vf subtitles=file.srt - burn subtitles using the subtitles filter
	//nolint:gosec // G204: arguments are internal file paths, not user input
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", fmt.Sprintf("subtitles=%s", subtitlePath),
		"-c:a", "copy",
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg subtitle burning failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
