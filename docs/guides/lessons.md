# Lessons and Courses

Course platforms such as Udemy, Coursera, and Thinkific organize a course into sections,
and each section into short lessons, each its own video, often with a quiz closing the
section. vac can record an HTML deck the same way: one video per lesson, re-recording
only the lessons that change.

## Terms

| Term | Meaning | In vac |
| --- | --- | --- |
| **Course** | A complete program of study | Outside vac: your course platform |
| **Section** | A group of lessons, often ending with a quiz | Typically one deck |
| **Lesson** | One video a learner watches and completes (Udemy: *lecture*) | A run of slides, starting at a `data-lesson` slide |
| **Deck** | The HTML file the slides and narration are written in | `--input` |
| **Slide** | One screen of the deck | An element with the `slide` class |
| **Clip** | The TTS audio for one slide's narration | `clips/<hash>.mp3` in the audio directory |
| **Line** | One sentence of narration | A transcript segment |
| **Chapter** | A timestamp marker inside a single video (as on YouTube) | Not a structural unit |

The deck is where you write; the lesson is what you record and publish. Keeping the whole
section in one deck keeps its style, terminology, and narration consistent, while
recording per lesson keeps edits, retakes, and re-uploads small.

## Marking Lessons

A slide with a `data-lesson` attribute starts a lesson. The attribute's value is the
lesson's title; if it is empty, the slide's `data-title` or ID is used.

```html
<section class="slide" data-id="title" data-vo="Welcome to the course.">…</section>

<section class="slide" data-id="sec-basics" data-lesson="The Basics"
  data-vo="Let's start with the basics.">…</section>
<section class="slide" data-id="basics-2" data-vo="…">…</section>

<section class="slide" data-id="sec-advanced" data-lesson="Going Further"
  data-vo="Now for the advanced material.">…</section>
```

- Slides before the first lesson start (a title slide, say) belong to the first lesson.
- A deck with no lesson starts is one lesson.
- If the deck's divider slides already share a class, `--lesson-class section` marks
  lessons by that class too, without editing the deck.

List a deck's lessons, with their slides, estimated length, and recording status:

```bash
vac html lessons --input deck/index.html
```

```
Building Web APIs · 24 slides · 5 lessons

  #  ID                     Title                              Slides   ~Min  Status
  1  sec-intro              Introduction                            4    2.4  recorded (2:31)
  2  sec-design             Designing Endpoints                     5    4.1  changed
  3  sec-errors             Handling Errors                         6    5.5  not recorded
```

## Recording Lessons

```bash
# Every lesson
vac html video --input deck/index.html --lessons all

# One or more lessons, by ID or number
vac html video --input deck/index.html --lessons sec-design,3

# Only lessons not yet recorded, or whose content changed since
vac html video --input deck/index.html --lessons changed
```

Each lesson is recorded separately: vac opens the deck at the lesson's first slide, steps
through the lesson's slides, and stops. Videos go to `lessons/` next to the deck (change
it with `--output-dir`), named by number and title, such as `02-designing-endpoints.mp4`.
With `--lang en-US,fr-FR`, each additional language gets its own file, such as
`02-designing-endpoints_fr-FR.mp4`.

### Opening the Deck at a Slide

To record a lesson, vac appends a URL fragment that opens the deck at the lesson's first
slide. The default is `#` followed by the slide's ID, for example
`index.html?capture=1#sec-design`, so the deck must jump to the slide whose `data-id`
matches the fragment when it loads:

```js
var id = decodeURIComponent(location.hash.slice(1));
var start = slides.findIndex(function (s) { return s.getAttribute('data-id') === id; });
show(start < 0 ? 0 : start);
```

For decks with another convention, set `--start-fragment`, where `{id}` is the slide's
ID, `{n}` its 1-based number, and `{i}` its 0-based index. For reveal.js, which uses
`#/<index>`, pass `--start-fragment "#/{i}"`.

## The Lesson Manifest

`lessons.json` in the output directory lists every lesson in deck order, for importing
into a course platform:

```json
{
  "title": "Building Web APIs",
  "defaultLanguage": "en-US",
  "lessons": [
    {
      "number": 1,
      "id": "sec-intro",
      "title": "Introduction",
      "slides": ["title", "sec-intro", "why-apis", "rest-basics"],
      "videos": { "en-US": "01-introduction.mp4" },
      "durationMs": 151000,
      "fingerprint": "3f9a1c0e5b7d2a64"
    }
  ]
}
```

The manifest is written after each lesson, so an interrupted run keeps the lessons it
finished. Lessons not recorded in a run keep their earlier entries.

### What Counts as Changed

Each recorded lesson stores a fingerprint of what it was recorded from:

- the HTML of its slides, including the narration,
- its narration in every language, including translations in the transcript,
- the voice, pronunciations, timing, and resolution.

`vac html lessons` reports a lesson as **changed** when its current fingerprint differs,
and `--lessons changed` re-records exactly those lessons, plus any not yet recorded.
Changes to the deck's stylesheets, scripts, or images are not detected; re-record the
affected lessons by ID, or all of them.

## Voiceover Clips Are Cached

Each slide's narration is synthesized once into a clip named by a hash of the spoken text
(after pronunciations), the language, and the voice settings. Re-recording a lesson
reuses every clip whose narration is unchanged, so editing one slide sends only that
slide's narration to the TTS provider. Clips live in the audio directory, shared by every
lesson and deck: `--audio-dir` if given, else `audio/` in `--workdir` (by default a
`vac-browser-video` folder in the system temp directory).

## Writing Lessons That Stand Alone

Learners jump straight to a lesson, rewatch one, or come back days later, so:

- Open each lesson with a slide that says what it covers; the lesson's first slide is
  its title card.
- Avoid narration that leans on another lesson, such as "as we saw on the last slide",
  across a lesson boundary.
- Close the section's last lesson by pointing to the quiz or what comes next.
- Aim for lessons of about 2 to 8 minutes; `vac html lessons` shows each lesson's
  estimated length.
