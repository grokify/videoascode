# Intro Example

A self-documenting presentation that introduces marp2video.

## Overview

This example demonstrates marp2video by creating a video about marp2video itself. It includes:

- 13 slides covering features and usage
- Multi-language transcripts (en-US, en-GB, es-ES)
- TTS voice settings optimized for YouTube

## Files

```
examples/intro/
├── presentation.md    # Marp source (13 slides)
├── transcript.json    # Multi-language transcript
└── output.mp4        # Generated video (after running)
```

## Presentation Content

The presentation covers:

1. Title slide
2. What is marp2video?
3. Why use marp2video?
4. Key features
5. How it works (pipeline)
6. Installation
7. Basic usage
8. Multi-language support
9. Customization options
10. Platform presets
11. Example use cases
12. Getting started
13. Thank you / call to action

## Running the Example

### Using Inline Comments (Default Language)

```bash
marp2video \
  --input examples/intro/presentation.md \
  --output examples/intro/output.mp4
```

### Using Transcript (American English)

```bash
marp2video \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --lang en-US \
  --output examples/intro/output_en-US.mp4
```

### Using Transcript (British English)

```bash
marp2video \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --lang en-GB \
  --output examples/intro/output_en-GB.mp4
```

### Using Transcript (Spanish)

```bash
marp2video \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --lang es-ES \
  --output examples/intro/output_es-ES.mp4
```

## Transcript Structure

The transcript.json includes three languages:

| Language | Locale | Voice |
|----------|--------|-------|
| American English | `en-US` | Adam |
| British English | `en-GB` | Daniel |
| Spanish | `es-ES` | Adam (multilingual) |

### Voice Settings

```json
{
  "metadata": {
    "defaultVoice": {
      "provider": "elevenlabs",
      "voiceId": "pNInz6obpgDQGcFmaJgB",
      "voiceName": "Adam",
      "model": "eleven_multilingual_v2",
      "stability": 0.5,
      "similarityBoost": 0.75,
      "style": 0.2
    },
    "defaultVenue": "youtube"
  }
}
```

## Customizing

### Change Voice

Edit the `defaultVoice` in transcript.json:

```json
{
  "defaultVoice": {
    "voiceId": "21m00Tcm4TlvDq8ikWAM",
    "voiceName": "Rachel"
  }
}
```

### Add a Language

Add a new locale to each slide's `transcripts` object:

```json
{
  "transcripts": {
    "en-US": { ... },
    "fr-FR": {
      "segments": [
        { "text": "Bienvenue dans marp2video." }
      ]
    }
  }
}
```

### Adjust Timing

Add pauses between segments:

```json
{
  "segments": [
    { "text": "First point.", "pause": 500 },
    { "text": "Second point.", "pause": 1000 }
  ]
}
```

## Expected Output

After running, you'll have:

- `output.mp4` - Full video with voiceover
- Individual slide videos in the working directory (if `--output-individual` specified)

Video specifications:

- Resolution: 1920x1080 (Full HD)
- Frame rate: 30 fps
- Audio: MP3 from ElevenLabs TTS
