# marp2video

Convert Marp presentations with voiceovers to video files.

This tool takes a Marp markdown presentation with voiceover text (inline comments or JSON transcript), generates speech using ElevenLabs TTS, and creates a synchronized video recording of the presentation.

## Features

- ✅ **Parse Marp presentations** with voiceover in HTML comments
- ✅ **JSON transcript support** for multi-language voiceovers
- ✅ **Text-to-speech** using ElevenLabs API (Adam voice by default)
- ✅ **Multi-language** support with per-slide voice configuration
- ✅ **Browser automation** with Rod to display slides
- ✅ **Screen recording** with synchronized audio using ffmpeg
- ✅ **Cross-platform** support (macOS, Linux, Windows)
- ✅ **Pause directives** like `[PAUSE:1000]` for timing control
- ✅ **Full orchestration** - entire process automated in Go
- ✅ **YouTube-ready** combined video output with optional transitions
- ✅ **Udemy-ready** individual slide videos for course lectures
- ✅ **Decoupled workflow** - generate audio and video separately

## Installation

### Prerequisites

1. **Go 1.21+**
   ```bash
   go version
   ```

2. **ffmpeg** (for video recording and processing)
   ```bash
   # macOS
   brew install ffmpeg

   # Linux
   sudo apt install ffmpeg

   # Windows
   # Download from https://ffmpeg.org/download.html
   ```

3. **Marp CLI** (for rendering presentations)
   ```bash
   npm install -g @marp-team/marp-cli
   ```

