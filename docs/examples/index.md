# Examples

Example presentations demonstrating vac features.

## Available Examples

| Example | Description | Features |
|---------|-------------|----------|
| [Intro](intro.md) | Self-documenting introduction | Multi-language (en-US, en-GB, es-ES) |

## Example Structure

Each example follows a consistent structure:

```
examples/<name>/
├── presentation.md       # Marp source (with inline voiceovers)
├── transcript.json       # Multi-language transcript
└── output.mp4           # Generated video (after running)
```

## Running Examples

### Using Inline Comments

```bash
vac \
  --input examples/intro/presentation.md \
  --output examples/intro/output.mp4
```

### Using Transcript (English)

```bash
vac \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --output examples/intro/output_en-US.mp4
```

### Using Transcript (Spanish)

```bash
vac \
  --input examples/intro/presentation.md \
  --transcript examples/intro/transcript.json \
  --lang es-ES \
  --output examples/intro/output_es-ES.mp4
```

## Creating Your Own Example

1. Create directory:
   ```bash
   mkdir examples/my-example
   ```

2. Create `presentation.md`:
   ```markdown
   ---
   marp: true
   ---

   # Slide 1

   <!-- Voiceover for slide 1 -->
   ```

3. Optionally create `transcript.json` for multi-language

4. Generate video:
   ```bash
   vac \
     --input examples/my-example/presentation.md \
     --output examples/my-example/output.mp4
   ```

## Contributing Examples

We welcome example contributions! Please:

1. Follow the standard structure
2. Include both inline comments and transcript.json
3. Add at least 2 languages if using transcript.json
4. Test video generation before submitting
