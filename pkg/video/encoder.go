package video

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// EncoderType represents the type of video encoder to use.
type EncoderType string

const (
	// EncoderSoftware uses libx264 (CPU-based, universal compatibility)
	EncoderSoftware EncoderType = "software"
	// EncoderHardware uses hardware acceleration when available
	EncoderHardware EncoderType = "hardware"
	// EncoderAuto automatically selects the best available encoder
	EncoderAuto EncoderType = "auto"
)

// EncoderConfig holds configuration for video encoding.
type EncoderConfig struct {
	// Type specifies which encoder to use
	Type EncoderType
	// Preset for encoding speed/quality tradeoff (for libx264: ultrafast, fast, medium, slow)
	Preset string
	// CRF (Constant Rate Factor) for quality (lower = better, 18-28 typical)
	CRF int
}

// DefaultEncoderConfig returns a default encoder configuration.
func DefaultEncoderConfig() EncoderConfig {
	return EncoderConfig{
		Type:   EncoderSoftware,
		Preset: "fast",
		CRF:    23,
	}
}

// FastEncoderConfig returns a configuration optimized for speed using hardware encoding.
func FastEncoderConfig() EncoderConfig {
	return EncoderConfig{
		Type:   EncoderHardware,
		Preset: "fast",
		CRF:    23,
	}
}

// Global encoder settings (can be set once at startup)
var (
	globalEncoderConfig = DefaultEncoderConfig()
	encoderConfigMu     sync.RWMutex
)

// SetGlobalEncoderConfig sets the global encoder configuration.
func SetGlobalEncoderConfig(config EncoderConfig) {
	encoderConfigMu.Lock()
	defer encoderConfigMu.Unlock()
	globalEncoderConfig = config
}

// GetGlobalEncoderConfig returns the current global encoder configuration.
func GetGlobalEncoderConfig() EncoderConfig {
	encoderConfigMu.RLock()
	defer encoderConfigMu.RUnlock()
	return globalEncoderConfig
}

// GetVideoCodec returns the FFmpeg video codec arguments based on the encoder config.
// Returns the codec name and any additional arguments.
func GetVideoCodec(config EncoderConfig) (codec string, args []string) {
	switch config.Type {
	case EncoderHardware:
		// Try hardware encoder
		if hwCodec, hwArgs, ok := getHardwareEncoder(); ok {
			return hwCodec, hwArgs
		}
		// Fall back to software
		return getSoftwareEncoder(config)
	case EncoderAuto:
		// Try hardware first, fall back to software
		if hwCodec, hwArgs, ok := getHardwareEncoder(); ok {
			return hwCodec, hwArgs
		}
		return getSoftwareEncoder(config)
	default:
		return getSoftwareEncoder(config)
	}
}

// getSoftwareEncoder returns the libx264 encoder settings.
func getSoftwareEncoder(config EncoderConfig) (string, []string) {
	preset := config.Preset
	if preset == "" {
		preset = "fast"
	}
	return "libx264", []string{"-preset", preset}
}

// Hardware encoder detection cache
var (
	hwEncoderOnce   sync.Once
	hwEncoderCodec  string
	hwEncoderArgs   []string
	hwEncoderAvail  bool
)

// getHardwareEncoder returns the best available hardware encoder.
func getHardwareEncoder() (codec string, args []string, available bool) {
	hwEncoderOnce.Do(func() {
		hwEncoderCodec, hwEncoderArgs, hwEncoderAvail = detectHardwareEncoder()
	})
	return hwEncoderCodec, hwEncoderArgs, hwEncoderAvail
}

// detectHardwareEncoder probes FFmpeg to find available hardware encoders.
func detectHardwareEncoder() (codec string, args []string, available bool) {
	ctx := context.Background()

	// Check platform-specific encoders
	switch runtime.GOOS {
	case "darwin":
		// macOS: VideoToolbox
		if checkEncoder(ctx, "h264_videotoolbox") {
			return "h264_videotoolbox", []string{"-q:v", "65"}, true
		}
	case "linux":
		// Linux: NVIDIA NVENC (most common), then VAAPI
		if checkEncoder(ctx, "h264_nvenc") {
			return "h264_nvenc", []string{"-preset", "p4", "-tune", "ll"}, true
		}
		if checkEncoder(ctx, "h264_vaapi") {
			return "h264_vaapi", []string{"-qp", "23"}, true
		}
	case "windows":
		// Windows: NVIDIA NVENC, then AMD AMF, then Intel QSV
		if checkEncoder(ctx, "h264_nvenc") {
			return "h264_nvenc", []string{"-preset", "p4", "-tune", "ll"}, true
		}
		if checkEncoder(ctx, "h264_amf") {
			return "h264_amf", []string{"-quality", "speed"}, true
		}
		if checkEncoder(ctx, "h264_qsv") {
			return "h264_qsv", []string{"-preset", "fast"}, true
		}
	}

	return "", nil, false
}

// checkEncoder tests if an encoder is available in FFmpeg.
func checkEncoder(ctx context.Context, encoder string) bool {
	// Try to get encoder info
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), encoder)
}

// GetEncoderDescription returns a human-readable description of the encoder being used.
func GetEncoderDescription(config EncoderConfig) string {
	codec, _ := GetVideoCodec(config)
	switch codec {
	case "h264_videotoolbox":
		return "VideoToolbox (macOS hardware)"
	case "h264_nvenc":
		return "NVENC (NVIDIA hardware)"
	case "h264_vaapi":
		return "VAAPI (Linux hardware)"
	case "h264_amf":
		return "AMF (AMD hardware)"
	case "h264_qsv":
		return "Quick Sync (Intel hardware)"
	case "libx264":
		return "libx264 (software)"
	default:
		return codec
	}
}
