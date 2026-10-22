package htmldeck

import (
	"fmt"
	"strconv"
	"strings"
)

// Lesson is a run of consecutive slides recorded as one video.
type Lesson struct {
	// Number is the lesson's 1-based position in the deck.
	Number int
	// ID is the ID of the slide that starts the lesson, or a slug of the
	// deck title when the deck has no lesson starts.
	ID string
	// Title is the lesson attribute's value, else the starting slide's
	// title, else its ID.
	Title string
	// Slides are the lesson's slides, in order. Each keeps its deck Index.
	Slides []Slide
}

// First returns the deck index of the lesson's first slide.
func (l Lesson) First() int { return l.Slides[0].Index }

// Name returns a file-friendly name such as "02-designing-endpoints", from the
// lesson number and title.
func (l Lesson) Name() string {
	s := slug(l.Title)
	if s == "" {
		s = slug(l.ID)
	}
	if s == "" {
		return fmt.Sprintf("%02d", l.Number)
	}
	return fmt.Sprintf("%02d-%s", l.Number, s)
}

// Lessons divides the deck into lessons. A lesson starts at each slide with
// StartsLesson set; slides before the first such slide join the first
// lesson. A deck with no lesson starts is a single lesson.
func (d *Deck) Lessons() []Lesson {
	ids := d.SlideIDs()
	var lessons []Lesson
	for i, s := range d.Slides {
		if len(lessons) == 0 || (s.StartsLesson && hasLessonStart(lessons[len(lessons)-1])) {
			lessons = append(lessons, Lesson{Number: len(lessons) + 1})
		}
		l := &lessons[len(lessons)-1]
		if s.StartsLesson && !hasLessonStart(*l) {
			l.ID = ids[i]
			l.Title = firstNonEmpty(s.LessonTitle, s.Title, ids[i])
		}
		l.Slides = append(l.Slides, s)
	}
	if len(lessons) == 1 && lessons[0].ID == "" {
		lessons[0].ID = firstNonEmpty(slug(d.Title), "deck")
		lessons[0].Title = firstNonEmpty(d.Title, "Deck")
	}
	return lessons
}

// hasLessonStart reports whether one of l's slides starts a lesson, so a
// later start begins a new lesson rather than naming this one.
func hasLessonStart(l Lesson) bool {
	for _, s := range l.Slides {
		if s.StartsLesson {
			return true
		}
	}
	return false
}

// Lesson finds a lesson by ID or 1-based number.
func (d *Deck) Lesson(ref string) (Lesson, error) {
	lessons := d.Lessons()
	if n, err := strconv.Atoi(ref); err == nil {
		if n < 1 || n > len(lessons) {
			return Lesson{}, fmt.Errorf("lesson %d: the deck has %d lessons", n, len(lessons))
		}
		return lessons[n-1], nil
	}
	for _, l := range lessons {
		if l.ID == ref {
			return l, nil
		}
	}
	ids := make([]string, len(lessons))
	for i, l := range lessons {
		ids[i] = l.ID
	}
	return Lesson{}, fmt.Errorf("no lesson %q (lessons: %s)", ref, strings.Join(ids, ", "))
}
