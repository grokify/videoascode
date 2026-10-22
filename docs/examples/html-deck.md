# HTML Deck

A four-slide native HTML deck in two lessons, in
[`examples/html-deck/`](https://github.com/grokify/videoascode/tree/main/examples/html-deck),
narrated from the deck itself.

## What It Shows

- Narration in a `data-vo` attribute, and in a reveal.js-style `<aside class="notes">`
- A `[PAUSE:600]` directive adding silence after a slide
- A pronunciation file, so TTS says "V A C" while subtitles show `vac`
- Two lessons, marked with `data-lesson`, and opening at the slide named in the URL
  fragment so each lesson can be recorded on its own
- A deck built to record well: a fixed 1280×720 stage, instant advance, and `?capture=1`
  hiding the page number

## Slide Markup

```html
<section class="slide" data-id="narration" data-title="Where Narration Lives"
  data-lesson="Writing Narration"
  data-vo="Narration lives in a data-vo attribute on each slide. [PAUSE:600]">
  …
</section>

<section class="slide" data-id="notes" data-title="Speaker Notes">
  …
  <aside class="notes" hidden>
    This slide's narration comes from an aside element with the notes class.
  </aside>
</section>
```

## Record It

```bash
vac html video \
  --input examples/html-deck/index.html \
  --pronunciations examples/html-deck/pronunciations.json \
  --output examples/html-deck/output.mp4 \
  --provider deepgram --subtitles
```

## Record It as Lessons

```bash
vac html lessons --input examples/html-deck/index.html

vac html video \
  --input examples/html-deck/index.html \
  --pronunciations examples/html-deck/pronunciations.json \
  --lessons all
```

The title slide joins the first lesson, so this writes `01-writing-narration.mp4`
(slides 1–3), `02-recording.mp4` (slide 4), and `lessons.json` to
`examples/html-deck/lessons/`. After editing a slide, `--lessons changed` re-records only
its lesson. See [Lessons and Courses](../guide/lessons.md).

## Regenerate the Transcript and Config

```bash
vac html init \
  --input examples/html-deck/index.html \
  --pronunciations examples/html-deck/pronunciations.json
```

This rewrites `voiceover/transcript.json` and `deck.video.json`, keeping translations
already in the transcript. See [HTML Slide Decks](../guide/html-decks.md).
