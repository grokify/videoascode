package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grokify/videoascode/pkg/browser"
	"github.com/grokify/videoascode/pkg/config"
	"github.com/spf13/cobra"
)

var browserRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record browser session (silent)",
	Long: `Execute browser automation steps and capture as video (no audio).

This command drives a web browser through a sequence of steps (navigate, click,
input, wait, etc.) and captures the session as a silent video file.

For browser recording with TTS voiceover, use 'vac browser video' instead.

The steps can be defined in a configuration file (JSON or YAML) or passed
directly via a steps file.

Examples:
  # Record from a steps file
  vac browser record --url https://example.com --steps demo-steps.json --output demo.mp4

  # Record from a config file
  vac browser record --config demo-config.yaml --output demo.mp4

  # Record with custom resolution and visible browser
  vac browser record --url https://example.com --steps demo-steps.json --output demo.mp4 --width 1920 --height 1080

  # Export timing data for later audio sync
  vac browser record --url https://example.com --steps demo-steps.json --output demo.mp4 --timing timing.json`,
	RunE: runBrowserRecord,
}

var (
	brConfigFile string
	brStepsFile  string
	brURL        string
	brOutputFile string
	brWidth      int
	brHeight     int
	brFPS        int
	brWorkDir    string
	brHeadless   bool
	brTimingFile string
	brTimeout    int
	brCleanup    bool
)

func init() {
	browserRecordCmd.Flags().StringVarP(&brConfigFile, "config", "c", "", "Configuration file (YAML/JSON) with segments")
	browserRecordCmd.Flags().StringVarP(&brStepsFile, "steps", "s", "", "Steps file (JSON/YAML) defining browser actions")
	browserRecordCmd.Flags().StringVarP(&brURL, "url", "u", "", "Starting URL for the browser")
	browserRecordCmd.Flags().StringVarP(&brOutputFile, "output", "o", "recording.mp4", "Output video file")
	browserRecordCmd.Flags().IntVar(&brWidth, "width", 1920, "Browser viewport width")
	browserRecordCmd.Flags().IntVar(&brHeight, "height", 1080, "Browser viewport height")
	browserRecordCmd.Flags().IntVar(&brFPS, "fps", 30, "Video frame rate")
	browserRecordCmd.Flags().StringVar(&brWorkDir, "workdir", "", "Working directory for temporary files")
	browserRecordCmd.Flags().BoolVar(&brHeadless, "headless", false, "Run browser in headless mode (no visible window)")
	browserRecordCmd.Flags().StringVarP(&brTimingFile, "timing", "t", "", "Output timing JSON file for transcript synchronization")
	browserRecordCmd.Flags().IntVar(&brTimeout, "timeout", 30000, "Default step timeout in milliseconds")
	browserRecordCmd.Flags().BoolVar(&brCleanup, "cleanup", true, "Clean up temporary files after recording")

	browserParentCmd.AddCommand(browserRecordCmd)
}

