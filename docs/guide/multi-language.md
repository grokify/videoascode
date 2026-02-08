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
