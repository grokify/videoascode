package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileSecure_ValidPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "writefile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	path := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err = writeFileSecure(path, content)
	if err != nil {
		t.Errorf("writeFileSecure() with valid path error = %v", err)
	}

	// Verify content was written
	got, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("writeFileSecure() content = %s, want %s", string(got), string(content))
	}
}

func TestWriteFileSecure_PathTraversal(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"parent directory", "../test.txt"},
		{"nested parent", "../../test.txt"},
		{"middle traversal", "/tmp/../etc/test.txt"},
		{"hidden traversal", "/tmp/foo/../bar/test.txt"},
		{"double dot only", ".."},
		{"trailing traversal", "/tmp/test/.."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := writeFileSecure(tc.path, []byte("test"))
			if err == nil {
				t.Errorf("writeFileSecure(%q) should return error for path traversal", tc.path)
			}
			if !strings.Contains(err.Error(), "..") {
				t.Errorf("writeFileSecure(%q) error should mention '..', got: %v", tc.path, err)
			}
		})
	}
}

func TestWriteFileSecure_CleanedPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "writefile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Path with redundant slashes should be cleaned
	path := filepath.Join(tmpDir, "subdir", "test.txt")
	dirtyPath := tmpDir + "//subdir///test.txt"

	// Create subdirectory
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	content := []byte("test content")
	err = writeFileSecure(dirtyPath, content)
	if err != nil {
		t.Errorf("writeFileSecure() with dirty path error = %v", err)
	}

	// Verify file exists at cleaned path
	got, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file at cleaned path: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("writeFileSecure() content = %s, want %s", string(got), string(content))
	}
}

func TestWriteFileSecure_InvalidDirectory(t *testing.T) {
	// Writing to non-existent directory should fail
	err := writeFileSecure("/nonexistent/directory/test.txt", []byte("test"))
	if err == nil {
		t.Error("writeFileSecure() to non-existent directory should return error")
	}
}

func TestCopyFile_PathTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Attempt to copy to path with traversal
	err = copyFile(srcPath, "../malicious.txt")
	if err == nil {
		t.Error("copyFile() to path with traversal should return error")
	}
}
