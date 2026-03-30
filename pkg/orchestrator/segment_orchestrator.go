package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grokify/videoascode/pkg/segment"
	"github.com/grokify/videoascode/pkg/source"
	"github.com/grokify/videoascode/pkg/tts"
	"github.com/grokify/videoascode/pkg/video"

	"github.com/grokify/mogo/fmt/progress"
)

// SegmentConfig holds configuration for the segment-based orchestrator.
type SegmentConfig struct {
	// Source is the content source (transcript, config, etc.)
	Source source.ContentSource

	// OutputFile is the final video output path
	OutputFile string

	// WorkDir is where temporary files are stored
	WorkDir string

	// AudioOutputDir is where to save audio tracks (if set, audio is preserved)
	// Structure: {AudioOutputDir}/{language}/combined.mp3
	AudioOutputDir string

	// Languages is the list of languages to generate (uses default if empty)
	Languages []string

	// Width and Height are video dimensions
	Width  int
	Height int

	// FrameRate is the video frame rate
	FrameRate int

	// TransitionDuration is crossfade duration between segments (seconds)
	TransitionDuration float64

	// Headless runs browser in headless mode
	Headless bool

	// ProgressWriter receives progress updates
	ProgressWriter io.Writer

	// Subtitles enables simple subtitle generation from voiceover timing
	Subtitles bool

	// SubtitlesSTT enables word-level subtitles using speech-to-text
	SubtitlesSTT bool

	// SubtitlesBurn burns subtitles into video (permanent)
	SubtitlesBurn bool

	// SubtitleFormat is "srt" or "vtt" (default: srt)
	SubtitleFormat string

	// DeepgramAPIKey for STT subtitles
	DeepgramAPIKey string

	// NoAudio generates video without audio (TTS still used for timing/subtitles)
	NoAudio bool

	// Parallel is the number of concurrent segment recordings (default 1 = sequential)
	Parallel int

	// FastEncoding uses hardware acceleration for video encoding when available
	FastEncoding bool

	// SegmentLimit limits processing to first N segments (0 = no limit)
	SegmentLimit int

	// StepLimit limits browser segments to first N steps (0 = no limit)
	StepLimit int
}

// Stage names for progress display
var stageNames = []string{
	"Loading content",
	"Generating audio",
	"Creating videos",
	"Combining videos",
}

// SegmentOrchestrator coordinates video generation using the segment abstraction.
// It works with any ContentSource (transcript JSON, config YAML) and
// any segment type (slides, browser demos).
type SegmentOrchestrator struct {
	config           SegmentConfig
	ttsGenerator     *tts.SegmentTTSGenerator
	videoRegistry    *video.ProviderRegistry
	combiner         *video.Combiner
	progressRenderer *progress.MultiStageRenderer
}

// NewSegmentOrchestrator creates a new segment-based orchestrator.
func NewSegmentOrchestrator(config SegmentConfig, ttsGen *tts.SegmentTTSGenerator) *SegmentOrchestrator {
	// Create video provider registry
	registry := video.NewProviderRegistry()

	// Register image provider for slides
	imageProvider := video.NewImageVideoProvider(config.Width, config.Height, config.FrameRate)
	registry.Register(imageProvider)

	// Register browser provider for browser segments
	browserWorkDir := filepath.Join(config.WorkDir, "browser")
	browserProvider := video.NewBrowserVideoProvider(config.Width, config.Height, config.FrameRate, config.Headless, browserWorkDir)
	registry.Register(browserProvider)

	// Create combiner (uses work dir for temp files)
	videoDir := filepath.Join(config.WorkDir, "video")
	combiner := video.NewCombiner(videoDir)

	// Set up progress renderer
	var progressRenderer *progress.MultiStageRenderer
	if config.ProgressWriter != nil {
		progressRenderer = progress.NewMultiStageRenderer(config.ProgressWriter).
			WithBarWidth(20).
			WithDescWidth(30)
	}

	return &SegmentOrchestrator{
		config:           config,
		ttsGenerator:     ttsGen,
		videoRegistry:    registry,
		combiner:         combiner,
		progressRenderer: progressRenderer,
	}
}

