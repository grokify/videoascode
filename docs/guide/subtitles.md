# Subtitle Generation

marp2video can generate subtitle files (SRT/VTT) from your presentation audio using speech-to-text.

## Quick Start

```bash
# Generate subtitles for a language
marp2video subtitle --audio audio/en-US/

# Output:
# subtitles/en-US.srt
# subtitles/en-US.vtt
```

## How It Works

1. **Audio input**: Reads MP3 files from `audio/{lang}/`
2. **Speech-to-text**: Uses Deepgram STT for word-level timing
3. **Subtitle generation**: Creates SRT and VTT files with accurate timestamps

### Timing Accuracy

The `subtitle` command uses Deepgram STT to transcribe the audio files and extract word-level timestamps. This provides accurate subtitle timing that matches the actual speech patterns, rather than estimating from text length.

## Command Reference

```
marp2video subtitle [flags]

Flags:
  -a, --audio string        Audio directory containing manifest.json (required)
  -o, --output string       Output directory for subtitle files (default "subtitles")
  -l, --lang string         Language code (auto-detected from manifest if not specified)
      --provider string     STT provider: deepgram or elevenlabs (default: deepgram)
      --individual          Also generate individual subtitle files per slide
```

### Examples

```bash
# Generate French subtitles
marp2video subtitle --audio audio/fr-FR/

# Custom output directory
marp2video subtitle --audio audio/zh-Hans/ --output subs/

# Keep individual slide subtitle files
marp2video subtitle --audio audio/en-US/ --individual
```

## Multi-Language Workflow

```bash
# Step 1: Generate audio for each language
marp2video tts --transcript transcript.json --output audio/en-US/ --lang en-US
marp2video tts --transcript transcript.json --output audio/fr-FR/ --lang fr-FR

# Step 2: Generate subtitles for each language
marp2video subtitle --audio audio/en-US/
marp2video subtitle --audio audio/fr-FR/

# Step 3: Generate videos
marp2video video --input slides.md --manifest audio/en-US/manifest.json --output video/en-US.mp4
marp2video video --input slides.md --manifest audio/fr-FR/manifest.json --output video/fr-FR.mp4
```

### Output Structure

```
project/
├── audio/
│   ├── en-US/
│   │   ├── manifest.json
│   │   └── slide_*.mp3
│   └── fr-FR/
│       ├── manifest.json
│       └── slide_*.mp3
├── subtitles/
│   ├── en-US.srt
│   ├── en-US.vtt
│   ├── fr-FR.srt
│   └── fr-FR.vtt
└── video/
    ├── en-US.mp4
    └── fr-FR.mp4
```

## Current Implementation: Standard Subtitles

The current implementation generates **standard subtitles** - the most professional and widely-used format:

```
┌─────────────────────────────────────┐
│                                     │
│         [Slide Content]             │
│                                     │
│  ─────────────────────────────────  │
│  Two types of AI users are emerging │
└─────────────────────────────────────┘
```

**Characteristics:**

- White text with black outline or semi-transparent background
- 1-2 lines maximum, positioned at bottom of screen
- Static display during speech segment
- Industry standard for professional content

**Used by:** Netflix, YouTube, broadcast TV, Udemy, Coursera

---

## Future Caption Styles

The following caption styles are planned for future implementation. They offer different trade-offs between engagement and professionalism.

### Karaoke Style (Word Highlight)

Words change color/highlight as they are spoken:

```
Two types of [AI users] are emerging
              ^^^^^^^^^ (highlighted in yellow as spoken)
```

**Characteristics:**

- Words highlight in sequence as spoken
- Requires word-level timestamps (available via Deepgram)
- More engaging than static subtitles
- Can be distracting for technical content

**Best for:** Music videos, language learning, accessibility features

**Implementation notes:**

- Requires word-level timing from STT (already available)
- Output format: ASS/SSA (Advanced SubStation Alpha) for styling
- Or: Burn into video using ffmpeg drawtext filter

### Word-by-Word Reveal (Social Media Style)

Words appear one at a time as spoken:

```
Frame 1: Two
Frame 2: Two types
Frame 3: Two types of
Frame 4: Two types of AI
...
```

**Characteristics:**

- Highly engaging, attention-grabbing
- Popular on TikTok, Instagram Reels, YouTube Shorts
- Less formal, not suitable for all content types
- Often combined with animations

**Best for:** Social media clips, promotional videos, short-form content

**Implementation notes:**

- Requires word-level timing
- Output: Burned into video (not separate subtitle file)
- Consider text positioning, font size, animations

### Animated Captions (CapCut/Premiere Style)

Words animate in with effects (pop, slide, bounce):

```
     ╭─────────────╮
     │  AI USERS   │  ← pops in with scale animation
     ╰─────────────╯
```

**Characteristics:**

- Very engaging, trendy aesthetic
- Complex to implement, requires video editing
- Not suitable for professional/corporate content
- Popular with content creators

**Best for:** Social media, entertainment, creator content

**Implementation notes:**

- Requires video compositing (ffmpeg complex filters or external tool)
- Template-based approach for consistency
- Consider offering preset animation styles

---

## Implementation Roadmap

| Style | Status | Priority | Notes |
|-------|--------|----------|-------|
| Standard (SRT/VTT) | ✅ Implemented | - | Current default |
| Burned-in standard | 🔲 Planned | High | ffmpeg subtitle overlay |
| Karaoke highlight | 🔲 Planned | Medium | ASS format + ffmpeg |
| Word-by-word reveal | 🔲 Planned | Medium | Social media use case |
| Animated captions | 🔲 Planned | Low | Complex, may use templates |

### Proposed CLI Extensions

```bash
# Burn subtitles into video (standard style)
marp2video video --input slides.md --manifest audio/en-US/manifest.json \
  --subtitles subtitles/en-US.srt --output video/en-US.mp4

# Karaoke style (future)
marp2video subtitle --audio audio/en-US/ --style karaoke --output subtitles/

# Word-by-word reveal (future)
marp2video video --input slides.md --manifest audio/en-US/manifest.json \
  --caption-style reveal --output video/en-US.mp4
```

---

## Technical Considerations

### Subtitle Formats

| Format | Extension | Features | Use Case |
|--------|-----------|----------|----------|
| SRT | `.srt` | Basic timing + text | Universal compatibility |
| VTT | `.vtt` | Timing + basic styling | Web video players |
| ASS/SSA | `.ass` | Full styling, positioning, effects | Karaoke, anime fansubs |
| TTML | `.ttml` | XML-based, broadcast standard | Broadcast, streaming |

### Burning Subtitles into Video

```bash
# Using ffmpeg with SRT
ffmpeg -i video.mp4 -vf "subtitles=subtitles.srt" output.mp4

# With custom styling
ffmpeg -i video.mp4 -vf "subtitles=subtitles.srt:force_style='FontSize=24,PrimaryColour=&HFFFFFF&'" output.mp4

# ASS format for full control
ffmpeg -i video.mp4 -vf "ass=subtitles.ass" output.mp4
```

### Word-Level Timing

Deepgram STT returns word-level timestamps:

```json
{
  "words": [
    {"word": "Two", "start": 0.0, "end": 0.2},
    {"word": "types", "start": 0.25, "end": 0.5},
    {"word": "of", "start": 0.52, "end": 0.6},
    {"word": "AI", "start": 0.65, "end": 0.9}
  ]
}
```

This data is already captured during subtitle generation and can be used for karaoke and word-reveal styles.
