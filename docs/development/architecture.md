# Architecture

Technical architecture and design decisions.

## Overview

vac follows a pipeline architecture with distinct stages:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Parser    │───▶│     TTS     │───▶│  Recorder   │───▶│  Combiner   │
│             │    │             │    │             │    │             │
│ Marp → HTML │    │ Text → MP3  │    │ HTML → MP4  │    │ Parts → MP4 │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

## Package Structure

```
vac/
├── cmd/vac/     # CLI entry point
│   └── main.go
├── pkg/
│   ├── orchestrator/   # Pipeline coordination
│   │   └── orchestrator.go
│   ├── parser/         # Marp markdown parsing
│   │   ├── marp_parser.go
│   │   └── marp_parser_test.go
│   ├── htmldeck/       # HTML slide deck parsing → transcript + browser config
│   │   ├── htmldeck.go
│   │   ├── generate.go
│   │   ├── lesson.go     # Lessons: runs of slides recorded as one video
│   │   └── manifest.go   # lessons.json and lesson fingerprints
│   ├── tts/            # Text-to-speech generation
│   │   ├── segment_generator.go  # Per-segment TTS with a content-addressed clip cache
│   │   └── elevenlabs.go
│   ├── avatar/         # Optional avatar presenter generation (OmniAvatar)
│   │   ├── generate.go
│   │   └── concat.go
│   ├── video/          # Video recording, combining, and avatar overlay
│   │   ├── recorder.go
│   │   ├── combiner.go
│   │   ├── overlay.go
│   │   └── combiner_test.go
│   └── transcript/     # Multi-language transcript types
│       ├── transcript.go
│       └── transcript.schema.json
└── examples/           # Example presentations
    └── intro/
```

## Core Components

### Orchestrator

The orchestrator (`pkg/orchestrator/orchestrator.go`) coordinates the pipeline:

```go
type Orchestrator struct {
    config    Config
    parser    *parser.MarpParser
    tts       *tts.ElevenLabsClient
    recorder  *video.Recorder
}

func (o *Orchestrator) Run(ctx context.Context) error {
    // 1. Parse Marp markdown
    // 2. Generate TTS audio for each slide
    // 3. Record screen for each slide
    // 4. Combine into final video
}
```

### Parser

The parser (`pkg/parser/marp_parser.go`) extracts slides and voiceover text:

```go
type Slide struct {
    Index   int
    Content string
    Notes   string  // Speaker notes / voiceover text
}

func (p *MarpParser) Parse(input string) ([]Slide, error)
```

Voiceover extraction priority:

1. HTML comments: `<!-- Voiceover text -->`
2. Speaker notes: Content after `---` separator

### HTML Deck Parser

The `htmldeck` package (`pkg/htmldeck`) reads native HTML slide decks. It parses the
page with `golang.org/x/net/html`, treats each element with a `slide` class as a slide
(without descending into it), and reads narration from `data-vo` or a child
`<aside class="notes">`:

```go
deck, err := htmldeck.Parse(r, htmldeck.Options{})
t := deck.Transcript(htmldeck.TranscriptOptions{Existing: previous})
cfg, err := deck.VideoConfig(t, htmldeck.ConfigOptions{URL: deckURL})
```

`Transcript` keeps other languages from an existing transcript, so regenerating never
discards translations. `VideoConfig` produces one browser segment with one step per slide
and a segment-level transcript with one segment per step, which the browser segment times
step by step. `vac html video` serves the deck directory on a loopback port and passes
the config straight to the browser recording pipeline.

A slide with a `data-lesson` attribute starts a lesson. `Deck.Lessons` divides the deck
into lessons, and `LessonVideoConfig` builds a config for one of them: the deck URL plus a
fragment (`#{id}` by default) that opens the deck at the lesson's first slide, and one
step per slide of the lesson. `LessonFingerprint` hashes the lesson's slide HTML and its
config (minus the URL), and `Manifest` (`lessons.json`) records each lesson's videos,
duration, and fingerprint, so `vac html video --lessons changed` re-records only lessons
whose content changed.

```go
for _, l := range deck.Lessons() {
    cfg, err := deck.LessonVideoConfig(t, l, htmldeck.ConfigOptions{URL: deckURL})
    fp := htmldeck.LessonFingerprint(l, cfg)
    // record cfg to l.Name()+".mp4", store fp in the manifest
}
```

### Voiceover Clip Cache

`tts.SegmentTTSGenerator` synthesizes each voiceover into
`{audio-dir}/{language}/clips/{key}.mp3`, where `tts.ClipKey` hashes the spoken text
(after pronunciations), the language, and the voice settings. A segment's clips and
pauses are assembled into `{segment}.mp3`, whose metadata stores a fingerprint of the
clip keys and pauses; the assembled file is reused only while the fingerprint matches.
Unchanged voiceovers are therefore never synthesized twice, across segments, lessons, and
runs, and edited narration or a changed voice can never reuse stale audio.

