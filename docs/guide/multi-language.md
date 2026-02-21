# Multi-Language Support

Create videos in multiple languages from a single presentation.

## Locale Codes

marp2video uses BCP-47 locale codes:

| Code | Language |
|------|----------|
| `en-US` | English (United States) |
| `en-GB` | English (United Kingdom) |
| `es-ES` | Spanish (Spain) |
| `es-MX` | Spanish (Mexico) |
| `fr-FR` | French (France) |
| `fr-CA` | French (Canada) |
| `de-DE` | German (Germany) |
| `zh-Hans` | Chinese (Simplified) |
| `zh-Hant` | Chinese (Traditional) |
| `ja-JP` | Japanese |
| `ko-KR` | Korean |
| `pt-BR` | Portuguese (Brazil) |

## Transcript Structure

Define transcripts for each locale:

```json
{
  "version": "1.0",
  "metadata": {
    "defaultLanguage": "en-US",
    "defaultVoice": {
      "provider": "elevenlabs",
      "voiceId": "pNInz6obpgDQGcFmaJgB"
    }
  },
  "slides": [
    {
      "index": 0,
      "transcripts": {
        "en-US": {
          "segments": [{ "text": "Hello, welcome!" }]
        },
        "en-GB": {
          "segments": [{ "text": "Hello, welcome!" }]
        },
        "es-ES": {
          "voice": {
            "voiceId": "onwK4e9ZLuTAKqWW03F9",
            "voiceName": "Daniel"
          },
          "segments": [{ "text": "¡Hola, bienvenido!" }]
        },
        "fr-CA": {
          "voice": {
            "voiceId": "XrExE9yKIg1WjnnlVkGX",
            "voiceName": "Mathieu"
          },
          "segments": [{ "text": "Bonjour, bienvenue!" }]
        }
      }
    }
  ]
}
```

## Voice Per Language

Each language can have its own voice configuration:

```json
{
  "en-US": {
    "voice": {
      "provider": "elevenlabs",
      "voiceId": "pNInz6obpgDQGcFmaJgB",
      "voiceName": "Adam",
      "stability": 0.5,
      "similarityBoost": 0.75
    },
    "segments": [...]
  },
  "es-ES": {
    "voice": {
      "provider": "elevenlabs",
      "voiceId": "onwK4e9ZLuTAKqWW03F9",
      "voiceName": "Daniel"
    },
    "segments": [...]
  }
}
```

## Generating Videos

Generate a video for each language:

```bash
# English (US) - default
marp2video \
  --input slides.md \
  --transcript transcript.json \
  --output video_en-US.mp4

# English (UK)
marp2video \
  --input slides.md \
  --transcript transcript.json \
  --lang en-GB \
  --output video_en-GB.mp4

# Spanish (Spain)
marp2video \
  --input slides.md \
  --transcript transcript.json \
  --lang es-ES \
  --output video_es-ES.mp4

# French (Canada)
marp2video \
  --input slides.md \
  --transcript transcript.json \
  --lang fr-CA \
  --output video_fr-CA.mp4
```

## Fallback Behavior

If a requested language is not available for a slide:

1. Try the exact locale (e.g., `en-US`)
2. Fall back to `defaultLanguage` from metadata
3. Error if neither is available

## Batch Generation

Generate all languages with a script:

```bash
#!/bin/bash
LANGUAGES=("en-US" "en-GB" "es-ES" "fr-CA")

for lang in "${LANGUAGES[@]}"; do
  marp2video \
    --input slides.md \
    --transcript transcript.json \
    --lang "$lang" \
    --output "video_${lang}.mp4"
done
```

---

## Browser Video Multi-Language

The `browser-video` command generates all language versions in a single run:

```bash
# Generate English, French, and Chinese versions at once
marp2video browser video --config demo.yaml --output demo.mp4 \
  --lang en-US,fr-FR,zh-Hans
```

### Output Files

```
demo.mp4          # Primary language (first in list: en-US)
demo_fr-FR.mp4    # French version (same video, French audio)
demo_zh-Hans.mp4  # Chinese version (same video, Chinese audio)
```

### Browser Config Format

Define voiceovers per language in your config:

```yaml
segments:
  - id: "segment_000"
    type: "browser"
    browser:
      url: "https://example.com"
      steps:
        - action: "wait"
          duration: 1000
          voiceover:
            en-US: "Welcome to our demo."
            fr-FR: "Bienvenue dans notre démo."
            zh-Hans: "欢迎来到我们的演示。"
        - action: "click"
          selector: "#button"
          voiceover:
            en-US: "Click the button to continue."
            fr-FR: "Cliquez sur le bouton pour continuer."
            zh-Hans: "点击按钮继续。"
```

### Pace to Longest Language

Different languages have different speech lengths for the same content:

| Language | Typical Length vs English |
|----------|---------------------------|
| English  | 1.0x (baseline)           |
| French   | 1.15-1.20x                |
| German   | 1.10-1.15x                |
| Spanish  | 1.10-1.15x                |
| Japanese | 0.90-0.95x                |
| Chinese  | 0.85-0.90x                |

marp2video automatically paces the video to the longest audio:

1. TTS audio is generated for all requested languages
2. For each step, the maximum voiceover duration is calculated
3. The browser action is timed to match the longest duration
4. This ensures all language versions sync correctly

**Example**: If English is 3 seconds but French is 4 seconds, the step will be 4 seconds long. The English audio will have 1 second of silence at the end.

### Audio Caching

Use `--audio-dir` to cache audio and speed up subsequent runs:

```bash
# First run: generates all TTS audio
marp2video browser video --config demo.yaml --output demo.mp4 \
  --audio-dir ./audio --lang en-US,fr-FR,zh-Hans

# Second run: reuses cached audio
marp2video browser video --config demo.yaml --output demo.mp4 \
  --audio-dir ./audio --lang en-US,fr-FR,zh-Hans
```

Cache structure:

```
audio/
├── en-US/
│   ├── segment_000.mp3
│   └── segment_000.json  # Timing metadata
├── fr-FR/
│   ├── segment_000.mp3
│   └── segment_000.json
└── zh-Hans/
    ├── segment_000.mp3
    └── segment_000.json
```
