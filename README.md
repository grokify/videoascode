# marp2video

Convert Marp presentations with voiceovers to video files.

This tool takes a Marp markdown presentation with voiceover text in HTML comments, generates speech using ElevenLabs TTS, and creates a synchronized video recording of the presentation.

## Features

- ✅ **Parse Marp presentations** with voiceover in HTML comments
- ✅ **Text-to-speech** using ElevenLabs API (Adam voice by default)
- ✅ **Browser automation** with Rod to display slides
- ✅ **Screen recording** with synchronized audio using ffmpeg
- ✅ **Cross-platform** support (macOS, Linux, Windows)
- ✅ **Pause directives** like `[PAUSE:1000]` for timing control
- ✅ **Full orchestration** - entire process automated in Go
- ✅ **YouTube-ready** combined video output with optional transitions
- ✅ **Udemy-ready** individual slide videos for course lectures

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

### Basic Usage

```bash
export ELEVENLABS_API_KEY="your-api-key-here"

./bin/marp2video \
  --input example_presentation.md \
  --output presentation.mp4
```

### Command-Line Options

```
--input string
    Input Marp markdown file (required)

--output string
    Output video file (default: "output.mp4")

--output-individual string
    Directory to save individual slide videos (for Udemy)
    Each slide will be saved as slide_000.mp4, slide_001.mp4, etc.

--transition float
    Transition duration between slides in seconds (default: 0 = no transitions)
    Uses crossfade effect for both video and audio

--api-key string
    ElevenLabs API key (or use ELEVENLABS_API_KEY env var)

--voice string
    ElevenLabs voice ID (default: "pNInz6obpgDQGcFmaJgB" - Adam voice)

--width int
    Video width in pixels (default: 1920)

--height int
    Video height in pixels (default: 1080)

--fps int
    Video frame rate (default: 30)

--workdir string
    Working directory for temporary files (default: system temp dir)

--screen-device string
    Screen capture device for macOS (auto-detected if empty)
    Use "ffmpeg -f avfoundation -list_devices true -i ''" to list devices

--check
    Check dependencies without running

--version
    Show version information
```

### Examples

**Generate combined video for YouTube:**
```bash
./bin/marp2video \
  --input presentation.md \
  --output youtube_video.mp4 \
  --transition 0.5
```

**Generate individual videos for Udemy:**
```bash
./bin/marp2video \
  --input presentation.md \
  --output combined.mp4 \
  --output-individual ./udemy_videos/
```

**Generate both with transitions:**
```bash
./bin/marp2video \
  --input presentation.md \
  --output youtube_video.mp4 \
  --output-individual ./udemy_videos/ \
  --transition 0.5
```

### Check Dependencies

```bash
./bin/marp2video --check
```

This will verify that all required tools (ffmpeg, marp) are installed.

## Presentation Format

### Voiceover Comments

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

### Pause Directives

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

## How It Works

### Pipeline Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│  INPUT: presentation.md (Marp markdown with <!-- voiceover comments -->) │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 1: Parse Markdown                                                 │
│  • Extract slides from Marp file                                        │
│  • Extract voiceover text from HTML comments                            │
│  • Parse [PAUSE:ms] timing directives                                   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 2: Generate Audio (ElevenLabs TTS)                                │
│  • Send voiceover text to ElevenLabs API                                │
│  • Receive MP3 audio files (one per slide)                              │
│  • Output: workdir/audio/slide_000.mp3, slide_001.mp3, ...              │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 3: Render HTML (Marp CLI)                                         │
│  • Execute: marp presentation.md -o presentation.html --html            │
│  • Creates navigable HTML presentation with all slides                  │
│  • Output: workdir/html/presentation.html                               │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 4: Record Slides (Browser + ffmpeg)                               │
│  • Launch headless browser via Rod (Chromium)                           │
│  • Load HTML presentation                                               │
│  • For each slide:                                                      │
│    ├─ Navigate to slide (keyboard: Home + Arrow keys)                   │
│    ├─ Start screen recording with audio overlay                         │
│    ├─ Record for: audio duration + pause directives                     │
│    └─ Save: workdir/video/slide_000.mp4, slide_001.mp4, ...             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 5: Combine Videos (ffmpeg)                                        │
│  • Concatenate all slide videos in sequence                             │
│  • Optional: Apply crossfade transitions (--transition flag)            │
│  • Output: presentation.mp4                                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 6: Export Individual Videos (Optional)                            │
│  • Copy individual slide videos to output directory                     │
│  • For Udemy courses: --output-individual ./lectures/                   │
│  • Output: lectures/slide_000.mp4, slide_001.mp4, ...                   │
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
├── cmd/marp2video/          # CLI entry point
├── pkg/
│   ├── parser/              # Marp markdown parser
│   ├── tts/                 # ElevenLabs text-to-speech
│   ├── renderer/            # Marp HTML renderer & browser control
│   ├── audio/               # Audio utilities
│   ├── video/               # Video recording & combination
│   └── orchestrator/        # Main workflow coordinator
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

The `examples/` directory contains self-contained examples, each in its own subdirectory:

```
examples/
├── intro/                    # Introduction to marp2video
│   ├── presentation.md       # Marp markdown source
│   ├── transcript.txt        # Human-readable voiceover transcript
│   └── output.mp4            # Generated video (after running)
└── README.md
```

### Running an Example

```bash
# Generate the intro video
marp2video \
  --input examples/intro/presentation.md \
  --output examples/intro/output.mp4
```

The `intro` example is a self-documenting presentation that explains what marp2video does - using marp2video itself.

### Full Example

See `example_presentation.md` for a complete example with:

- Custom Marp theme
- Voiceover comments on each slide
- Pause directives for timing
- 30+ slides demonstrating various features

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

- [ ] Add support for custom voice settings (speed, stability, etc.)
- [x] Implement video transitions between slides
- [x] Individual slide video export (for Udemy)
- [ ] Add progress bar during conversion
- [ ] Support for background music
- [ ] Batch processing of multiple presentations
- [ ] Web UI for easier configuration
- [ ] Export to different video formats
- [ ] Add subtitle/caption generation
