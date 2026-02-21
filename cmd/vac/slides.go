package main

import (
	"github.com/spf13/cobra"
)

var slidesCmd = &cobra.Command{
	Use:   "slides",
	Short: "Commands for Marp slide presentations",
	Long: `Commands for converting Marp markdown presentations to video.

Subcommands:
  video    Generate video from Marp presentation (full pipeline)
  tts      Generate audio from transcript JSON

Examples:
  # Full pipeline with inline voiceovers
  marp2video slides video --input slides.md --output video.mp4

  # Generate audio only
  marp2video slides tts --transcript transcript.json --output audio/

  # Use pre-generated audio
  marp2video slides video --input slides.md --manifest audio/manifest.json --output video.mp4`,
}

func init() {
	rootCmd.AddCommand(slidesCmd)
}
