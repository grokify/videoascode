# Contributing

Guidelines for contributing to vac.

## Getting Started

### Prerequisites

- Go 1.21+
- FFmpeg
- Marp CLI
- ElevenLabs API key (for testing TTS)

### Clone and Build

```bash
git clone https://github.com/grokify/videoascode.git
cd vac
go mod download
go build -o vac ./cmd/vac
```

### Run Tests

```bash
go test -v ./...
```

### Run Linter

```bash
golangci-lint run
```

## Code Style

### Formatting

Use `gofmt`:

```bash
gofmt -w .
```

### Linting

All code must pass `golangci-lint`:

```bash
golangci-lint run
```

### Error Handling

Follow this priority order:

1. **Panic**: For programming errors / invariant violations
2. **Return**: If the function can return an error
3. **Log**: If error cannot be returned, use slog via context:

```go
import "github.com/grokify/mogo/log/slogutil"

logger := slogutil.LoggerFromContext(ctx, nil)
logger.Error("operation failed", "error", err)
```

4. **Report**: If none of the above, report to the user

Never silently discard errors.

### Logging

Use structured logging via `slog`:

```go
logger := slogutil.LoggerFromContext(ctx, nil)

// Good
logger.Info("processing slide", "index", i, "duration", duration)

// Avoid
log.Printf("processing slide %d", i)
```

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Formatting (no code change) |
| `refactor` | Code restructuring |
| `perf` | Performance improvement |
| `test` | Adding tests |
| `build` | Build system changes |
| `ci` | CI configuration |
| `chore` | Maintenance |

Examples:

```
feat(tts): add Deepgram provider support

fix: handle empty slide notes gracefully

docs: add transcript schema reference

test: add parser edge case tests
```

## Pull Request Process

1. **Fork** the repository
2. **Create branch**: `git checkout -b feat/my-feature`
3. **Make changes** with tests
4. **Run checks**:
   ```bash
   go test -v ./...
   golangci-lint run
   ```
5. **Commit** with conventional commit message
6. **Push**: `git push origin feat/my-feature`
7. **Open PR** against `main`

### PR Checklist

- [ ] Tests pass locally
- [ ] Linter passes
- [ ] Commit messages follow conventions
- [ ] Documentation updated (if applicable)
- [ ] No breaking changes (or documented in PR)

## Adding Features

### New TTS Provider

1. Create `pkg/tts/<provider>.go`
2. Implement the TTS interface:
   ```go
   type TTSClient interface {
       Synthesize(ctx context.Context, text string) ([]byte, error)
   }
   ```
3. Add CLI flag for provider selection
4. Update documentation

### New Output Format

1. Update `pkg/video/combiner.go`
2. Add format-specific FFmpeg options
3. Add CLI flag
4. Update documentation

### New Transcript Feature

1. Update `pkg/transcript/transcript.go`
2. Update `pkg/transcript/transcript.schema.json`
3. Update orchestrator to use new feature
4. Add example in `examples/`
5. Update documentation

## Testing

### Unit Tests

Place tests alongside code:

```
pkg/parser/
├── marp_parser.go
└── marp_parser_test.go
```

Example:

```go
func TestParseMarpContent(t *testing.T) {
    input := `---
marp: true
---
# Slide 1
<!-- Voiceover text -->
`
    slides, err := parser.ParseMarpContent(input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(slides) != 1 {
        t.Errorf("expected 1 slide, got %d", len(slides))
    }
}
```

### Integration Tests

For tests requiring external services:

```go
func TestElevenLabsSynthesize(t *testing.T) {
    if os.Getenv("ELEVENLABS_API_KEY") == "" {
        t.Skip("ELEVENLABS_API_KEY not set")
    }
    // ...
}
```

## Documentation

### Code Comments

- Document exported functions and types
- Keep comments concise and useful
- Don't state the obvious

### MkDocs

Documentation lives in `docs/`:

```bash
# Install MkDocs
pip install mkdocs-material

# Preview locally
mkdocs serve

# Build
mkdocs build
```

## Release Process

1. Update version in code
2. Update CHANGELOG.md
3. Create PR with version bump
4. After merge, tag release:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

## Questions?

- Open an issue for bugs or feature requests
- Discussions for general questions
