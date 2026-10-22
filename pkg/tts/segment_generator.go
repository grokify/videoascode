package tts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/grokify/videoascode/pkg/media"
	omnitts "github.com/grokify/videoascode/pkg/omnivoice/tts"
	"github.com/grokify/videoascode/pkg/segment"
	"github.com/grokify/videoascode/pkg/transcript"
)

// segmentMetadata records a segment's assembled audio so a later run can
// reuse it. Fingerprint identifies the clips and pauses it was built from.
type segmentMetadata struct {
	SegmentID          string         `json:"segmentId"`
	Fingerprint        string         `json:"fingerprint,omitempty"`
	TotalDuration      int            `json:"totalDuration"`
	VoiceoverDurations map[int]int    `json:"voiceoverDurations"`
	VoiceoverFiles     map[int]string `json:"voiceoverFiles"`
}

// ClipDir is the subdirectory of a language's audio directory that holds
// voiceover clips.
const ClipDir = "clips"

// ClipKey names the audio clip for one voiceover. It is a hash of the text
// sent to the TTS provider (after pronunciation substitutions), the
// language, and the voice settings, so a clip is reused only when all of
// them are unchanged: editing narration, a pronunciation, or the voice
// produces a new clip, and unchanged voiceovers are never synthesized twice.
func ClipKey(spokenText, language string, voice transcript.VoiceConfig) string {
	b, err := json.Marshal(struct {
		Text     string                 `json:"text"`
		Language string                 `json:"language"`
		Voice    transcript.VoiceConfig `json:"voice"`
	}{spokenText, language, voice})
	if err != nil {
		panic(fmt.Sprintf("marshal clip key: %v", err)) // plain strings and numbers always marshal
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

// SegmentTTSGenerator generates TTS audio for segments.
// It works with the segment.Segment interface, making it usable
// for both slides and browser demos.
type SegmentTTSGenerator struct {
	provider     *omnitts.Provider
	defaultVoice transcript.VoiceConfig
	progressFunc ProgressFunc
	pronouncer   *Pronouncer

	// synth replaces the provider when set; tests use it to count calls.
	synth func(ctx context.Context, text string, voice transcript.VoiceConfig) ([]byte, error)
}

// NewSegmentTTSGenerator creates a new segment-aware TTS generator.
func NewSegmentTTSGenerator(provider *omnitts.Provider, defaultVoice transcript.VoiceConfig) *SegmentTTSGenerator {
	return &SegmentTTSGenerator{
		provider:     provider,
		defaultVoice: defaultVoice,
	}
}

// SetProgressFunc sets a callback for progress updates during TTS generation.
func (g *SegmentTTSGenerator) SetProgressFunc(fn ProgressFunc) {
	g.progressFunc = fn
}

// SetPronunciations configures TTS pronunciation substitutions (see
// PronunciationDictionary). Applied only to the text sent to the TTS
// provider; voiceover/subtitle/manifest text elsewhere is unaffected. A nil
// dict clears any previously set pronunciations.
func (g *SegmentTTSGenerator) SetPronunciations(dict *PronunciationDictionary) {
	g.pronouncer = NewPronouncer(dict)
}

// reportProgress calls the progress function if set.
func (g *SegmentTTSGenerator) reportProgress(current, total int, desc string) {
	if g.progressFunc != nil {
		g.progressFunc(current, total, desc)
	}
}

// SegmentAudioResult contains TTS results for a segment.
type SegmentAudioResult struct {
	// SegmentID matches segment.GetID()
	SegmentID string

	// AudioFile is the path to the combined audio file for this segment
	AudioFile string

	// VoiceoverAudioFiles maps voiceover index to individual audio file
	VoiceoverAudioFiles map[int]string

	// Duration is the total audio duration in milliseconds
	Duration int

	// VoiceoverDurations maps voiceover index to duration in milliseconds
	VoiceoverDurations map[int]int
}

// GenerateForSegment generates TTS audio for a single segment.
//
// Each voiceover is synthesized once into {outputDir}/clips/{ClipKey}.mp3
// and reused on later runs while its text and voice are unchanged. The
// segment's clips and pauses are assembled into {outputDir}/{segmentID}.mp3,
// which is rebuilt whenever any of them changes.
func (g *SegmentTTSGenerator) GenerateForSegment(
	ctx context.Context,
	seg segment.Segment,
	language string,
	outputDir string,
) (*SegmentAudioResult, error) {
	return g.generate(ctx, seg, language, outputDir, func(int, string) {})
}

// generate implements GenerateForSegment. tick is called with the number of
// voiceovers just handled and a description, for progress reporting.
func (g *SegmentTTSGenerator) generate(
	ctx context.Context,
	seg segment.Segment,
	language string,
	outputDir string,
	tick func(n int, desc string),
) (*SegmentAudioResult, error) {
	voiceovers := seg.GetVoiceovers(language)
	if len(voiceovers) == 0 {
		return nil, fmt.Errorf("segment %s has no voiceovers for language %s", seg.GetID(), language)
	}

	clipDir := filepath.Join(outputDir, ClipDir)
	if err := os.MkdirAll(clipDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Work out each voiceover's spoken text, voice, and clip before touching
	// the provider, so the segment's fingerprint is known up front.
	type clip struct {
		vo     segment.Voiceover
		spoken string
		voice  transcript.VoiceConfig
		path   string
	}
	clips := make([]clip, len(voiceovers))
	fp := sha256.New()
	for i, vo := range voiceovers {
		voice := g.defaultVoice
		if vo.Voice != nil {
			voice = *vo.Voice
		}
		// Pronunciation substitutions affect only what's sent to TTS;
		// vo.Text itself (used for subtitles/manifests) is untouched.
		spoken := g.pronouncer.Apply(vo.Text, language)
		key := ClipKey(spoken, language, voice)
		clips[i] = clip{vo: vo, spoken: spoken, voice: voice, path: filepath.Join(clipDir, key+".mp3")}
		fmt.Fprintf(fp, "%d:%s:%d\n", vo.Index, key, vo.Pause)
	}
	fingerprint := hex.EncodeToString(fp.Sum(nil)[:8])

	combinedPath := filepath.Join(outputDir, seg.GetID()+".mp3")
	metadataPath := filepath.Join(outputDir, seg.GetID()+".json")

	if meta, err := loadSegmentMetadata(metadataPath); err == nil && meta.Fingerprint == fingerprint {
		audio := combinedPath
		if len(clips) == 1 {
			audio = clips[0].path
		}
		if info, err := os.Stat(audio); err == nil && info.Size() > 0 {
			tick(len(clips), seg.GetID()+" (cached)")
			return &SegmentAudioResult{
				SegmentID:           seg.GetID(),
				AudioFile:           audio,
				Duration:            meta.TotalDuration,
				VoiceoverDurations:  meta.VoiceoverDurations,
				VoiceoverAudioFiles: meta.VoiceoverFiles,
			}, nil
		}
	}

	result := &SegmentAudioResult{
		SegmentID:           seg.GetID(),
		VoiceoverAudioFiles: make(map[int]string),
		VoiceoverDurations:  make(map[int]int),
	}
	for _, c := range clips {
		tick(1, fmt.Sprintf("%s vo%d", seg.GetID(), c.vo.Index))
		duration, err := g.ensureClip(ctx, c.path, c.spoken, c.voice)
		if err != nil {
			return nil, fmt.Errorf("voiceover %d: %w", c.vo.Index, err)
		}
		result.VoiceoverAudioFiles[c.vo.Index] = c.path
		result.VoiceoverDurations[c.vo.Index] = duration
		result.Duration += duration + c.vo.Pause
	}

	if len(clips) == 1 {
		result.AudioFile = clips[0].path
	} else {
		if err := g.combineAudioFiles(ctx, result.VoiceoverAudioFiles, voiceovers, combinedPath); err != nil {
			return nil, fmt.Errorf("failed to combine audio files: %w", err)
		}
		result.AudioFile = combinedPath
	}

	meta := segmentMetadata{
		SegmentID:          seg.GetID(),
		Fingerprint:        fingerprint,
		TotalDuration:      result.Duration,
		VoiceoverDurations: result.VoiceoverDurations,
		VoiceoverFiles:     result.VoiceoverAudioFiles,
	}
	if err := saveSegmentMetadata(metadataPath, meta); err != nil {
		// The audio is complete; without metadata the next run reassembles
		// it from the cached clips instead of reusing it.
		fmt.Printf("Warning: failed to save audio metadata: %v\n", err)
	}
	return result, nil
}

// ensureClip returns the duration of the clip at path, synthesizing it first
// if it does not exist or cannot be read.
func (g *SegmentTTSGenerator) ensureClip(ctx context.Context, path, spoken string, voice transcript.VoiceConfig) (int, error) {
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		if d, err := media.GetAudioDurationMs(path); err == nil {
			return d, nil
		}
		// Unreadable (for example a partial write): synthesize it again.
	}
	audioData, err := g.synthesize(ctx, spoken, voice)
	if err != nil {
		return 0, fmt.Errorf("TTS failed: %w", err)
	}
	// Write to a temporary name first so an interrupted run never leaves a
	// truncated clip under its final name.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, audioData, 0600); err != nil {
		return 0, fmt.Errorf("failed to write audio file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return 0, fmt.Errorf("failed to write audio file: %w", err)
	}
	d, err := media.GetAudioDurationMs(path)
	if err != nil {
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}
	return d, nil
}

// loadSegmentMetadata loads cached metadata from a JSON file
func loadSegmentMetadata(path string) (*segmentMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta segmentMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// saveSegmentMetadata saves metadata to a JSON file
func saveSegmentMetadata(path string, meta segmentMetadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// GenerateForSegments generates TTS for multiple segments.
func (g *SegmentTTSGenerator) GenerateForSegments(
	ctx context.Context,
	segments []segment.Segment,
	language string,
	outputDir string,
) (map[string]*SegmentAudioResult, error) {
	results := make(map[string]*SegmentAudioResult)

	for _, seg := range segments {
		result, err := g.GenerateForSegment(ctx, seg, language, outputDir)
		if err != nil {
			return nil, fmt.Errorf("segment %s: %w", seg.GetID(), err)
		}
		results[seg.GetID()] = result
	}

	return results, nil
}

// GenerateMultiLanguage generates TTS for all segments in multiple languages.
// Returns a map of language -> segment ID -> result.
func (g *SegmentTTSGenerator) GenerateMultiLanguage(
	ctx context.Context,
	segments []segment.Segment,
	languages []string,
	outputDir string,
) (map[string]map[string]*SegmentAudioResult, error) {
	results := make(map[string]map[string]*SegmentAudioResult)

	// Count total voiceovers for progress reporting
	totalVoiceovers := 0
	for _, seg := range segments {
		for _, lang := range languages {
			totalVoiceovers += len(seg.GetVoiceovers(lang))
		}
	}

	currentVoiceover := 0
	for _, lang := range languages {
		langDir := filepath.Join(outputDir, lang)
		results[lang] = make(map[string]*SegmentAudioResult)

		for _, seg := range segments {
			result, err := g.generateForSegmentWithProgress(ctx, seg, lang, langDir, &currentVoiceover, totalVoiceovers)
			if err != nil {
				return nil, fmt.Errorf("language %s, segment %s: %w", lang, seg.GetID(), err)
			}
			results[lang][seg.GetID()] = result
		}
	}

	return results, nil
}

// generateForSegmentWithProgress generates TTS for a segment and updates progress.
func (g *SegmentTTSGenerator) generateForSegmentWithProgress(
	ctx context.Context,
	seg segment.Segment,
	language string,
	outputDir string,
	currentVoiceover *int,
	totalVoiceovers int,
) (*SegmentAudioResult, error) {
	return g.generate(ctx, seg, language, outputDir, func(n int, desc string) {
		*currentVoiceover += n
		g.reportProgress(*currentVoiceover, totalVoiceovers, desc)
	})
}

// CalculateMaxDurations finds the maximum duration per segment across all languages.
// This is used to pace browser recordings so all languages fit.
func CalculateMaxDurations(results map[string]map[string]*SegmentAudioResult) map[string]int {
	maxDurations := make(map[string]int)

	for _, langResults := range results {
		for segID, result := range langResults {
			if result.Duration > maxDurations[segID] {
				maxDurations[segID] = result.Duration
			}
		}
	}

	return maxDurations
}

// CalculateMaxVoiceoverDurations finds the maximum duration per voiceover across languages.
// This is used for browser steps where each voiceover maps to a step.
func CalculateMaxVoiceoverDurations(results map[string]map[string]*SegmentAudioResult, segmentID string) map[int]int {
	maxDurations := make(map[int]int)

	for _, langResults := range results {
		if result, ok := langResults[segmentID]; ok {
			for idx, duration := range result.VoiceoverDurations {
				if duration > maxDurations[idx] {
					maxDurations[idx] = duration
				}
			}
		}
	}

	return maxDurations
}

// synthesize calls the TTS provider.
func (g *SegmentTTSGenerator) synthesize(ctx context.Context, text string, voice transcript.VoiceConfig) ([]byte, error) {
	if g.synth != nil {
		return g.synth(ctx, text, voice)
	}
	// Use the omnitts.Provider wrapper which handles config conversion
	return g.provider.Synthesize(ctx, text, voice)
}

// combineAudioFiles concatenates multiple audio files with optional pauses.
func (g *SegmentTTSGenerator) combineAudioFiles(
	ctx context.Context,
	audioFiles map[int]string,
	voiceovers []segment.Voiceover,
	outputPath string,
) error {
	// Create a concat file for ffmpeg
	concatPath := outputPath + ".concat.txt"
	defer os.Remove(concatPath)

	var lines []string
	for _, vo := range voiceovers {
		audioPath, ok := audioFiles[vo.Index]
		if !ok {
			continue
		}

		// Use absolute path to avoid ffmpeg concat relative path issues
		absAudioPath, err := filepath.Abs(audioPath)
		if err != nil {
			absAudioPath = audioPath // Fall back to original if Abs fails
		}

		// Add audio file
		lines = append(lines, fmt.Sprintf("file '%s'", absAudioPath))

		// Add silence for pause if specified
		if vo.Pause > 0 {
			// Generate silence file
			silencePath := outputPath + fmt.Sprintf(".silence_%d.mp3", vo.Index)
			if err := generateSilence(ctx, vo.Pause, silencePath); err != nil {
				return fmt.Errorf("failed to generate silence: %w", err)
			}
			defer os.Remove(silencePath)
			// Use absolute path for silence too
			absSilencePath, err := filepath.Abs(silencePath)
			if err != nil {
				absSilencePath = silencePath
			}
			lines = append(lines, fmt.Sprintf("file '%s'", absSilencePath))
		}
	}

	concatContent := ""
	for _, line := range lines {
		concatContent += line + "\n"
	}

	if err := os.WriteFile(concatPath, []byte(concatContent), 0600); err != nil {
		return fmt.Errorf("failed to write concat file: %w", err)
	}

	// Run ffmpeg concat
	args := []string{
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", concatPath,
		"-c:a", "libmp3lame",
		"-q:a", "2",
		outputPath,
	}

	if output, err := runFFmpeg(ctx, args...); err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// generateSilence creates a silent audio file of the specified duration.
func generateSilence(ctx context.Context, durationMs int, outputPath string) error {
	durationSec := float64(durationMs) / 1000.0

	args := []string{
		"-y",
		"-f", "lavfi",
		"-i", fmt.Sprintf("anullsrc=r=44100:cl=stereo:d=%.3f", durationSec),
		"-c:a", "libmp3lame",
		"-q:a", "2",
		outputPath,
	}

	if output, err := runFFmpeg(ctx, args...); err != nil {
		return fmt.Errorf("ffmpeg silence generation failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// runFFmpeg executes an ffmpeg command.
func runFFmpeg(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	return cmd.CombinedOutput()
}
