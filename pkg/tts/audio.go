package tts

import (
	"time"

	"github.com/grokify/ffutil"
)

// getAudioDuration uses ffprobe to get the duration of an audio file
func getAudioDuration(filePath string) (time.Duration, error) {
	return ffutil.Duration(filePath)
}
