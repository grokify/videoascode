# marp2video vs marptalk

A detailed comparison of [marp2video](https://github.com/grokify/marp2video) and [marptalk](https://github.com/imjasonh/marptalk), two tools for creating narrated videos from Marp presentations.

## Overview

| Aspect | marp2video | marptalk |
|--------|------------|----------|
| **Language** | Go | Node.js |
| **License** | MIT | Apache-2.0 |
| **CLI Framework** | Cobra | Commander |
| **Primary TTS** | ElevenLabs (via OmniVoice) | Google Cloud TTS |
| **Primary STT** | Deepgram (via OmniVoice) | N/A (duration-based timing) |

## TTS Provider Support

| Feature | marp2video | marptalk |
|---------|------------|----------|
| **ElevenLabs** | :white_check_mark: Primary | :x: |
| **Google Cloud TTS** | :x: | :white_check_mark: Primary |
| **Deepgram TTS** | :white_check_mark: Secondary | :x: |
| **Browser TTS Fallback** | :x: | :white_check_mark: (Web Speech API) |
| **Provider Abstraction** | :white_check_mark: OmniVoice | :x: Single provider |
| **Voice Cloning** | :white_check_mark: (ElevenLabs) | :x: |

## Multi-Language Support

| Feature | marp2video | marptalk |
|---------|------------|----------|
| **Multi-language transcripts** | :white_check_mark: JSON with per-slide locales | :x: Single language per run |
| **Locale codes** | BCP-47 (en-US, zh-Hans, etc.) | Language codes (en-US, es-ES) |
| **Per-slide voice override** | :white_check_mark: | :x: |
| **Mixed TTS providers per video** | :white_check_mark: (e.g., ElevenLabs + Deepgram) | :x: |

## Voiceover Input Format

| Feature | marp2video | marptalk |
|---------|------------|----------|
| **Inline HTML comments** | :white_check_mark: `<!-- voiceover text -->` | :white_check_mark: `<!-- speaker notes -->` |
| **JSON transcript** | :white_check_mark: Structured, multi-language | :x: |
| **Pause directives** | :white_check_mark: `[PAUSE:1000]` | :x: |
| **Per-segment voice settings** | :white_check_mark: | :x: |

## Subtitle Generation

| Feature | marp2video | marptalk |
|---------|------------|----------|
| **SRT output** | :white_check_mark: | :white_check_mark: |
| **VTT output** | :white_check_mark: | :x: |
| **Generation method** | STT transcription (actual audio) | Word count estimation (150 wpm) |
| **Word-level timestamps** | :white_check_mark: | :x: |
| **Dictionary case correction** | :white_check_mark: (tech terms, custom JSON) | :x: |
| **YouTube chapters** | :x: | :white_check_mark: |

## Video Generation

| Feature | marp2video | marptalk |
|---------|------------|----------|
| **Method** | Image-based (Marp PNG export) | Static slide screenshots |
| **Audio sync** | Manifest-based (actual durations) | Audio file duration analysis |
| **Soft subtitles** | :white_check_mark: | :white_check_mark: |
| **Hard subtitles (burned-in)** | :white_check_mark: | :white_check_mark: |
| **Crossfade transitions** | :white_check_mark: `--transition` | :x: |
| **Individual slide export** | :white_check_mark: (Udemy-ready) | :x: |
| **Mixed audio sample rates** | :white_check_mark: (filter_complex concat) | N/A (single provider) |

## Workflow

| Feature | marp2video | marptalk |
|---------|------------|----------|
| **Decoupled TTS/Video** | :white_check_mark: Separate `tts` and `video` commands | :white_check_mark: `--generate-tts` flag |
| **Audio manifest** | :white_check_mark: JSON with timing info | :x: |
| **Resume/skip existing** | :white_check_mark: `--force` flag | :white_check_mark: `--no-generate-tts` |
| **Debug mode** | :white_check_mark: `MARP2VIDEO_DEBUG=1` | :white_check_mark: `DEBUG=1` |

## Unique Features

### marp2video only

- **OmniVoice provider abstraction** - Swap TTS/STT providers without code changes
- **Mixed TTS providers** - Use ElevenLabs for some slides, Deepgram for others
- **STT-based subtitles** - Word-level timestamps from actual audio transcription
- **Dictionary case correction** - Fix capitalization of tech terms in subtitles
- **JSON transcripts** - Per-slide language and voice overrides
- **Crossfade transitions** - Smooth transitions between slides
- **Individual slide export** - Ready for Udemy course uploads

### marptalk only

- **Browser TTS fallback** - No API costs for testing (Web Speech API)
- **YouTube chapter markers** - Auto-generated chapter timestamps
- **LLM-assisted drafting** - Generate presentations via GitHub Issues
- **Self-playing HTML** - Presentation with playback controls
- **Zero API key iteration** - Develop without any API credentials

## Architecture Comparison

=== "marp2video (Go)"

    ```
    marp2video/
    ├── cmd/marp2video/
    │   ├── tts.go
    │   ├── video.go
    │   └── subtitle.go
    ├── pkg/
    │   ├── parser/
    │   ├── transcript/
    │   ├── omnivoice/
    │   ├── video/
    │   └── orchestrator/
    └── go.mod
    ```

=== "marptalk (Node.js)"

    ```
    marptalk/
    ├── src/
    │   ├── generate.js (main)
    │   ├── extract-notes.js
    │   ├── generate-audio.js
    │   ├── generate-html.js
    │   ├── generate-subtitles.js
    │   └── generate-video.js
    └── package.json
    ```

## When to Use Which

| Use Case | Recommended |
|----------|-------------|
| Quick prototyping without API costs | **marptalk** (browser TTS fallback) |
| High-quality production voices | **marp2video** (ElevenLabs) |
| Multi-language presentations | **marp2video** (JSON transcripts) |
| YouTube publishing with chapters | **marptalk** |
| Udemy course creation | **marp2video** (individual slide export) |
| Mixed TTS providers | **marp2video** (OmniVoice) |
| Google Cloud ecosystem | **marptalk** |
| Voice cloning | **marp2video** (ElevenLabs) |
| LLM-assisted content creation | **marptalk** (GitHub Issues workflow) |

## Summary

**marp2video** is more feature-rich for production workflows with its provider abstraction layer (OmniVoice), multi-language support, and STT-based subtitle generation. It handles complex scenarios like mixed audio sample rates from different TTS providers.

**marptalk** excels at rapid iteration with its browser TTS fallback and YouTube-focused features (chapter markers). Its LLM-assisted drafting workflow via GitHub Issues is innovative for content creation.

## Links

- [marp2video on GitHub](https://github.com/grokify/marp2video)
- [marptalk on GitHub](https://github.com/imjasonh/marptalk)
- [OmniVoice - TTS/STT abstraction layer](https://github.com/agentplexus/omnivoice)
- [Marp - Markdown Presentation Ecosystem](https://marp.app/)
