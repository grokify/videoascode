package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grokify/marp2video/pkg/transcript"
	"github.com/grokify/marp2video/pkg/tts"
	"github.com/grokify/mogo/fmt/progress"
	"github.com/spf13/cobra"
)

var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "Generate audio from transcript",
	Long: `Generate audio files from a transcript JSON file using ElevenLabs TTS.

This command processes a transcript.json file and generates:
  - One MP3 audio file per slide
  - A manifest.json with timing information

The manifest can be used by the 'video' command to synchronize
slide recordings with the pre-generated audio.

Examples:
  # Generate audio for default language
  marp2video tts --transcript transcript.json --output audio/

  # Generate audio for specific language
  marp2video tts --transcript transcript.json --output audio/ --lang es-ES

  # Use custom API key
  marp2video tts --transcript transcript.json --output audio/ --api-key YOUR_KEY`,
	RunE: runTTS,
}

var (
	ttsTranscriptFile string
	ttsOutputDir      string
	ttsLanguage       string
	ttsAPIKey         string
)

func init() {
	ttsCmd.Flags().StringVarP(&ttsTranscriptFile, "transcript", "t", "", "Transcript JSON file (required)")
	ttsCmd.Flags().StringVarP(&ttsOutputDir, "output", "o", "audio", "Output directory for audio files")
	ttsCmd.Flags().StringVarP(&ttsLanguage, "lang", "l", "", "Language/locale code (e.g., en-US, es-ES)")
	ttsCmd.Flags().StringVarP(&ttsAPIKey, "api-key", "k", "", "ElevenLabs API key (or use ELEVENLABS_API_KEY env var)")

	if err := ttsCmd.MarkFlagRequired("transcript"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(ttsCmd)
}

func runTTS(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate transcript file exists
	if _, err := os.Stat(ttsTranscriptFile); os.IsNotExist(err) {
		return fmt.Errorf("transcript file does not exist: %s", ttsTranscriptFile)
	}

	// Get API key
	apiKey := ttsAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("ELEVENLABS_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("ElevenLabs API key required: use --api-key flag or ELEVENLABS_API_KEY env var")
	}

	// Load transcript
	t, err := transcript.LoadFromFile(ttsTranscriptFile)
	if err != nil {
		return fmt.Errorf("failed to load transcript: %w", err)
	}

	// Determine language
	language := ttsLanguage
	if language == "" {
		language = t.Metadata.DefaultLanguage
	}
	if language == "" {
		return fmt.Errorf("no language specified and no default language in transcript")
	}

	fmt.Printf("Generating audio from transcript\n")
	fmt.Printf("  Transcript: %s\n", ttsTranscriptFile)
	fmt.Printf("  Language:   %s\n", language)
	fmt.Printf("  Output:     %s\n", ttsOutputDir)
	fmt.Printf("  Slides:     %d\n\n", len(t.Slides))

	// Create progress renderer
	renderer := progress.NewSingleStageRenderer(os.Stdout).WithBarWidth(30)

	// Progress callback
	progressFn := func(current, total int, name string) {
		renderer.Update(current, total, name)
	}

	// Create generator with progress callback
	generator, err := tts.NewTranscriptGenerator(tts.TranscriptGeneratorConfig{
		APIKey:       apiKey,
		OutputDir:    ttsOutputDir,
		ProgressFunc: progressFn,
	})
	if err != nil {
		return fmt.Errorf("failed to create TTS generator: %w", err)
	}

	// Generate audio
	manifest, err := generator.GenerateFromTranscript(ctx, t, language)

	// Clear progress line
	renderer.Done("")

	if err != nil {
		return fmt.Errorf("failed to generate audio: %w", err)
	}

	// Save manifest
	manifestPath := filepath.Join(ttsOutputDir, "manifest.json")
	if err := manifest.SaveToFile(manifestPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Print summary
	fmt.Printf("\n✓ Generated %d audio files\n", len(manifest.Slides))
	fmt.Printf("✓ Manifest saved to: %s\n", manifestPath)
	fmt.Printf("✓ Total duration: %.1f seconds\n", float64(manifest.GetTotalDuration())/1000)

	return nil
}
