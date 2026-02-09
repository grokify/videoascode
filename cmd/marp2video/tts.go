package main

import (
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

By default, existing audio files are skipped (not regenerated). This allows
resuming interrupted runs without re-generating already completed slides.
Use --force to regenerate all audio files.

Examples:
  # Generate audio for default language (skips existing files)
  marp2video tts --transcript transcript.json --output audio/

  # Resume an interrupted run (automatically skips completed slides)
  marp2video tts --transcript transcript.json --output audio/

  # Force regeneration of all audio files
  marp2video tts --transcript transcript.json --output audio/ --force

  # Generate audio for specific language
  marp2video tts --transcript transcript.json --output audio/ --lang es-ES`,
	RunE: runTTS,
}

var (
	ttsTranscriptFile string
	ttsOutputDir      string
	ttsLanguage       string
	ttsAPIKey         string
	ttsForce          bool
)

func init() {
	ttsCmd.Flags().StringVarP(&ttsTranscriptFile, "transcript", "t", "", "Transcript JSON file (required)")
	ttsCmd.Flags().StringVarP(&ttsOutputDir, "output", "o", "audio", "Output directory for audio files")
	ttsCmd.Flags().StringVarP(&ttsLanguage, "lang", "l", "", "Language/locale code (e.g., en-US, es-ES)")
	ttsCmd.Flags().StringVarP(&ttsAPIKey, "api-key", "k", "", "ElevenLabs API key (or use ELEVENLABS_API_KEY env var)")
	ttsCmd.Flags().BoolVarP(&ttsForce, "force", "f", false, "Regenerate audio even if files already exist")

	if err := ttsCmd.MarkFlagRequired("transcript"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(ttsCmd)
}

func runTTS(cmd *cobra.Command, args []string) error {
	ctx := newContext()

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

	// Create generator config
	genConfig := tts.TranscriptGeneratorConfig{
		APIKey:    apiKey,
		OutputDir: ttsOutputDir,
		Force:     ttsForce,
	}

	// Only show progress bar when not in verbose mode (logs provide progress info)
	var renderer *progress.SingleStageRenderer
	if !verbose {
		renderer = progress.NewSingleStageRenderer(os.Stderr).WithBarWidth(30)
		genConfig.ProgressFunc = func(current, total int, name string) {
			renderer.Update(current, total, name)
		}
	}

	generator, err := tts.NewTranscriptGenerator(genConfig)
	if err != nil {
		return fmt.Errorf("failed to create TTS generator: %w", err)
	}

	// Generate audio
	manifest, err := generator.GenerateFromTranscript(ctx, t, language)

	// Clear progress line if we were showing it
	if renderer != nil {
		renderer.Done("")
	}

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
