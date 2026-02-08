# Examples

This directory contains example presentations demonstrating marp2video.

## Structure

Each example is in its own subdirectory with a consistent structure:

```
examples/
├── intro/                    # Introduction to marp2video
│   ├── presentation.md       # Marp markdown source (with inline voiceovers)
│   ├── transcript.json       # Structured transcript (multi-language, TTS settings)
│   └── output.mp4            # Generated video (after running)
├── <future-example>/
│   ├── presentation.md
│   ├── transcript.json
│   └── output.mp4
└── README.md                 # This file
```

## Transcript Formats

marp2video supports two voiceover formats:

### 1. Inline Comments (Simple)

Voiceover text embedded directly in markdown:

```markdown
# My Slide

<!-- This text will be spoken for this slide.
     [PAUSE:500]
     Pause directives control timing. -->
```

### 2. Transcript JSON (Advanced)

Separate JSON file with multi-language support, TTS settings, and timing:

```json
{
  "version": "1.0",
  "metadata": {
    "defaultLanguage": "en-US",
    "defaultVoice": { "provider": "elevenlabs", "voiceId": "..." }
  },
  "slides": [
    {
      "index": 0,
      "transcripts": {
        "en-US": { "segments": [{ "text": "Hello", "pause": 500 }] },
        "en-GB": { "segments": [{ "text": "Hello", "pause": 500 }] },
        "es-ES": { "segments": [{ "text": "Hola", "pause": 500 }] },
        "fr-CA": { "segments": [{ "text": "Bonjour", "pause": 500 }] }
      }
    }
  ]
}
```

See `pkg/transcript/transcript.schema.json` for the full schema.

## Available Examples

| Example | Description | Slides | Languages |
|---------|-------------|--------|-----------|
| [intro](./intro/) | Self-documenting introduction to marp2video | 13 | en-US, en-GB, es-ES |

## Running an Example

```bash
# Using inline comments (from presentation.md)
marp2video \
  --input examples/intro/presentation.md \
  --output examples/intro/output.mp4

# Using transcript.json (American English - default)
marp2video \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --output examples/intro/output_en-US.mp4

# Using transcript.json (British English)
marp2video \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --lang en-GB \
  --output examples/intro/output_en-GB.mp4

# Using transcript.json (Spanish - Spain)
marp2video \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --lang es-ES \
  --output examples/intro/output_es-ES.mp4
```

## Creating New Examples

1. Create a new subdirectory: `mkdir examples/my-example`
2. Add `presentation.md` with Marp frontmatter
3. Choose your voiceover approach:
   - **Simple**: Add `<!-- voiceover comments -->` inline
   - **Advanced**: Create `transcript.json` with multi-language support
4. Run marp2video to generate `output.mp4`

## Transcript JSON Features

The transcript.json format supports:

- **Multi-language**: Transcripts in multiple languages per slide
- **Voice per language**: Different TTS voices for each language
- **Segments with timing**: Fine-grained control over pauses and pacing
- **SSML hints**: Emphasis, prosody, pronunciation control
- **Venue optimization**: Pre-configured settings for YouTube, Udemy, etc.
- **Avatar support**: Future integration with HeyGen, Synthesia, etc.
