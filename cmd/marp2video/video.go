package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grokify/marp2video/pkg/orchestrator"
	"github.com/grokify/marp2video/pkg/renderer"
	"github.com/grokify/marp2video/pkg/video"
	"github.com/spf13/cobra"
)

var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "Generate video from Marp presentation",
	Long: `Generate a video from a Marp markdown presentation with AI voiceovers.

This command runs the full pipeline:
  1. Parse Marp markdown and extract voiceover text
  2. Generate audio using ElevenLabs TTS
  3. Render presentation to HTML
  4. Record each slide with synchronized audio
  5. Combine into final video

You can also use a pre-generated audio manifest (from 'marp2video tts')
to skip the TTS generation step.

Examples:
  # Full pipeline with inline voiceovers
  marp2video video --input slides.md --output video.mp4

  # Use pre-generated audio manifest
  marp2video video --input slides.md --manifest audio/manifest.json --output video.mp4

  # Custom voice and resolution
  marp2video video --input slides.md --output video.mp4 --voice 21m00Tcm4TlvDq8ikWAM --width 1280 --height 720`,
	RunE: runVideo,
}

var (
	videoInputFile          string
	videoOutputFile         string
	videoAPIKey             string
	videoVoiceID            string
	videoWidth              int
	videoHeight             int
	videoFPS                int
	videoWorkDir            string
	videoOutputIndividual   string
	videoTransitionDuration float64
	videoScreenDevice       string
	videoManifest           string
	videoCheckDeps          bool
	videoSubtitles          string
	videoSubtitlesLang      string
)

func init() {
	videoCmd.Flags().StringVarP(&videoInputFile, "input", "i", "", "Input Marp markdown file (required)")
	videoCmd.Flags().StringVarP(&videoOutputFile, "output", "o", "output.mp4", "Output video file")
	videoCmd.Flags().StringVarP(&videoAPIKey, "api-key", "k", "", "ElevenLabs API key (or use ELEVENLABS_API_KEY env var)")
	videoCmd.Flags().StringVarP(&videoVoiceID, "voice", "v", "pNInz6obpgDQGcFmaJgB", "ElevenLabs voice ID (default: Adam)")
	videoCmd.Flags().IntVar(&videoWidth, "width", 1920, "Video width")
	videoCmd.Flags().IntVar(&videoHeight, "height", 1080, "Video height")
	videoCmd.Flags().IntVar(&videoFPS, "fps", 30, "Video frame rate")
	videoCmd.Flags().StringVar(&videoWorkDir, "workdir", "", "Working directory for temporary files")
	videoCmd.Flags().StringVar(&videoOutputIndividual, "output-individual", "", "Directory to save individual slide videos")
	videoCmd.Flags().Float64Var(&videoTransitionDuration, "transition", 0, "Transition duration between slides in seconds")
	videoCmd.Flags().StringVar(&videoScreenDevice, "screen-device", "", "Screen capture device (macOS only)")
	videoCmd.Flags().StringVarP(&videoManifest, "manifest", "m", "", "Audio manifest file (from 'marp2video tts')")
	videoCmd.Flags().BoolVar(&videoCheckDeps, "check", false, "Check dependencies and exit")
	videoCmd.Flags().StringVar(&videoSubtitles, "subtitles", "", "Subtitle file to embed (SRT or VTT)")
	videoCmd.Flags().StringVar(&videoSubtitlesLang, "subtitles-lang", "", "Subtitle language code, e.g., en-US (auto-detected from filename if not specified)")

	if err := videoCmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(videoCmd)
}

func runVideo(cmd *cobra.Command, args []string) error {
	// Check dependencies
	if videoCheckDeps {
		return checkDependencies()
	}

	// Validate input file exists
	if _, err := os.Stat(videoInputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", videoInputFile)
	}

	// Get API key (only required if not using manifest)
	apiKey := videoAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("ELEVENLABS_API_KEY")
	}
	if apiKey == "" && videoManifest == "" {
		return fmt.Errorf("ElevenLabs API key required: use --api-key flag or ELEVENLABS_API_KEY env var")
	}

	// Set working directory
	workDir := videoWorkDir
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "marp2video")
	}

	// Check dependencies before starting
	if err := checkDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Create orchestrator config with progress output
	config := orchestrator.Config{
		InputFile:           videoInputFile,
		OutputFile:          videoOutputFile,
		WorkDir:             workDir,
		ElevenLabsAPIKey:    apiKey,
		VoiceID:             videoVoiceID,
		Width:               videoWidth,
		Height:              videoHeight,
		FrameRate:           videoFPS,
		OutputIndividualDir: videoOutputIndividual,
		TransitionDuration:  videoTransitionDuration,
		ScreenDevice:        videoScreenDevice,
		AudioManifest:       videoManifest,
		ProgressWriter:      os.Stdout,
	}

	// Create and run orchestrator
	orch := orchestrator.NewOrchestrator(config)

	ctx := newContext()
	if err := orch.Process(ctx); err != nil {
		return err
	}

	// Embed subtitles if provided
	if videoSubtitles != "" {
		if _, err := os.Stat(videoSubtitles); os.IsNotExist(err) {
			return fmt.Errorf("subtitle file does not exist: %s", videoSubtitles)
		}

		// Determine language code
		lang := videoSubtitlesLang
		if lang == "" {
			// Auto-detect from filename (e.g., "en-US.srt" -> "en-US")
			lang = video.DetectLanguageFromSubtitlePath(videoSubtitles)
		}
		if lang == "" {
			lang = "en-US" // Default fallback
		}

		// Convert BCP-47 to ISO 639-2 for ffmpeg
		isoLang := video.BCP47ToISO639(lang)

		fmt.Printf("\nEmbedding subtitles from: %s\n", videoSubtitles)
		fmt.Printf("  Language: %s (%s)\n", lang, isoLang)

		// Create temp output file
		tempOutput := videoOutputFile + ".with-subs.mp4"

		// Run ffmpeg to embed subtitles
		if err := video.EmbedSubtitles(videoOutputFile, videoSubtitles, isoLang, tempOutput); err != nil {
			return fmt.Errorf("failed to embed subtitles: %w", err)
		}

		// Replace original with subtitled version
		if err := os.Remove(videoOutputFile); err != nil {
			return fmt.Errorf("failed to remove original video: %w", err)
		}
		if err := os.Rename(tempOutput, videoOutputFile); err != nil {
			return fmt.Errorf("failed to rename subtitled video: %w", err)
		}

		fmt.Printf("✓ Subtitles embedded\n")
	}

	fmt.Printf("\n✓ Success! Video saved to: %s\n", videoOutputFile)
	return nil
}

// checkDependencies verifies all required tools are installed
func checkDependencies() error {
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
