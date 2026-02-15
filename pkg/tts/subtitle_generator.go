package tts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	omnistt "github.com/grokify/marp2video/pkg/omnivoice/stt"
	"github.com/grokify/marp2video/pkg/transcript"
	"github.com/grokify/mogo/log/slogutil"
)

// SubtitleFormat specifies the subtitle output format.
type SubtitleFormat string

const (
	FormatSRT SubtitleFormat = "srt"
	FormatVTT SubtitleFormat = "vtt"
)

// SubtitleGeneratorConfig holds configuration for subtitle generation.
type SubtitleGeneratorConfig struct {
	omnistt.ProviderConfig        // Embedded provider config (ElevenLabsAPIKey, DeepgramAPIKey)
	DefaultProvider        string // "elevenlabs" or "deepgram" (default: deepgram for STT)
	OutputDir              string // Directory for subtitle files (SRT/VTT output)
	AudioDir               string // Audio directory (for timestamps.json)
	Format                 SubtitleFormat
	Language               string // Language hint for transcription
	ProgressFunc           ProgressFunc
	UseOriginalText        bool                   // Use original transcript text with STT timestamps
	OriginalTranscript     *transcript.Transcript // Original transcript for text alignment
	SaveTimestamps         bool                   // Save raw STT timestamps to JSON for reuse
	TimestampsFile         string                 // Load timestamps from file instead of calling STT
	DictionaryPaths        []string               // Additional dictionary files for case correction
	NoBuiltInDictionary    bool                   // Disable built-in dictionary corrections
	CaseCorrector          *CaseCorrector         // Pre-loaded case corrector (optional)
}

// SubtitleGenerator generates subtitles from audio files.
type SubtitleGenerator struct {
	config        SubtitleGeneratorConfig
	factory       *omnistt.Factory
	provider      *omnistt.Provider
	caseCorrector *CaseCorrector
}

// NewSubtitleGenerator creates a new subtitle generator.
func NewSubtitleGenerator(config SubtitleGeneratorConfig) (*SubtitleGenerator, error) {
	var factory *omnistt.Factory
	var provider *omnistt.Provider

	// Only create STT provider if not loading from timestamps file
	if config.TimestampsFile == "" {
		factory = omnistt.NewFactory(config.ProviderConfig)

		if config.DefaultProvider != "" {
			factory.SetFallback(config.DefaultProvider)
		}

		var err error
		provider, err = factory.Get("")
		if err != nil {
			return nil, fmt.Errorf("failed to create STT provider: %w", err)
		}
	}

	// Set defaults
	if config.Format == "" {
		config.Format = FormatSRT
	}

	// Load case corrector
	var caseCorrector *CaseCorrector
	if config.CaseCorrector != nil {
		caseCorrector = config.CaseCorrector
	} else {
		// Load dictionaries
		loader := NewDictionaryLoader().
			WithBuiltIn(!config.NoBuiltInDictionary).
			WithAdditionalPaths(config.DictionaryPaths)

		dict, err := loader.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load dictionaries: %w", err)
		}
		caseCorrector = NewCaseCorrector(dict)
	}

	return &SubtitleGenerator{
		config:        config,
		factory:       factory,
		provider:      provider,
		caseCorrector: caseCorrector,
	}, nil
}

// getOriginalTextForSlide retrieves the original transcript text for a slide.
func (g *SubtitleGenerator) getOriginalTextForSlide(slideIndex int) string {
	if g.config.OriginalTranscript == nil {
		return ""
	}

	content, err := g.config.OriginalTranscript.GetSlideTranscript(slideIndex, g.config.Language)
	if err != nil {
		return ""
	}

	return content.GetFullText()
}

