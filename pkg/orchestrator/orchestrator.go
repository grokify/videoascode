package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/grokify/marp2video/pkg/audio"
	"github.com/grokify/marp2video/pkg/parser"
	"github.com/grokify/marp2video/pkg/renderer"
	"github.com/grokify/marp2video/pkg/tts"
	"github.com/grokify/marp2video/pkg/video"
	"github.com/grokify/mogo/fmt/progress"
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
	OutputIndividualDir string    // Directory for individual slide videos (Udemy)
	TransitionDuration  float64   // Duration of transitions between slides in seconds
	ScreenDevice        string    // Screen capture device (macOS, auto-detected if empty)
	AudioManifest       string    // Path to audio manifest file (from 'marp2video tts')
	ProgressWriter      io.Writer // Writer for progress output (nil to disable)
}

// Orchestrator coordinates the entire video generation process
type Orchestrator struct {
	config   Config
	progress *progress.MultiStageRenderer
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

	o := &Orchestrator{config: config}

	// Initialize progress renderer if writer is provided
	if config.ProgressWriter != nil {
		o.progress = progress.NewMultiStageRenderer(config.ProgressWriter).
			WithBarWidth(20).
			WithDescWidth(30)
	}

	return o
}

// totalStages returns the number of stages (5 or 6 if exporting individual videos)
func (o *Orchestrator) totalStages() int {
	if o.config.OutputIndividualDir != "" {
		return 6
	}
	return 5
}

// updateProgress updates the progress display if enabled
func (o *Orchestrator) updateProgress(stage int, desc string, current, total int, done bool) {
	if o.progress == nil {
		return
	}
	o.progress.Update(progress.StageInfo{
		Stage:       stage,
		TotalStages: o.totalStages(),
		Description: desc,
		Current:     current,
		Total:       total,
		Done:        done,
	})
}

