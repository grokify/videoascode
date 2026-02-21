# CLI Reference

Complete command-line interface reference.

## Command Structure

vac uses a hierarchical command structure:

```
vac
├── slides              # Marp slide presentations
│   ├── video          # Full pipeline: parse, TTS, record, combine
│   └── tts            # Generate audio from transcript
├── browser            # Browser automation recordings
│   ├── video          # Record with TTS voiceover
│   └── record         # Silent recording (no audio)
└── subtitle           # Generate subtitles from audio
```

---

## slides video

Generate video from Marp presentation (full pipeline).

```bash
vac slides video [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-i, --input` | string | *required* | Input Marp markdown file |
| `-o, --output` | string | `output.mp4` | Output video file |
| `-m, --manifest` | string | | Audio manifest file (from `slides tts`) |
| `-k, --api-key` | string | `$ELEVENLABS_API_KEY` | ElevenLabs API key |
| `-v, --voice` | string | `pNInz6obpgDQGcFmaJgB` | ElevenLabs voice ID (Adam) |
| `--width` | int | `1920` | Video width in pixels |
| `--height` | int | `1080` | Video height in pixels |
| `--fps` | int | `30` | Video frame rate |
| `--transition` | float | `0` | Transition duration (seconds) |
| `--subtitles` | string | | Subtitle file to embed (SRT or VTT) |
| `--subtitles-lang` | string | auto-detect | Subtitle language code |
| `--output-individual` | string | | Directory for individual slide videos |
| `--screen-device` | string | auto-detect | macOS screen capture device |
| `--workdir` | string | system temp | Working directory for temp files |
| `--check` | bool | | Verify dependencies and exit |

### Examples

```bash
# Full pipeline with inline voiceovers
vac slides video --input slides.md --output video.mp4

# Use pre-generated audio
vac slides video --input slides.md --manifest audio/manifest.json --output video.mp4

# With transitions and custom resolution
vac slides video --input slides.md --output video.mp4 \
  --transition 0.5 --width 1280 --height 720

# Generate individual slide videos for Udemy
vac slides video --input slides.md --output combined.mp4 \
  --output-individual ./lectures/

# Check dependencies
vac slides video --check
```

---

## slides tts

Generate audio files from a transcript JSON file.

```bash
vac slides tts [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-t, --transcript` | string | *required* | Transcript JSON file |
| `-o, --output` | string | `audio` | Output directory for audio files |
| `-l, --lang` | string | from transcript | Language/locale code (e.g., `en-US`) |
| `--provider` | string | auto-detect | TTS provider: `elevenlabs` or `deepgram` |
| `--elevenlabs-api-key` | string | `$ELEVENLABS_API_KEY` | ElevenLabs API key |
| `--deepgram-api-key` | string | `$DEEPGRAM_API_KEY` | Deepgram API key |
| `-f, --force` | bool | `false` | Regenerate audio even if files exist |

### Examples

```bash
# Generate English audio
vac slides tts --transcript transcript.json --output audio/en-US/ --lang en-US

# Generate Spanish audio with Deepgram
vac slides tts --transcript transcript.json --output audio/es-ES/ \
  --lang es-ES --provider deepgram

# Force regeneration
vac slides tts --transcript transcript.json --output audio/ --force
```

---

## browser video

Record browser-driven demos with AI-generated voiceover.

```bash
vac browser video [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | string | *required* | Configuration file (YAML/JSON) |
| `-o, --output` | string | `output.mp4` | Output video file |
| `-a, --audio-dir` | string | | Directory to save/reuse audio tracks |
| `-p, --provider` | string | auto-detect | TTS provider: `elevenlabs` or `deepgram` |
| `-v, --voice` | string | from config | TTS voice ID |
| `-l, --lang` | string | `en-US` | Languages to generate (comma-separated) |
| `--elevenlabs-api-key` | string | `$ELEVENLABS_API_KEY` | ElevenLabs API key |
| `--deepgram-api-key` | string | `$DEEPGRAM_API_KEY` | Deepgram API key |
| `--width` | int | `1920` | Video width in pixels |
| `--height` | int | `1080` | Video height in pixels |
| `--fps` | int | `30` | Video frame rate |
| `--transition` | float | `0` | Transition duration (seconds) |
| `--headless` | bool | `false` | Run browser in headless mode |
| `--subtitles` | bool | `false` | Generate subtitles from voiceover timing |
| `--subtitles-stt` | bool | `false` | Generate word-level subtitles using STT |
| `--subtitles-burn` | bool | `false` | Burn subtitles into video (requires FFmpeg with libass) |
| `--no-audio` | bool | `false` | Generate video without audio (TTS used for timing/subtitles) |
| `--fast` | bool | `false` | Use hardware-accelerated encoding (VideoToolbox on macOS) |
| `--limit` | int | `0` | Limit to first N segments (0 = no limit, for testing) |
| `--limit-steps` | int | `0` | Limit browser segments to first N steps (0 = no limit, for testing) |
| `--workdir` | string | system temp | Working directory for temp files |

### Examples

```bash
# Basic browser demo
vac browser video --config demo.yaml --output demo.mp4

