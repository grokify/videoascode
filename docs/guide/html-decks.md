# HTML Slide Decks

vac records native HTML slide decks as well as Marp presentations. Write the narration
into the deck itself, and vac generates the transcript and video config, serves the deck,
and records it one slide at a time, timing each slide to its own narration.

## What vac Expects

A deck is an HTML page whose slides are elements with a `slide` class, advanced one slide
per key press (`ArrowRight` by default). Each slide's narration comes from:

1. its `data-vo` attribute, or
2. a child `<aside class="notes">` element (the reveal.js speaker-notes convention).

A slide's identifier comes from `data-id` (or `id`), and its title from `data-title`. Both
are optional; they label the steps and transcript entries.

```html
<section class="slide" data-id="intro" data-title="Introduction"
  data-vo="Welcome. This deck explains the protocol in five minutes. [PAUSE:500]">
  <h1>Introduction</h1>
</section>

<section class="slide" id="flow">
  <h2>The Flow</h2>
  <aside class="notes">The agent signs every request.</aside>
</section>
```

`[PAUSE:ms]` directives work as they do in Marp voiceover comments: they are removed from
the spoken text and added as silence after the slide.

Slides nested inside another slide are not counted separately. The class, narration
attribute, and next key can all be changed with flags (`--slide-class`,
`--narration-attr`, `--next-key`).

## Record a Deck

```bash
vac html video --input deck/index.html --output deck.mp4 --provider deepgram
```

vac serves the deck's directory on a free loopback port, so relative assets, fonts, and
scripts load as they would on a web server. It opens `index.html?capture=1` and presses
the next key once per slide. The `capture=1` query lets a deck hide its on-screen controls
during recording; change it with `--query`, or pass `--query ""` for none.

`vac html video` accepts the same recording, TTS, and subtitle flags as
`vac browser video`, including `--audio-dir`, `--lang`, `--subtitles`, `--headless`, and
`--limit-steps`.

## Generate the Transcript and Config

To keep the narration in a file you can translate, review, or render later:

```bash
vac html init --input deck/index.html --provider deepgram --voice aura-asteria-en
```

This writes two files next to the deck:

| File | Contents |
| --- | --- |
| `voiceover/transcript.json` | The narration in the [transcript format](../reference/transcript-schema.md), one entry per slide, split into sentences |
| `deck.video.json` | A config for `vac browser video`, one browser step per slide |

Re-run `vac html init` whenever the narration changes. It keeps what you added since:

- **Translations.** Other languages in the transcript are kept, matched to slides by
  position. A language is added to the config once it covers every slide.
- **Voice.** The transcript's voice is kept unless you pass `--provider` or `--voice`.
- **Pronunciations.** The config's pronunciations are kept unless you pass
  `--pronunciations` with a JSON file of the same shape as the config's
  [pronunciations](pronunciation.md) map.

`vac html video` reads the same transcript when it exists, so translations added there are
recorded too:

```bash
vac html video --input deck/index.html --output deck.mp4 --lang en-US,fr-FR
```

To record from the generated config instead, serve the deck yourself and use
`vac browser video`:

```bash
cd deck && python3 -m http.server 8000 &
vac browser video --config deck/deck.video.json --output deck.mp4
```

## Lessons

A slide with a `data-lesson` attribute starts a lesson, and `vac html video --lessons`
records each lesson as its own video with a `lessons.json` manifest, re-recording only
the lessons that change. See [Lessons and Courses](lessons.md).

```bash
vac html lessons --input deck/index.html
vac html video --input deck/index.html --lessons all
```

## Timing

Each slide stays on screen for exactly its own narration plus its pause, which is how the
audio track is assembled. This applies to any browser config whose segment-level
transcript has one segment per step; configs whose transcript doesn't line up with the
steps keep spreading the narration across steps proportionally.

## Making a Deck Record Well

- Render at a fixed size (for example a 1280×720 stage scaled to fit the window), so every
  frame has the same geometry.
- Advance instantly. Disable slide transitions and entrance animations in capture mode,
  since each step is captured as soon as the key is pressed.
- Respond to `?capture=1` by hiding progress bars, page numbers, and notes panels.
- Keep each slide's narration on the slide element, so the deck is its single source.
- Open at the slide named in the URL fragment (`#<data-id>`), so vac can record one
  lesson at a time.
