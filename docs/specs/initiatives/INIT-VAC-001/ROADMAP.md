# ROADMAP — Reusable narration audio

**Initiative:** `INIT-VAC-001`
**Repository:** `github.com/grokify/videoascode`

Commits implementing an item carry `Refs: RMI-VAC-<NNN>`.

## Phase 1 — Content-Addressed Audio

**Theme:** Audio identified by what it says

- [ ] `RMI-VAC-001` Content-addressed TTS clip cache shared by slide, browser and narration commands
- [ ] `RMI-VAC-002` Voiceover IDs on browser steps, transcript segments and slide notes
- [ ] `RMI-VAC-003` Audio manifest written beside `--audio-dir` output
  - Depends on: `RMI-VAC-001`, `RMI-VAC-002`

## Phase 2 — Narration Export

**Theme:** Audio for every unit, language and variant

- [ ] `RMI-VAC-004` Narration set and audio manifest Go types with generated, embedded JSON Schemas
  - Depends on: `RMI-VAC-003`
- [ ] `RMI-VAC-005` `vac narration tts` with language and variant selection, concurrency and dry-run cost report
  - Depends on: `RMI-VAC-004`
- [ ] `RMI-VAC-006` Word timings in the manifest where the provider returns alignment
  - Depends on: `RMI-VAC-005`
- [ ] `RMI-VAC-007` Narration audio guide, changelog and release notes
  - Depends on: `RMI-VAC-005`
