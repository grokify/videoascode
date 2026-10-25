# Voiceover Formats

vac supports two formats for defining voiceovers.

## Inline Comments (Simple)

Add voiceover text directly in your Marp markdown using HTML comments:

```markdown
---
marp: true
---

# My Slide

Content here.

<!-- This text will be converted to speech
     and played while this slide is shown. -->

---

# Next Slide

<!-- Voiceover for the second slide. -->
```

### Pause Directives

Control timing with `[PAUSE:milliseconds]` directives:

```markdown
<!-- Welcome to this presentation.
     [PAUSE:1000]
     Let's explore the first topic.
     [PAUSE:500]
     Here we go! -->
```

| Directive | Duration |
|-----------|----------|
| `[PAUSE:500]` | 0.5 seconds |
| `[PAUSE:1000]` | 1 second |
| `[PAUSE:2000]` | 2 seconds |

Pause directives are removed from the spoken text automatically.

### Pros and Cons

| Pros | Cons |
|------|------|
| Simple, all-in-one file | Single language only |
| Easy to edit | Limited TTS control |
| No extra files | No voice customization per slide |

---

## JSON Transcripts (Advanced)

Use a separate `transcript.json` file for advanced features:

```json
{
  "version": "1.0",
  "metadata": {
    "title": "My Presentation",
    "defaultLanguage": "en-US",
    "defaultVoice": {
      "provider": "elevenlabs",
      "voiceId": "pNInz6obpgDQGcFmaJgB",
      "voiceName": "Adam"
    }
  },
  "slides": [
    {
      "index": 0,
      "transcripts": {
        "en-US": {
          "segments": [
            { "text": "Welcome to this presentation.", "pause": 1000 },
            { "text": "Let's explore the first topic." }
          ]
        },
        "es-ES": {
          "segments": [
            { "text": "Bienvenido a esta presentación.", "pause": 1000 },
            { "text": "Exploremos el primer tema." }
          ]
        }
      }
    }
  ]
}
```

### Usage

```bash
vac \
  --input slides.md \
  --transcript transcript.json \
  --lang es-ES \
  --output video_spanish.mp4
```

### Features

| Feature | Description |
|---------|-------------|
| **Multi-language** | Multiple languages per slide |
| **Voice per language** | Different voices for each language |
| **Segments** | Fine-grained control over text chunks |
| **Pause per segment** | Precise timing control |
| **SSML hints** | Emphasis, prosody, pronunciation |
| **Venue settings** | Platform-specific voice tuning |

### Pros and Cons

| Pros | Cons |
|------|------|
| Multi-language support | Separate file to maintain |
| Full TTS control | More complex structure |
| Voice per slide/segment | Requires JSON knowledge |
| SSML support | Must keep in sync with slides |

---

## Comparison

| Feature | Inline Comments | JSON Transcript |
|---------|-----------------|-----------------|
| Multi-language | :x: | :white_check_mark: |
| Pause control | :white_check_mark: | :white_check_mark: |
| Voice per slide | :x: | :white_check_mark: |
| SSML hints | :x: | :white_check_mark: |
| Segment-level control | :x: | :white_check_mark: |
| Single file workflow | :white_check_mark: | :x: |

## When to Use Which

- **Inline Comments**: Quick prototypes, single-language videos
- **JSON Transcript**: Production content, courses, multi-language
