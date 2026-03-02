# vac vs marptalk

Two tools solving the same problem: converting [Marp](https://marp.app/) markdown presentations into narrated videos with AI-generated voiceovers. Both emerged from the same insight—that presentations are more effective when you can see slides and hear narration together—but took different paths to get there.

## Origins

### marptalk

[Jason Hall](https://www.linkedin.com/in/imjasonh/) created marptalk after observing that short presentations with spoken narration and visual elements are an effective way to learn, but "not every bit of information on the internet is presented the way I prefer." His solution: a Node.js tool that takes Marp presentations with embedded speaker notes, generates audio via Google Cloud TTS, and stitches everything into a video with subtitles using Puppeteer and ffmpeg. The result processes in about a minute at minimal cost. Hall also added browser-based TTS fallback using the Web Speech API, enabling rapid iteration without any API costs during development.

See Jason's [original LinkedIn post](https://www.linkedin.com/posts/imjasonh_i-find-short-presentations-to-be-a-really-activity-7378140962096693248-h_Aw) introducing marptalk.

### vac

vac grew from production needs at [PlexusOne](https://github.com/plexusone), which publishes many Marp presentations and wanted to turn them into narrated videos. The project also served as a way to exercise the [OmniVoice](https://github.com/plexusone/omnivoice) libraries for multi-language and multi-provider TTS/STT workflows. Rather than being locked into a single provider, vac uses OmniVoice as a unified abstraction layer—allowing different providers for different languages or slides. A Chinese slide might use Deepgram while English slides use ElevenLabs, all in the same video. The tool also generates subtitles from actual audio transcription (STT) rather than estimating from word count, producing word-level timestamps and proper capitalization via dictionary-based correction.

## Philosophy

The projects reflect different design philosophies:

**marptalk** optimizes for rapid iteration and accessibility. Browser TTS fallback means you can preview presentations instantly without API keys. YouTube chapter markers are auto-generated. LLM-assisted drafting via GitHub Issues lets you generate first drafts from a topic description. It's designed for quick experimentation.

**vac** optimizes for production workflows and flexibility. The decoupled architecture separates audio generation from video creation. JSON transcripts support per-slide voice overrides and multi-language content in a single file. Individual slide export targets platforms like Udemy. It's designed for complex, repeatable pipelines.

Both tools demonstrate that with modern AI voice services, the gap between "slides with speaker notes" and "polished video content" can be bridged automatically.

---

## Feature Comparison

### Overview

| Aspect | vac | marptalk |
|--------|------------|----------|
| **Language** | Go | Node.js |
| **License** | MIT | Apache-2.0 |
| **CLI Framework** | Cobra | Commander |
| **Primary TTS** | ElevenLabs (via OmniVoice) | Google Cloud TTS |
| **Primary STT** | Deepgram (via OmniVoice) | N/A (duration-based timing) |

### TTS Provider Support

| Feature | vac | marptalk |
|---------|------------|----------|
| **ElevenLabs** | :white_check_mark: Primary | :x: |
| **Google Cloud TTS** | :x: | :white_check_mark: Primary |
| **Deepgram TTS** | :white_check_mark: Secondary | :x: |
| **Browser TTS Fallback** | :x: | :white_check_mark: (Web Speech API) |
| **Provider Abstraction** | :white_check_mark: OmniVoice | :x: Single provider |
| **Voice Cloning** | :white_check_mark: (ElevenLabs) | :x: |

### Multi-Language Support

| Feature | vac | marptalk |
|---------|------------|----------|
| **Multi-language transcripts** | :white_check_mark: JSON with per-slide locales | :x: Single language per run |
| **Locale codes** | BCP-47 (en-US, zh-Hans, etc.) | Language codes (en-US, es-ES) |
| **Per-slide voice override** | :white_check_mark: | :x: |
| **Mixed TTS providers per video** | :white_check_mark: (e.g., ElevenLabs + Deepgram) | :x: |

### Voiceover Input Format

| Feature | vac | marptalk |
|---------|------------|----------|
| **Inline HTML comments** | :white_check_mark: `<!-- voiceover text -->` | :white_check_mark: `<!-- speaker notes -->` |
| **JSON transcript** | :white_check_mark: Structured, multi-language | :x: |
| **Pause directives** | :white_check_mark: `[PAUSE:1000]` | :x: |
| **Per-segment voice settings** | :white_check_mark: | :x: |

### Subtitle Generation

| Feature | vac | marptalk |
|---------|------------|----------|
| **SRT output** | :white_check_mark: | :white_check_mark: |
| **VTT output** | :white_check_mark: | :x: |
| **Generation method** | STT transcription (actual audio) | Word count estimation (150 wpm) |
| **Word-level timestamps** | :white_check_mark: | :x: |
| **Dictionary case correction** | :white_check_mark: (tech terms, custom JSON) | :x: |
| **YouTube chapters** | :x: | :white_check_mark: |

### Video Generation

| Feature | vac | marptalk |
|---------|------------|----------|
| **Method** | Image-based (Marp PNG export) | Static slide screenshots |
| **Audio sync** | Manifest-based (actual durations) | Audio file duration analysis |
| **Soft subtitles** | :white_check_mark: | :white_check_mark: |
| **Hard subtitles (burned-in)** | :white_check_mark: | :white_check_mark: |
| **Crossfade transitions** | :white_check_mark: `--transition` | :x: |
| **Individual slide export** | :white_check_mark: (Udemy-ready) | :x: |
| **Mixed audio sample rates** | :white_check_mark: (filter_complex concat) | N/A (single provider) |

### Workflow

| Feature | vac | marptalk |
|---------|------------|----------|
| **Decoupled TTS/Video** | :white_check_mark: Separate `tts` and `video` commands | :white_check_mark: `--generate-tts` flag |
| **Audio manifest** | :white_check_mark: JSON with timing info | :x: |
| **Resume/skip existing** | :white_check_mark: `--force` flag | :white_check_mark: `--no-generate-tts` |
| **Debug mode** | :white_check_mark: `MARP2VIDEO_DEBUG=1` | :white_check_mark: `DEBUG=1` |

## Unique Features

### vac only

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

=== "vac (Go)"

    ```
    vac/
    ├── cmd/vac/
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
| High-quality production voices | **vac** (ElevenLabs) |
| Multi-language presentations | **vac** (JSON transcripts) |
| YouTube publishing with chapters | **marptalk** |
| Udemy course creation | **vac** (individual slide export) |
| Mixed TTS providers | **vac** (OmniVoice) |
| Google Cloud ecosystem | **marptalk** |
| Voice cloning | **vac** (ElevenLabs) |
| LLM-assisted content creation | **marptalk** (GitHub Issues workflow) |

## Summary

**vac** is more feature-rich for production workflows with its provider abstraction layer (OmniVoice), multi-language support, and STT-based subtitle generation. It handles complex scenarios like mixed audio sample rates from different TTS providers.

**marptalk** excels at rapid iteration with its browser TTS fallback and YouTube-focused features (chapter markers). Its LLM-assisted drafting workflow via GitHub Issues is innovative for content creation.

## Links

- [vac on GitHub](https://github.com/grokify/videoascode)
- [marptalk on GitHub](https://github.com/imjasonh/marptalk)
- [OmniVoice - TTS/STT abstraction layer](https://github.com/plexusone/omnivoice)
- [Marp - Markdown Presentation Ecosystem](https://marp.app/)
