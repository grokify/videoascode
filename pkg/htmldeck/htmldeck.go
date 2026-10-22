// Package htmldeck reads native HTML slide decks and produces the transcript
// and browser video config that videoascode records them with.
//
// A deck is any HTML page whose slides are elements carrying a "slide" class,
// advanced one at a time by a key press (ArrowRight by default). Each slide's
// narration comes from its data-vo attribute or, failing that, from a child
// <aside class="notes"> element (the reveal.js speaker-notes convention). A
// slide's identifier comes from data-id or id, and its title from data-title.
//
//	<section class="slide" data-id="intro" data-title="Introduction"
//	  data-vo="Welcome. [PAUSE:500] Let's begin.">
//	  …
//	</section>
//
// Narration may contain [PAUSE:ms] directives, as in Marp voiceover comments;
// they are removed from the spoken text and added as a pause after the slide.
//
// # Lessons
//
// A deck can be divided into lessons, each recorded as its own video. A
// slide with a data-lesson attribute starts a lesson; the attribute's value,
// if any, is the lesson's title. Slides before the first lesson start (a
// title slide, say) belong to the first lesson, and a deck with no lesson
// starts is one lesson. Typically one deck is a course section, and each
// lesson is one video in it:
//
//	<section class="slide section" data-id="sec-design" data-lesson="Designing Endpoints">
//
// Options.LessonClass also marks lesson starts by class, for decks whose
// divider slides already share one.
package htmldeck

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Default attribute and class names.
const (
	DefaultSlideClass     = "slide"
	DefaultNarrationAttr  = "data-vo"
	DefaultIDAttr         = "data-id"
	DefaultTitleAttr      = "data-title"
	DefaultNotesClass     = "notes"
	DefaultLessonAttr     = "data-lesson"
	DefaultNextKey        = "ArrowRight"
	DefaultInitialWaitMs  = 800
	DefaultLanguage       = "en-US"
	DefaultTranscriptVers = "1.0"
)

// Slide is one slide of a deck.
type Slide struct {
	// Index is the slide's zero-based position in the deck.
	Index int
	// ID is the slide's identifier, from data-id or id. It may be empty.
	ID string
	// Title is the slide's title, from data-title. It may be empty.
	Title string
	// Narration is the text to speak, with [PAUSE:ms] directives removed.
	Narration string
	// Pause is the total of the slide's [PAUSE:ms] directives, in
	// milliseconds, spoken as silence after the narration.
	Pause int
	// StartsLesson reports whether a lesson starts at this slide.
	StartsLesson bool
	// LessonTitle is the value of the slide's lesson attribute, if any.
	LessonTitle string
	// ContentHash is a hash of the slide's HTML, used to tell when a
	// recorded lesson is out of date.
	ContentHash string
}

// Deck is a parsed HTML deck.
type Deck struct {
	// Title is the page's <title> text.
	Title  string
	Slides []Slide
}

// Options selects how slides and narration are found. The zero value uses
// the defaults.
type Options struct {
	SlideClass    string // class marking a slide element; default "slide"
	NarrationAttr string // attribute holding narration; default "data-vo"
	IDAttr        string // attribute holding the slide ID; default "data-id"
	TitleAttr     string // attribute holding the slide title; default "data-title"
	NotesClass    string // class of the fallback notes element; default "notes"
	LessonAttr    string // attribute marking a lesson start; default "data-lesson"
	LessonClass   string // class also marking a lesson start; default none
}

func (o Options) withDefaults() Options {
	if o.SlideClass == "" {
		o.SlideClass = DefaultSlideClass
	}
	if o.NarrationAttr == "" {
		o.NarrationAttr = DefaultNarrationAttr
	}
	if o.IDAttr == "" {
		o.IDAttr = DefaultIDAttr
	}
	if o.TitleAttr == "" {
		o.TitleAttr = DefaultTitleAttr
	}
	if o.NotesClass == "" {
		o.NotesClass = DefaultNotesClass
	}
	if o.LessonAttr == "" {
		o.LessonAttr = DefaultLessonAttr
	}
	return o
}