// applyCaseCorrection applies dictionary-based case correction to a transcription result.
func (g *SubtitleGenerator) applyCaseCorrection(result *omnistt.TranscriptionResult) *omnistt.TranscriptionResult {
	if g.caseCorrector == nil {
		return result
	}

	// Create a copy to avoid modifying the original
	corrected := &omnistt.TranscriptionResult{
		Text:     g.caseCorrector.Correct(result.Text),
		Language: result.Language,
		Duration: result.Duration,
		Segments: make([]omnistt.Segment, len(result.Segments)),
	}

	for i, seg := range result.Segments {
		corrected.Segments[i] = omnistt.Segment{
			Text:       g.caseCorrector.Correct(seg.Text),
			StartTime:  seg.StartTime,
			EndTime:    seg.EndTime,
			Confidence: seg.Confidence,
			Speaker:    seg.Speaker,
			Words:      make([]omnistt.Word, len(seg.Words)),
		}

		// Correct individual words
		for j, word := range seg.Words {
			corrected.Segments[i].Words[j] = omnistt.Word{
				Text:       g.caseCorrector.CorrectWord(word.Text),
				StartTime:  word.StartTime,
				EndTime:    word.EndTime,
				Confidence: word.Confidence,
				Speaker:    word.Speaker,
			}
		}
	}

	return corrected
}

// SubtitleResult contains generated subtitle information.
type SubtitleResult struct {
	SlideIndex   int
	SubtitleFile string
	WordCount    int
	Duration     time.Duration
}

// SubtitleManifest contains information about generated subtitles.
type SubtitleManifest struct {
	Language    string           `json:"language"`
	Format      string           `json:"format"`
	GeneratedAt time.Time        `json:"generatedAt"`
	Subtitles   []SubtitleResult `json:"subtitles"`
	CombinedSRT string           `json:"combinedSrt,omitempty"`
	CombinedVTT string           `json:"combinedVtt,omitempty"`
}

// TimestampsData stores raw STT timestamps for later reuse.
type TimestampsData struct {
	Version     string                    `json:"version"`
	Language    string                    `json:"language"`
	GeneratedAt time.Time                 `json:"generatedAt"`
	Provider    string                    `json:"provider"`
	Slides      []SlideTimestamps         `json:"slides"`
}

// SlideTimestamps stores timestamps for a single slide.
type SlideTimestamps struct {
	Index      int                        `json:"index"`
	AudioFile  string                     `json:"audioFile"`
	Duration   time.Duration              `json:"durationNs"`
	Result     *omnistt.TranscriptionResult `json:"result"`
}

// SaveTimestampsData saves timestamps to a JSON file.
func SaveTimestampsData(data *TimestampsData, path string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal timestamps: %w", err)
	}
	if err := os.WriteFile(path, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write timestamps file: %w", err)
	}
	return nil
}

// LoadTimestampsData loads timestamps from a JSON file.
func LoadTimestampsData(path string) (*TimestampsData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read timestamps file: %w", err)
	}
	var timestamps TimestampsData
	if err := json.Unmarshal(data, &timestamps); err != nil {
		return nil, fmt.Errorf("failed to parse timestamps JSON: %w", err)
	}
	return &timestamps, nil
}