4. **ElevenLabs API Key**
   - Sign up at [ElevenLabs](https://elevenlabs.io/)
   - Get your API key from the dashboard

### Build from Source

```bash
git clone https://github.com/grokify/marp2video
cd marp2video
go build -o bin/marp2video ./cmd/marp2video
```

## Usage

marp2video provides two subcommands for flexible workflows:

- `marp2video video` - Full pipeline or video generation with pre-generated audio
- `marp2video tts` - Generate audio from JSON transcript

### Quick Start (Full Pipeline)

```bash
export ELEVENLABS_API_KEY="your-api-key-here"

# Using inline voiceover comments
marp2video video --input slides.md --output video.mp4
```

### Two-Step Workflow (Recommended for Multi-Language)

```bash
# Step 1: Generate audio from transcript
marp2video tts --transcript transcript.json --output audio/ --lang en-US

# Step 2: Generate video with pre-generated audio
marp2video video --input slides.md --manifest audio/manifest.json --output video.mp4
```

### Command: `marp2video tts`

Generate audio files from a JSON transcript.

```
marp2video tts [flags]

Flags:
  -t, --transcript string   Transcript JSON file (required)
  -o, --output string       Output directory for audio files (default "audio")
  -l, --lang string         Language/locale code (e.g., en-US, es-ES)
  -k, --api-key string      ElevenLabs API key (or use ELEVENLABS_API_KEY env var)
```

**Output:**

- `audio/slide_000.mp3`, `slide_001.mp3`, ... (one per slide)
- `audio/manifest.json` (timing information for video recording)

**Example:**

```bash
# Generate audio for Spanish
marp2video tts --transcript transcript.json --output audio_es/ --lang es-ES
```

### Command: `marp2video video`

Generate video from Marp presentation.

```
marp2video video [flags]

Flags:
  -i, --input string              Input Marp markdown file (required)
  -o, --output string             Output video file (default "output.mp4")
  -m, --manifest string           Audio manifest file (from 'marp2video tts')
  -k, --api-key string            ElevenLabs API key (or use ELEVENLABS_API_KEY env var)
  -v, --voice string              ElevenLabs voice ID (default: Adam)
      --width int                 Video width (default 1920)
      --height int                Video height (default 1080)
      --fps int                   Frame rate (default 30)
      --transition float          Transition duration in seconds
      --output-individual string  Directory for individual slide videos
      --workdir string            Working directory for temp files
      --screen-device string      Screen capture device (macOS)
      --check                     Check dependencies and exit
```

### Examples

**Full pipeline with inline voiceovers:**

```bash
marp2video video \
  --input presentation.md \
  --output youtube_video.mp4 \
  --transition 0.5
```

**Multi-language workflow:**

```bash
# Generate audio for each language
marp2video tts --transcript transcript.json --output audio_en/ --lang en-US
marp2video tts --transcript transcript.json --output audio_es/ --lang es-ES

# Generate videos
marp2video video --input slides.md --manifest audio_en/manifest.json --output video_en.mp4
marp2video video --input slides.md --manifest audio_es/manifest.json --output video_es.mp4
```

**Generate individual videos for Udemy:**

```bash
marp2video video \
  --input presentation.md \
  --output combined.mp4 \
  --output-individual ./udemy_videos/
```

### Check Dependencies

```bash
marp2video video --check
```

This will verify that all required tools (ffmpeg, marp) are installed.

## Voiceover Formats

marp2video supports two voiceover formats:

1. **Inline HTML comments** - Simple, single-language
2. **JSON transcript** - Multi-language, advanced TTS control

### Option 1: Inline Voiceover Comments

Add voiceover text in HTML comments before or after slide content:

```markdown
---
marp: true
---

<!--
This is the voiceover for the first slide.
It will be converted to speech using ElevenLabs.
[PAUSE:1000]
You can add pause directives for timing control.
-->

# First Slide

This is the visible content

---

<!--
Voiceover for slide 2...
-->

# Second Slide

More content
```

#### Pause Directives

Use `[PAUSE:milliseconds]` to add pauses in the voiceover:

```markdown
<!--
First sentence.
[PAUSE:1000]
Second sentence after a 1-second pause.
[PAUSE:2000]
Third sentence after a 2-second pause.
-->
```

The pause directives are automatically removed from the spoken text.

### Option 2: JSON Transcript

For multi-language support and advanced TTS configuration, use a JSON transcript file:

```json
{
  "version": "1.0",
  "metadata": {
    "title": "My Presentation",
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
  },
  "slides": [
    {
      "index": 0,
      "title": "Title Slide",
      "transcripts": {
        "en-US": {
          "segments": [
            { "text": "Welcome to the presentation.", "pause": 500 },
            { "text": "Let's get started." }
          ]
        },
        "es-ES": {
          "voice": {
            "voiceId": "onwK4e9ZLuTAKqWW03F9",
            "voiceName": "Daniel"
          },
          "segments": [
            { "text": "Bienvenido a la presentación.", "pause": 500 },
            { "text": "Comencemos." }
          ]
        }
      }
    }
  ]
}
```

#### Transcript Features

| Feature | Description |
|---------|-------------|
| **Multi-language** | Per-slide transcripts for each locale (en-US, es-ES, etc.) |
| **Voice override** | Different voice per language or segment |
| **Pause control** | Pause after each segment (milliseconds) |
| **Venue presets** | Optimized settings for YouTube, Udemy, Coursera |
| **TTS parameters** | Stability, similarity boost, style exaggeration |

#### Audio Manifest

When using `marp2video tts`, a manifest is generated with timing info:

```json
{
  "version": "1.0",
  "language": "en-US",
  "generatedAt": "2024-01-01T12:00:00Z",
  "slides": [
    {
      "index": 0,
      "audioFile": "slide_000.mp3",
      "audioDurationMs": 5200,
      "pauseDurationMs": 500,
      "totalDurationMs": 5700
    }
  ]
}
```

This manifest is used by `marp2video video --manifest` for precise slide timing.

## How It Works

### Pipeline Overview

marp2video supports two workflows:

**Workflow A: Full Pipeline (inline voiceovers)**
```
presentation.md → Parse → TTS → Render → Record → Combine → video.mp4
```

**Workflow B: Two-Step (JSON transcript)**
```
Step 1: transcript.json → marp2video tts → audio/*.mp3 + manifest.json
Step 2: presentation.md + manifest.json → marp2video video → video.mp4
```

### Detailed Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│  INPUT OPTIONS                                                          │
│  ┌─────────────────────────┐    ┌─────────────────────────────────────┐ │
│  │ A: presentation.md      │ OR │ B: transcript.json (multi-language) │ │
│  │    (inline voiceovers)  │    │    + presentation.md                │ │
│  └─────────────────────────┘    └─────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 1: Parse / Load Transcript                                        │
│  • A: Extract voiceover from HTML comments + parse [PAUSE:ms]           │
│  • B: Load transcript.json, select language, resolve voice config       │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 2: Generate Audio (ElevenLabs TTS)                                │
│  • Send voiceover text to ElevenLabs API                                │
│  • Apply voice settings (stability, similarity, style)                  │
│  • Output: audio/slide_000.mp3, slide_001.mp3, ...                      │
│  • Output: audio/manifest.json (timing for video recording)             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 3: Render HTML (Marp CLI)                                         │
│  • Execute: marp presentation.md -o presentation.html --html            │
│  • Creates navigable HTML presentation with all slides                  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 4: Record Slides (Browser + ffmpeg)                               │
│  • Launch headless browser via Rod (Chromium)                           │
│  • Load HTML presentation                                               │
│  • For each slide:                                                      │
│    ├─ Navigate to slide                                                 │
│    ├─ Record for: audioDurationMs + pauseDurationMs (from manifest)     │
│    └─ Save: video/slide_000.mp4, slide_001.mp4, ...                     │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 5: Combine Videos (ffmpeg)                                        │
│  • Concatenate all slide videos in sequence                             │
│  • Optional: Apply crossfade transitions (--transition flag)            │
│  • Output: video.mp4                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 6: Export Individual Videos (Optional)                            │
│  • Copy individual slide videos to output directory                     │
│  • For Udemy courses: --output-individual ./lectures/                   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Step Details

| Step | Component | Tool | Input | Output |
|------|-----------|------|-------|--------|
| 1 | Parser | Go | `slides.md` | Slides + voiceovers |
| 2 | TTS | ElevenLabs API | Voiceover text | `slide_*.mp3` |
| 3 | Renderer | Marp CLI | `slides.md` | `presentation.html` |
| 4 | Recorder | Rod + ffmpeg | HTML + MP3 | `slide_*.mp4` |
| 5 | Combiner | ffmpeg | `slide_*.mp4` | `output.mp4` |
| 6 | Exporter | Go | `slide_*.mp4` | Individual files |

## Architecture

```
marp2video/
├── cmd/marp2video/          # CLI (Cobra-based)
│   ├── main.go              # Entry point
│   ├── root.go              # Root command
│   ├── tts.go               # TTS subcommand
│   └── video.go             # Video subcommand
├── pkg/
│   ├── parser/              # Marp markdown parser
│   ├── transcript/          # JSON transcript types
│   ├── tts/                 # ElevenLabs TTS + manifest
│   ├── renderer/            # Marp HTML renderer & browser control
│   ├── audio/               # Audio utilities
│   ├── video/               # Video recording & combination
│   └── orchestrator/        # Main workflow coordinator
├── examples/                # Example presentations
│   └── intro/               # Self-documenting example
│       ├── presentation.md
│       ├── transcript.json
│       └── README.md
└── docs/                    # MkDocs documentation
```

## Platform-Specific Recording

### macOS (including Apple Silicon M1/M2/M3)

Fully compatible with Apple Silicon Macs. Uses `avfoundation` for screen capture:
```bash
ffmpeg -f avfoundation -i "<device>:none" ...
```

**Screen device auto-detection**: The tool automatically detects the correct screen capture device. On Macs with external displays or connected iPhones, the device number varies. To list available devices:
```bash
ffmpeg -f avfoundation -list_devices true -i ""
```

You can manually specify the device if needed:
```bash
./bin/marp2video --input slides.md --output video.mp4 --screen-device "4:none"
```

### Linux
Uses `x11grab` for screen capture:
```bash
ffmpeg -f x11grab -i ":0.0" ...
```

### Windows
Uses `gdigrab` for screen capture:
```bash
ffmpeg -f gdigrab -i "desktop" ...
```

## Output Format & Platform Compatibility

Videos are encoded with settings optimized for direct upload to YouTube and Udemy - no re-encoding required.

### Video Specifications

| Setting | Value | Notes |
|---------|-------|-------|
| **Container** | MP4 | Universal compatibility |
| **Video Codec** | H.264 (libx264) | Required by YouTube & Udemy |
| **Resolution** | 1920x1080 | Full HD (configurable) |
| **Frame Rate** | 30fps | Standard (configurable) |
| **Quality** | CRF 23 | Good quality/size balance |
| **Pixel Format** | yuv420p | Maximum compatibility |
| **Audio Codec** | AAC | Required by both platforms |
| **Audio Bitrate** | 192kbps | Clear speech audio |

### YouTube Upload

The combined video (`--output`) is ready for direct upload:
- Includes optional crossfade transitions (`--transition 0.5`)
- Single file containing all slides with narration
- No processing or re-encoding needed

### Udemy Upload

Individual slide videos (`--output-individual`) are designed for Udemy courses:
- Each slide saved as separate file (slide_000.mp4, slide_001.mp4, etc.)
- Upload as individual lectures in your course curriculum
- Sequential naming for easy organization

**Tip for Udemy**: Udemy recommends lectures be 2+ minutes. For short slides, consider:
- Adding longer pause directives (`[PAUSE:5000]`)
- Combining related slides into single lectures
- Using more detailed voiceover scripts

## Examples

The `examples/` directory contains self-contained examples:

```
examples/
├── intro/                    # Introduction to marp2video
│   ├── presentation.md       # Marp markdown source (13 slides)
│   ├── transcript.json       # Multi-language transcript (en-US, en-GB, es-ES)
│   ├── README.md             # Detailed usage instructions
│   └── audio/                # Generated audio (after running tts)
│       ├── manifest.json
│       └── slide_*.mp3
└── README.md
```

### Running the Intro Example

**Option A: Full pipeline (inline voiceovers)**

```bash
marp2video video \
  --input examples/intro/presentation.md \
  --output examples/intro/output.mp4
```

**Option B: Two-step with transcript (multi-language)**

```bash
# Generate audio for English
marp2video tts \
  --transcript examples/intro/transcript.json \
  --output examples/intro/audio/ \
  --lang en-US

# Generate video
marp2video video \
  --input examples/intro/presentation.md \
  --manifest examples/intro/audio/manifest.json \
  --output examples/intro/output.mp4

# Generate Spanish version
marp2video tts \
  --transcript examples/intro/transcript.json \
  --output examples/intro/audio_es/ \
  --lang es-ES

marp2video video \
  --input examples/intro/presentation.md \
  --manifest examples/intro/audio_es/manifest.json \
  --output examples/intro/output_es.mp4
```

The `intro` example is a self-documenting presentation that explains what marp2video does - using marp2video itself.

### Additional Example

See `example_presentation.md` for a complete example with:

- Custom Marp theme
- Voiceover comments on each slide
- Pause directives for timing

## Troubleshooting

### "ffmpeg not found"
Install ffmpeg using your package manager (see Prerequisites)

### "marp CLI not found"
Install Marp CLI: `npm install -g @marp-team/marp-cli`

### "ElevenLabs API error"
- Verify your API key is correct
- Check your ElevenLabs account has sufficient credits
- Ensure you have access to the voice ID you specified

### Recording issues
- Ensure the browser window is visible during recording
- On macOS, you may need to grant screen recording permissions
- Try reducing video resolution if performance is poor

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/marp2video ./cmd/marp2video
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [Marp](https://marp.app/) - Markdown presentation ecosystem
- [ElevenLabs](https://elevenlabs.io/) - AI voice generation
- [Rod](https://github.com/go-rod/rod) - Browser automation framework
- [ffmpeg](https://ffmpeg.org/) - Multimedia processing

## Roadmap

- [x] Custom voice settings (stability, similarity, style)
- [x] Video transitions between slides
- [x] Individual slide video export (for Udemy)
- [x] JSON transcript for multi-language support
- [x] Decoupled TTS workflow (separate audio generation)
- [x] Audio manifest with timing information
- [ ] Add progress bar during conversion
- [ ] Support for background music
- [ ] Batch processing of multiple presentations
- [ ] Web UI for easier configuration
- [ ] Export to different video formats
- [ ] Add subtitle/caption generation
- [ ] Avatar integration (HeyGen, Synthesia)
