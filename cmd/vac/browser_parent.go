package main

import (
	"github.com/spf13/cobra"
)

var browserParentCmd = &cobra.Command{
	Use:   "browser",
	Short: "Commands for browser-based video recording",
	Long: `Commands for recording browser automation sessions as video.

Subcommands:
  record   Record browser session (silent, no voiceover)
  video    Record browser session with TTS voiceover

Examples:
  # Silent recording from steps file
  marp2video browser record --url https://example.com --steps demo.json --output demo.mp4

  # Recording with voiceover from config
  marp2video browser video --config demo.yaml --output demo.mp4

  # Multi-language with audio caching
  marp2video browser video --config demo.yaml --output demo.mp4 --audio-dir ./audio --lang en-US,fr-FR`,
}

func init() {
	rootCmd.AddCommand(browserParentCmd)
}
