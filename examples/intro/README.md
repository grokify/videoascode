# Intro Example

A self-documenting presentation that introduces vac.

## Files

| File | Description |
|------|-------------|
| `presentation.md` | Marp markdown source (13 slides) |
| `transcript.json` | Multi-language voiceover transcript |
| `audio/` | Generated audio files (after Step 1) |
| `audio/manifest.json` | Audio timing manifest (after Step 1) |
| `output.mp4` | Final video (after full pipeline) |

## Step 1: Generate Audio from Transcript

Use the `vac tts` command to generate audio files from the transcript.

### Using CLI (Recommended)

```bash
# Set API key
export ELEVENLABS_API_KEY="your-api-key"

# Generate audio for default language (en-US)
vac tts --transcript transcript.json --output audio/

# Generate audio for specific language
vac tts --transcript transcript.json --output audio/ --lang es-ES

# Generate audio for British English
vac tts --transcript transcript.json --output audio/ --lang en-GB
```

This generates:

- `audio/slide_000.mp3`, `slide_001.mp3`, ... (one per slide)
- `audio/manifest.json` (timing information for video recorder)

### Manifest Output

The manifest.json contains timing data for use by the video recorder:

```json
{
  "version": "1.0",
  "language": "en-US",
  "generatedAt": "2024-01-01T12:00:00Z",
  "slides": [
    {
      "index": 0,
      "title": "Title Slide",
      "audioFile": "slide_000.mp3",
      "audioDurationMs": 5200,
      "pauseDurationMs": 500,
      "totalDurationMs": 5700
    }
  ]
}
```

### Using Go Library (Programmatic)

```go
package main

import (
    "context"
    "github.com/grokify/videoascode/pkg/transcript"
    "github.com/grokify/videoascode/pkg/tts"
)

func main() {
    ctx := context.Background()

    // Load transcript
    t, _ := transcript.LoadFromFile("transcript.json")

    // Create generator
    generator, _ := tts.NewTranscriptGenerator(tts.TranscriptGeneratorConfig{
        APIKey:    os.Getenv("ELEVENLABS_API_KEY"),
        OutputDir: "audio",
    })

    // Generate audio and get manifest
    manifest, _ := generator.GenerateFromTranscript(ctx, t, "en-US")

    // Save manifest
    manifest.SaveToFile("audio/manifest.json")
}
```

## Step 2: Generate Video

Use the `vac video` command with the audio manifest.

### Using Pre-generated Audio (Recommended)

```bash
vac video \
  --input presentation.md \
  --manifest audio/manifest.json \
  --output output.mp4
```

### Using Inline Voiceovers (Full Pipeline)

```bash
vac video \
  --input presentation.md \
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
    "title": "Introduction to vac",
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
        { "text": "Welcome to vac...", "pause": 500 },
        { "text": "In this presentation..." }
      ]
    },
    "es-ES": {
      "voice": {
        "voiceId": "onwK4e9ZLuTAKqWW03F9",
        "voiceName": "Daniel"
      },
      "segments": [
        { "text": "Bienvenido a vac...", "pause": 500 },
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

# 2. Generate audio (using Go script above or vac)
go run generate_audio.go

# 3. Generate video
vac \
  --input presentation.md \
  --output output.mp4

# 4. For different languages
vac \
  --input presentation.md \
  --transcript transcript.json \
  --lang es-ES \
  --output output_es.mp4
```
