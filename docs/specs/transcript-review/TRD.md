# TRD: Transcript Review

## Overview

`vac html review` lets a person read and approve a deck's narration, slide by slide and
language by language, before it is sent to a TTS provider. `vac html video
--require-review` refuses to record narration that has not been approved.

## Problem

Narration lives in `data-vo` attributes and transcript JSON, where it is hard to read in
order and impossible to read next to its slide. Errors show up only after TTS and
recording: an acronym spelled out wrongly, a symbol read aloud, a translation missing
on one slide. Each one costs a TTS run and a re-recording, and some are missed until a
viewer reports them.

## Goals

- Show each slide next to its narration, in any language in the transcript, with a
  second language alongside for checking translations.
- Show the text TTS will actually receive, after pronunciation substitutions.
- Flag text TTS tends to misread, without blocking on it.
- Record a decision per slide and language that goes stale when the text changes.
- Let the recording step require approval.
- Produce a read-only snapshot to share or archive.

## Non-goals

- Editing narration on the page. Narration stays in the deck (and translations in the
  transcript), which keeps a single source; the page reloads from disk.
- Playing the TTS provider's audio. Reviewing that audio is a later, separate step; the
  page previews with the browser's own speech instead (see Browser preview).
- Presenter mode for live talks. The same model and page could support it later.
- Multi-user review. The server is local (loopback) and single-reviewer; decisions are a
  file that can be committed and reviewed like any other.

## Design

### Package `pkg/review`

| Part | Purpose |
|------|---------|
| `File`, `Entry`, `TextHash` | The review file: decisions keyed by slide ID and language |
| `Check` | Findings at three levels: error, warning, info |
| `Build` → `Model` | Deck, transcript, pronunciations and review file joined into one JSON model |
| `NewHandler` | Serves the page, the model, decisions, and the deck's directory |
| `Export` | Writes the page with the model embedded, read-only |
| `NewMultiHandler` | Serves several decks: a deck list at `/`, each deck's page under `/d/<slug>/` |

The command layer (`cmd/vac/html_review.go`) only parses flags, builds the model through
the same deck loader as `html init` and `html video`, and starts the server.

### Staleness

A decision stores `TextHash`: a hash of the spoken text (after pronunciation
substitutions) and the pause that follows it. State is derived, never stored:

| Condition | State |
|-----------|-------|
| No decision | `unreviewed` |
| Hash differs from current text | `stale` |
| Hash matches, `approved` | `approved` |
| Hash matches, `changes` | `changes` |

Hashing the spoken form means that changing a pronunciation entry also invalidates the
approval, because the audio would change.

Decisions are keyed by slide ID (`data-id`, else `id`, else `slide-N`), so reordering
slides keeps them. Slides without IDs lose their decisions when slides move.

### Review file

`voiceover/review.json` next to the transcript, formatted for clean diffs. Writes are
serialized in the handler; each write reloads the model so the stored hash is the hash of
the text on disk at that moment, not of what the page last showed.

### Page

One HTML template with the CSS and JS inlined, shared by the live page and the
snapshot. The live page loads the model from `/api/model`; the snapshot embeds it.
JSON is marshalled with HTML escaping, so the embedded model cannot close its script
element.

The deck is shown in a frame laid out at the recording size and scaled to fit. It moves
by fragment (`location.replace` with a new hash, which also works across origins for
snapshots opened from `file://`), or by reloading with `--nav reload` for decks that
read their fragment only on load. A shield over the frame keeps clicks from moving the
deck out of step with the page.

### Several decks

A course is several decks. `NewMultiHandler` mounts one `NewHandler` per deck under
`/d/<slug>/` and lists them at `/` with each deck's progress. The page uses relative URLs
only (`api/model`, `deck/`), so the same page works at the root or under a prefix. Each
deck's model carries `Decks`, the links for the deck switcher. Slugs come from the deck's
file name, or its directory for `index.html`, made unique.

### Browser preview

The page reads narration aloud with the Web Speech API, through `@plexusone/speech`, a
browser speech library published from the SystemForge Web packages. Its standalone build
is copied to `pkg/review/static/vendor/speech.js` at a pinned version by
`make vendor-speech` and inlined in the page, so the live page and the snapshot work
offline and without a JavaScript build in this repository. The copy is never edited by
hand; upgrading means changing `SPEECH_VERSION` and copying again.

Each `Segment` in the model carries `Spoken`, its text after pronunciation substitutions,
so the preview speaks what TTS would receive, segment by segment, with each segment's
pause. The library splits long text into shorter utterances, because some browser voices
stop partway through long ones.

### Recording gate

`vac html video --require-review` builds the same model for the `--lang` languages and
fails, before any TTS call, listing every slide that is not `approved`. With
`--lessons`, only the selected lessons' slides are checked.

## Checks

| Code | Level | Rule |
|------|-------|------|
| `no-narration` | error | No text and no pause |
| `missing-translation` | error | Source language has text; this language has none |
| `unknown-directive` | warn | A bracketed directive other than `[PAUSE:ms]` remains |
| `url` | warn | URL, email address, or bare domain |
| `symbol` | warn | One of `& / + = # @ ~ * _ \| < > ^ \ { }` |
| `long-sentence` | warn | More than 40 words |
| `numbers` | info | Digits, currency, percentages, `k`/`M`/`B` suffixes |
| `acronym` | info | All-caps term with no pronunciation entry for the language |
| `long-slide` | info | More than 90 seconds at 140 words a minute |
| `silent` | info | A pause and no text |

A silent slide in the source language is silent in every language and is not reported
as a missing translation.

## Later

- Play cached TTS clips per slide, for a second review pass on audio.
- Preview with the configured TTS provider on request, for slides where the browser
  voice is not good enough to judge.
- Presenter mode: speaker notes, next slide, and timer in one window, slides in another.
- `vac html lessons` showing review state per lesson.
