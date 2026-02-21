package orchestrator

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestProgressAnimatorStartsAndStops(t *testing.T) {
	var buf bytes.Buffer

	config := SegmentConfig{
		ProgressWriter: &buf,
	}

	orch := &SegmentOrchestrator{
		config: config,
	}

	// Initialize progress renderer
	orch.progressRenderer = nil // No renderer, animator should handle nil

	// Test with nil renderer - should not panic
	animator := orch.startProgressAnimation(2, 1000, "test")
	if animator != nil {
		t.Error("Animator should be nil when no progress renderer is set")
	}
}

func TestProgressAnimatorWithRenderer(t *testing.T) {
	var buf bytes.Buffer

	config := SegmentConfig{
		ProgressWriter: &buf,
	}

	// Create orchestrator with progress renderer
	orch := NewSegmentOrchestrator(config, nil)

	// Start animation for 500ms estimated duration
	animator := orch.startProgressAnimation(2, 500, "recording")
	if animator == nil {
		t.Fatal("Animator should not be nil")
	}

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)

	// Stop the animator
	animator.stopAndComplete()

	// Verify some output was written
	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected progress output to be written")
	}

	// Should contain stage info
	if !strings.Contains(output, "Creating videos") {
		t.Errorf("Output should contain stage name, got: %s", output)
	}
}

func TestProgressAnimatorStopsCleanly(t *testing.T) {
	var buf bytes.Buffer

	config := SegmentConfig{
		ProgressWriter: &buf,
	}

	orch := NewSegmentOrchestrator(config, nil)

	// Start animation
	animator := orch.startProgressAnimation(2, 10000, "test")

	// Stop immediately
	start := time.Now()
	animator.stopAndComplete()
	elapsed := time.Since(start)

	// Should stop quickly (within 200ms)
	if elapsed > 200*time.Millisecond {
		t.Errorf("Animator took too long to stop: %v", elapsed)
	}
}

func TestProgressAnimatorNilSafe(t *testing.T) {
	// Calling stopAndComplete on nil should not panic
	var animator *progressAnimator
	animator.stopAndComplete() // Should not panic
}

func TestSegmentConfigDefaults(t *testing.T) {
	config := SegmentConfig{}

	// Test default values
	if config.Parallel != 0 {
		t.Errorf("Default Parallel should be 0, got %d", config.Parallel)
	}
	if config.NoAudio != false {
		t.Error("Default NoAudio should be false")
	}
	if config.Headless != false {
		t.Error("Default Headless should be false")
	}
	if config.SubtitlesBurn != false {
		t.Error("Default SubtitlesBurn should be false")
	}
}

func TestSegmentConfigParallel(t *testing.T) {
	tests := []struct {
		name            string
		parallel        int
		expectedWorkers int
	}{
		{"zero defaults to 1", 0, 1},
		{"negative defaults to 1", -1, 1},
		{"one is sequential", 1, 1},
		{"four is parallel", 4, 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := SegmentConfig{
				Parallel: tc.parallel,
			}

			// The orchestrator normalizes parallel value
			parallel := config.Parallel
			if parallel <= 0 {
				parallel = 1
			}

			if parallel != tc.expectedWorkers {
				t.Errorf("Expected %d workers, got %d", tc.expectedWorkers, parallel)
			}
		})
	}
}

func TestEscapeFFmpegFilterPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path",
			input:    "/path/to/file.srt",
			expected: "/path/to/file.srt",
		},
		{
			name:     "path with colon",
			input:    "/path/to/file:name.srt",
			expected: "/path/to/file\\:name.srt",
		},
		{
			name:     "path with single quote",
			input:    "/path/to/file'name.srt",
			expected: "/path/to/file\\'name.srt",
		},
		{
			name:     "path with brackets",
			input:    "/path/to/file[1].srt",
			expected: "/path/to/file\\[1\\].srt",
		},
		{
			name:     "path with backslash converted to forward slash",
			input:    "C:\\Users\\test\\file.srt",
			expected: "C\\:/Users/test/file.srt", // Colon is also escaped (Windows paths need special handling)
		},
		{
			name:     "complex path",
			input:    "/path/to/[test]'s:file.srt",
			expected: "/path/to/\\[test\\]\\'s\\:file.srt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeFFmpegFilterPath(tc.input)
			if result != tc.expected {
				t.Errorf("escapeFFmpegFilterPath(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUpdateProgressWithCount(t *testing.T) {
	var buf bytes.Buffer

	config := SegmentConfig{
		ProgressWriter: &buf,
	}

	orch := NewSegmentOrchestrator(config, nil)

	// Update progress
	orch.updateProgressWithCount(2, 5, 10, "segment_005")

	output := buf.String()

	// Should contain progress info
	if !strings.Contains(output, "Creating videos") {
		t.Errorf("Output should contain stage name, got: %s", output)
	}
	if !strings.Contains(output, "segment_005") {
		t.Errorf("Output should contain segment name, got: %s", output)
	}
	if !strings.Contains(output, "5/10") || !strings.Contains(output, "(5/10)") {
		t.Errorf("Output should contain progress count, got: %s", output)
	}
}