var (
	pauseRe      = regexp.MustCompile(`\[PAUSE:(\d+)\]`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

// Parse reads an HTML deck. Slides nested inside another slide are not
// counted separately. It returns an error if the page has no slides.
func Parse(r io.Reader, opts Options) (*Deck, error) {
	opts = opts.withDefaults()
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	deck := &Deck{}
	var walk func(n *html.Node) error
	walk = func(n *html.Node) error {
		if n.Type == html.ElementNode {
			if n.Data == "title" && deck.Title == "" {
				deck.Title = normalizeSpace(textContent(n))
			}
			if hasClass(n, opts.SlideClass) {
				slide, err := readSlide(n, len(deck.Slides), opts)
				if err != nil {
					return err
				}
				deck.Slides = append(deck.Slides, slide)
				return nil // do not descend into a slide
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(doc); err != nil {
		return nil, err
	}
	if len(deck.Slides) == 0 {
		return nil, fmt.Errorf("no elements with class %q found", opts.SlideClass)
	}
	return deck, nil
}

func readSlide(n *html.Node, index int, opts Options) (Slide, error) {
	s := Slide{Index: index, Title: attr(n, opts.TitleAttr)}
	s.ID = attr(n, opts.IDAttr)
	if s.ID == "" {
		s.ID = attr(n, "id")
	}

	raw, ok := attrOK(n, opts.NarrationAttr)
	if !ok {
		if notes := findClass(n, opts.NotesClass); notes != nil {
			raw = textContent(notes)
		}
	}

	for _, m := range pauseRe.FindAllStringSubmatch(raw, -1) {
		ms, err := strconv.Atoi(m[1])
		if err != nil {
			return Slide{}, fmt.Errorf("slide %d: invalid pause %q: %w", index+1, m[0], err)
		}
		s.Pause += ms
	}
	s.Narration = normalizeSpace(pauseRe.ReplaceAllString(raw, " "))

	s.LessonTitle, s.StartsLesson = attrOK(n, opts.LessonAttr)
	s.LessonTitle = normalizeSpace(s.LessonTitle)
	if opts.LessonClass != "" && hasClass(n, opts.LessonClass) {
		s.StartsLesson = true
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, n); err != nil {
		return Slide{}, fmt.Errorf("slide %d: render: %w", index+1, err)
	}
	sum := sha256.Sum256(buf.Bytes())
	s.ContentHash = hex.EncodeToString(sum[:8])
	return s, nil
}

// SlideIDs returns each slide's ID, or "slide-N" (1-based) when it has none.
func (d *Deck) SlideIDs() []string {
	ids := make([]string, len(d.Slides))
	for i, s := range d.Slides {
		ids[i] = s.ID
		if ids[i] == "" {
			ids[i] = "slide-" + strconv.Itoa(i+1)
		}
	}
	return ids
}

// Unnarrated returns the 1-based numbers of slides with no narration.
func (d *Deck) Unnarrated() []int {
	var out []int
	for i, s := range d.Slides {
		if s.Narration == "" {
			out = append(out, i+1)
		}
	}
	return out
}

func hasClass(n *html.Node, class string) bool {
	return slices.Contains(strings.Fields(attr(n, "class")), class)
}

func findClass(n *html.Node, class string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && hasClass(c, class) {
			return c
		}
		if found := findClass(c, class); found != nil {
			return found
		}
	}
	return nil
}

func attr(n *html.Node, key string) string {
	v, _ := attrOK(n, key) // absent and empty both read as ""
	return v
}

func attrOK(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Namespace == "" && a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func normalizeSpace(s string) string {
	return strings.TrimSpace(whitespaceRe.ReplaceAllString(s, " "))
}
