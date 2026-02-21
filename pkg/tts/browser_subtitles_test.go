package tts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSRT(t *testing.T) {
	generator := NewBrowserSubtitleGenerator(FormatSRT)

	timings := []VoiceoverTiming{
		{Index: 1, Text: "Hello world", StartMs: 0, EndMs: 2000, Duration: 2000},
		{Index: 2, Text: "This is a test", StartMs: 2500, EndMs: 5000, Duration: 2500},
		{Index: 3, Text: "Final line", StartMs: 5500, EndMs: 7000, Duration: 1500},
	}

	// Create temp file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.srt")

	err := generator.GenerateFromTimings(timings, outputPath)
	if err != nil {
		t.Fatalf("GenerateFromTimings failed: %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	srt := string(content)

	// Verify SRT format
	if !strings.Contains(srt, "1\n") {
		t.Error("SRT should contain subtitle index 1")
	}
	if !strings.Contains(srt, "00:00:00,000 --> 00:00:02,000") {
		t.Error("SRT should contain first timestamp")
	}
	if !strings.Contains(srt, "Hello world") {
		t.Error("SRT should contain first subtitle text")
	}
	if !strings.Contains(srt, "00:00:02,500 --> 00:00:05,000") {
		t.Error("SRT should contain second timestamp")
	}
	if !strings.Contains(srt, "This is a test") {
		t.Error("SRT should contain second subtitle text")
	}
}

func TestGenerateVTT(t *testing.T) {
	generator := NewBrowserSubtitleGenerator(FormatVTT)

	timings := []VoiceoverTiming{
		{Index: 1, Text: "Hello world", StartMs: 0, EndMs: 2000, Duration: 2000},
		{Index: 2, Text: "This is a test", StartMs: 2500, EndMs: 5000, Duration: 2500},
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.vtt")

	err := generator.GenerateFromTimings(timings, outputPath)
	if err != nil {
		t.Fatalf("GenerateFromTimings failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	vtt := string(content)

	// Verify VTT format
	if !strings.HasPrefix(vtt, "WEBVTT") {
		t.Error("VTT should start with WEBVTT header")
	}
	// VTT uses period instead of comma for milliseconds
	if !strings.Contains(vtt, "00:00:00.000 --> 00:00:02.000") {
		t.Error("VTT should contain first timestamp with period separator")
	}
	if !strings.Contains(vtt, "Hello world") {
		t.Error("VTT should contain first subtitle text")
	}
}

func TestFormatSRTTime(t *testing.T) {
	tests := []struct {
		ms       int
		expected string
	}{
		{0, "00:00:00,000"},
		{1000, "00:00:01,000"},
		{60000, "00:01:00,000"},
		{3600000, "01:00:00,000"},
		{3661500, "01:01:01,500"},
		{90061234, "25:01:01,234"},
	}

	for _, tc := range tests {
		result := formatSRTTime(tc.ms)
		if result != tc.expected {
			t.Errorf("formatSRTTime(%d) = %q, want %q", tc.ms, result, tc.expected)
		}
	}
}

func TestFormatVTTTime(t *testing.T) {
	tests := []struct {
		ms       int
		expected string
	}{
		{0, "00:00:00.000"},
		{1000, "00:00:01.000"},
		{60000, "00:01:00.000"},
		{3600000, "01:00:00.000"},
		{3661500, "01:01:01.500"},
	}

	for _, tc := range tests {
		result := formatVTTTime(tc.ms)
		if result != tc.expected {
			t.Errorf("formatVTTTime(%d) = %q, want %q", tc.ms, result, tc.expected)
		}
	}
}

func TestDefaultFormat(t *testing.T) {
	// Empty format should default to SRT
	generator := NewBrowserSubtitleGenerator("")
	if generator.format != FormatSRT {
		t.Errorf("Default format should be SRT, got %q", generator.format)
	}
}

func TestWrapTextRespectsLineLimits(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		maxChars       int
		maxLines       int
		expectAllWords bool // Whether all words should be preserved
	}{
		{
			name:           "short text fits completely",
			text:           "Hello world",
			maxChars:       42,
			maxLines:       2,
			expectAllWords: true,
		},
		{
			name:           "text fits in two lines",
			text:           "This is a test of subtitle wrapping",
			maxChars:       42,
			maxLines:       2,
			expectAllWords: true,
		},
		{
			name:           "long text is truncated",
			text:           "This is a very long piece of text that will definitely need more than two lines to display properly",
			maxChars:       42,
			maxLines:       2,
			expectAllWords: false, // wrapText truncates; use splitTextIntoChunks for full content
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := wrapText(tc.text, tc.maxChars, tc.maxLines)

			// Verify max lines not exceeded
			lines := strings.Split(result, "\n")
			if len(lines) > tc.maxLines {
				t.Errorf("Result has %d lines, expected max %d\nResult: %q", len(lines), tc.maxLines, result)
			}

			// Verify each line respects character limit
			for i, line := range lines {
				if len(line) > tc.maxChars {
					t.Errorf("Line %d exceeds max chars (%d > %d): %q", i+1, len(line), tc.maxChars, line)
				}
			}

			// Check word preservation based on expected behavior
			if tc.expectAllWords {
				originalWords := strings.Fields(tc.text)
				resultWords := strings.Fields(result)
				if len(resultWords) != len(originalWords) {
					t.Errorf("Word count mismatch: original has %d words, result has %d words\nOriginal: %q\nResult: %q",
						len(originalWords), len(resultWords), tc.text, result)
				}
			}
		})
	}
}

func TestSplitTextIntoChunksPreservesAllWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxChars int
		maxLines int
	}{
		{
			name:     "short text no split needed",
			text:     "Hello world",
			maxChars: 42,
			maxLines: 2,
		},
		{
			name:     "medium text no split",
			text:     "This is a test of the subtitle wrapping",
			maxChars: 42,
			maxLines: 2,
		},
		{
			name:     "long text needs multiple chunks",
			text:     "This is a very long piece of text that will definitely need to be split into multiple chunks because it contains way too many words to fit into just two lines of forty two characters each and we absolutely must ensure that every single word is preserved in the output",
			maxChars: 42,
			maxLines: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks := splitTextIntoChunks(tc.text, tc.maxChars, tc.maxLines)

			// Combine all chunks and count words
			allChunkText := strings.Join(chunks, " ")
			originalWords := strings.Fields(tc.text)
			resultWords := strings.Fields(allChunkText)

			if len(resultWords) != len(originalWords) {
				t.Errorf("Word count mismatch: original has %d words, chunks have %d words\nOriginal: %q\nChunks: %v",
					len(originalWords), len(resultWords), tc.text, chunks)
			}

			// Verify all original words are present
			for _, word := range originalWords {
				found := false
				for _, chunk := range chunks {
					if strings.Contains(chunk, word) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Word %q is missing from chunks\nOriginal: %q\nChunks: %v", word, tc.text, chunks)
				}
			}
		})
	}
}

