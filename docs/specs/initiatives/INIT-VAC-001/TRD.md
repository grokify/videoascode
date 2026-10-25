# TRD — Reusable narration audio

**Initiative:** `INIT-VAC-001`
**Repository:** `github.com/grokify/videoascode`

## Types

```go
// pkg/narration
type Set struct {
    Version         string                                  `json:"version"`
    DefaultLanguage string                                  `json:"defaultLanguage"`
    Voices          map[string]transcript.VoiceConfig       `json:"voices,omitempty"` // by language
    Pronunciations  map[string]map[string]string            `json:"pronunciations,omitempty"`
    Units           []Unit                                  `json:"units"`
}
type Unit struct {
    ID      string            `json:"id"`                // e.g. "federated-authorization/challenge"
    Variant string            `json:"variant,omitempty"` // opaque, e.g. "curl"
    Texts   map[string]string `json:"texts"`             // by language
}

// pkg/audio
type Manifest struct {
    Version string `json:"version"`
    Format  string `json:"format"` // mp3
    Clips   []Clip `json:"clips"`
}
type Clip struct {
    ID         string `json:"id"`
    Variant    string `json:"variant,omitempty"`
    Language   string `json:"language"`
    File       string `json:"file"`       // relative to the manifest
    DurationMs int    `json:"durationMs"`
    TextSHA256 string `json:"textSha256"` // hex, of the text as written
    Voice      string `json:"voice"`      // "<provider>/<voiceId>"
    Words      []Word `json:"words,omitempty"`
}
type Word struct {
    Text    string `json:"text"`
    StartMs int    `json:"startMs"`
    EndMs   int    `json:"endMs"`
}
```

- Unit IDs match `^[A-Za-z0-9][A-Za-z0-9._/-]*$`; `(id, variant, language)` is unique in a set and in a manifest.
- Schemas are generated with `invopop/jsonschema` into `schema/` and linted with `--property-case camelCase`.

## Clip Cache (`pkg/tts/cache`)

- `Key(provider, model, voice VoiceConfig, language, spokenText) string`: SHA-256 over a canonical JSON encoding of those fields, hex.
- Layout: `<cache-dir>/<key[0:2]>/<key>.<format>` plus `<key>.json` with duration and word timings, written atomically (temp file and rename).
- `SegmentTTSGenerator` and the slide generator look up the cache before calling the provider. The existing `--audio-dir` files are still written (copied from the cache) so current workflows are unchanged.

## Commands

- `vac narration tts --input set.json --output audio/ [--lang ...] [--variant ...] [--concurrency N] [--force] [--dry-run]`.
- `browser video` and `slides video`: when `--audio-dir` is set, also write `audio-manifest.json` there, with IDs from steps, transcript segments or slide notes (`<!-- id: ... -->` in notes), and clips copied into `clips/`.

## Compatibility

- `Step` gains `ID string \`json:"id,omitempty"\``; `transcript.Segment` gains the same. Absent IDs produce clips without IDs, omitted from the manifest.
- The `slides tts` manifest is unchanged.
