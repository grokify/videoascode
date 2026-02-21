# Voice Settings

TTS voice configuration reference.

## ElevenLabs Voices

### Popular Voices

| Voice ID | Name | Gender | Accent | Best For |
|----------|------|--------|--------|----------|
| `pNInz6obpgDQGcFmaJgB` | Adam | Male | American | Professional, clear |
| `21m00Tcm4TlvDq8ikWAM` | Rachel | Female | American | Warm, natural |
| `onwK4e9ZLuTAKqWW03F9` | Daniel | Male | British | Authoritative |
| `XrExE9yKIg1WjnnlVkGX` | Matilda | Female | American | Friendly |
| `ErXwobaYiN019PkySvjV` | Antoni | Male | American | Conversational |
| `MF3mGyEYCl7XYWbV9V6O` | Elli | Female | American | Young, energetic |

### Voice Parameters

| Parameter | Range | Default | Description |
|-----------|-------|---------|-------------|
| `stability` | 0.0 - 1.0 | 0.5 | Voice consistency |
| `similarityBoost` | 0.0 - 1.0 | 0.75 | Closeness to original voice |
| `style` | 0.0 - 1.0 | 0.0 | Expressiveness |

### Stability

- **Low (0.0 - 0.3)**: More expressive, varied intonation
- **Medium (0.3 - 0.7)**: Balanced
- **High (0.7 - 1.0)**: Consistent, predictable

### Similarity Boost

- **Low (0.0 - 0.3)**: More variation from original voice
- **Medium (0.3 - 0.7)**: Balanced
- **High (0.7 - 1.0)**: Very close to original voice

## Venue Presets

Pre-configured settings for different platforms:

### YouTube

```json
{
  "stability": 0.45,
  "similarityBoost": 0.8,
  "style": 0.2
}
```

Optimized for sustained viewer attention.

### Udemy

```json
{
  "stability": 0.5,
  "similarityBoost": 0.75,
  "style": 0.0
}
```

Clear, consistent for learning.

### Coursera

```json
{
  "stability": 0.7,
  "similarityBoost": 0.85,
  "style": 0.2
}
```

Academic, engaging.

### TikTok

```json
{
  "stability": 0.3,
  "similarityBoost": 0.85,
  "style": 0.45
}
```

Energetic, immediate engagement.

### Instagram

```json
{
  "stability": 0.4,
  "similarityBoost": 0.85,
  "style": 0.35
}
```

Polished, engaging.

## Models

| Model ID | Description |
|----------|-------------|
| `eleven_multilingual_v2` | Best quality, multi-language |
| `eleven_monolingual_v1` | English only, faster |
| `eleven_turbo_v2` | Fastest, good quality |

## Output Formats

| Format | Sample Rate | Use Case |
|--------|-------------|----------|
| `mp3_22050_32` | 22050 Hz | Small files |
| `mp3_44100_128` | 44100 Hz | Standard quality |
| `pcm_16000` | 16000 Hz | Real-time |
| `pcm_44100` | 44100 Hz | High quality |

## OmniVoice Compatibility

vac's VoiceConfig aligns with OmniVoice's SynthesisConfig:

```go
type SynthesisConfig struct {
    VoiceID        string
    Model          string
    OutputFormat   string
    SampleRate     int
    Speed          float64
    Pitch          float64
    Stability      float64
    SimilarityBoost float64
}
```

This enables provider switching (ElevenLabs, Deepgram, etc.) without changing transcript format.