// GenerateFromManifest generates subtitles for all audio files in a TTS manifest.
func (g *SubtitleGenerator) GenerateFromManifest(ctx context.Context, manifest *Manifest, audioDir string) (*SubtitleManifest, error) {
	logger := slogutil.LoggerFromContext(ctx, slogutil.Null())

	// Create output directory
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	result := &SubtitleManifest{
		Language:    g.config.Language,
		Format:      string(g.config.Format),
		GeneratedAt: time.Now().UTC(),
	}

	numSlides := len(manifest.Slides)
	var allResults []*omnistt.TranscriptionResult
	var slideOffsets []time.Duration

	// Track cumulative offset for combined subtitle
	var cumulativeOffset time.Duration

	// Load pre-saved timestamps if provided or auto-detect in audio directory
	var savedTimestamps *TimestampsData
	timestampsFile := g.config.TimestampsFile

	// Auto-detect timestamps.json in audio directory if not explicitly provided
	if timestampsFile == "" && audioDir != "" {
		autoPath := filepath.Join(audioDir, "timestamps.json")
		if _, err := os.Stat(autoPath); err == nil {
			timestampsFile = autoPath
			logger.Info("auto-detected timestamps file", "file", autoPath)
		}
	}

	if timestampsFile != "" {
		var err error
		savedTimestamps, err = LoadTimestampsData(timestampsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load timestamps: %w", err)
		}
		logger.Info("loaded timestamps from file",
			"file", timestampsFile,
			"slides", len(savedTimestamps.Slides))
	}

	// Prepare timestamps data for saving
	var timestampsToSave *TimestampsData
	if g.config.SaveTimestamps {
		providerName := ""
		if g.provider != nil {
			providerName = g.provider.Name()
		}
		timestampsToSave = &TimestampsData{
			Version:     "1.0",
			Language:    g.config.Language,
			GeneratedAt: time.Now().UTC(),
			Provider:    providerName,
			Slides:      make([]SlideTimestamps, 0, numSlides),
		}
	}

	// Process each slide
	for i, slide := range manifest.Slides {
		// Report progress
		if g.config.ProgressFunc != nil {
			g.config.ProgressFunc(i+1, numSlides, slide.Title)
		}

		audioPath := filepath.Join(audioDir, slide.AudioFile)

		// Check if audio file exists (unless using saved timestamps)
		if savedTimestamps == nil {
			if _, err := os.Stat(audioPath); os.IsNotExist(err) {
				logger.Warn("skipping slide without audio file",
					"index", slide.Index,
					"file", audioPath)
				continue
			}
		}

		var transcription *omnistt.TranscriptionResult

		// Use saved timestamps or call STT
		if savedTimestamps != nil {
			// Find timestamps for this slide
			found := false
			for _, st := range savedTimestamps.Slides {
				if st.Index == slide.Index {
					transcription = st.Result
					found = true
					logger.Info("using saved timestamps",
						"slide", slide.Index)
					break
				}
			}
			if !found {
				logger.Warn("no saved timestamps for slide",
					"index", slide.Index)
				continue
			}
		} else {
			logger.Info("transcribing audio",
				"slide", slide.Index,
				"file", audioPath,
				"provider", g.provider.Name())

			// Transcribe audio
			transcriptionConfig := omnistt.TranscriptionConfig{
				Language:             g.config.Language,
				EnablePunctuation:    true,
				EnableWordTimestamps: true,
			}

			var err error
			transcription, err = g.provider.TranscribeFile(ctx, audioPath, transcriptionConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to transcribe slide %d: %w", slide.Index, err)
			}

			// Save raw timestamps if enabled
			if timestampsToSave != nil {
				timestampsToSave.Slides = append(timestampsToSave.Slides, SlideTimestamps{
					Index:     slide.Index,
					AudioFile: slide.AudioFile,
					Duration:  transcription.Duration,
					Result:    transcription,
				})
			}
		}

		// Align with original text if enabled
		if g.config.UseOriginalText && g.config.OriginalTranscript != nil {
			originalText := g.getOriginalTextForSlide(slide.Index)
			if originalText != "" {
				aligned := alignTranscriptionWithOriginal(transcription, originalText)
				if aligned != transcription {
					logger.Info("aligned transcription with original text",
						"slide", slide.Index,
						"originalWords", len(splitIntoWords(originalText)))
					transcription = aligned
				} else {
					logger.Warn("alignment failed, using STT transcription",
						"slide", slide.Index)
				}
			}
		}

		// Apply dictionary-based case correction
		if g.caseCorrector != nil {
			transcription = g.applyCaseCorrection(transcription)
		}

		// Store for combined subtitle generation
		allResults = append(allResults, transcription)
		slideOffsets = append(slideOffsets, cumulativeOffset)

		// Generate individual subtitle file
		subtitlePath := filepath.Join(g.config.OutputDir, fmt.Sprintf("slide_%03d.%s", slide.Index, g.config.Format))

		opts := omnistt.DefaultSubtitleOptions()
		var subtitleContent string
		if g.config.Format == FormatVTT {
			subtitleContent = omnistt.GenerateVTTFromResult(transcription, opts)
		} else {
			subtitleContent = omnistt.GenerateSRTFromResult(transcription, opts)
		}

		if err := os.WriteFile(subtitlePath, []byte(subtitleContent), 0600); err != nil {
			return nil, fmt.Errorf("failed to write subtitle file for slide %d: %w", slide.Index, err)
		}

		// Count words
		wordCount := 0
		for _, seg := range transcription.Segments {
			wordCount += len(seg.Words)
		}

		result.Subtitles = append(result.Subtitles, SubtitleResult{
			SlideIndex:   slide.Index,
			SubtitleFile: filepath.Base(subtitlePath),
			WordCount:    wordCount,
			Duration:     transcription.Duration,
		})

		// Update cumulative offset (include pause duration)
		cumulativeOffset += time.Duration(slide.TotalDuration) * time.Millisecond

		logger.Info("generated subtitle",
			"slide", slide.Index,
			"words", wordCount,
			"duration", transcription.Duration)
	}

	// Generate combined subtitle file
	if len(allResults) > 0 {
		combinedSRT, combinedVTT := g.generateCombinedSubtitles(allResults, slideOffsets)

		// Save combined SRT
		combinedSRTPath := filepath.Join(g.config.OutputDir, "combined.srt")
		if err := os.WriteFile(combinedSRTPath, []byte(combinedSRT), 0600); err != nil {
			logger.Warn("failed to write combined SRT", "error", err)
		} else {
			result.CombinedSRT = "combined.srt"
		}

		// Save combined VTT
		combinedVTTPath := filepath.Join(g.config.OutputDir, "combined.vtt")
		if err := os.WriteFile(combinedVTTPath, []byte(combinedVTT), 0600); err != nil {
			logger.Warn("failed to write combined VTT", "error", err)
		} else {
			result.CombinedVTT = "combined.vtt"
		}
	}

	// Save timestamps to audio directory if enabled
	if timestampsToSave != nil && len(timestampsToSave.Slides) > 0 && g.config.AudioDir != "" {
		timestampsPath := filepath.Join(g.config.AudioDir, "timestamps.json")
		if err := SaveTimestampsData(timestampsToSave, timestampsPath); err != nil {
			logger.Warn("failed to save timestamps", "error", err)
		} else {
			logger.Info("saved timestamps",
				"file", timestampsPath,
				"slides", len(timestampsToSave.Slides))
		}
	}

	return result, nil
}

