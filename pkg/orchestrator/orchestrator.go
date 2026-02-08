package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grokify/marp2video/pkg/audio"
	"github.com/grokify/marp2video/pkg/parser"
	"github.com/grokify/marp2video/pkg/renderer"
	"github.com/grokify/marp2video/pkg/tts"
	"github.com/grokify/marp2video/pkg/video"
	"github.com/grokify/mogo/log/slogutil"
)

// Config holds orchestrator configuration
type Config struct {
	InputFile           string
	OutputFile          string
	WorkDir             string
	ElevenLabsAPIKey    string
	VoiceID             string
	Width               int
	Height              int
	FrameRate           int
	OutputIndividualDir string  // Directory for individual slide videos (Udemy)
	TransitionDuration  float64 // Duration of transitions between slides in seconds
	ScreenDevice        string  // Screen capture device (macOS, auto-detected if empty)
	AudioManifest       string  // Path to audio manifest file (from 'marp2video tts')
}

// Orchestrator coordinates the entire video generation process
type Orchestrator struct {
	config Config
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(config Config) *Orchestrator {
	// Set defaults
	if config.Width == 0 {
		config.Width = 1920
	}
	if config.Height == 0 {
		config.Height = 1080
	}
	if config.FrameRate == 0 {
		config.FrameRate = 30
	}
	if config.WorkDir == "" {
		config.WorkDir = filepath.Join(os.TempDir(), "marp2video")
	}

	return &Orchestrator{config: config}
}

// Process orchestrates the entire conversion process
func (o *Orchestrator) Process(ctx context.Context) error {
	logger := slogutil.LoggerFromContext(ctx, nil)
	logger.Info("Starting Marp to Video conversion...")

	// Create working directories
	audioDir := filepath.Join(o.config.WorkDir, "audio")
	videoDir := filepath.Join(o.config.WorkDir, "video")
	htmlDir := filepath.Join(o.config.WorkDir, "html")

	for _, dir := range []string{audioDir, videoDir, htmlDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Step 1: Parse Marp markdown
	logger.Info("Step 1: Parsing Marp markdown...")
	content, err := os.ReadFile(o.config.InputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	presentation, err := parser.ParseMarpFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse Marp file: %w", err)
	}
	logger.Info("Found slides", "count", len(presentation.Slides))

	// Step 2: Generate audio for each slide
	logger.Info("Step 2: Generating audio with ElevenLabs...")
	ttsGen := tts.NewGenerator(tts.Config{
		APIKey:    o.config.ElevenLabsAPIKey,
		VoiceID:   o.config.VoiceID,
		OutputDir: audioDir,
	})

	audioPlayer := audio.NewPlayer()
	audioFiles := make([]string, len(presentation.Slides))
	slideDurations := make([]time.Duration, len(presentation.Slides))

	for i, slide := range presentation.Slides {
		if slide.Voiceover == "" {
			logger.Info("Slide: No voiceover, skipping audio generation", "slide", i)
			continue
		}

		logger.Info("Slide: Generating audio...", "slide", i)
		audioResult, err := ttsGen.GenerateAudio(ctx, slide.Voiceover, i)
		if err != nil {
			return fmt.Errorf("failed to generate audio for slide %d: %w", i, err)
		}

		audioFiles[i] = audioResult.FilePath

		// Get actual audio duration using ffprobe
		duration, err := audioPlayer.GetDuration(audioResult.FilePath)
		if err != nil {
			return fmt.Errorf("failed to get audio duration for slide %d: %w", i, err)
		}

		// Add pause durations
		totalDuration := duration + time.Duration(slide.TotalPauseDuration)*time.Millisecond
		slideDurations[i] = totalDuration

		logger.Info("Slide: Audio generated", "slide", i, "duration", totalDuration.Seconds())
	}

	// Step 3: Render Marp to HTML
	logger.Info("Step 3: Rendering Marp to HTML...")
	marpRenderer := renderer.NewMarpRenderer()
	if err := marpRenderer.CheckMarpCLI(); err != nil {
		return err
	}

	htmlPath, err := marpRenderer.RenderToHTML(o.config.InputFile, htmlDir)
	if err != nil {
		return fmt.Errorf("failed to render HTML: %w", err)
	}
	logger.Info("HTML presentation created", "path", htmlPath)

	// Step 4: Open browser and record each slide
	logger.Info("Step 4: Recording slides...")
	browserCtrl, err := renderer.NewBrowserController(o.config.Width, o.config.Height)
	if err != nil {
		return fmt.Errorf("failed to create browser controller: %w", err)
	}
	defer func() {
		if err := browserCtrl.Close(); err != nil {
			logger.Warn("failed to close browser", "error", err)
		}
	}()

	if err := browserCtrl.LoadPresentation(htmlPath); err != nil {
		return fmt.Errorf("failed to load presentation: %w", err)
	}

	recorder := video.NewRecorder(video.RecorderConfig{
		OutputDir:    videoDir,
		Width:        o.config.Width,
		Height:       o.config.Height,
		FrameRate:    o.config.FrameRate,
		ScreenDevice: o.config.ScreenDevice,
	})

	videoFiles := make([]string, 0, len(presentation.Slides))

	for i := range presentation.Slides {
		if audioFiles[i] == "" {
			logger.Info("Slide: Skipping (no audio)", "slide", i)
			continue
		}

		logger.Info("Slide: Recording...", "slide", i)

		// Navigate to slide
		if err := browserCtrl.NavigateToSlide(i); err != nil {
			return fmt.Errorf("failed to navigate to slide %d: %w", i, err)
		}

		// Wait a moment for slide to settle
		time.Sleep(500 * time.Millisecond)

		// Record slide with audio
		videoPath, err := recorder.RecordSlide(ctx, i, audioFiles[i], slideDurations[i])
		if err != nil {
			return fmt.Errorf("failed to record slide %d: %w", i, err)
		}

		videoFiles = append(videoFiles, videoPath)
		logger.Info("Slide: Recorded", "slide", i, "path", videoPath)
	}

	// Step 5: Save individual videos if requested (for Udemy)
	if o.config.OutputIndividualDir != "" {
		logger.Info("Step 5a: Saving individual slide videos...")
		if err := os.MkdirAll(o.config.OutputIndividualDir, 0755); err != nil {
			return fmt.Errorf("failed to create individual output directory: %w", err)
		}

		for i, videoPath := range videoFiles {
			// Find the slide index from the video filename
			baseName := filepath.Base(videoPath)
			destPath := filepath.Join(o.config.OutputIndividualDir, baseName)

			// Copy the video file
			data, err := os.ReadFile(videoPath)
			if err != nil {
				return fmt.Errorf("failed to read video %d: %w", i, err)
			}
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write video %d: %w", i, err)
			}
			logger.Info("Saved individual video", "path", destPath)
		}
	}

	// Step 6: Combine all videos (for YouTube)
	logger.Info("Step 6: Combining videos...")
	combiner := video.NewCombiner(videoDir)

	var combineErr error
	if o.config.TransitionDuration > 0 {
		combineErr = combiner.CombineVideosWithTransitions(ctx, videoFiles, o.config.OutputFile, o.config.TransitionDuration)
	} else {
		combineErr = combiner.CombineVideos(ctx, videoFiles, o.config.OutputFile)
	}
	if combineErr != nil {
		return fmt.Errorf("failed to combine videos: %w", combineErr)
	}

	logger.Info("Video generation complete", "output", o.config.OutputFile)
	if o.config.OutputIndividualDir != "" {
		logger.Info("Individual videos saved", "dir", o.config.OutputIndividualDir)
	}
	return nil
}
