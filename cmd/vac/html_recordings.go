package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/academyosio/course-spec/deck"

	"github.com/grokify/videoascode/pkg/recording"
	"github.com/grokify/videoascode/pkg/source"
)

// planRecordingStatus sets each plan's status from its course recording,
// for --audio-only: recorded when current in every --lang, changed when
// out of date in any, not recorded otherwise.
func planRecordingStatus(plans []lessonPlan) error {
	sectionID, err := recording.SectionID(hdInput)
	if err != nil {
		return err
	}
	for _, p := range plans {
		first := p.entry.Slides[0]
		status := statusRecorded
		for _, lang := range bvLanguages {
			st, err := recording.Status(hdInput, sectionID, first, lang, hdLanguage)
			if err != nil {
				return fmt.Errorf("lesson %d: %w", p.lesson.Number, err)
			}
			switch {
			case st == deck.RecordingStale && status == statusRecorded:
				status = statusChanged
			case st == deck.RecordingMissing:
				status = statusNotRecorded
			}
		}
		plans[p.lesson.Number-1].status = status
	}
	return nil
}

// writeLessonAudio synthesizes the lesson's audio in each --lang, reading
// cached clips when they exist, and returns the files of its YouTube
// chapters by language. With videoBase, the lesson video's path without
// ".mp4", chapters are written beside the video. With write, the audio,
// captions and timings go into the deck's recordings/.
func writeLessonAudio(ctx context.Context, in *htmlDeckInputs, p lessonPlan, videoBase string, write bool) (map[string]string, error) {
	ts, err := newTTSSetup(p.cfg)
	if err != nil {
		return nil, err
	}
	segments, err := source.NewConfigSource(p.cfg).Load()
	if err != nil {
		return nil, err
	}
	if len(segments) != 1 {
		return nil, fmt.Errorf("a lesson config has %d segments, not one", len(segments))
	}
	seg := segments[0]
	results, err := ts.gen.GenerateMultiLanguage(ctx, segments, bvLanguages, ts.audioDir)
	if err != nil {
		return nil, fmt.Errorf("synthesize the narration: %w", err)
	}
	sectionID, err := recording.SectionID(hdInput)
	if err != nil {
		return nil, err
	}
	ids := in.deck.SlideIDs()

	chapters := map[string]string{}
	for i, lang := range bvLanguages {
		res := results[lang][seg.GetID()]
		if res == nil {
			return nil, fmt.Errorf("no %s audio", lang)
		}
		vos := seg.GetVoiceovers(lang)
		if len(vos) != len(p.lesson.Slides) {
			return nil, fmt.Errorf("%s: %d voiceovers for %d slides", lang, len(vos), len(p.lesson.Slides))
		}
		slides := make([]recording.Slide, len(p.lesson.Slides))
		for j, s := range p.lesson.Slides {
			tslide := in.transcript.Slides[s.Index]
			var sentences []string
			for _, g := range tslide.Transcripts[lang].Segments {
				if t := strings.TrimSpace(g.Text); t != "" {
					sentences = append(sentences, t)
				}
			}
			slides[j] = recording.Slide{
				ID:        ids[s.Index],
				Title:     tslide.Title,
				Sentences: sentences,
				SpeechMs:  res.VoiceoverDurations[vos[j].Index],
				PauseMs:   vos[j].Pause,
			}
		}

		if videoBase != "" {
			if lines := recording.Chapters(slides, 0); lines != nil {
				path := videoBase + ".chapters.txt"
				if i > 0 {
					path = videoBase + "_" + lang + ".chapters.txt"
				}
				if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil { //nolint:gosec // output meant to be read
					return nil, err
				}
				chapters[lang] = filepath.Base(path)
			}
		}

		if !write {
			continue
		}
		rec, err := recording.Write(recording.Lesson{
			DeckPath:        hdInput,
			SectionID:       sectionID,
			Language:        lang,
			DefaultLanguage: hdLanguage,
			AudioFile:       res.AudioFile,
			Slides:          slides,
			Voice:           deck.RecordingVoice{Provider: ts.voice.Provider, VoiceID: ts.voice.VoiceID, Name: ts.voice.VoiceName, Model: ts.voice.Model},
		})
		if err != nil {
			return nil, fmt.Errorf("%s recording: %w", lang, err)
		}
		fmt.Printf("  %s: recordings/%s (%d:%02d)\n", lang, rec.Audio, rec.DurationMs/60000, rec.DurationMs/1000%60)
	}
	if len(chapters) == 0 {
		return nil, nil
	}
	return chapters, nil
}
