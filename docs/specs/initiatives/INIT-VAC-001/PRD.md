# PRD — Reusable narration audio

**Initiative:** `INIT-VAC-001`
**Repository:** `github.com/grokify/videoascode`

## Problem

videoascode synthesizes narration to make videos, and that audio is useful beyond the video. A web walkthrough of the same content (for example, a PIDL protocol viewer) can play the same clips when a reader steps forward or presses play, so the website and the video sound identical and the text-to-speech cost is paid once.

Today the audio cannot be reused that way:

- **Clips are named by position** (`segment_000/voiceover_001.mp3`). Nothing records which piece of content a clip speaks, so another tool cannot find the clip for a given step.
- **Stale audio is reused.** The cache is keyed by position, so editing a voiceover keeps the old audio until someone deletes the file by hand.
- **Only video inputs produce audio.** A web player needs a clip for every step in every language, including content that is not in any video, and content that comes in several forms (the same step explained for HTTP, cURL or Go code).
- **There is no manifest across commands.** `slides tts` writes a per-slide manifest; browser videos write timing files per segment. Neither carries content identity, text hashes or variants.

## Goals

- Any tool can hand videoascode a set of narration units and get back one clip per unit, language and variant, with a manifest that names each clip by the unit's ID.
- Audio is cached by what is spoken and how, so unchanged text is never synthesized twice, edited text is always synthesized again, and video builds and web exports share clips.
- Videos can name their voiceovers with the same IDs, so the audio made for a video is also usable on the web.

## Non-Goals

- Playing audio in web pages. Players (such as the PIDL viewer) read the manifest themselves.
- Knowing what a variant means. Variants are opaque keys chosen by the producer (PIDL uses representations such as `http`, `curl`, `go`).
- Translating text. Producers supply the text for each language.

## Requirements

1. **Content-addressed clip cache.** A clip's key is the SHA-256 of the provider, model, voice settings, language and the spoken text after pronunciations are applied. Clips are stored once per key in a cache directory (`--cache-dir`, default under the user cache directory) and reused by `slides`, `browser` and `narration` commands. A changed text or voice yields a new key, so stale audio is never reused. `--no-cache` forces synthesis.
2. **Voiceover IDs.** Browser steps, segment transcripts and slide notes MAY carry an `id`. IDs flow into timing metadata and the audio manifest.
3. **Audio manifest** (`audio-manifest.json`): `version`, `format`, and `clips`, each with `id`, `variant` (optional), `language`, `file` (relative path), `durationMs`, `textSha256` (of the text as written, before pronunciations), and `voice` (provider and voice ID). Written by `browser video` and `slides video` when `--audio-dir` is set, and by `narration tts`.
4. **Narration set input.** A JSON document listing `units`, each with an `id`, optional `variant`, and `texts` keyed by language, plus `defaultLanguage`, per-language `voices`, and `pronunciations` in the existing pronunciation shape.
5. **`vac narration tts`.** Synthesizes every unit × language × variant in a narration set into `<out>/clips/<key>.<format>` and writes `<out>/audio-manifest.json`. Identical texts share one clip. Options: `--lang` to limit languages, `--variant` to limit variants, `--concurrency`, `--force`, `--dry-run` (reports how many clips would be synthesized and their character count, for cost).
6. **Word timings.** When the provider returns alignment, the manifest MAY carry `words` (`text`, `startMs`, `endMs`) per clip, for captions and highlighting.
7. **Schema.** Go types are the source of truth for the narration set and audio manifest; JSON Schemas are generated from them, camelCase, and embedded.

## Success Criteria

- Running `vac narration tts` twice with no changes synthesizes nothing the second time.
- Editing one unit's French text synthesizes exactly one new clip.
- A browser video built with voiceover IDs and `--audio-dir` writes a manifest whose clips a web player can look up by ID and language.
- A web player and a video built from the same text play byte-identical audio.

## Decisions

- **Key by content, name by ID.** The cache key guarantees freshness; the manifest's IDs give other tools a stable handle.
- **Variants are opaque.** videoascode stays protocol-agnostic.
- **The existing `--audio-dir` layout stays** for compatibility; the manifest is added beside it.
