package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/grokify/videoascode/pkg/config"
	"github.com/grokify/videoascode/pkg/segment"
	"github.com/grokify/videoascode/pkg/source"
	"github.com/grokify/videoascode/pkg/tts"
)

// planRetakes maps the slide IDs of --retake to their lessons: the result
// holds, by plan index, the positions of the slides to re-take within the
// lesson. Every slide must be in one of the selected lessons.
func planRetakes(plans []lessonPlan, selected []int, refs []string) (map[int][]int, error) {
	type place struct{ plan, pos int }
	where := map[string]place{}
	for i, p := range plans {
		for j, id := range p.entry.Slides {
			where[id] = place{i, j}
		}
	}
	isSelected := map[int]bool{}
	for _, i := range selected {
		isSelected[i] = true
	}

	out := map[int][]int{}
	seen := map[string]bool{}
	for _, ref := range refs {
		id := strings.TrimSpace(ref)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		w, ok := where[id]
		if !ok {
			return nil, fmt.Errorf("--retake: the deck has no slide %q", id)
		}
		if !isSelected[w.plan] {
			l := plans[w.plan].lesson
			return nil, fmt.Errorf("--retake: slide %q is in lesson %d (%s), which --lessons does not select", id, l.Number, l.ID)
		}
		out[w.plan] = append(out[w.plan], w.pos)
	}
	return out, nil
}

// lessonSegment loads the one segment of a lesson's video config.
func lessonSegment(cfg *config.VideoConfig) (segment.Segment, error) {
	segments, err := source.NewConfigSource(cfg).Load()
	if err != nil {
		return nil, err
	}
	if len(segments) != 1 {
		return nil, fmt.Errorf("a lesson config has %d segments, not one", len(segments))
	}
	return segments[0], nil
}

// retakeLessonClips removes the cached voiceover clips of the slides at
// positions in the lesson, in each language, so they are synthesized again
// while the lesson's other clips are reused. audioDir is the generator's
// audio cache, holding a directory per language.
func retakeLessonClips(w io.Writer, gen *tts.SegmentTTSGenerator, audioDir string, p lessonPlan, positions []int, languages []string) error {
	seg, err := lessonSegment(p.cfg)
	if err != nil {
		return err
	}
	for _, lang := range languages {
		vos := seg.GetVoiceovers(lang)
		if len(vos) != len(p.lesson.Slides) {
			return fmt.Errorf("%s: %d voiceovers for %d slides", lang, len(vos), len(p.lesson.Slides))
		}
		indexes := make([]int, len(positions))
		ids := make([]string, len(positions))
		for i, pos := range positions {
			indexes[i] = vos[pos].Index
			ids[i] = p.entry.Slides[pos]
		}
		clips, err := gen.RemoveClips(seg, lang, filepath.Join(audioDir, lang), indexes)
		if err != nil {
			return fmt.Errorf("%s: %w", lang, err)
		}
		// RemoveClips returns a clip per index, in order.
		for i, clip := range clips {
			fmt.Fprintf(w, "  re-take %s (%s): %s\n", ids[i], lang, filepath.Base(clip))
		}
	}
	return nil
}