// Process runs the complete video generation pipeline.
func (o *SegmentOrchestrator) Process(ctx context.Context) error {
	// Configure video encoder (hardware or software)
	if o.config.FastEncoding {
		video.SetGlobalEncoderConfig(video.FastEncoderConfig())
	} else {
		video.SetGlobalEncoderConfig(video.DefaultEncoderConfig())
	}

	// Ensure work directory exists
	if err := os.MkdirAll(o.config.WorkDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Stage 1: Load segments
	o.startStage(0)
	segments, err := o.config.Source.Load()
	if err != nil {
		return fmt.Errorf("failed to load content: %w", err)
	}
	o.completeStage(0)

	if len(segments) == 0 {
		return fmt.Errorf("no segments found in source")
	}

	// Apply segment limit if specified
	if o.config.SegmentLimit > 0 && o.config.SegmentLimit < len(segments) {
		fmt.Printf("Limiting to first %d of %d segments\n", o.config.SegmentLimit, len(segments))
		segments = segments[:o.config.SegmentLimit]
	}

	// Apply step limit to browser segments if specified
	if o.config.StepLimit > 0 {
		for _, seg := range segments {
			if browserSeg, ok := seg.(*segment.BrowserSegment); ok {
				browserSeg.LimitSteps(o.config.StepLimit)
			}
		}
	}

	// Get metadata
	metadata := o.config.Source.GetMetadata()
	languages := o.config.Languages
	if len(languages) == 0 {
		languages = []string{metadata.DefaultLanguage}
	}

	// Stage 2: Generate TTS for all segments/languages
	o.startStage(1)
	// Use AudioOutputDir if specified, otherwise use work directory
	// This allows reusing previously generated audio files
	audioDir := o.config.AudioOutputDir
	if audioDir == "" {
		audioDir = filepath.Join(o.config.WorkDir, "audio")
	}

	// Set up progress callback for TTS generation (shows "n/m" format)
	o.ttsGenerator.SetProgressFunc(func(current, total int, desc string) {
		o.updateProgressWithCount(1, current, total, desc)
	})

	audioResults, err := o.ttsGenerator.GenerateMultiLanguage(ctx, segments, languages, audioDir)
	if err != nil {
		return fmt.Errorf("TTS generation failed: %w", err)
	}
	o.completeStage(1)

	// Calculate max durations for pacing (for browser segments)
	maxDurations := tts.CalculateMaxDurations(audioResults)

	// Stage 3: Create videos for each segment
	o.startStage(2)
	videoDir := filepath.Join(o.config.WorkDir, "video")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		return fmt.Errorf("failed to create video directory: %w", err)
	}

	videoPaths := make([]string, len(segments))
	primaryLang := languages[0]

	// Determine parallelism (default to 1 = sequential)
	parallel := o.config.Parallel
	if parallel <= 0 {
		parallel = 1
	}

	if parallel == 1 {
		// Sequential processing
		// For single segment or few segments, use time-based progress animation
		// Estimate: video recording takes roughly 1.2x the audio duration
		var estimatedDurationMs int
		for _, dur := range maxDurations {
			estimatedDurationMs += dur
		}
		estimatedDurationMs = int(float64(estimatedDurationMs) * 1.2) // Add 20% buffer

		// Use time-based animation if we have few segments
		useTimeAnimation := len(segments) <= 2 && estimatedDurationMs > 5000

		var animator *progressAnimator
		if useTimeAnimation {
			animator = o.startProgressAnimation(2, estimatedDurationMs, "recording")
		}

		for i, seg := range segments {
			videoPath, err := o.createSegmentVideo(ctx, seg, audioResults, primaryLang, maxDurations, videoDir)
			if err != nil {
				if animator != nil {
					animator.stopAndComplete()
				}
				return err
			}
			videoPaths[i] = videoPath

			// Only update count-based progress if not using animation
			if !useTimeAnimation {
				o.updateProgressWithCount(2, i+1, len(segments), seg.GetID())
			}
		}

		if animator != nil {
			animator.stopAndComplete()
		}
	} else {
		// Parallel processing
		var wg sync.WaitGroup
		var firstErr atomic.Value
		var completed int64

		// Create a semaphore to limit concurrency
		sem := make(chan struct{}, parallel)

		for i, seg := range segments {
			wg.Add(1)
			go func(idx int, s segment.Segment) {
				defer wg.Done()

				// Acquire semaphore
				sem <- struct{}{}
				defer func() { <-sem }()

				// Skip if we already have an error
				if firstErr.Load() != nil {
					return
				}

				videoPath, err := o.createSegmentVideo(ctx, s, audioResults, primaryLang, maxDurations, videoDir)
				if err != nil {
					firstErr.CompareAndSwap(nil, err)
					return
				}

				videoPaths[idx] = videoPath
				count := atomic.AddInt64(&completed, 1)
				o.updateProgressWithCount(2, int(count), len(segments), s.GetID())
			}(i, seg)
		}

		wg.Wait()

		// Check for errors
		if err := firstErr.Load(); err != nil {
			return err.(error)
		}
	}
	o.completeStage(2)

	// Stage 4: Combine all videos
	o.startStage(3)
	if o.config.TransitionDuration > 0 {
		if err := o.combiner.CombineVideosWithTransitions(ctx, videoPaths, o.config.OutputFile, o.config.TransitionDuration); err != nil {
			return fmt.Errorf("failed to combine videos: %w", err)
		}
	} else {
		if err := o.combiner.CombineVideos(ctx, videoPaths, o.config.OutputFile); err != nil {
			return fmt.Errorf("failed to combine videos: %w", err)
		}
	}
	o.completeStage(3)

	// Generate additional language versions if needed
	if len(languages) > 1 {
		if err := o.generateLanguageVersions(ctx, segments, audioResults, languages[1:]); err != nil {
			return fmt.Errorf("failed to generate language versions: %w", err)
		}
	}

	// Save audio tracks if AudioOutputDir is specified
	if o.config.AudioOutputDir != "" {
		if err := o.saveAudioTracks(ctx, segments, audioResults, languages); err != nil {
			return fmt.Errorf("failed to save audio tracks: %w", err)
		}
	}

	// Generate subtitles if requested
	if o.config.Subtitles || o.config.SubtitlesSTT {
		primaryLang := languages[0]
		if err := o.generateSubtitles(ctx, segments, audioResults, primaryLang); err != nil {
			return fmt.Errorf("failed to generate subtitles: %w", err)
		}
	}

	return nil
}