// generateCombinedSubtitles creates combined SRT and VTT with adjusted timestamps.
func (g *SubtitleGenerator) generateCombinedSubtitles(results []*omnistt.TranscriptionResult, offsets []time.Duration) (string, string) {
	// Combine all results with adjusted timestamps
	combined := &omnistt.TranscriptionResult{}

	for i, result := range results {
		offset := offsets[i]

		for _, seg := range result.Segments {
			adjustedSeg := omnistt.Segment{
				Text:       seg.Text,
				StartTime:  seg.StartTime + offset,
				EndTime:    seg.EndTime + offset,
				Confidence: seg.Confidence,
				Speaker:    seg.Speaker,
			}

			for _, w := range seg.Words {
				adjustedSeg.Words = append(adjustedSeg.Words, omnistt.Word{
					Text:       w.Text,
					StartTime:  w.StartTime + offset,
					EndTime:    w.EndTime + offset,
					Confidence: w.Confidence,
					Speaker:    w.Speaker,
				})
			}

			combined.Segments = append(combined.Segments, adjustedSeg)
		}
	}

	opts := omnistt.DefaultSubtitleOptions()
	srt := omnistt.GenerateSRTFromResult(combined, opts)
	vtt := omnistt.GenerateVTTFromResult(combined, opts)

	return srt, vtt
}

// alignTranscriptionWithOriginal replaces STT transcription text with original text
// while preserving STT word-level timestamps. This preserves proper capitalization
// (AI, I, Claude Code, etc.) from the original transcript.
func alignTranscriptionWithOriginal(result *omnistt.TranscriptionResult, originalText string) *omnistt.TranscriptionResult {
	// Extract all words with timestamps from STT result
	var sttWords []omnistt.Word
	for _, seg := range result.Segments {
		sttWords = append(sttWords, seg.Words...)
	}

	// Split original text into words
	originalWords := splitIntoWords(originalText)

	// If word counts match, do direct positional mapping
	if len(originalWords) == len(sttWords) {
		return createAlignedResult(originalWords, sttWords, result.Duration, result.Language)
	}

	// Word counts don't match - try fuzzy alignment
	alignedWords := fuzzyAlignWords(originalWords, sttWords)
	if len(alignedWords) > 0 {
		return createAlignedResult(alignedWords, sttWords[:len(alignedWords)], result.Duration, result.Language)
	}

	// Alignment failed - return original STT result
	return result
}