// Process orchestrates the entire conversion process
func (o *Orchestrator) Process(ctx context.Context) error {
	logger := slogutil.LoggerFromContext(ctx, slogutil.Null())
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

	// Create output directory if needed
	if outputDir := filepath.Dir(o.config.OutputFile); outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
		}
	}

	// Step 1: Parse Marp markdown
	o.updateProgress(1, "Parsing Marp markdown", 0, 0, false)
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
	o.updateProgress(1, "Parsing Marp markdown", 0, 0, true)

	// Step 2: Get audio for each slide (from manifest or generate)
	numSlides := len(presentation.Slides)
	audioFiles := make([]string, numSlides)
	slideDurations := make([]time.Duration, numSlides)

	if o.config.AudioManifest != "" {
		// Use pre-generated audio from manifest
		logger.Info("Step 2: Loading audio from manifest...", "manifest", o.config.AudioManifest)
		o.updateProgress(2, "Loading audio manifest", 0, 0, false)

		manifest, err := tts.LoadManifest(o.config.AudioManifest)
		if err != nil {
			return fmt.Errorf("failed to load audio manifest: %w", err)
		}

		manifestDir := filepath.Dir(o.config.AudioManifest)
		for i := 0; i < numSlides; i++ {
			o.updateProgress(2, "Loading audio manifest", i+1, numSlides, false)

			slideAudio, err := manifest.GetSlide(i)
			if err != nil {
				logger.Warn("No audio in manifest for slide", "slide", i)
				continue
			}

			audioFiles[i] = filepath.Join(manifestDir, slideAudio.AudioFile)
			slideDurations[i] = time.Duration(slideAudio.TotalDuration) * time.Millisecond

			logger.Info("Slide: Using pre-generated audio", "slide", i, "duration", slideDurations[i].Seconds())
		}
		o.updateProgress(2, "Loading audio manifest", numSlides, numSlides, true)
	} else {
		// Generate audio with ElevenLabs
		logger.Info("Step 2: Generating audio with ElevenLabs...")
		ttsGen := tts.NewGenerator(tts.Config{
			APIKey:    o.config.ElevenLabsAPIKey,
			VoiceID:   o.config.VoiceID,
			OutputDir: audioDir,
		})

		audioPlayer := audio.NewPlayer()

		for i, slide := range presentation.Slides {
			o.updateProgress(2, "Generating audio", i+1, numSlides, false)

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
		o.updateProgress(2, "Generating audio", numSlides, numSlides, true)
	}

	// Step 3: Render Marp to HTML
	o.updateProgress(3, "Rendering HTML", 0, 0, false)
	logger.Info("Step 3: Rendering Marp to HTML...")
	marpRenderer := renderer.NewMarpRenderer()
	if err := marpRenderer.CheckMarpCLI(); err != nil {
		return err
	}

	htmlPath, err := marpRenderer.RenderToHTML(o.config.InputFile, htmlDir)
	if err != nil {
		return fmt.Errorf("failed to render HTML: %w", err)
	}
	_ = htmlPath // HTML created for reference; using images for video
	logger.Info("HTML presentation created", "path", htmlPath)
	o.updateProgress(3, "Rendering HTML", 0, 0, true)

	// Step 4: Render slides to images and create videos
	o.updateProgress(4, "Creating slide videos", 0, numSlides, false)
	logger.Info("Step 4: Rendering slides to images...")

	// Render Marp to PNG images (one per slide)
	imagesDir := filepath.Join(o.config.WorkDir, "images")
	imagePaths, err := marpRenderer.RenderToImages(o.config.InputFile, imagesDir)
	if err != nil {
		return fmt.Errorf("failed to render images: %w", err)
	}
	logger.Info("Generated slide images", "count", len(imagePaths))

	// Sort image paths to ensure correct order
	sort.Strings(imagePaths)

	// Create video converter
	converter := video.NewImageVideoConverter(video.ImageVideoConfig{
		OutputDir: videoDir,
		Width:     o.config.Width,
		Height:    o.config.Height,
		FrameRate: o.config.FrameRate,
	})

	videoFiles := make([]string, 0, len(presentation.Slides))

	// Create video for each slide with audio
	for i := range presentation.Slides {
		o.updateProgress(4, "Creating slide videos", i+1, numSlides, false)

		if audioFiles[i] == "" {
			logger.Info("Slide: Skipping (no audio)", "slide", i)
			continue
		}

		// Find corresponding image (marp generates slide.001.png, slide.002.png, etc.)
		if i >= len(imagePaths) {
			return fmt.Errorf("no image found for slide %d", i)
		}
		imagePath := imagePaths[i]

		logger.Info("Creating video for slide", "slide", i, "image", imagePath, "audio", audioFiles[i])

		// Create video from image + audio
		videoPath, err := converter.CreateSlideVideoWithSize(ctx, i, imagePath, audioFiles[i], slideDurations[i], o.config.Width, o.config.Height)
		if err != nil {
			return fmt.Errorf("failed to create video for slide %d: %w", i, err)
		}

		videoFiles = append(videoFiles, videoPath)
		logger.Info("Slide: Video created", "slide", i, "path", videoPath)
	}
	o.updateProgress(4, "Creating slide videos", numSlides, numSlides, true)

	// Step 5: Combine all videos (for YouTube)
	combineStage := 5
	o.updateProgress(combineStage, "Combining videos", 0, 0, false)
	logger.Info("Step 5: Combining videos...")
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
	o.updateProgress(combineStage, "Combining videos", 0, 0, true)

	// Step 6: Save individual videos if requested (for Udemy)
	if o.config.OutputIndividualDir != "" {
		exportStage := 6
		numVideos := len(videoFiles)
		o.updateProgress(exportStage, "Exporting individual videos", 0, numVideos, false)
		logger.Info("Step 6: Saving individual slide videos...")

		if err := os.MkdirAll(o.config.OutputIndividualDir, 0755); err != nil {
			return fmt.Errorf("failed to create individual output directory: %w", err)
		}

		for i, videoPath := range videoFiles {
			o.updateProgress(exportStage, "Exporting individual videos", i+1, numVideos, false)

			// Find the slide index from the video filename
			baseName := filepath.Base(videoPath)
			destPath := filepath.Join(o.config.OutputIndividualDir, baseName)

			// Copy the video file
			data, err := os.ReadFile(videoPath)
			if err != nil {
				return fmt.Errorf("failed to read video %d: %w", i, err)
			}
			if err := os.WriteFile(destPath, data, 0600); err != nil {
				return fmt.Errorf("failed to write video %d: %w", i, err)
			}
			logger.Info("Saved individual video", "path", destPath)
		}
		o.updateProgress(exportStage, "Exporting individual videos", numVideos, numVideos, true)
	}

	logger.Info("Video generation complete", "output", o.config.OutputFile)
	if o.config.OutputIndividualDir != "" {
		logger.Info("Individual videos saved", "dir", o.config.OutputIndividualDir)
	}
	return nil
}