// createSegmentVideo creates a video for a single segment.
func (o *SegmentOrchestrator) createSegmentVideo(
	ctx context.Context,
	seg segment.Segment,
	audioResults map[string]map[string]*tts.SegmentAudioResult,
	primaryLang string,
	maxDurations map[string]int,
	videoDir string,
) (string, error) {
	audioResult, ok := audioResults[primaryLang][seg.GetID()]
	if !ok {
		return "", fmt.Errorf("no audio result for segment %s", seg.GetID())
	}

	videoPath := filepath.Join(videoDir, seg.GetID()+".mp4")

	// For browser segments, use timing-aware recording then combine with audio
	if seg.GetSourceType() == segment.SourceTypeBrowser {
		voiceoverDurations := tts.CalculateMaxVoiceoverDurations(audioResults, seg.GetID())

		// Create a new browser provider for this segment (needed for parallel recording)
		browserWorkDir := filepath.Join(o.config.WorkDir, "browser", seg.GetID())
		browserProvider := video.NewBrowserVideoProvider(
			o.config.Width, o.config.Height, o.config.FrameRate,
			o.config.Headless, browserWorkDir,
		)

		if o.config.NoAudio {
			// Create video without audio (silent) - paced to TTS timing
			if _, err := browserProvider.CreateVideoWithTiming(ctx, seg, voiceoverDurations, videoPath); err != nil {
				return "", fmt.Errorf("failed to create video for segment %s: %w", seg.GetID(), err)
			}
		} else {
			// Create silent video paced to TTS timing
			silentVideoPath := filepath.Join(videoDir, seg.GetID()+"_silent.mp4")
			if _, err := browserProvider.CreateVideoWithTiming(ctx, seg, voiceoverDurations, silentVideoPath); err != nil {
				return "", fmt.Errorf("failed to create video for segment %s: %w", seg.GetID(), err)
			}

			// Combine with audio
			if err := video.AddAudioToVideo(ctx, silentVideoPath, audioResult.AudioFile, videoPath); err != nil {
				return "", fmt.Errorf("failed to add audio to video for segment %s: %w", seg.GetID(), err)
			}

			// Clean up silent video
			os.Remove(silentVideoPath)
		}
	} else {
		// For slides, use standard approach with max duration padding
		duration := maxDurations[seg.GetID()]
		if slideSeg, ok := seg.(*segment.SlideSegment); ok {
			// Add pause duration
			duration += slideSeg.GetTotalPauseDuration(primaryLang)
		}

		imageProvider := o.videoRegistry.GetProvider(segment.SourceTypeSlide).(*video.ImageVideoProvider)
		if o.config.NoAudio {
			// Create video without audio
			if _, err := imageProvider.CreateVideoWithDuration(ctx, seg, "", videoPath, duration); err != nil {
				return "", fmt.Errorf("failed to create video for segment %s: %w", seg.GetID(), err)
			}
		} else {
			if _, err := imageProvider.CreateVideoWithDuration(ctx, seg, audioResult.AudioFile, videoPath, duration); err != nil {
				return "", fmt.Errorf("failed to create video for segment %s: %w", seg.GetID(), err)
			}
		}
	}

	return videoPath, nil
}

