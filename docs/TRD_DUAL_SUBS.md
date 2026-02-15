# TRD: Dual Language Subtitles for Language Learning

## Overview

Support for dual-language subtitles in marp2video presentations to enable language learning use cases. Users want to see subtitles in both their native language and the language they're learning simultaneously.

## Problem Statement

When learning a language via video content, users benefit from seeing:
1. Subtitles in the target language (what they're learning)
2. Subtitles in their native language (for comprehension)

Current marp2video supports single-language subtitle tracks, but not dual-language display.

## Use Cases

1. **Language Learner**: Native English speaker learning French wants to watch French audio with both French and English subtitles visible
2. **Bilingual Presentation**: Presenter wants to share content accessible to multiple language audiences simultaneously

## Technical Challenges

### Duration Mismatch

Different languages have different speech durations for the same content:

| Slide | English (s) | French (s) | Difference |
|-------|-------------|------------|------------|
| 0     | 34.7        | 49.1       | +14.3s     |
| 7     | 88.7        | 122.7      | +33.9s     |
| 9     | 92.4        | 119.0      | +26.6s     |
| Total | 16.0 min    | 20.3 min   | +4.3 min   |

This means subtitles timed to one language's audio won't sync properly with another language's audio.

### Video Player Support

- **VLC Secondary Subtitles**: VLC has a "Secondary Sub Track" feature, but testing shows it may only display one track at a time (not simultaneously)
- **Embedded Dual Tracks**: MP4 can contain multiple subtitle tracks, but player support for simultaneous display varies

## Proposed Solutions

### Option 1: Stacked Bilingual Subtitles (Recommended for MVP)

Merge two subtitle files into one with stacked text:

```srt
1
00:00:00,080 --> 00:00:04,179
Two kinds of AI users are emerging
Deux types d'utilisateurs de l'IA émergent
```

**Pros:**
- Works in any video player
- Simple implementation
- Single subtitle track to manage

**Cons:**
- More screen space used
- Can't toggle languages independently
- Timing must match one audio track (other language's subtitles may be slightly off)

### Option 2: Extended Duration Video

Create a video with max(lang1, lang2) duration per slide:

1. Compare manifests to get max duration per slide
2. Generate video with extended slide durations
3. Pad shorter audio with silence or use longer audio
4. Time both subtitle tracks to the extended duration

**Pros:**
- Both subtitle tracks perfectly synced
- Can toggle languages independently (in supporting players)

**Cons:**
- More complex implementation
- Video duration increases
- May feel slow for native speakers of the shorter language

### Option 3: Interactive Web Player

Build a custom HTML5 video player that:
- Supports displaying multiple subtitle tracks
- Allows toggling each language on/off
- Syncs to primary audio track

**Pros:**
- Full control over display
- Best user experience

**Cons:**
- Requires web hosting
- Significant development effort
- Not portable as MP4

## Implementation Plan

### Phase 1: Stacked Bilingual Subtitles (MVP)

1. Add `marp2video subtitle merge` command:
   ```bash
   marp2video subtitle merge \
     --primary subtitles/fr-FR.srt \
     --secondary subtitles/en-US.srt \
     --output subtitles/fr-en-bilingual.srt
   ```

2. Options:
   - `--primary`: Primary language (timing source)
   - `--secondary`: Secondary language (translations)
   - `--position`: top/bottom stacking order
   - `--separator`: Optional visual separator between languages

### Phase 2: Extended Duration Video

1. Add `marp2video video` options:
   ```bash
   marp2video video \
     --input presentation.md \
     --manifests audio/en-US/manifest.json,audio/fr-FR/manifest.json \
     --duration-strategy max \
     --output bilingual.mp4
   ```

2. Duration strategies:
   - `max`: Use maximum duration per slide
   - `primary`: Use first manifest's duration
   - `padded`: Pad shorter audio with silence

### Phase 3: Multi-Track Subtitle Embedding

1. Enhance `--subtitles` flag to accept multiple files:
   ```bash
   marp2video video \
     --subtitles subtitles/en-US.srt:eng,subtitles/fr-FR.srt:fra
   ```

2. Embed multiple subtitle tracks with proper language metadata

## Related Changes in This Session

### marp2video Changes

1. **Dictionary-based case correction** (`pkg/tts/dictionary.go`):
   - Built-in dictionary with 200+ tech terms
   - Layered loading: built-in → user config → CLI flags → project local
   - Month/day name corrections added

2. **`--provider` override** (`pkg/tts/transcript_generator.go`):
   - `--provider` now overrides voice config's provider setting
   - Clears provider-specific settings (model, voiceID) when switching providers

3. **Timestamps caching** (`pkg/tts/subtitle_generator.go`):
   - Auto-saves `timestamps.json` to audio directory
   - Auto-detects cached timestamps to skip STT API calls

### Files Modified

- `cmd/marp2video/tts.go` - Provider override flag
- `cmd/marp2video/subtitle.go` - Dictionary flags, auto-detection
- `pkg/tts/transcript_generator.go` - ForceProvider logic, clear provider settings
- `pkg/tts/subtitle_generator.go` - Timestamps caching, case correction
- `pkg/tts/dictionary.go` - New file: dictionary system
- `pkg/tts/dictionary_test.go` - New file: dictionary tests

## References

- VLC Subtitle Documentation: https://wiki.videolan.org/Subtitles/
- SRT Format Specification: https://en.wikipedia.org/wiki/SubRip
- BCP-47 Language Tags: https://www.rfc-editor.org/info/bcp47