### TTS Client

The TTS client (`pkg/tts/elevenlabs.go`) generates audio:

```go
type ElevenLabsClient struct {
    client  *elevenlabs.Client
    voiceID string
}

func (c *ElevenLabsClient) Synthesize(ctx context.Context, text string) ([]byte, error)
```

Uses `github.com/plexusone/go-elevenlabs` with OmniVoice compatibility.

### Video Recorder

The recorder (`pkg/video/recorder.go`) captures screen:

```go
type Recorder struct {
    config RecorderConfig
}

func (r *Recorder) RecordSlide(ctx context.Context, htmlPath string, duration time.Duration) (string, error)
```

Uses:

- Marp CLI to serve HTML presentation
- FFmpeg with AVFoundation for screen capture (macOS)

### Video Combiner

The combiner (`pkg/video/combiner.go`) joins videos:

```go
func CombineVideos(ctx context.Context, inputPaths []string, outputPath string) error
func CombineVideosWithTransitions(ctx context.Context, inputPaths []string, outputPath string, transition float64) error
```

Uses FFmpeg `filter_complex` concat for seamless joining.

#### Mixed Audio Sample Rate Handling

Different TTS providers output audio at different sample rates:

| Provider | Sample Rate |
|----------|-------------|
| ElevenLabs | 44100 Hz |
| Deepgram | 22050 Hz |
| OpenAI | 24000 Hz |

The combiner uses `filter_complex` concat instead of the concat demuxer to properly handle mixed sample rates. This approach:

1. Decodes all input audio streams
2. Concatenates them in the filter graph
3. Re-encodes to consistent 44100 Hz AAC output

This ensures videos can seamlessly combine audio from multiple TTS providers (e.g., ElevenLabs for some languages, Deepgram for others).

## Data Flow

### Without Transcript

```
slides.md
    │
    ▼
┌─────────────────────────────────┐
│ Parser.Parse()                  │
│ Extract slides + inline notes   │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ TTS.Synthesize() × N slides     │
│ Generate audio/slide_N.mp3      │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Recorder.RecordSlide() × N      │
│ Generate video/slide_N.mp4      │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Combiner.CombineVideos()        │
│ Generate output.mp4             │
└─────────────────────────────────┘
```

### With Transcript

```
slides.md + transcript.json
    │
    ▼
┌─────────────────────────────────┐
│ Load transcript for locale      │
│ Override voice settings         │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ TTS.Synthesize() with segments  │
│ Apply pauses, emphasis          │
└─────────────────────────────────┘
    │
    ▼
[Same recording/combining flow]
```

## External Dependencies

### Required

| Dependency | Purpose |
|------------|---------|
| FFmpeg | Video recording and combining |
| Marp CLI | Markdown to HTML conversion |
| ElevenLabs API | Text-to-speech synthesis |

### Go Modules

| Module | Purpose |
|--------|---------|
| `github.com/plexusone/go-elevenlabs` | ElevenLabs API client |
| `github.com/grokify/mogo` | Utilities (slog context) |
| `github.com/spf13/cobra` | CLI framework |

## Design Decisions

### Why Pipeline Architecture?

- **Modularity**: Each stage can be tested independently
- **Flexibility**: Easy to swap TTS providers or video tools
- **Debugging**: Intermediate files aid troubleshooting

### Why Separate Transcript Format?

- **Multi-language**: Single source, multiple outputs
- **Rich TTS Control**: Pauses, emphasis, voice switching
- **Future Avatar Support**: HeyGen/Synthesia integration ready

### Why FFmpeg Screen Recording?

- **Cross-platform**: Works on macOS, Linux, Windows
- **Reliable**: Mature, well-tested tool
- **Flexible**: Supports many output formats

### Why ElevenLabs?

- **Quality**: Natural-sounding voices
- **Multilingual**: Same voice across languages
- **API**: Clean REST API, good Go client

## Future Considerations

### Avatar Integration

The transcript schema includes `AvatarConfig` for future integration:

```go
type AvatarConfig struct {
    Provider  string `json:"provider"`  // heygen, synthesia, d-id
    AvatarID  string `json:"avatarId"`
    Position  string `json:"position"`
    Size      string `json:"size"`
}
```

### Provider Abstraction

OmniVoice compatibility enables provider switching:

```go
// Current: ElevenLabs
tts := elevenlabs.NewClient(apiKey)

// Future: Deepgram, OpenAI, etc.
tts := omnivoice.NewClient("deepgram", apiKey)
```

### Caching

Potential optimizations:

- Cache TTS audio by text hash
- Cache slide renders by content hash
- Skip unchanged slides on re-generation