// generateLanguageVersions creates video versions for additional languages.
// It reuses the video track and swaps in different audio.
func (o *SegmentOrchestrator) generateLanguageVersions(
	ctx context.Context,
	segments []segment.Segment,
	audioResults map[string]map[string]*tts.SegmentAudioResult,
	languages []string,
) error {
	for _, lang := range languages {
		// Create combined audio for this language
		langAudioDir := filepath.Join(o.config.WorkDir, "audio", lang)
		combinedAudioPath := filepath.Join(langAudioDir, "combined.mp3")

		// Collect audio files in order
		var audioFiles []string
		for _, seg := range segments {
			if result, ok := audioResults[lang][seg.GetID()]; ok {
				audioFiles = append(audioFiles, result.AudioFile)
			}
		}

		// Combine audio files
		if err := combineAudioFiles(ctx, audioFiles, combinedAudioPath); err != nil {
			return fmt.Errorf("failed to combine audio for %s: %w", lang, err)
		}

		// Create output path for this language
		ext := filepath.Ext(o.config.OutputFile)
		base := o.config.OutputFile[:len(o.config.OutputFile)-len(ext)]
		langOutputPath := fmt.Sprintf("%s_%s%s", base, lang, ext)

		// Combine video with language audio
		if err := video.ReplaceAudio(ctx, o.config.OutputFile, combinedAudioPath, langOutputPath); err != nil {
			return fmt.Errorf("failed to create %s version: %w", lang, err)
		}
	}

	return nil
}

