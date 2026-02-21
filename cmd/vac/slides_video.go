package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grokify/videoascode/pkg/orchestrator"
	"github.com/grokify/videoascode/pkg/renderer"
	"github.com/grokify/videoascode/pkg/video"
	"github.com/spf13/cobra"
)

var slidesVideoCmd = &cobra.Command{
	Use:   "video",
	Short: "Generate video from Marp presentation",
	Long: `Generate a video from a Marp markdown presentation with AI voiceovers.

This command runs the full pipeline:
  1. Parse Marp markdown and extract voiceover text
  2. Generate audio using ElevenLabs TTS
  3. Render presentation to HTML
  4. Record each slide with synchronized audio
  5. Combine into final video

You can also use a pre-generated audio manifest (from 'vac slides tts')
to skip the TTS generation step.

Examples:
  # Full pipeline with inline voiceovers
  vac slides video --input slides.md --output video.mp4

  # Use pre-generated audio manifest
  vac slides video --input slides.md --manifest audio/manifest.json --output video.mp4

  # Custom voice and resolution
  vac slides video --input slides.md --output video.mp4 --voice 21m00Tcm4TlvDq8ikWAM --width 1280 --height 720`,
	RunE: runSlidesVideo,
}

var (
	svInputFile          string
	svOutputFile         string
	svAPIKey             string
	svVoiceID            string
	svWidth              int
	svHeight             int
	svFPS                int
	svWorkDir            string
	svOutputIndividual   string
	svTransitionDuration float64
	svScreenDevice       string
	svManifest           string
	svCheckDeps          bool
	svSubtitles          string
	svSubtitlesLang      string
)

func init() {
	slidesVideoCmd.Flags().StringVarP(&svInputFile, "input", "i", "", "Input Marp markdown file (required)")
	slidesVideoCmd.Flags().StringVarP(&svOutputFile, "output", "o", "output.mp4", "Output video file")
	slidesVideoCmd.Flags().StringVarP(&svAPIKey, "api-key", "k", "", "ElevenLabs API key (or use ELEVENLABS_API_KEY env var)")
	slidesVideoCmd.Flags().StringVarP(&svVoiceID, "voice", "v", "pNInz6obpgDQGcFmaJgB", "ElevenLabs voice ID (default: Adam)")
	slidesVideoCmd.Flags().IntVar(&svWidth, "width", 1920, "Video width")
	slidesVideoCmd.Flags().IntVar(&svHeight, "height", 1080, "Video height")
	slidesVideoCmd.Flags().IntVar(&svFPS, "fps", 30, "Video frame rate")
	slidesVideoCmd.Flags().StringVar(&svWorkDir, "workdir", "", "Working directory for temporary files")
	slidesVideoCmd.Flags().StringVar(&svOutputIndividual, "output-individual", "", "Directory to save individual slide videos")
	slidesVideoCmd.Flags().Float64Var(&svTransitionDuration, "transition", 0, "Transition duration between slides in seconds")
	slidesVideoCmd.Flags().StringVar(&svScreenDevice, "screen-device", "", "Screen capture device (macOS only)")
	slidesVideoCmd.Flags().StringVarP(&svManifest, "manifest", "m", "", "Audio manifest file (from 'vac slides tts')")
	slidesVideoCmd.Flags().BoolVar(&svCheckDeps, "check", false, "Check dependencies and exit")
	slidesVideoCmd.Flags().StringVar(&svSubtitles, "subtitles", "", "Subtitle file to embed (SRT or VTT)")
	slidesVideoCmd.Flags().StringVar(&svSubtitlesLang, "subtitles-lang", "", "Subtitle language code, e.g., en-US (auto-detected from filename if not specified)")

	if err := slidesVideoCmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}

	slidesCmd.AddCommand(slidesVideoCmd)
}

func runSlidesVideo(cmd *cobra.Command, args []string) error {
	// Check dependencies
	if svCheckDeps {
		return checkSlidesVideoDependencies()
	}

	// Validate input file exists
	if _, err := os.Stat(svInputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", svInputFile)
	}

	// Get API key (only required if not using manifest)
	apiKey := svAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("ELEVENLABS_API_KEY")
	}
	if apiKey == "" && svManifest == "" {
		return fmt.Errorf("ElevenLabs API key required: use --api-key flag or ELEVENLABS_API_KEY env var")
	}

	// Set working directory
	workDir := svWorkDir
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "vac")
	}

	// Check dependencies before starting
	if err := checkSlidesVideoDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Create orchestrator config with progress output
	config := orchestrator.Config{
		InputFile:           svInputFile,
		OutputFile:          svOutputFile,
		WorkDir:             workDir,
		ElevenLabsAPIKey:    apiKey,
		VoiceID:             svVoiceID,
		Width:               svWidth,
		Height:              svHeight,
		FrameRate:           svFPS,
		OutputIndividualDir: svOutputIndividual,
		TransitionDuration:  svTransitionDuration,
		ScreenDevice:        svScreenDevice,
		AudioManifest:       svManifest,
		ProgressWriter:      os.Stdout,
	}

	// Create and run orchestrator
	orch := orchestrator.NewOrchestrator(config)

	ctx := newContext()
	if err := orch.Process(ctx); err != nil {
		return err
	}

	// Embed subtitles if provided
	if svSubtitles != "" {
		if _, err := os.Stat(svSubtitles); os.IsNotExist(err) {
			return fmt.Errorf("subtitle file does not exist: %s", svSubtitles)
		}

		// Determine language code
		lang := svSubtitlesLang
		if lang == "" {
			// Auto-detect from filename (e.g., "en-US.srt" -> "en-US")
			lang = video.DetectLanguageFromSubtitlePath(svSubtitles)
		}
		if lang == "" {
			lang = "en-US" // Default fallback
		}

		// Convert BCP-47 to ISO 639-2 for ffmpeg
		isoLang := video.BCP47ToISO639(lang)

		fmt.Printf("\nEmbedding subtitles from: %s\n", svSubtitles)
		fmt.Printf("  Language: %s (%s)\n", lang, isoLang)

		// Create temp output file
		tempOutput := svOutputFile + ".with-subs.mp4"

		// Run ffmpeg to embed subtitles
		if err := video.EmbedSubtitles(svOutputFile, svSubtitles, isoLang, tempOutput); err != nil {
			return fmt.Errorf("failed to embed subtitles: %w", err)
		}

		// Replace original with subtitled version
		if err := os.Remove(svOutputFile); err != nil {
			return fmt.Errorf("failed to remove original video: %w", err)
		}
		if err := os.Rename(tempOutput, svOutputFile); err != nil {
			return fmt.Errorf("failed to rename subtitled video: %w", err)
		}

		fmt.Printf("✓ Subtitles embedded\n")
	}

	fmt.Printf("\n✓ Success! Video saved to: %s\n", svOutputFile)
	return nil
}

// checkSlidesVideoDependencies verifies all required tools are installed
func checkSlidesVideoDependencies() error {
	// Check ffmpeg
	if err := video.CheckFFmpeg(); err != nil {
		return fmt.Errorf("ffmpeg not found: %w\nInstall: https://ffmpeg.org/download.html", err)
	}
	fmt.Println("✓ ffmpeg found")

	// Check Marp CLI
	marpRenderer := renderer.NewMarpRenderer()
	if err := marpRenderer.CheckMarpCLI(); err != nil {
		return fmt.Errorf("marp CLI not found: %w\nInstall: npm install -g @marp-team/marp-cli", err)
	}
	fmt.Println("✓ marp CLI found")

	return nil
}
