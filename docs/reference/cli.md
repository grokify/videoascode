# CLI Reference

Complete command-line interface reference.

## Synopsis

```bash
marp2video [options]
```

## Options

### Input/Output

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--input` | string | *required* | Input Marp markdown file |
| `--output` | string | `output.mp4` | Output video file |
| `--output-individual` | string | | Directory for individual slide videos |
| `--transcript` | string | | JSON transcript file for multi-language |
| `--lang` | string | | Language/locale code (e.g., `en-US`, `es-ES`) |
| `--workdir` | string | system temp | Working directory for temp files |

### TTS Settings

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--api-key` | string | `$ELEVENLABS_API_KEY` | ElevenLabs API key |
| `--voice` | string | `pNInz6obpgDQGcFmaJgB` | ElevenLabs voice ID (Adam) |

### Video Settings

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--width` | int | `1920` | Video width in pixels |
| `--height` | int | `1080` | Video height in pixels |
| `--fps` | int | `30` | Video frame rate |
| `--transition` | float | `0` | Transition duration (seconds) |

### Platform-Specific

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--screen-device` | string | auto-detected | macOS screen capture device |

### Utility

| Flag | Type | Description |
|------|------|-------------|
| `--check` | bool | Verify dependencies and exit |
| `--version` | bool | Show version and exit |
| `--help` | bool | Show help and exit |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ELEVENLABS_API_KEY` | ElevenLabs API key (alternative to `--api-key`) |

## Examples

### Basic Usage

```bash
# Minimal - uses inline comments
marp2video --input slides.md --output video.mp4

# With API key
marp2video --input slides.md --output video.mp4 \
  --api-key "your-api-key"
```

### Multi-Language

```bash
# English (default from transcript)
marp2video --input slides.md \
  --transcript transcript.json \
  --output video_en.mp4

# Spanish
marp2video --input slides.md \
  --transcript transcript.json \
  --lang es-ES \
  --output video_es.mp4
```

### Platform-Specific Output

```bash
# YouTube (combined with transitions)
marp2video --input slides.md \
  --output youtube.mp4 \
  --transition 0.5

# Udemy (individual + combined)
marp2video --input slides.md \
  --output combined.mp4 \
  --output-individual ./lectures/
```

### Custom Video Settings

```bash
# 720p at 24fps
marp2video --input slides.md \
  --output video.mp4 \
  --width 1280 --height 720 --fps 24

# 4K at 60fps
marp2video --input slides.md \
  --output video.mp4 \
  --width 3840 --height 2160 --fps 60
```

### Custom Voice

```bash
# Use Rachel voice
marp2video --input slides.md \
  --output video.mp4 \
  --voice 21m00Tcm4TlvDq8ikWAM
```

### Dependency Check

```bash
marp2video --check
```

Output:

```
✓ ffmpeg found
✓ Marp CLI found
✓ ELEVENLABS_API_KEY set
All dependencies OK!
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Missing dependencies |
| 3 | Invalid input file |
| 4 | TTS generation failed |
| 5 | Recording failed |
| 6 | Video combination failed |
