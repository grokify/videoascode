# PLAN — Reusable narration audio

**Initiative:** `INIT-VAC-001`

1. **Cache first.** The content-addressed cache fixes stale audio for existing users and is the foundation for everything else.
2. **IDs and manifest.** Add voiceover IDs and write the manifest from existing video commands.
3. **Narration export.** Add the narration set types, schema and `vac narration tts`.
4. **Word timings** where providers return alignment (ElevenLabs first).
5. **Docs and release.**

Testing uses a fake TTS provider that returns deterministic audio and durations, so cache hits, misses and manifest contents are unit tested without API keys.

The first consumer is PIDL (`INIT-PIDL-005`), which exports narration sets from protocol walkthroughs and plays the clips in its viewer.