func runBrowserRecord(cmd *cobra.Command, args []string) error {
	ctx := newContext()

	// Load configuration
	var steps []browser.Step
	var startURL string

	if brConfigFile != "" {
		// Load from config file
		cfg, err := config.LoadFromFile(brConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Find browser segments and extract steps
		for _, seg := range cfg.Segments {
			if seg.Type == config.SourceTypeBrowser {
				if seg.URL != "" && startURL == "" {
					startURL = seg.URL
				}
				steps = append(steps, seg.Steps...)
			}
		}

		if len(steps) == 0 {
			return fmt.Errorf("no browser segments found in config file")
		}

		// Apply config defaults
		if brWidth == 1920 && cfg.Resolution.Width > 0 {
			brWidth = cfg.Resolution.Width
		}
		if brHeight == 1080 && cfg.Resolution.Height > 0 {
			brHeight = cfg.Resolution.Height
		}
		if brFPS == 30 && cfg.FrameRate > 0 {
			brFPS = cfg.FrameRate
		}
	} else if brStepsFile != "" {
		// Load steps from file
		seq, err := browser.LoadStepSequence(brStepsFile)
		if err != nil {
			return fmt.Errorf("failed to load steps: %w", err)
		}
		steps = seq.Steps
		startURL = seq.URL
	} else {
		return fmt.Errorf("either --config or --steps is required")
	}

	// Override URL if provided via flag
	if brURL != "" {
		startURL = brURL
	}

	if startURL == "" {
		return fmt.Errorf("starting URL is required (use --url or specify in config/steps file)")
	}

	// Validate steps
	for i, step := range steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step %d: %w", i, err)
		}
	}

	// Set up working directory
	workDir := brWorkDir
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "vac-browser")
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	fmt.Printf("Recording browser session (silent)...\n")
	fmt.Printf("  URL: %s\n", startURL)
	fmt.Printf("  Steps: %d\n", len(steps))
	fmt.Printf("  Resolution: %dx%d\n", brWidth, brHeight)
	fmt.Printf("  Headless: %v\n", brHeadless)
	fmt.Printf("  Output: %s\n", brOutputFile)
	fmt.Println()

	// Create recorder
	recorderCfg := browser.RecorderConfig{
		Width:            brWidth,
		Height:           brHeight,
		OutputDir:        workDir,
		FrameRate:        brFPS,
		Headless:         brHeadless,
		DefaultTimeout:   brTimeout,
		CaptureEveryStep: true,
	}

	recorder, err := browser.NewRecorder(recorderCfg)
	if err != nil {
		return fmt.Errorf("failed to create recorder: %w", err)
	}
	defer recorder.Close()

	// Launch browser
	fmt.Printf("Launching browser...\n")
	if err := recorder.Launch(); err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	// Navigate to starting URL
	fmt.Printf("Navigating to %s...\n", startURL)
	if err := recorder.Navigate(startURL); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Execute steps
	fmt.Printf("Executing %d steps...\n", len(steps))
	results, err := recorder.RecordSteps(ctx, steps)
	if err != nil {
		// Print partial results for debugging
		fmt.Printf("\nRecording stopped after %d steps due to error:\n", len(results))
		fmt.Printf("  Error: %v\n", err)
	}

	// Print step results
	fmt.Println()
	for i, result := range results {
		status := "✓"
		if !result.Success {
			status = "✗"
		}
		desc := result.Step.Description
		if desc == "" {
			desc = string(result.Step.Action)
			if result.Step.Selector != "" {
				desc += " " + result.Step.Selector
			}
		}
		fmt.Printf("  %s Step %d: %s (%dms)\n", status, i+1, desc, result.Duration)
		if result.Error != "" {
			fmt.Printf("      Error: %s\n", result.Error)
		}
	}
	fmt.Println()

	// Generate video
	fmt.Printf("Generating video...\n")
	videoPath, err := recorder.GenerateVideo(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate video: %w", err)
	}

	// Move video to final destination
	if videoPath != brOutputFile {
		// Ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(brOutputFile), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Read and write to handle cross-device moves
		data, err := os.ReadFile(videoPath)
		if err != nil {
			return fmt.Errorf("failed to read generated video: %w", err)
		}
		if err := os.WriteFile(brOutputFile, data, 0600); err != nil { //nolint:gosec // G703: Path from CLI flag
			return fmt.Errorf("failed to write output video: %w", err)
		}
	}

	// Save timing data if requested
	if brTimingFile != "" {
		if err := recorder.SaveTimingData(); err != nil {
			fmt.Printf("Warning: failed to save timing data: %v\n", err)
		} else {
			timingPath := filepath.Join(workDir, fmt.Sprintf("timing_%s.json", recorder.GetTimingData().SessionID))

			// Copy to requested location
			data, err := os.ReadFile(timingPath)
			if err != nil {
				fmt.Printf("Warning: failed to read timing data: %v\n", err)
			} else {
				if err := os.WriteFile(brTimingFile, data, 0600); err != nil { //nolint:gosec // G703: Path from CLI flag
					fmt.Printf("Warning: failed to write timing data: %v\n", err)
				} else {
					fmt.Printf("✓ Timing data saved to: %s\n", brTimingFile)
				}
			}
		}
	}

	// Cleanup
	if brCleanup {
		if err := recorder.Cleanup(); err != nil {
			fmt.Printf("Warning: cleanup failed: %v\n", err)
		}
	}

	fmt.Printf("\n✓ Success! Recording saved to: %s\n", brOutputFile)

	// Print timing summary
	timing := recorder.GetTimingData()
	if timing != nil {
		fmt.Printf("\nRecording summary:\n")
		fmt.Printf("  Total duration: %dms (%.1fs)\n", timing.TotalDuration, float64(timing.TotalDuration)/1000)
		fmt.Printf("  Frames captured: %d\n", timing.FrameCount)
		fmt.Printf("  Steps completed: %d/%d\n", len(results), len(steps))
	}

	return nil
}
