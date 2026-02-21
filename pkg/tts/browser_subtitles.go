package tts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/marp2video/pkg/segment"
)

// BrowserSubtitleGenerator generates subtitles for browser segments using voiceover timing.
// This is a simple approach that doesn't require STT - it uses the known voiceover text
// and TTS-generated timing data.
type BrowserSubtitleGenerator struct {
	format SubtitleFormat
}

// NewBrowserSubtitleGenerator creates a new browser subtitle generator.
func NewBrowserSubtitleGenerator(format SubtitleFormat) *BrowserSubtitleGenerator {
	if format == "" {
		format = FormatSRT
	}
	return &BrowserSubtitleGenerator{format: format}
}

// VoiceoverTiming holds timing information for a single voiceover.
type VoiceoverTiming struct {
	Index    int
	Text     string
	StartMs  int
	EndMs    int
	Duration int
}

// GenerateFromSegment generates subtitles for a browser segment using audio results.
func (g *BrowserSubtitleGenerator) GenerateFromSegment(
	seg segment.Segment,
	audioResult *SegmentAudioResult,
	language string,
	outputPath string,
) error {
	voiceovers := seg.GetVoiceovers(language)
	if len(voiceovers) == 0 {
		return fmt.Errorf("no voiceovers found for segment %s in language %s", seg.GetID(), language)
	}

	// Build timing data from voiceover durations
	timings := make([]VoiceoverTiming, 0, len(voiceovers))
	currentTimeMs := 0

	for _, vo := range voiceovers {
		duration := audioResult.VoiceoverDurations[vo.Index]
		if duration == 0 {
			// Fallback: estimate 150 words per minute
			wordCount := len(strings.Fields(vo.Text))
			duration = wordCount * 400 // ~400ms per word
		}

		timings = append(timings, VoiceoverTiming{
			Index:    vo.Index,
			Text:     vo.Text,
			StartMs:  currentTimeMs,
			EndMs:    currentTimeMs + duration,
			Duration: duration,
		})

		currentTimeMs += duration
		// Add pause if specified
		if vo.Pause > 0 {
			currentTimeMs += vo.Pause
		}
	}

	// Split long voiceovers into 2-line chunks for readability
	chunkedTimings := splitTimingsIntoChunks(timings)

	// Generate subtitle content
	var content string
	switch g.format {
	case FormatVTT:
		content = g.generateVTT(chunkedTimings)
	default:
		content = g.generateSRT(chunkedTimings)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write subtitle file
	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write subtitle file: %w", err)
	}

	return nil
}

// GenerateFromTimings generates subtitles from pre-computed timing data.
func (g *BrowserSubtitleGenerator) GenerateFromTimings(timings []VoiceoverTiming, outputPath string) error {
	// Split long voiceovers into 2-line chunks
	chunkedTimings := splitTimingsIntoChunks(timings)

	var content string
	switch g.format {
	case FormatVTT:
		content = g.generateVTT(chunkedTimings)
	default:
		content = g.generateSRT(chunkedTimings)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(content), 0600)
}

// splitTimingsIntoChunks splits long voiceover text into smaller subtitle chunks.
// Each chunk is limited to 2 lines of ~42 characters each (standard for 1080p video).
// This ensures no line exceeds the display width when burned into video.
func splitTimingsIntoChunks(timings []VoiceoverTiming) []VoiceoverTiming {
	const maxCharsPerLine = 42
	const maxLines = 2

	var result []VoiceoverTiming

	for _, t := range timings {
		text := strings.TrimSpace(t.Text)
		if text == "" {
			continue
		}

		// Always use splitTextIntoChunks to properly handle word boundaries.
		// Character count alone doesn't determine how many lines are needed
		// because words may not break evenly at line boundaries.
		chunks := splitTextIntoChunks(text, maxCharsPerLine, maxLines)
		if len(chunks) == 0 {
			continue
		}

		// If only one chunk, use original timing
		if len(chunks) == 1 {
			result = append(result, VoiceoverTiming{
				Index:    len(result) + 1,
				Text:     chunks[0],
				StartMs:  t.StartMs,
				EndMs:    t.EndMs,
				Duration: t.Duration,
			})
			continue
		}

		// Distribute timing proportionally across multiple chunks based on word count
		// Word count is more accurate than character count since TTS timing is more
		// consistent per word than per character
		totalWords := len(strings.Fields(text))
		if totalWords == 0 {
			totalWords = 1
		}
		currentStartMs := t.StartMs

		for i, chunk := range chunks {
			// Calculate proportional duration based on word count
			chunkWords := len(strings.Fields(strings.ReplaceAll(chunk, "\n", " ")))
			if chunkWords == 0 {
				chunkWords = 1
			}
			chunkDuration := (t.Duration * chunkWords) / totalWords

			// Ensure minimum duration of 500ms per chunk for readability
			if chunkDuration < 500 && t.Duration >= 500 {
				chunkDuration = 500
			}

			// Ensure last chunk ends at the original end time
			endMs := currentStartMs + chunkDuration
			if i == len(chunks)-1 {
				endMs = t.EndMs
			}

			result = append(result, VoiceoverTiming{
				Index:    len(result) + 1,
				Text:     chunk,
				StartMs:  currentStartMs,
				EndMs:    endMs,
				Duration: endMs - currentStartMs,
			})

			currentStartMs = endMs
		}
	}

	return result
}