func TestSplitTimingsPreservesAllWords(t *testing.T) {
	// Test that the full pipeline preserves all words
	timings := []VoiceoverTiming{
		{
			Index:    1,
			Text:     "This is a very long voiceover text that contains many words and will need to be split into multiple subtitle entries to display properly on screen without taking up too much space",
			StartMs:  0,
			EndMs:    10000,
			Duration: 10000,
		},
	}

	result := splitTimingsIntoChunks(timings)

	// Combine all text from result
	var allText []string
	for _, t := range result {
		allText = append(allText, t.Text)
	}
	combinedText := strings.Join(allText, " ")

	originalWords := strings.Fields(timings[0].Text)
	resultWords := strings.Fields(combinedText)

	if len(resultWords) != len(originalWords) {
		t.Errorf("Word count mismatch: original has %d words, result has %d words",
			len(originalWords), len(resultWords))
	}

	// Verify timing continuity
	for i := 1; i < len(result); i++ {
		if result[i].StartMs != result[i-1].EndMs {
			t.Errorf("Timing gap between chunk %d (ends %d) and chunk %d (starts %d)",
				i-1, result[i-1].EndMs, i, result[i].StartMs)
		}
	}

	// Verify first and last timing
	if result[0].StartMs != timings[0].StartMs {
		t.Errorf("First chunk should start at %d, got %d", timings[0].StartMs, result[0].StartMs)
	}
	if result[len(result)-1].EndMs != timings[0].EndMs {
		t.Errorf("Last chunk should end at %d, got %d", timings[0].EndMs, result[len(result)-1].EndMs)
	}
}
