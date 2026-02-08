# Quick Start

Create your first video in 5 minutes.

## Step 1: Create a Presentation

Create a file called `slides.md`:

```markdown
---
marp: true
theme: default
---

# Welcome

This is my first marp2video presentation!

<!-- Welcome to this presentation.
     We'll create a video with AI narration. -->

---

# Features

- Easy to use
- AI-powered voiceovers
- Multiple output formats

<!-- This slide shows the key features.
     [PAUSE:500]
     Let's explore each one. -->

---

# Thank You

Questions?

<!-- Thanks for watching! -->
```

## Step 2: Set Your API Key

```bash
export ELEVENLABS_API_KEY="your-api-key"
```

## Step 3: Generate the Video

```bash
marp2video --input slides.md --output my_video.mp4
```

## What Happens

1. **Parsing** - marp2video extracts voiceover text from `<!-- comments -->`
2. **TTS Generation** - Text is sent to ElevenLabs, audio files created
3. **HTML Rendering** - Marp CLI converts markdown to HTML
4. **Recording** - Each slide is recorded with synchronized audio
5. **Combining** - All slides are concatenated into `my_video.mp4`

## Next Steps

- Add [pause directives](../guide/voiceover-formats.md#pause-directives) for timing control
- Use [transcript.json](../guide/voiceover-formats.md#json-transcripts) for multi-language
- Configure [output options](../guide/output-options.md) for different platforms
