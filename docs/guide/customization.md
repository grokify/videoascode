# Customization

Customize voice, video, and output settings.

## Voice Settings

### Default Voice

Set via command line:

```bash
marp2video \
  --input slides.md \
  --voice pNInz6obpgDQGcFmaJgB \
  --output video.mp4
```

### ElevenLabs Voices

Popular voices:

| Voice ID | Name | Description |
|----------|------|-------------|
| `pNInz6obpgDQGcFmaJgB` | Adam | Clear, professional male |
| `21m00Tcm4TlvDq8ikWAM` | Rachel | Warm, natural female |
| `onwK4e9ZLuTAKqWW03F9` | Daniel | British male |
| `XrExE9yKIg1WjnnlVkGX` | Matilda | American female |

### Voice Settings in transcript.json

Fine-tune voice parameters:

```json
{
  "defaultVoice": {
    "provider": "elevenlabs",
    "voiceId": "pNInz6obpgDQGcFmaJgB",
    "voiceName": "Adam",
    "model": "eleven_multilingual_v2",
    "stability": 0.5,
    "similarityBoost": 0.75,
    "style": 0.2
  }
}
```

| Parameter | Range | Description |
|-----------|-------|-------------|
| `stability` | 0.0 - 1.0 | Voice consistency (higher = more stable) |
| `similarityBoost` | 0.0 - 1.0 | Voice similarity (higher = closer to original) |
| `style` | 0.0 - 1.0 | Style exaggeration (higher = more expressive) |

## Venue Presets

Optimize for different platforms:

| Venue | Stability | Similarity | Style | Use Case |
|-------|-----------|------------|-------|----------|
| YouTube | 0.45 | 0.8 | 0.2 | Sustained attention |
| Udemy | 0.5 | 0.75 | 0.0 | Clear, consistent |
| Coursera | 0.7 | 0.85 | 0.2 | Academic, engaging |
| TikTok | 0.3 | 0.85 | 0.45 | Energetic, immediate |

Set in transcript.json:

```json
{
  "metadata": {
    "defaultVenue": "youtube"
  }
}
```

## Video Settings

### Resolution

```bash
# Full HD (default)
--width 1920 --height 1080

# 720p
--width 1280 --height 720

# 4K UHD
--width 3840 --height 2160
```

### Frame Rate

```bash
--fps 24  # Cinematic
--fps 30  # Standard (default)
--fps 60  # Smooth
```

### Transitions

```bash
--transition 0.5  # 0.5 second crossfade
--transition 1.0  # 1 second crossfade
--transition 0    # No transitions (default)
```

## Screen Recording

### macOS Screen Device

Auto-detected, but can be overridden:

```bash
# List available devices
ffmpeg -f avfoundation -list_devices true -i ""

# Use specific device
marp2video --input slides.md --output video.mp4 \
  --screen-device "4:none"
```

## Working Directory

Customize temp file location:

```bash
marp2video --input slides.md --output video.mp4 \
  --workdir /path/to/workdir
```

Files created:

```
workdir/
├── audio/slide_*.mp3
├── video/slide_*.mp4
└── html/presentation.html
```
