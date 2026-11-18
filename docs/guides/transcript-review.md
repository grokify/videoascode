# Transcript Review

Read and approve a deck's narration, slide by slide, before any of it is sent to a TTS
provider. `vac html review` serves a page with the deck on one side and its transcript
on the other. Each approval is tied to the exact text it was made on, and
`vac html video --require-review` refuses to record until every slide is approved.

Review is cheaper than re-recording: a mispronounced acronym, a number read the wrong
way, or a missing translation caught here costs nothing, while the same mistake found in
a finished video costs a TTS run and a recording.

## Start a Review

```bash
vac html review --input deck/index.html --open
```

The page opens at `http://127.0.0.1:8765/`. Use `--port 0` for any free port. It takes the
same deck flags as `vac html init`, including `--pronunciations`, `--transcript`, and
`--start-fragment`.

```bash
# Record who made each decision, and use the deck's pronunciation list
vac html review --input deck/index.html --reviewer alex --pronunciations pronunciations.json

# Review only some languages
vac html review --input deck/index.html --lang en-US,fr-FR
```

## Review Several Decks Together

Name more than one deck, with `--input` and as arguments, to review a whole course, or
several courses, in one session:

```bash
vac html review courses/*/sections/*/index.html --pronunciations pronunciations.json --open
```

The first page lists every deck with its progress per language. Each deck's page has a
deck switcher and an **All decks** link. Each deck keeps its own transcript and review
file, so `--transcript`, `--review-file`, and `--export` work with one deck only.
`--summary` prints one row per deck and language.

## What the Page Shows

For the current slide, in the selected language:

- **The narration**, sentence by sentence, with each pause shown as a marker. Terms
  with a pronunciation entry are highlighted; hover over one to see how it is said.
- **What TTS receives**: the text after pronunciation substitutions, shown when it
  differs from the narration as written.
- **Words and estimated seconds** at 140 words a minute, plus the pause.
- **Checks** for text that TTS tends to read badly:

| Level | Check | Example |
|-------|-------|---------|
| Error | No narration and no pause | The slide would flash past |
| Error | Missing translation | The source language has narration; this language has none |
| Warning | Symbol | `&`, `/`, `+` may be read aloud or dropped |
| Warning | Address | `example.com` or an email address, read letter by letter |
| Warning | Unknown directive | `[BREAK:2]` would be spoken |
| Warning | Long sentence | More than 40 words |
| Info | Numbers | `80%`, `1M`, `$5`: listen for how they are read |
| Info | Acronym | An all-caps term with no pronunciation entry in this language |
| Info | Long slide | More than 90 seconds of narration |
| Info | Silent slide | A pause and no narration, which is often intended |

**Compare** shows a second language next to the first, for checking a translation against
its source. The outline on the left groups slides by lesson and marks each slide's state.

## Listen Before Recording

**Browser preview**, under the deck, reads the narration aloud with the browser's own
voice, at no cost and with no TTS key:

- **This slide** (`P`) reads the current slide, highlighting each sentence as it is
  spoken and keeping its pauses.
- **From here** (`Shift+P`) reads on through the following slides, moving the deck with
  it, so a whole lesson can be heard in order.
- `Esc`, the same button, or moving to another slide stops it.

The preview reads what TTS receives, after pronunciation substitutions, in the selected
language. Pick a voice and speed next to the buttons; the page remembers them for each
language. Voices come from the browser and the operating system, so they differ between
machines, and a voice marked *online* needs the network.

The preview is for wording, pronunciation and pacing. It does not sound like the
recording voice, and a pronunciation spelled for the TTS provider may sound slightly
different in the browser.

## Decisions

For each slide and language, either:

- **Approve** (`A`): the narration is ready for TTS. The page moves to the next slide
  that needs review.
- **Request changes** (`C`, type a note, then `Ctrl+Enter`): the note says what should
  change.

| State | Meaning |
|-------|---------|
| Not reviewed | No decision yet |
| Approved | The current text is approved |
| Changes requested | Changes were requested on the current text |
| Stale | A decision was made on text that has since changed |

A decision records a hash of the text TTS will receive and the pause after it. Editing the
narration, its pause, or a pronunciation it uses makes the decision stale, so the slide
needs review again. Decisions are saved to `voiceover/review.json` next to the transcript
(set another file with `--review-file`):

```json
{
  "version": "1.0",
  "slides": {
    "intro": {
      "en-US": {
        "status": "changes",
        "note": "Say \"five dollars a month\", not \"$5/mo\"",
        "textHash": "c1fed69be4d7",
        "reviewer": "alex",
        "updatedAt": "2026-10-26T11:56:51Z"
      }
    }
  }
}
```

Narration is edited where it lives: in the deck's `data-vo` attributes, or in the
transcript for translations. The page reads the deck and transcript from disk each time it
loads, so save the file and press `R` to see the change.

## Keyboard

| Key | Action |
|-----|--------|
| `←` `→` | Previous and next slide |
| `N` | Next slide that needs review |
| `A` | Approve |
| `C` | Write a note; `Ctrl+Enter` requests changes |
| `R` | Reload the deck and transcript |
| `P` | Listen to this slide with the browser's voice |
| `Shift+P` | Listen from this slide on |
| `Esc` | Stop listening |

## Require Review Before Recording

```bash
vac html video --input deck/index.html --lessons all --require-review
```

With `--require-review`, `vac html video` stops before any TTS call unless every slide it
would record is approved on its current text in every `--lang`. With `--lessons`, only
the slides of the selected lessons are checked. The error lists the slides that still
need review.

## Progress and Snapshots

```bash
# Review progress per language
vac html review --input deck/index.html --summary

# A read-only page for sharing the transcript or archiving a review
vac html review --input deck/index.html --export deck/transcript.html
```

A snapshot embeds the transcript and decisions and loads the deck by a path relative to
the snapshot, so keep the two together when you share them.

## Deck Navigation

The page shows the deck in a frame and moves it by changing its URL fragment
(`--start-fragment`, default `#{id}`), the same way `vac html video` opens a deck at a
lesson. Decks that follow `hashchange` events move without reloading. For a deck that
reads its fragment only when it loads, use `--nav reload`. The frame is laid out at
`--width` by `--height` (default 1920×1080) and scaled to fit.
