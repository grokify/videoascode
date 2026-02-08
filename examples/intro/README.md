# Intro Example

A self-documenting presentation that introduces marp2video.

## Files

| File | Description |
|------|-------------|
| `presentation.md` | Marp markdown source (13 slides) |
| `transcript.json` | Multi-language voiceover transcript |
| `audio/` | Generated audio files (after Step 1) |
| `output.mp4` | Final video (after full pipeline) |

## Step 1: Generate Audio from Transcript

Convert the transcript segments to audio using OmniVoice TTS providers.

### Using ElevenLabs (via OmniVoice)

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/agentplexus/go-elevenlabs/omnivoice/tts"
    omnitts "github.com/agentplexus/omnivoice/tts"
)

func main() {
    ctx := context.Background()

    // Initialize ElevenLabs provider
    provider, err := tts.New(tts.WithAPIKey(os.Getenv("ELEVENLABS_API_KEY")))
    if err != nil {
        panic(err)
    }

    // Load transcript
    data, _ := os.ReadFile("transcript.json")
    var transcript Transcript
    json.Unmarshal(data, &transcript)

    // Process each slide for the target language
    lang := "en-US" // or "en-GB", "es-ES"
    for _, slide := range transcript.Slides {
        content, ok := slide.Transcripts[lang]
        if !ok {
            continue
        }

        // Combine segments into full text
        var text string
        for _, seg := range content.Segments {
            text += seg.Text + " "
        }

        // Determine voice config
        voiceID := transcript.Metadata.DefaultVoice.VoiceID
        if content.Voice != nil && content.Voice.VoiceID != "" {
            voiceID = content.Voice.VoiceID
        }

        // Synthesize audio
        config := omnitts.SynthesisConfig{
            VoiceID:         voiceID,
            Model:           "eleven_multilingual_v2",
            OutputFormat:    "mp3",
            SampleRate:      44100,
            Stability:       0.5,
            SimilarityBoost: 0.75,
        }

        result, err := provider.Synthesize(ctx, text, config)
        if err != nil {
            panic(err)
        }

        // Save audio file
        outPath := filepath.Join("audio", fmt.Sprintf("slide_%03d.mp3", slide.Index))
        os.MkdirAll("audio", 0755)
        os.WriteFile(outPath, result.Audio, 0644)

        fmt.Printf("Generated %s (%d bytes)\n", outPath, len(result.Audio))
    }
}

// Transcript types (simplified)
type Transcript struct {
    Metadata struct {
        DefaultVoice struct {
            VoiceID string `json:"voiceId"`
        } `json:"defaultVoice"`
    } `json:"metadata"`
    Slides []struct {
        Index       int                        `json:"index"`
        Transcripts map[string]LanguageContent `json:"transcripts"`
    } `json:"slides"`
}

type LanguageContent struct {
    Voice    *VoiceConfig `json:"voice"`
    Segments []Segment    `json:"segments"`
}

type VoiceConfig struct {
    VoiceID string `json:"voiceId"`
}

type Segment struct {
    Text  string `json:"text"`
    Pause int    `json:"pause"`
}
```

### Using OmniVoice Client with Fallback

```go
// Create client with multiple providers (when available)
client := omnitts.NewClient(
    elevenLabsProvider,
    // deepgramProvider,  // Future: Deepgram TTS
    // openaiProvider,    // Future: OpenAI TTS
)

// Synthesize with automatic fallback
result, err := client.Synthesize(ctx, text, config)
```

## Step 2: Generate Video

Once audio files are generated, use marp2video to create the video.

### Using Inline Voiceovers

```bash
marp2video \
  --input presentation.md \
  --output output.mp4
```

### Using Pre-generated Audio

```bash
# Audio files in audio/ directory will be used automatically
marp2video \
  --input presentation.md \
  --audio-dir audio/ \
  --output output.mp4
```

## Transcript Structure

### Root Object

```json
{
  "version": "1.0",
  "metadata": { ... },
  "slides": [ ... ]
}
```

### Metadata

```json
{
  "metadata": {
    "title": "Introduction to marp2video",
    "defaultLanguage": "en-US",
    "defaultVoice": {
      "provider": "elevenlabs",
      "voiceId": "pNInz6obpgDQGcFmaJgB",
      "voiceName": "Adam",
      "model": "eleven_multilingual_v2",
      "stability": 0.5,
      "similarityBoost": 0.75
    },
    "defaultVenue": "youtube"
  }
}
```

### Slide with Multi-Language Transcripts

```json
{
  "index": 0,
  "title": "Title Slide",
  "transcripts": {
    "en-US": {
      "segments": [
        { "text": "Welcome to marp2video...", "pause": 500 },
        { "text": "In this presentation..." }
      ]
    },
    "es-ES": {
      "voice": {
        "voiceId": "onwK4e9ZLuTAKqWW03F9",
        "voiceName": "Daniel"
      },
      "segments": [
        { "text": "Bienvenido a marp2video...", "pause": 500 },
        { "text": "En esta presentación..." }
      ]
    }
  }
}
```

## Available Languages

| Locale | Language | Voice |
|--------|----------|-------|
| `en-US` | American English | Adam (default) |
| `en-GB` | British English | Adam |
| `es-ES` | Spanish | Daniel |

## ElevenLabs Voice IDs

| Voice ID | Name | Accent |
|----------|------|--------|
| `pNInz6obpgDQGcFmaJgB` | Adam | American |
| `21m00Tcm4TlvDq8ikWAM` | Rachel | American |
| `onwK4e9ZLuTAKqWW03F9` | Daniel | British |
| `XrExE9yKIg1WjnnlVkGX` | Matilda | American |

## OmniVoice SynthesisConfig

| Field | Type | Description |
|-------|------|-------------|
| `VoiceID` | string | Provider-specific voice ID |
| `Model` | string | TTS model (e.g., `eleven_multilingual_v2`) |
| `OutputFormat` | string | Audio format (`mp3`, `pcm`, `wav`) |
| `SampleRate` | int | Sample rate in Hz (22050, 44100) |
| `Stability` | float64 | Voice consistency (0.0 - 1.0) |
| `SimilarityBoost` | float64 | Voice similarity (0.0 - 1.0) |

## Handling Pauses

Pauses specified in segments can be handled by:

1. **Silence injection**: Add silence to the audio file
2. **Slide timing**: Extend slide duration during recording
3. **Post-processing**: Add silence between audio clips

```go
// Calculate total duration with pauses
var totalPauseMs int
for _, seg := range content.Segments {
    totalPauseMs += seg.Pause
}

// Audio duration + pause duration = slide duration
slideDuration := result.DurationMs + totalPauseMs
```

## Future: Deepgram TTS

When Deepgram OmniVoice provider is available:

```go
import deepgramtts "github.com/agentplexus/go-deepgram/omnivoice/tts"

provider, _ := deepgramtts.New(
    deepgramtts.WithAPIKey(os.Getenv("DEEPGRAM_API_KEY")),
)

config := omnitts.SynthesisConfig{
    VoiceID:      "aura-asteria-en",  // Deepgram Aura voice
    OutputFormat: "mp3",
    SampleRate:   24000,
}
```

## Complete Workflow

```bash
# 1. Set API key
export ELEVENLABS_API_KEY="your-key"

# 2. Generate audio (using Go script above or marp2video)
go run generate_audio.go

# 3. Generate video
marp2video \
  --input presentation.md \
  --output output.mp4

# 4. For different languages
marp2video \
  --input presentation.md \
  --transcript transcript.json \
  --lang es-ES \
  --output output_es.mp4
```