// splitIntoWords splits text into words, preserving punctuation attached to words.
func splitIntoWords(text string) []string {
	var words []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// stripPunctuation removes leading/trailing punctuation for word comparison.
func stripPunctuation(word string) string {
	runes := []rune(word)
	start := 0
	end := len(runes)

	// Strip leading punctuation
	for start < end && !unicode.IsLetter(runes[start]) && !unicode.IsDigit(runes[start]) {
		start++
	}

	// Strip trailing punctuation
	for end > start && !unicode.IsLetter(runes[end-1]) && !unicode.IsDigit(runes[end-1]) {
		end--
	}

	if start >= end {
		return word // Return original if all punctuation
	}

	return string(runes[start:end])
}

// wordsMatch compares two words, ignoring case and punctuation.
func wordsMatch(original, stt string) bool {
	origClean := strings.ToLower(stripPunctuation(original))
	sttClean := strings.ToLower(stripPunctuation(stt))
	return origClean == sttClean
}

// fuzzyAlignWords attempts to align original words with STT words,
// handling minor mismatches like contractions or number formatting.
func fuzzyAlignWords(originalWords []string, sttWords []omnistt.Word) []string {
	result := make([]string, 0, len(sttWords))
	origIdx := 0

	for sttIdx := 0; sttIdx < len(sttWords) && origIdx < len(originalWords); sttIdx++ {
		sttWord := sttWords[sttIdx].Text

		// Direct match
		if wordsMatch(originalWords[origIdx], sttWord) {
			// Preserve original word with its punctuation
			result = append(result, originalWords[origIdx])
			origIdx++
			continue
		}

		// Check if STT split a word (e.g., "A.I." -> "A" "I")
		// Look ahead in STT to see if combined matches original
		if sttIdx+1 < len(sttWords) {
			combined := sttWord + sttWords[sttIdx+1].Text
			if wordsMatch(originalWords[origIdx], combined) {
				// Use first half timing, skip next STT word
				result = append(result, originalWords[origIdx])
				origIdx++
				sttIdx++ // Skip next STT word
				continue
			}
		}

		// Check if original has a word that STT combined
		// (e.g., original "do not" but STT heard "don't")
		if origIdx+1 < len(originalWords) {
			combined := originalWords[origIdx] + originalWords[origIdx+1]
			if wordsMatch(combined, sttWord) {
				// Use original combined
				result = append(result, originalWords[origIdx]+" "+originalWords[origIdx+1])
				origIdx += 2
				continue
			}
		}

		// No match found - use STT word but this is a signal alignment may fail
		result = append(result, sttWord)
		origIdx++
	}

	// If we used all original words and matched all STT words, success
	if origIdx == len(originalWords) && len(result) == len(sttWords) {
		return result
	}

	// Alignment was incomplete
	return nil
}

// createAlignedResult creates a new TranscriptionResult with original words and STT timestamps.
func createAlignedResult(originalWords []string, sttWords []omnistt.Word, duration time.Duration, language string) *omnistt.TranscriptionResult {
	// Create a single segment with all aligned words
	words := make([]omnistt.Word, len(sttWords))
	var fullText strings.Builder

	for i, sttWord := range sttWords {
		if i > 0 {
			fullText.WriteString(" ")
		}
		originalWord := originalWords[i]
		fullText.WriteString(originalWord)

		words[i] = omnistt.Word{
			Text:       originalWord,
			StartTime:  sttWord.StartTime,
			EndTime:    sttWord.EndTime,
			Confidence: sttWord.Confidence,
			Speaker:    sttWord.Speaker,
		}
	}

	// Determine segment start/end from words
	var startTime, endTime time.Duration
	if len(words) > 0 {
		startTime = words[0].StartTime
		endTime = words[len(words)-1].EndTime
	}

	return &omnistt.TranscriptionResult{
		Text:     fullText.String(),
		Language: language,
		Duration: duration,
		Segments: []omnistt.Segment{
			{
				Text:       fullText.String(),
				StartTime:  startTime,
				EndTime:    endTime,
				Confidence: averageConfidence(words),
				Words:      words,
			},
		},
	}
}

// averageConfidence calculates the average confidence across words.
func averageConfidence(words []omnistt.Word) float64 {
	if len(words) == 0 {
		return 0
	}
	var sum float64
	for _, w := range words {
		sum += w.Confidence
	}
	return sum / float64(len(words))
}