// saveAudioTracks saves combined audio tracks for each language to the specified directory.
func (o *SegmentOrchestrator) saveAudioTracks(
	ctx context.Context,
	segments []segment.Segment,
	audioResults map[string]map[string]*tts.SegmentAudioResult,
	languages []string,
) error {
	if err := os.MkdirAll(o.config.AudioOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create audio output directory: %w", err)
	}

	for _, lang := range languages {
		// Create language subdirectory
		langDir := filepath.Join(o.config.AudioOutputDir, lang)
		if err := os.MkdirAll(langDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", lang, err)
		}

		// Collect audio files in order
		var audioFiles []string
		for _, seg := range segments {
			if result, ok := audioResults[lang][seg.GetID()]; ok {
				audioFiles = append(audioFiles, result.AudioFile)
			}
		}

		if len(audioFiles) == 0 {
			continue
		}

		// Create combined audio file
		combinedPath := filepath.Join(langDir, "combined.mp3")
		if err := combineAudioFiles(ctx, audioFiles, combinedPath); err != nil {
			return fmt.Errorf("failed to combine audio for %s: %w", lang, err)
		}

		// Also copy individual segment audio files
		for _, seg := range segments {
			if result, ok := audioResults[lang][seg.GetID()]; ok {
				segAudioPath := filepath.Join(langDir, seg.GetID()+".mp3")
				if err := copyFile(result.AudioFile, segAudioPath); err != nil {
					return fmt.Errorf("failed to copy audio for segment %s: %w", seg.GetID(), err)
				}
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeFileSecure(dst, data)
}

// totalStages returns the number of stages
func (o *SegmentOrchestrator) totalStages() int {
	return 4
}

// updateProgress updates the progress display if enabled
func (o *SegmentOrchestrator) updateProgressInfo(stage int, desc string, current, total int, done bool) {
	if o.progressRenderer == nil {
		return
	}
	info := progress.StageInfo{
		Stage:       stage,
		TotalStages: o.totalStages(),
		Description: desc,
		Current:     current,
		Total:       total,
		Done:        done,
	}
	o.progressRenderer.Update(info)
}

// Progress helpers
func (o *SegmentOrchestrator) startStage(index int) {
	name := ""
	if index < len(stageNames) {
		name = stageNames[index]
	}
	o.updateProgressInfo(index+1, name, 0, 1, false)
}

func (o *SegmentOrchestrator) completeStage(index int) {
	name := ""
	if index < len(stageNames) {
		name = stageNames[index]
	}
	o.updateProgressInfo(index+1, name, 1, 1, true)
}

// updateProgressWithCount updates progress with actual item counts (displays as "n/m")
func (o *SegmentOrchestrator) updateProgressWithCount(stage, current, total int, desc string) {
	name := ""
	if stage < len(stageNames) {
		name = stageNames[stage]
	}
	if desc != "" {
		name = fmt.Sprintf("%s: %s", name, desc)
	}
	o.updateProgressInfo(stage+1, name, current, total, false)
}

// progressAnimator animates progress over an estimated duration.
// Returns a stop function that should be called when the work is complete.
type progressAnimator struct {
	stop    chan struct{}
	stopped chan struct{}
	stage   int
	orch    *SegmentOrchestrator
	desc    string
}

// startProgressAnimation starts a time-based progress animation for a stage.
// estimatedDurationMs is how long we expect the work to take.
// Returns an animator that should be stopped when work completes.
// nolint:unparam // stage parameter is for future extensibility (other stages beyond video creation)
func (o *SegmentOrchestrator) startProgressAnimation(stage int, estimatedDurationMs int, desc string) *progressAnimator {
	if o.progressRenderer == nil {
		return nil
	}

	pa := &progressAnimator{
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		stage:   stage,
		orch:    o,
		desc:    desc,
	}

	go func() {
		defer close(pa.stopped)

		startTime := time.Now()
		estimatedDuration := time.Duration(estimatedDurationMs) * time.Millisecond
		ticker := time.NewTicker(100 * time.Millisecond) // Update every 100ms
		defer ticker.Stop()

		for {
			select {
			case <-pa.stop:
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				// Calculate progress percentage (cap at 95% to leave room for completion)
				pct := int((float64(elapsed) / float64(estimatedDuration)) * 95)
				if pct > 95 {
					pct = 95
				}
				if pct < 1 {
					pct = 1
				}

				name := ""
				if stage < len(stageNames) {
					name = stageNames[stage]
				}
				if pa.desc != "" {
					name = fmt.Sprintf("%s: %s", name, pa.desc)
				}
				pa.orch.updateProgressInfo(stage+1, name, pct, 100, false)
			}
		}
	}()

	return pa
}

// stopAndComplete stops the animation and marks progress as complete.
func (pa *progressAnimator) stopAndComplete() {
	if pa == nil {
		return
	}
	close(pa.stop)
	<-pa.stopped // Wait for goroutine to finish
}

// generateSubtitles generates subtitle files for the video.
func (o *SegmentOrchestrator) generateSubtitles(
	ctx context.Context,
	segments []segment.Segment,
	audioResults map[string]map[string]*tts.SegmentAudioResult,
	language string,
) error {
	// Determine output path for subtitles (same as video but with subtitle extension)
	ext := filepath.Ext(o.config.OutputFile)
	base := o.config.OutputFile[:len(o.config.OutputFile)-len(ext)]

	subtitleFormat := tts.SubtitleFormat(o.config.SubtitleFormat)
	if subtitleFormat == "" {
		subtitleFormat = tts.FormatSRT
	}

	subtitleExt := ".srt"
	if subtitleFormat == tts.FormatVTT {
		subtitleExt = ".vtt"
	}

	subtitlePath := base + subtitleExt

	// Calculate cumulative timing across all segments
	var allTimings []tts.VoiceoverTiming
	cumulativeTimeMs := 0

	for _, seg := range segments {
		langResults, ok := audioResults[language]
		if !ok {
			continue
		}
		audioResult, ok := langResults[seg.GetID()]
		if !ok {
			continue
		}

		voiceovers := seg.GetVoiceovers(language)
		for _, vo := range voiceovers {
			duration := audioResult.VoiceoverDurations[vo.Index]
			if duration == 0 {
				// Fallback estimation
				wordCount := len(strings.Fields(vo.Text))
				duration = wordCount * 400 // ~400ms per word
			}

			allTimings = append(allTimings, tts.VoiceoverTiming{
				Index:    len(allTimings) + 1,
				Text:     vo.Text,
				StartMs:  cumulativeTimeMs,
				EndMs:    cumulativeTimeMs + duration,
				Duration: duration,
			})

			cumulativeTimeMs += duration
			if vo.Pause > 0 {
				cumulativeTimeMs += vo.Pause
			}
		}
	}

	if len(allTimings) == 0 {
		return fmt.Errorf("no voiceovers found for subtitle generation")
	}

	// Generate subtitles using the browser subtitle generator
	generator := tts.NewBrowserSubtitleGenerator(subtitleFormat)
	if err := generator.GenerateFromTimings(allTimings, subtitlePath); err != nil {
		return fmt.Errorf("failed to generate subtitles: %w", err)
	}

	// If burning subtitles, re-encode the video with subtitles
	if o.config.SubtitlesBurn {
		if err := o.burnSubtitles(ctx, subtitlePath); err != nil {
			return fmt.Errorf("failed to burn subtitles: %w", err)
		}
	}

	return nil
}

// checkFFmpegSubtitleSupport verifies that FFmpeg has the subtitles filter available.
// The subtitles filter requires libass to be compiled into FFmpeg.
func checkFFmpegSubtitleSupport(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-filters")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to query FFmpeg filters: %w", err)
	}

	if !strings.Contains(string(output), "subtitles") {
		return fmt.Errorf(`FFmpeg subtitles filter not available.

Your FFmpeg installation was built without libass support, which is required
for burning subtitles into videos.

To fix this, reinstall FFmpeg with libass support:

  # Option 1: Use the homebrew-ffmpeg tap (recommended)
  brew tap homebrew-ffmpeg/ffmpeg
  brew install homebrew-ffmpeg/ffmpeg/ffmpeg --with-libass

  # Option 2: Build from source
  brew uninstall ffmpeg
  brew install ffmpeg --build-from-source

After reinstalling, run your command again.

Alternatively, use --subtitles without --subtitles-burn to generate
a separate .srt file that can be loaded by video players.`)
	}

	return nil
}

// burnSubtitles re-encodes the video with subtitles burned in.
func (o *SegmentOrchestrator) burnSubtitles(ctx context.Context, subtitlePath string) error {
	// Check if FFmpeg has subtitles filter support
	if err := checkFFmpegSubtitleSupport(ctx); err != nil {
		return err
	}

	// Verify input files exist
	if _, err := os.Stat(o.config.OutputFile); err != nil {
		return fmt.Errorf("video file not found: %s", o.config.OutputFile)
	}
	if _, err := os.Stat(subtitlePath); err != nil {
		return fmt.Errorf("subtitle file not found: %s", subtitlePath)
	}

	// Create temporary output file
	ext := filepath.Ext(o.config.OutputFile)
	base := o.config.OutputFile[:len(o.config.OutputFile)-len(ext)]
	tempOutput := base + "_subtitled" + ext

	// Get absolute path for subtitle file (ffmpeg filter requires it for reliability)
	absSubtitlePath, err := filepath.Abs(subtitlePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for subtitles: %w", err)
	}

	// Escape special characters for ffmpeg's subtitles filter
	// FFmpeg filter syntax requires escaping: \ : ' [ ]
	escapedPath := escapeFFmpegFilterPath(absSubtitlePath)

	// Build the filter string with fps conversion for proper subtitle timing
	// VFR videos (few frames with long durations) need CFR conversion for subtitles to sync
	// We convert to 30fps, apply subtitles, which ensures subtitle timing matches real time
	filterStr := fmt.Sprintf("fps=30,subtitles='%s'", escapedPath)

	// Use ffmpeg to burn subtitles
	args := []string{
		"-y",
		"-i", o.config.OutputFile,
		"-vf", filterStr,
		"-c:a", "copy",
		tempOutput,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up temp file if it was created
		os.Remove(tempOutput)

		// Provide detailed error message
		return fmt.Errorf(`ffmpeg subtitle burn failed: %w

Command: ffmpeg %s
Filter:  %s
Subtitle file: %s

FFmpeg output:
%s

If the error mentions parsing issues, try:
1. Ensure FFmpeg has libass support: ffmpeg -filters | grep subtitles
2. Check the subtitle file is valid: head -20 %s`,
			err,
			strings.Join(args, " "),
			filterStr,
			absSubtitlePath,
			string(output),
			subtitlePath)
	}

	// Replace original with subtitled version
	if err := os.Remove(o.config.OutputFile); err != nil {
		os.Remove(tempOutput) // Clean up
		return fmt.Errorf("failed to remove original video: %w", err)
	}
	if err := os.Rename(tempOutput, o.config.OutputFile); err != nil {
		return fmt.Errorf("failed to rename subtitled video: %w", err)
	}

	return nil
}

// escapeFFmpegFilterPath escapes special characters in a path for use in FFmpeg filters.
// FFmpeg's filter syntax requires escaping: backslash, colon, single quote, and brackets.
// The path will be wrapped in single quotes by the caller.
func escapeFFmpegFilterPath(path string) string {
	// Order matters: escape backslashes first, then other special chars
	// These are FFmpeg filter escaping rules, not shell escaping
	path = strings.ReplaceAll(path, "\\", "/") // Use forward slashes (works on all platforms)
	path = strings.ReplaceAll(path, "'", "\\'")
	path = strings.ReplaceAll(path, ":", "\\:")
	path = strings.ReplaceAll(path, "[", "\\[")
	path = strings.ReplaceAll(path, "]", "\\]")
	return path
}

// combineAudioFiles concatenates audio files into a single file.
func combineAudioFiles(ctx context.Context, inputFiles []string, outputPath string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	if len(inputFiles) == 1 {
		// Just copy the single file
		data, err := os.ReadFile(inputFiles[0])
		if err != nil {
			return err
		}
		return writeFileSecure(outputPath, data)
	}

	// Create concat file
	concatPath := outputPath + ".concat.txt"
	defer os.Remove(concatPath)

	var content string
	for _, f := range inputFiles {
		content += fmt.Sprintf("file '%s'\n", f)
	}

	if err := os.WriteFile(concatPath, []byte(content), 0600); err != nil {
		return err
	}

	// Run ffmpeg
	args := []string{
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", concatPath,
		"-c:a", "copy",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
