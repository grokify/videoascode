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
	if err := os.WriteFile(srcPath, []byte("fake video content"), 0600); err != nil {
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
	if err := os.WriteFile(srcPath, content, 0600); err != nil {
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

func TestCombineVideos_SingleVideo(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "combiner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a fake source video file
	srcPath := filepath.Join(tmpDir, "source.mp4")
	if err := os.WriteFile(srcPath, []byte("fake video content"), 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test single video case (should just copy)
	c := NewCombiner(tmpDir)
	dstPath := filepath.Join(tmpDir, "output.mp4")
	err = c.CombineVideos(context.Background(), []string{srcPath}, dstPath)
	if err != nil {
		t.Errorf("CombineVideos() single video error = %v", err)
	}

	// Verify the file was copied
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("CombineVideos() single video did not create output file")
	}

	// Verify content matches
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}
	if string(got) != "fake video content" {
		t.Errorf("CombineVideos() single video content mismatch")
	}
}

// TestCombineVideos_FilterComplexConcat documents that CombineVideos uses
// ffmpeg's filter_complex concat filter instead of the concat demuxer.
// This is important because:
//   - Different TTS providers output different audio sample rates
//     (e.g., ElevenLabs: 44100 Hz, Deepgram: 22050 Hz)
//   - The concat demuxer with -c copy cannot handle mixed sample rates
//   - filter_complex concat properly decodes and re-encodes all streams
//
// See: https://trac.ffmpeg.org/wiki/Concatenate
func TestCombineVideos_FilterComplexConcat(t *testing.T) {
	// This is a documentation test - actual ffmpeg invocation would require
	// real video files. The implementation uses filter_complex concat:
	//
	// ffmpeg -i input1.mp4 -i input2.mp4 ... \
	//   -filter_complex "[0:v][0:a][1:v][1:a]...concat=n=N:v=1:a=1[outv][outa]" \
	//   -map "[outv]" -map "[outa]" \
	//   -c:v libx264 -c:a aac -ar 44100 output.mp4
	//
	// This ensures all audio is normalized to 44100 Hz regardless of input.

	// Verify the combiner is constructed correctly
	c := NewCombiner("/tmp/test")
	if c.outputDir != "/tmp/test" {
		t.Errorf("NewCombiner() outputDir = %s, want /tmp/test", c.outputDir)
	}
}

// TestMixedAudioSampleRates_Documentation documents the mixed sample rate issue
// and its solution. This test serves as documentation for future maintainers.
func TestMixedAudioSampleRates_Documentation(t *testing.T) {
	// Problem:
	// When generating videos with audio from multiple TTS providers,
	// the audio streams may have different sample rates:
	//   - ElevenLabs: 44100 Hz (high quality)
	//   - Deepgram: 22050 Hz (standard)
	//   - OpenAI: 24000 Hz (standard)
	//
	// The ffmpeg concat demuxer (-f concat) with -c copy cannot handle
	// streams with different parameters. It treats all streams as having
	// the same format as the first stream, causing:
	//   - Corrupted audio data
	//   - "Sample rate index does not match" errors
	//   - Audio duration mismatches (e.g., audio half the video length)
	//
	// Solution:
	// Use filter_complex concat instead of the concat demuxer.
	// This approach:
	//   1. Decodes all input streams
	//   2. Concatenates them in the filter graph
	//   3. Re-encodes to consistent output format (44100 Hz AAC)
	//
	// Trade-off:
	// Re-encoding takes more time than stream copying, but ensures
	// compatibility across all TTS providers.

	t.Log("Mixed audio sample rate handling is documented in combiner.go")
}
