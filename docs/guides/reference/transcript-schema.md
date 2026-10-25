# Transcript Schema

JSON schema reference for multi-language transcripts.

## Schema Location

```
pkg/transcript/transcript.schema.json
```

## Structure Overview

```json
{
  "version": "1.0",
  "metadata": { ... },
  "slides": [ ... ]
}
```

## Root Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | ✅ | Schema version (e.g., `"1.0"`) |
| `metadata` | Metadata | ✅ | Presentation-level settings |
| `slides` | Slide[] | ✅ | Array of slide transcripts |

## Metadata

```json
{
  "metadata": {
    "title": "My Presentation",
    "description": "Optional description",
    "defaultLanguage": "en-US",
    "defaultVoice": {
      "provider": "elevenlabs",
      "voiceId": "pNInz6obpgDQGcFmaJgB",
      "voiceName": "Adam"
    },
    "defaultVenue": "youtube",
    "tags": ["tutorial", "demo"]
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | ✅ | Presentation title |
| `description` | string | | Optional description |
| `defaultLanguage` | string | ✅ | BCP-47 locale (e.g., `en-US`) |
| `defaultVoice` | VoiceConfig | ✅ | Default TTS voice settings |
| `defaultVenue` | string | | Target platform |
| `tags` | string[] | | Organization tags |
| `custom` | object | | User-defined key-value pairs |

### Venue Options

| Value | Platform |
|-------|----------|
| `youtube` | YouTube |
| `udemy` | Udemy |
| `coursera` | Coursera |
| `edx` | edX |
| `instagram` | Instagram |
| `tiktok` | TikTok |
| `general` | General purpose |

## VoiceConfig

```json
{
  "provider": "elevenlabs",
  "voiceId": "pNInz6obpgDQGcFmaJgB",
  "voiceName": "Adam",
  "model": "eleven_multilingual_v2",
  "outputFormat": "mp3",
  "sampleRate": 44100,
  "speed": 1.0,
  "pitch": 0.0,
  "stability": 0.5,
  "similarityBoost": 0.75,
  "style": 0.2
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `provider` | string | | TTS provider (elevenlabs, deepgram, etc.) |
| `voiceId` | string | ✅ | Provider-specific voice ID |
| `voiceName` | string | | Human-readable name |
| `model` | string | | Provider-specific model |
| `outputFormat` | string | `mp3` | Audio format |
| `sampleRate` | int | | Sample rate (Hz) |
| `speed` | float | 1.0 | Speech speed (0.25 - 4.0) |
| `pitch` | float | 0.0 | Pitch adjustment (-1.0 to 1.0) |
| `stability` | float | | Voice consistency (ElevenLabs) |
| `similarityBoost` | float | | Voice similarity (ElevenLabs) |
| `style` | float | | Style exaggeration (ElevenLabs) |

## Slide

```json
{
  "index": 0,
  "title": "Welcome Slide",
  "transcripts": {
    "en-US": { ... },
    "es-ES": { ... }
  },
  "avatar": { ... },
  "notes": "Internal notes"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `index` | int | ✅ | Slide index (0-based) |
| `title` | string | | Slide title for reference |
| `transcripts` | object | ✅ | Locale to LanguageContent map |
| `avatar` | AvatarConfig | | Virtual avatar settings |
| `notes` | string | | Internal notes (not spoken) |

## LanguageContent

```json
{
  "en-US": {
    "voice": { ... },
    "segments": [ ... ],
    "timing": { ... }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `voice` | VoiceConfig | | Override voice for this language |
| `segments` | Segment[] | ✅ | Text segments |
| `timing` | TimingInfo | | Populated after TTS generation |

## Segment

```json
{
  "text": "Welcome to the presentation.",
  "pause": 500,
  "emphasis": "moderate",
  "rate": "medium",
  "pitch": "+2st",
  "ssml": { ... }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `text` | string | ✅ | Text to speak |
| `pause` | int | | Pause after segment (ms) |
| `emphasis` | string | | `none`, `moderate`, `strong` |
| `rate` | string | | `x-slow`, `slow`, `medium`, `fast`, `x-fast` |
| `pitch` | string | | Pitch adjustment |
| `voice` | VoiceConfig | | Override voice for segment |
| `ssml` | SSMLHints | | Additional SSML hints |

## SSMLHints

```json
{
  "breaks": ["400ms", "1s"],
  "emphasis": ["important", "keyword"],
  "prosody": "rate=\"slow\"",
  "sayAs": "date",
  "phoneme": "ˈɛksəmpl̩",
  "subAlias": "HTML"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `breaks` | string[] | Break durations |
| `emphasis` | string[] | Words to emphasize |
| `prosody` | string | Custom prosody |
| `sayAs` | string | Interpretation (date, time, etc.) |
| `phoneme` | string | IPA pronunciation |
| `subAlias` | string | Substitution text |

## TimingInfo

Populated after TTS generation:

```json
{
  "audioDuration": 5200,
  "pauseDuration": 1000,
  "totalDuration": 6200
}
```

| Field | Type | Description |
|-------|------|-------------|
| `audioDuration` | int | Audio duration (ms) |
| `pauseDuration` | int | Total pause duration (ms) |
| `totalDuration` | int | Total slide duration (ms) |

## AvatarConfig

For future HeyGen/Synthesia integration:

```json
{
  "provider": "heygen",
  "avatarId": "avatar_001",
  "position": "bottom-right",
  "size": "medium",
  "style": "professional"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `provider` | string | ✅ | heygen, synthesia, d-id |
| `avatarId` | string | ✅ | Provider-specific ID |
| `position` | string | | bottom-right, bottom-left, etc. |
| `size` | string | | small, medium, large |
| `style` | string | | Visual style |

## Complete Example

See `examples/intro/transcript.json` for a full working example.