# Multi-language with audio caching
vac browser video --config demo.yaml --output demo.mp4 \
  --audio-dir ./audio --lang en-US,fr-FR,zh-Hans

# With subtitles burned in (requires FFmpeg with libass)
vac browser video --config demo.yaml --output demo.mp4 \
  --subtitles --subtitles-burn

# Silent video with burned subtitles (no audio track)
vac browser video --config demo.yaml --output demo.mp4 \
  --subtitles --subtitles-burn --no-audio

# Headless mode for CI/CD
vac browser video --config demo.yaml --output demo.mp4 --headless

# Using Deepgram TTS
vac browser video --config demo.yaml --output demo.mp4 --provider deepgram

# Fast encoding with hardware acceleration (macOS VideoToolbox)
vac browser video --config demo.yaml --output demo.mp4 --fast

# Test with limited segments (faster iteration)
vac browser video --config demo.yaml --output demo.mp4 --limit 2

# Test with limited browser steps (faster iteration)
vac browser video --config demo.yaml --output demo.mp4 --limit-steps 3
```

### Audio Caching

When using `--audio-dir`, vac caches generated TTS audio:

- Audio files stored as `{audio-dir}/{language}/segment_XXX.mp3`
- Metadata JSON files store per-voiceover timing information
- Subsequent runs skip TTS generation if cached audio exists

### Multi-Language Timing

When generating multiple languages, the video is paced to the longest audio:

1. TTS audio is generated for all requested languages
2. Per-voiceover durations are compared across languages
3. Each browser step uses the maximum duration
4. All language versions sync with the same video

---

## browser record

Record browser session without audio (silent recording).

```bash
vac browser record [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | string | | Configuration file (YAML/JSON) |
| `-s, --steps` | string | | Steps file defining browser actions |
| `-u, --url` | string | | Starting URL for the browser |
| `-o, --output` | string | `recording.mp4` | Output video file |
| `--width` | int | `1920` | Browser viewport width |
| `--height` | int | `1080` | Browser viewport height |
| `--fps` | int | `30` | Video frame rate |
| `--headless` | bool | `false` | Run browser in headless mode |
| `-t, --timing` | string | | Output timing JSON file |
| `--timeout` | int | `30000` | Default step timeout (ms) |
| `--workdir` | string | system temp | Working directory |
| `--cleanup` | bool | `true` | Clean up temp files after recording |

### Examples

```bash
# Record from steps file
vac browser record --url https://example.com --steps demo.json --output demo.mp4

# Record from config file
vac browser record --config demo.yaml --output demo.mp4

# Export timing data for later audio sync
vac browser record --url https://example.com --steps demo.json \
  --output demo.mp4 --timing timing.json

# Headless mode
vac browser record --url https://example.com --steps demo.json \
  --output demo.mp4 --headless
```

---

## subtitle

Generate subtitles from audio files using speech-to-text.

```bash
vac subtitle [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-a, --audio` | string | *required* | Audio directory containing manifest.json |
| `-o, --output` | string | `subtitles` | Output directory for subtitle files |
| `-l, --lang` | string | from manifest | Language code |
| `--provider` | string | `deepgram` | STT provider: `deepgram` or `elevenlabs` |
| `--individual` | bool | `false` | Also generate per-slide subtitle files |

### Examples

```bash
# Generate subtitles (language auto-detected)
vac subtitle --audio audio/en-US/

# Custom output directory
vac subtitle --audio audio/fr-FR/ --output subs/

# Keep individual slide subtitles
vac subtitle --audio audio/en-US/ --individual
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ELEVENLABS_API_KEY` | ElevenLabs API key for TTS |
| `DEEPGRAM_API_KEY` | Deepgram API key for TTS/STT |

---

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
