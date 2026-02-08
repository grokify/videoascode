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
	cmd := exec.Command(r.marpCLIPath,
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

// CheckMarpCLI verifies that Marp CLI is installed
func (r *MarpRenderer) CheckMarpCLI() error {
	cmd := exec.Command(r.marpCLIPath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("marp CLI not found. Install with: npm install -g @marp-team/marp-cli")
	}
	return nil
}
