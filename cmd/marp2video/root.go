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
	Short: "Convert Marp presentations to video with AI voiceovers",
	Long: `marp2video transforms Marp markdown presentations into professional videos
with AI-generated voiceovers using ElevenLabs text-to-speech.

Use subcommands to run specific stages of the pipeline:
  tts    - Generate audio from transcript
  video  - Generate video from presentation (full pipeline)

Examples:
  # Full pipeline (original behavior)
  marp2video video --input slides.md --output video.mp4

  # Generate audio only from transcript
  marp2video tts --transcript transcript.json --output audio/ --lang en-US

  # Generate video using pre-generated audio
  marp2video video --input slides.md --manifest audio/manifest.json --output video.mp4`,
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
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
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
