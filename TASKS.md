# Tasks

Pending tasks for VideoAsCode.

## Testing

- [ ] Add tests for `pkg/source` (currently 0% coverage)
- [ ] Add integration tests for `pkg/media` (requires ffmpeg/ffprobe)
- [ ] Add integration tests for `pkg/video` (requires ffmpeg)
- [ ] Increase `pkg/browser` coverage (currently 2.4%)

## Documentation

- [ ] Update CHANGELOG.md for next release

## Coverage Summary

| Package | Coverage | Notes |
|---------|----------|-------|
| `pkg/parser` | 96.9% | Good |
| `pkg/transcript` | 94.3% | Good |
| `pkg/config` | 72.0% | Good |
| `pkg/segment` | 63.5% | Acceptable |
| `pkg/tts` | 34.3% | Needs improvement |
| `pkg/omnivoice/tts` | 32.1% | Factory requires API keys |
| `pkg/omnivoice/stt` | 32.5% | Factory requires API keys |
| `pkg/video` | 26.6% | Needs ffmpeg integration tests |
| `pkg/orchestrator` | 13.1% | Complex integration |
| `pkg/browser` | 2.4% | Needs chromedp mocking |
| `pkg/source` | 0% | Needs unit tests |
| `pkg/media` | 0% | Needs ffmpeg integration tests |
| `pkg/audio` | 0% | Needs unit tests |
| `pkg/renderer` | 0% | Needs marp-cli integration tests |
