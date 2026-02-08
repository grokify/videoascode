package video

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCombiner(t *testing.T) {
	c := NewCombiner("/tmp/video")
	if c.outputDir != "/tmp/video" {
		t.Errorf("NewCombiner() outputDir = %s, want /tmp/video", c.outputDir)
	}
}

func TestCombineVideos_EmptyList(t *testing.T) {
	c := NewCombiner("/tmp")
	err := c.CombineVideos(context.Background(), []string{}, "/tmp/output.mp4")
	if err == nil {
		t.Error("CombineVideos() with empty list should return error")
	}
}

func TestCombineVideosWithTransitions_EmptyList(t *testing.T) {
	c := NewCombiner("/tmp")
	err := c.CombineVideosWithTransitions(context.Background(), []string{}, "/tmp/output.mp4", 0.5)
	if err == nil {
		t.Error("CombineVideosWithTransitions() with empty list should return error")
	}
}

func TestCombineVideosWithTransitions_SingleVideo(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "combiner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a fake source video file
	srcPath := filepath.Join(tmpDir, "source.mp4")
	if err := os.WriteFile(srcPath, []byte("fake video content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test single video case (should just copy)
	c := NewCombiner(tmpDir)
	dstPath := filepath.Join(tmpDir, "output.mp4")
	err = c.CombineVideosWithTransitions(context.Background(), []string{srcPath}, dstPath, 0.5)
	if err != nil {
		t.Errorf("CombineVideosWithTransitions() single video error = %v", err)
	}

	// Verify the file was copied
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("CombineVideosWithTransitions() single video did not create output file")
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "copyfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy the file
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Errorf("copyFile() error = %v", err)
	}

	// Verify content
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("Failed to read dest file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("copyFile() content = %s, want %s", string(got), string(content))
	}
}

func TestCopyFile_SourceNotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = copyFile("/nonexistent/file.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("copyFile() with nonexistent source should return error")
	}
}
