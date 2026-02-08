---
marp: true
theme: default
paginate: true
backgroundColor: #1a1a2e
color: #eaeaea
style: |
  section {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  }
  h1 {
    color: #00d4ff;
  }
  h2 {
    color: #7b68ee;
  }
  code {
    background-color: #16213e;
    color: #00ff88;
    padding: 2px 8px;
    border-radius: 4px;
  }
  pre {
    background-color: #16213e;
    border-radius: 8px;
  }
  ul li {
    margin: 0.5em 0;
  }
  .highlight {
    color: #00d4ff;
    font-weight: bold;
  }
---

# marp2video

## Transform Markdown Presentations into Videos with AI Voiceovers

<!--
Welcome to marp2video, a command-line tool that transforms your Marp markdown presentations into professional videos with AI-generated voiceovers.
[PAUSE:500]
In this presentation, we'll explore what marp2video can do and how to use it.
-->

---

# What is marp2video?

- **Marp** presentations + **ElevenLabs** TTS = **Video**
- Automates the entire workflow:
  - Markdown to HTML rendering
  - Text-to-speech generation
  - Screen recording with audio sync
  - Video concatenation with transitions

<!--
Marp2video combines the power of Marp presentations with ElevenLabs text-to-speech technology to create polished videos automatically.
[PAUSE:300]
It handles the complete pipeline: rendering your slides, generating speech from your script, recording each slide, and combining everything into a final video.
-->

---

# The Problem It Solves

Creating video content from presentations typically requires:

1. Recording yourself speaking
2. Screen capture software
3. Video editing tools
4. Hours of manual work

**marp2video automates all of this.**

<!--
Creating video tutorials or course content traditionally requires significant effort.
[PAUSE:300]
You need to record audio, capture your screen, and then spend hours editing everything together.
[PAUSE:500]
Marp2video eliminates this manual work by automating the entire process.
-->

---

# How It Works

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Marp MD   │ ──▶ │  ElevenLabs │ ──▶ │   ffmpeg    │
│  + Comments │     │     TTS     │     │  Recording  │
└─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │ Final Video │
                                        │    (.mp4)   │
                                        └─────────────┘
```

<!--
Here's the workflow.
[PAUSE:300]
You write your presentation in Marp markdown, adding voiceover scripts as HTML comments.
[PAUSE:300]
Marp2video sends the text to ElevenLabs to generate natural-sounding speech.
[PAUSE:300]
Then it uses ffmpeg to record each slide synchronized with the audio, and combines everything into your final video.
-->

---

# Writing Voiceovers

Add voiceover text as HTML comments in your slides:

```markdown
# My Slide Title

Some bullet points here.

<!-- This text will be converted to speech
     and played while this slide is shown. -->
```

<!--
Writing voiceovers is simple.
[PAUSE:300]
Just add HTML comments to your slides containing the text you want spoken.
[PAUSE:300]
The tool extracts these comments and converts them to speech using AI.
-->

---

# Pause Directives

Control timing with pause directives:

```markdown
<!-- Welcome to my presentation.
     [PAUSE:1000]
     Let's get started with the first topic. -->
```

- `[PAUSE:500]` = half second pause
- `[PAUSE:1000]` = one second pause
- `[PAUSE:2000]` = two second pause

<!--
You can add pauses to your voiceover for better pacing.
[PAUSE:500]
Use the pause directive with the duration in milliseconds.
[PAUSE:300]
This helps create natural breaks in the narration.
-->

---

# Installation

```bash
# Install the CLI
go install github.com/grokify/marp2video/cmd/marp2video@latest

# Requirements:
# - Marp CLI (npm install -g @marp-team/marp-cli)
# - ffmpeg
# - ElevenLabs API key
```

<!--
To install marp2video, use Go's install command.
[PAUSE:300]
You'll also need the Marp CLI from npm, ffmpeg for video processing, and an ElevenLabs API key for the text-to-speech functionality.
-->

---

# Basic Usage

```bash
# Set your API key
export ELEVENLABS_API_KEY="your-key-here"

# Generate video from presentation
marp2video --input slides.md --output video.mp4
```

<!--
Using marp2video is straightforward.
[PAUSE:300]
Set your ElevenLabs API key as an environment variable, then run the command with your input markdown file and desired output path.
-->

---

# Output Options

```bash
# Combined video for YouTube
marp2video --input slides.md --output video.mp4

# Individual videos for Udemy courses
marp2video --input slides.md \
           --output combined.mp4 \
           --output-individual ./lessons/

# With slide transitions
marp2video --input slides.md \
           --output video.mp4 \
           --transition 0.5
```

<!--
Marp2video supports multiple output options.
[PAUSE:300]
Generate a single combined video for YouTube, or export individual slide videos for platforms like Udemy.
[PAUSE:300]
You can also add smooth crossfade transitions between slides.
-->

---

# Customization

| Flag | Description |
|------|-------------|
| `--voice` | ElevenLabs voice ID |
| `--width` | Video width (default: 1920) |
| `--height` | Video height (default: 1080) |
| `--fps` | Frame rate (default: 30) |
| `--transition` | Transition duration in seconds |

<!--
The tool offers several customization options.
[PAUSE:300]
Choose different voices, adjust video resolution, set the frame rate, and configure transition effects between slides.
-->

---

# Use Cases

- **Online Courses** - Udemy, Coursera, Skillshare
- **YouTube Tutorials** - Technical walkthroughs
- **Product Demos** - Feature presentations
- **Training Materials** - Corporate learning
- **Documentation** - Animated guides

<!--
Marp2video is perfect for creating online course content, YouTube tutorials, product demonstrations, training materials, and animated documentation.
[PAUSE:500]
Any scenario where you need to turn slides into video content.
-->

---

# This Presentation

**Fun fact:** This video was created using marp2video!

```bash
# Using inline comments:
marp2video \
  --input examples/intro/presentation.md \
  --output examples/intro/output.mp4

# Or using transcript.json for multi-language:
marp2video \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --lang es \
  --output examples/intro/output_es.mp4
```

A self-documenting example.

<!--
Here's something meta.
[PAUSE:300]
This very presentation was converted to video using marp2video itself.
[PAUSE:500]
It's a self-contained, self-documenting example of what the tool can do.
-->

---

# Get Started

1. Install the prerequisites
2. Write your Marp presentation
3. Add voiceover comments
4. Run marp2video
5. Share your video!

**Repository:** `github.com/grokify/marp2video`

<!--
Getting started is easy.
[PAUSE:300]
Install the prerequisites, write your presentation with voiceover comments, run the tool, and you'll have a professional video ready to share.
[PAUSE:500]
Check out the repository on GitHub for more examples and documentation.
[PAUSE:300]
Thanks for watching!
-->
