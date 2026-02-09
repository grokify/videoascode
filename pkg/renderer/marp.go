package renderer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// MarpRenderer handles conversion of Marp markdown to HTML
type MarpRenderer struct {
	marpCLIPath string
}

// NewMarpRenderer creates a new Marp renderer
func NewMarpRenderer() *MarpRenderer {
	return &MarpRenderer{
		marpCLIPath: "marp", // assumes marp is in PATH
	}
}

// RenderToHTML converts Marp markdown to HTML presentation
func (r *MarpRenderer) RenderToHTML(inputPath, outputDir string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate output path
	outputPath := filepath.Join(outputDir, "presentation.html")

	// Run marp CLI
	cmd := exec.Command(r.marpCLIPath, // #nosec G204 -- marpCLIPath is fixed, paths are user-provided intentionally
		inputPath,
		"-o", outputPath,
		"--html",
		"--allow-local-files",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("marp CLI failed: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}

// RenderToImages converts Marp markdown to PNG images (one per slide)
func (r *MarpRenderer) RenderToImages(inputPath, outputDir string) ([]string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate output path pattern
	outputPath := filepath.Join(outputDir, "slide.png")

	// Run marp CLI with --images png
	cmd := exec.Command(r.marpCLIPath, // #nosec G204 -- marpCLIPath is fixed, paths are user-provided intentionally
		inputPath,
		"-o", outputPath,
		"--images", "png",
		"--allow-local-files",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("marp CLI failed: %w\nOutput: %s", err, string(output))
	}

	// Find all generated images (marp creates slide.001.png, slide.002.png, etc.)
	pattern := filepath.Join(outputDir, "slide.*.png")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find generated images: %w", err)
	}

	// Sort by filename to ensure correct order
	// The files are named slide.001.png, slide.002.png, etc.
	// Glob returns them sorted, but let's be explicit
	if len(matches) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	return matches, nil
}

// CheckMarpCLI verifies that Marp CLI is installed
func (r *MarpRenderer) CheckMarpCLI() error {
	cmd := exec.Command(r.marpCLIPath, "--version") // #nosec G204 -- marpCLIPath is fixed to "marp"
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("marp CLI not found. Install with: npm install -g @marp-team/marp-cli")
	}
	return nil
}
