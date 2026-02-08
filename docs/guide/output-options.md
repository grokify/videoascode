# Output Options

Configure video output for different platforms.

## Combined Video (YouTube)

Generate a single video file with all slides:

```bash
marp2video \
  --input slides.md \
  --output video.mp4
```

### With Transitions

Add crossfade transitions between slides:

```bash
marp2video \
  --input slides.md \
  --output video.mp4 \
  --transition 0.5
```

The `--transition` value is in seconds (0.5 = half second crossfade).

## Individual Videos (Udemy)

Export each slide as a separate video file:

```bash
marp2video \
  --input slides.md \
  --output combined.mp4 \
  --output-individual ./lectures/
```

This creates:

```
lectures/
├── slide_000.mp4
├── slide_001.mp4
├── slide_002.mp4
└── ...
```

!!! tip "Udemy Tip"
    Udemy recommends lectures be 2+ minutes. For short slides, add longer pause directives or combine related slides.

## Both Outputs

Generate combined and individual videos:

```bash
marp2video \
  --input slides.md \
  --output youtube_video.mp4 \
  --output-individual ./udemy_videos/ \
  --transition 0.5
```

## Video Specifications

All videos are encoded with platform-optimized settings:

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

## Custom Resolution

```bash
# 720p
marp2video --input slides.md --output video.mp4 \
  --width 1280 --height 720

# 4K
marp2video --input slides.md --output video.mp4 \
  --width 3840 --height 2160
```

## Custom Frame Rate

```bash
# 24fps (cinematic)
marp2video --input slides.md --output video.mp4 --fps 24

# 60fps (smooth)
marp2video --input slides.md --output video.mp4 --fps 60
```

## Platform Compatibility

| Platform | Combined | Individual | Transitions |
|----------|----------|------------|-------------|
| YouTube | ✅ | - | ✅ |
| Udemy | ✅ | ✅ | ✅ |
| Coursera | ✅ | ✅ | ✅ |
| Vimeo | ✅ | - | ✅ |
| LinkedIn Learning | ✅ | ✅ | ✅ |
