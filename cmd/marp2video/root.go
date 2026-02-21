package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/grokify/mogo/log/slogutil"
	"github.com/spf13/cobra"
)

const version = "0.2.0"

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "marp2video",
	Short: "Convert Marp presentations and browser demos to video with AI voiceovers",
	Long: `marp2video creates professional videos with AI-generated voiceovers.

Two main workflows:

  slides   - Marp markdown presentations
  browser  - Browser automation recordings

Additional commands:

  subtitle - Generate subtitles from audio

Examples:
  # Marp slides: full pipeline
  marp2video slides video --input slides.md --output video.mp4

  # Marp slides: generate audio only
  marp2video slides tts --transcript transcript.json --output audio/

  # Browser: record with voiceover
  marp2video browser video --config demo.yaml --output demo.mp4

  # Browser: silent recording
  marp2video browser record --url https://example.com --steps demo.json --output demo.mp4

  # Generate subtitles
  marp2video subtitle --audio audio/en-US/`,
	Version: version,
}

func init() {
	rootCmd.SetVersionTemplate("marp2video version {{.Version}}\n")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Enable verbose logging")
}

// newContext returns a context with a logger if verbose mode is enabled
func newContext() context.Context {
	ctx := context.Background()
	if verbose {
		// Use stderr for logs to avoid disrupting progress bar on stdout
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		ctx = slogutil.ContextWithLogger(ctx, logger)
	}
	return ctx
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
