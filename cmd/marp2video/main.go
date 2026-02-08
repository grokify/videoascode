package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/grokify/marp2video/pkg/orchestrator"
	"github.com/grokify/marp2video/pkg/renderer"
	"github.com/grokify/marp2video/pkg/video"
)

const version = "0.1.0"

func main() {
	// Define CLI flags
	inputFile := flag.String("input", "", "Input Marp markdown file (required)")
	outputFile := flag.String("output", "output.mp4", "Output video file")
	apiKey := flag.String("api-key", "", "ElevenLabs API key (or use ELEVENLABS_API_KEY env var)")
	voiceID := flag.String("voice", "pNInz6obpgDQGcFmaJgB", "ElevenLabs voice ID (default: Adam)")
	width := flag.Int("width", 1920, "Video width")
	height := flag.Int("height", 1080, "Video height")
	fps := flag.Int("fps", 30, "Video frame rate")
	workDir := flag.String("workdir", "", "Working directory for temporary files")
	outputIndividual := flag.String("output-individual", "", "Directory to save individual slide videos (for Udemy)")
	transitionDuration := flag.Float64("transition", 0, "Transition duration between slides in seconds (0 = no transitions)")
	screenDevice := flag.String("screen-device", "", "Screen capture device (macOS only, auto-detected if empty)")
	showVersion := flag.Bool("version", false, "Show version")
	checkDeps := flag.Bool("check", false, "Check dependencies")

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("marp2video version %s\n", version)
		os.Exit(0)
	}

	// Check dependencies
	if *checkDeps {
		if err := checkDependencies(); err != nil {
			log.Fatalf("Dependency check failed: %v", err)
		}
		fmt.Println("✓ All dependencies are installed")
		os.Exit(0)
	}

	// Validate required flags
	if *inputFile == "" {
		fmt.Println("Error: --input is required")
		flag.Usage()
		os.Exit(1)
	}

	// Check if input file exists
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		log.Fatalf("Input file does not exist: %s", *inputFile)
	}

	// Get API key from flag or environment
	elevenLabsAPIKey := *apiKey
	if elevenLabsAPIKey == "" {
		elevenLabsAPIKey = os.Getenv("ELEVENLABS_API_KEY")
	}
	if elevenLabsAPIKey == "" {
		log.Fatal("ElevenLabs API key required: use --api-key flag or ELEVENLABS_API_KEY env var")
	}

	// Set working directory
	workingDir := *workDir
	if workingDir == "" {
		workingDir = filepath.Join(os.TempDir(), "marp2video")
	}

	// Create orchestrator config
	config := orchestrator.Config{
		InputFile:           *inputFile,
		OutputFile:          *outputFile,
		WorkDir:             workingDir,
		ElevenLabsAPIKey:    elevenLabsAPIKey,
		VoiceID:             *voiceID,
		Width:               *width,
		Height:              *height,
		FrameRate:           *fps,
		OutputIndividualDir: *outputIndividual,
		TransitionDuration:  *transitionDuration,
		ScreenDevice:        *screenDevice,
	}

	// Check dependencies before starting
	if err := checkDependencies(); err != nil {
		log.Fatalf("Dependency check failed: %v\nRun 'marp2video --check' for details", err)
	}

	// Create and run orchestrator
	orch := orchestrator.NewOrchestrator(config)

	ctx := context.Background()
	if err := orch.Process(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("\n✓ Success! Video saved to: %s\n", *outputFile)
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
