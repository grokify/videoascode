# CLAUDE.md - VideoAsCode (vac)

Project-specific instructions for Claude Code.

## Project Overview

VideoAsCode converts Marp presentations with voiceovers to video files. It parses Marp markdown, generates speech using TTS providers (via OmniVoice), and creates synchronized video with optional subtitles.

## Build and Run

```bash
# Build
go build -o bin/vac ./cmd/vac

# Run tests
go test ./...

# Run linter
golangci-lint run

# Example usage
./bin/vac generate slides.md --output video.mp4
```

## Package Structure

| Package | Responsibility |
|---------|----------------|
| `cmd/vac` | CLI entry point, command handling |
| `pkg/config` | Video configuration, resolution, segments |
| `pkg/parser` | Marp markdown parsing, voiceover extraction |
| `pkg/transcript` | Multi-language transcript data structures |
| `pkg/segment` | Slide and browser segment abstractions |
| `pkg/tts` | TTS generation, segment audio, subtitles |
| `pkg/omnivoice/tts` | OmniVoice TTS provider wrapper |
| `pkg/omnivoice/stt` | OmniVoice STT provider wrapper |
| `pkg/browser` | Browser automation for demo recording |
| `pkg/video` | Video generation, ffmpeg integration |
| `pkg/orchestrator` | End-to-end workflow coordination |
| `pkg/media` | Audio/video utilities (duration, etc.) |
| `pkg/renderer` | Marp slide rendering to images |
| `pkg/source` | Source file handling |
| `pkg/audio` | Audio processing utilities |

## Testing

### Using OmniVoice Mocks

For TTS/STT tests, use omnivoice-core mock providers instead of real API calls:

```go
import "github.com/plexusone/omnivoice-core/tts/providertest"

// Provider-specific mocks
mock := providertest.NewElevenLabsMock()
mock := providertest.NewDeepgramMock()
mock := providertest.NewOpenAIMock()

// Configurable behaviors
mock := providertest.NewMockProviderWithOptions(
    providertest.WithLatency(100 * time.Millisecond),
    providertest.WithError(providertest.ErrMockRateLimit),
    providertest.WithFailAfterN(3, providertest.ErrMockQuotaExceeded),
)

// WAV fixtures
fixture := providertest.GenerateWAVFixture(1000, 22050)
```

### Test Categories

- **Unit tests**: Use mocks, no external dependencies
- **Integration tests**: Require ffmpeg, ffprobe, or API keys
- Packages with ffmpeg dependencies (`pkg/media`, `pkg/video`) need integration tests

## Key Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/plexusone/omnivoice` | Unified TTS/STT interface |
| `github.com/plexusone/omnivoice-core` | Core types, mock providers |
| `github.com/chromedp/chromedp` | Browser automation |
| `github.com/spf13/cobra` | CLI framework |

## Path Security

User-provided paths in `cmd/` use nolint comments. Library code in `pkg/` validates paths:

```go
// In pkg/ - validate before file operations
if strings.Contains(path, "..") {
    return fmt.Errorf("invalid path: contains '..' traversal sequence")
}
```

## External Tools

The following tools must be installed for full functionality:

- **ffmpeg/ffprobe**: Video/audio processing
- **marp-cli**: Slide rendering to PNG
- **Chrome/Chromium**: Browser demo recording (via chromedp)

## Environment Variables

```bash
ELEVENLABS_API_KEY=...  # TTS generation
DEEPGRAM_API_KEY=...    # STT/subtitle generation
```

## Workflow

1. Parse Marp markdown → extract slides and voiceovers
2. Generate TTS audio for each voiceover (cached)
3. Render slides to PNG images
4. (Optional) Record browser demos with voiceover sync
5. Combine images/video + audio → final video
6. (Optional) Generate subtitles via STT
