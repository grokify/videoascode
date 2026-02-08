# marp2video

**Transform Marp Markdown Presentations into Videos with AI Voiceovers**

marp2video is a command-line tool that automates the conversion of [Marp](https://marp.app/) presentations into professional videos with AI-generated narration.

## Features

- :material-file-document: **Parse Marp presentations** with voiceover in HTML comments or JSON transcripts
- :material-microphone: **Text-to-speech** using ElevenLabs, Deepgram, and other providers via OmniVoice
- :material-web: **Browser automation** with Rod to display slides
- :material-video: **Screen recording** with synchronized audio using ffmpeg
- :material-earth: **Multi-language support** with BCP-47 locale codes (en-US, en-GB, fr-CA, etc.)
- :material-television: **Platform-optimized** output for YouTube, Udemy, Coursera
- :material-transition: **Crossfade transitions** between slides

## Quick Example

```bash
# Simple: inline voiceover comments
marp2video --input slides.md --output video.mp4

# Advanced: multi-language transcript
marp2video --input slides.md \
           --transcript transcript.json \
           --lang es-ES \
           --output video_spanish.mp4
```

## How It Works

```mermaid
flowchart LR
    A[Marp MD] --> B[Parse]
    B --> C[TTS Audio]
    C --> D[Record Slides]
    D --> E[Combine]
    E --> F[Video.mp4]
```

1. **Parse** - Extract slides and voiceover from Marp markdown
2. **Generate Audio** - Convert text to speech via ElevenLabs/OmniVoice
3. **Render HTML** - Use Marp CLI to create HTML presentation
4. **Record** - Screen capture each slide with audio sync
5. **Combine** - Concatenate slides with optional transitions

## Getting Started

<div class="grid cards" markdown>

- :material-download: **[Installation](getting-started/installation.md)**

    Install marp2video and its dependencies

- :material-rocket-launch: **[Quick Start](getting-started/quick-start.md)**

    Create your first video in minutes

- :material-book-open-variant: **[User Guide](guide/pipeline.md)**

    Learn about the full pipeline

- :material-code-json: **[Transcript Schema](reference/transcript-schema.md)**

    Multi-language JSON format reference

</div>

## Use Cases

| Platform | Use Case | Features |
|----------|----------|----------|
| **YouTube** | Tutorials, demos | Combined video with transitions |
| **Udemy** | Course lectures | Individual slide videos |
| **Coursera** | Academic content | Professional voice settings |
| **Documentation** | Animated guides | Multi-language support |