// splitTextIntoChunks splits text into chunks where each chunk fits within maxLines
// of maxCharsPerLine. This properly tracks line boundaries based on word wrapping,
// not just character count.
func splitTextIntoChunks(text string, maxCharsPerLine, maxLines int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var chunks []string
	var chunkLines []string  // Lines in current chunk
	var currentLine []string // Words in current line
	currentLineLen := 0

	for _, word := range words {
		wordLen := len(word)

		// Calculate length if we add this word to current line
		newLineLen := currentLineLen + wordLen
		if currentLineLen > 0 {
			newLineLen++ // Account for space
		}

		if newLineLen > maxCharsPerLine && len(currentLine) > 0 {
			// Word doesn't fit on current line - finalize this line
			chunkLines = append(chunkLines, strings.Join(currentLine, " "))

			// Check if we've filled this chunk
			if len(chunkLines) >= maxLines {
				// Save this chunk and start a new one
				chunks = append(chunks, strings.Join(chunkLines, "\n"))
				chunkLines = nil
			}

			// Start new line with current word
			currentLine = []string{word}
			currentLineLen = wordLen
		} else {
			// Word fits on current line
			currentLine = append(currentLine, word)
			currentLineLen = newLineLen
		}
	}

	// Don't forget remaining content
	if len(currentLine) > 0 {
		chunkLines = append(chunkLines, strings.Join(currentLine, " "))
	}
	if len(chunkLines) > 0 {
		chunks = append(chunks, strings.Join(chunkLines, "\n"))
	}

	return chunks
}

// wrapText wraps text to fit within maxCharsPerLine, up to maxLines.
// If text requires more than maxLines, only the first maxLines are returned.
// Use splitTextIntoChunks to get all content split into properly-sized chunks.
func wrapText(text string, maxCharsPerLine, maxLines int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	var currentLine []string
	currentLen := 0

	for _, word := range words {
		wordLen := len(word)

		newLen := currentLen + wordLen
		if currentLen > 0 {
			newLen++ // Account for space
		}

		if newLen > maxCharsPerLine && len(currentLine) > 0 {
			// Finalize current line
			lines = append(lines, strings.Join(currentLine, " "))

			// Stop if we've hit max lines
			if len(lines) >= maxLines {
				return strings.Join(lines, "\n")
			}

			// Start new line
			currentLine = []string{word}
			currentLen = wordLen
		} else {
			currentLine = append(currentLine, word)
			currentLen = newLen
		}
	}

	// Don't forget the last line
	if len(currentLine) > 0 && len(lines) < maxLines {
		lines = append(lines, strings.Join(currentLine, " "))
	}

	return strings.Join(lines, "\n")
}

// generateSRT creates SRT format subtitle content.
func (g *BrowserSubtitleGenerator) generateSRT(timings []VoiceoverTiming) string {
	var sb strings.Builder

	for i, t := range timings {
		// SRT index (1-based)
		sb.WriteString(fmt.Sprintf("%d\n", i+1))

		// Timestamps: HH:MM:SS,mmm --> HH:MM:SS,mmm
		sb.WriteString(fmt.Sprintf("%s --> %s\n",
			formatSRTTime(t.StartMs),
			formatSRTTime(t.EndMs)))

		// Text (can span multiple lines)
		sb.WriteString(t.Text)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// generateVTT creates WebVTT format subtitle content.
func (g *BrowserSubtitleGenerator) generateVTT(timings []VoiceoverTiming) string {
	var sb strings.Builder

	// VTT header
	sb.WriteString("WEBVTT\n\n")

	for i, t := range timings {
		// Optional cue identifier
		sb.WriteString(fmt.Sprintf("%d\n", i+1))

		// Timestamps: HH:MM:SS.mmm --> HH:MM:SS.mmm
		sb.WriteString(fmt.Sprintf("%s --> %s\n",
			formatVTTTime(t.StartMs),
			formatVTTTime(t.EndMs)))

		// Text
		sb.WriteString(t.Text)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// formatSRTTime formats milliseconds as SRT timestamp (HH:MM:SS,mmm).
func formatSRTTime(ms int) string {
	hours := ms / 3600000
	ms %= 3600000
	minutes := ms / 60000
	ms %= 60000
	seconds := ms / 1000
	millis := ms % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, millis)
}

// formatVTTTime formats milliseconds as VTT timestamp (HH:MM:SS.mmm).
func formatVTTTime(ms int) string {
	hours := ms / 3600000
	ms %= 3600000
	minutes := ms / 60000
	ms %= 60000
	seconds := ms / 1000
	millis := ms % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}
