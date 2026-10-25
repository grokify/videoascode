# VideoAsCode (vac)

**Transform Marp Presentations and HTML Slide Decks into Videos with AI Voiceovers**

vac is a command-line tool that automates the conversion of [Marp](https://marp.app/) presentations and native HTML slide decks into professional videos with AI-generated narration.

## Features

- :material-file-document: **Parse Marp presentations** with voiceover in HTML comments or JSON transcripts
- :material-presentation: **Record HTML slide decks** with narration in `data-vo` attributes or reveal.js speaker notes — served and recorded in one command, each slide timed to its own narration
- :material-microphone: **Text-to-speech** using ElevenLabs, Deepgram, local F5-TTS, and other providers via OmniVoice
- :material-web: **Browser automation** with Rod to display slides and record demos
- :material-video: **Screen recording** with synchronized audio using ffmpeg
- :material-earth: **Multi-language support** with BCP-47 locale codes (en-US, en-GB, fr-CA, etc.)
- :material-television: **Platform-optimized** output for YouTube, Udemy, Coursera
- :material-transition: **Crossfade transitions** between slides
- :material-subtitles: **Subtitle generation** from voiceover timing or speech-to-text (Deepgram or local Whisper)
- :material-school: **Course lessons** — record a deck as one video per lesson, with a manifest for your course platform, re-recording only the lessons that changed
- :material-cached: **Audio caching** per voiceover, keyed by text and voice, so only edited narration is sent to TTS again
- :material-account-circle: **Avatar presenter overlay** (optional) via [OmniAvatar](https://github.com/plexusone/omniavatar) — HeyGen, Tavus, bitHuman
- :material-shield-lock: **Local, offline TTS & STT** (optional) — F5-TTS + Whisper on Apple Silicon via OmniVoice MLX providers, no API keys or cloud calls

## Quick Example

=== "Marp Slides"

    ```bash
    # Simple: inline voiceover comments
    vac slides video --input slides.md --output video.mp4

    # Advanced: multi-language transcript
    vac slides video --input slides.md \
               --transcript transcript.json \
               --lang es-ES \
               --output video_spanish.mp4
    ```

=== "HTML Deck"

    ```bash
    # Serve and record an HTML deck; narration comes from each slide's data-vo
    vac html video --input deck/index.html --output deck.mp4

    # Write a transcript (for translations) and a reusable video config
    vac html init --input deck/index.html --provider deepgram --voice aura-asteria-en

    # One video per lesson (slides with data-lesson start lessons)
    vac html video --input deck/index.html --lessons all
    ```

=== "Browser Demo"

    ```bash
    # Record browser demo with voiceover
    vac browser video --config demo.yaml --output demo.mp4

    # Multi-language with audio caching
    vac browser video --config demo.yaml --output demo.mp4 \
               --audio-dir ./audio --lang en-US,fr-FR,zh-Hans

    # With subtitles burned in
    vac browser video --config demo.yaml --output demo.mp4 \
               --subtitles --subtitles-burn
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

- :material-download: **[Installation](guides/getting-started/installation.md)**

    Install vac and its dependencies

- :material-rocket-launch: **[Quick Start](guides/getting-started/quick-start.md)**

    Create your first video in minutes

- :material-book-open-variant: **[User Guide](guides/pipeline.md)**

    Learn about the Marp slides pipeline

- :material-presentation: **[HTML Slide Decks](guides/html-decks.md)**

    Narrate and record native HTML decks

- :material-school: **[Lessons and Courses](guides/lessons.md)**

    One video per lesson, Udemy- and Coursera-style

- :material-web: **[Browser Video](guides/browser-video.md)**

    Record browser demos with voiceover

- :material-code-json: **[Transcript Schema](guides/reference/transcript-schema.md)**

    Multi-language JSON format reference

- :material-subtitles: **[Subtitles](guides/subtitles.md)**

    Generate SRT/VTT subtitles

</div>

## Use Cases

| Platform | Use Case | Command | Features |
|----------|----------|---------|----------|
| **YouTube** | Tutorials, demos | `slides video` | Combined video with transitions |
| **Udemy** | Course lectures | `slides video`, `html video --lessons` | One video per lesson, re-record only what changed |
| **Coursera** | Academic content | `slides video` | Professional voice settings |
| **Product Demos** | SaaS walkthroughs | `browser video` | Automated browser recording |
| **Talks & Briefings** | HTML decks, reveal.js | `html video` | Per-slide narration timing, no build step |
| **Documentation** | Animated guides | Both | Multi-language support |
