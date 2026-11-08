package review

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Level is how serious a check finding is.
type Level string

const (
	// LevelError blocks a slide from being recorded as intended.
	LevelError Level = "error"
	// LevelWarn is text TTS is likely to read badly.
	LevelWarn Level = "warn"
	// LevelInfo is worth listening for once the audio exists.
	LevelInfo Level = "info"
)

// Finding is one check result for one slide in one language.
type Finding struct {
	Level   Level  `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Check thresholds.
const (
	// LongSentenceWords is the sentence length past which TTS phrasing
	// tends to flatten.
	LongSentenceWords = 40
	// LongSlideSeconds is the narration length past which a slide is
	// likely to hold the screen too long.
	LongSlideSeconds = 90
	// WordsPerMinute estimates narration length.
	WordsPerMinute = 140
)

var (
	directiveRe = regexp.MustCompile(`\[[A-Z]+(:[^\]]*)?\]`)
	urlRe       = regexp.MustCompile(`(?i)\bhttps?://\S+|\b[\w.+-]+@[\w-]+\.[\w.]+\b|\b[\w-]+\.(com|io|dev|org|net|ai|app|co)\b`)
	numberRe    = regexp.MustCompile(`[$€£]?\b\d[\d,.]*(%|[kKmMbB]\b|x\b)?`)
	acronymRe   = regexp.MustCompile(`\b[A-Z][A-Z0-9]{1,}s?\b`)
	symbolRe    = regexp.MustCompile(`[&/+=#@~*_|<>^\\{}]`)
	sentenceRe  = regexp.MustCompile(`[^.?!]+[.?!]*`)
)

// CheckInput is the narration of one slide in one language.
type CheckInput struct {
	// Text is the narration as written, with pause directives removed.
	Text string
	// PauseMs is the silence after the narration.
	PauseMs int
	// Language is the narration's BCP-47 language.
	Language string
	// Pronunciations maps terms to spoken forms, per language.
	Pronunciations map[string]map[string]string
	// Missing reports that the slide has narration in the deck's default
	// language but none in this one.
	Missing bool
}

// Check returns findings about narration that TTS is likely to read badly,
// most serious first.
func Check(in CheckInput) []Finding {
	var out []Finding
	add := func(l Level, code, format string, args ...any) {
		out = append(out, Finding{Level: l, Code: code, Message: fmt.Sprintf(format, args...)})
	}
	text := strings.TrimSpace(in.Text)

	if in.Missing {
		add(LevelError, "missing-translation", "No %s narration for this slide", in.Language)
		return out
	}
	if text == "" {
		if in.PauseMs > 0 {
			add(LevelInfo, "silent", "Silent slide: %d ms pause and no narration", in.PauseMs)
		} else {
			add(LevelError, "no-narration", "No narration and no pause; the slide will flash past")
		}
		return out
	}

	if m := directiveRe.FindAllString(text, -1); len(m) > 0 {
		add(LevelWarn, "unknown-directive", "Directive TTS will read aloud: %s", strings.Join(unique(m), ", "))
	}
	if m := urlRe.FindAllString(text, -1); len(m) > 0 {
		add(LevelWarn, "url", "Address read letter by letter or skipped: %s", strings.Join(unique(m), ", "))
	}
	// Text with addresses, directives and pronunciation terms removed, so
	// they are not also reported as symbols, numbers or acronyms.
	plain := withoutTerms(directiveRe.ReplaceAllString(urlRe.ReplaceAllString(text, " "), " "), in.Language, in.Pronunciations)
	if m := symbolRe.FindAllString(plain, -1); len(m) > 0 {
		add(LevelWarn, "symbol", "Symbol may be read aloud or dropped: %s", strings.Join(unique(m), " "))
	}
	for _, s := range sentenceRe.FindAllString(text, -1) {
		if n := len(strings.Fields(s)); n > LongSentenceWords {
			add(LevelWarn, "long-sentence", "%d-word sentence; consider splitting: %q", n, clip(s, 60))
		}
	}

	if m := numberRe.FindAllString(plain, -1); len(m) > 0 {
		add(LevelInfo, "numbers", "Check how numbers are read: %s", strings.Join(unique(m), ", "))
	}
	var unknown []string
	for _, a := range unique(acronymRe.FindAllString(plain, -1)) {
		if !hasPronunciation(in.Pronunciations, strings.TrimSuffix(a, "s"), in.Language) &&
			!hasPronunciation(in.Pronunciations, a, in.Language) {
			unknown = append(unknown, a)
		}
	}
	if len(unknown) > 0 {
		add(LevelInfo, "acronym", "Acronym with no pronunciation entry: %s", strings.Join(unknown, ", "))
	}
	if secs := Seconds(text, in.PauseMs); secs > LongSlideSeconds {
		add(LevelInfo, "long-slide", "About %.0f seconds on one slide", secs)
	}

	slices.SortStableFunc(out, func(a, b Finding) int { return rank(a.Level) - rank(b.Level) })
	return out
}

// Words counts the words in text.
func Words(text string) int { return len(strings.Fields(text)) }

// Seconds estimates how long text takes to say, plus its pause.
func Seconds(text string, pauseMs int) float64 {
	return float64(Words(text))/WordsPerMinute*60 + float64(pauseMs)/1000
}

// withoutTerms blanks out every term with a pronunciation in lang, longest
// first, so "AI-DLC" is not also checked as "AI" and "DLC".
func withoutTerms(text, lang string, p map[string]map[string]string) string {
	terms := make([]string, 0, len(p))
	for t, byLang := range p {
		if _, ok := byLang[lang]; ok && t != "" {
			terms = append(terms, t)
		}
	}
	slices.SortFunc(terms, func(a, b string) int { return len(b) - len(a) })
	for _, t := range terms {
		text = strings.ReplaceAll(text, t, " ")
	}
	return text
}

func hasPronunciation(p map[string]map[string]string, term, lang string) bool {
	_, ok := p[term][lang]
	return ok
}

func rank(l Level) int {
	switch l {
	case LevelError:
		return 0
	case LevelWarn:
		return 1
	default:
		return 2
	}
}

func unique(in []string) []string {
	var out []string
	for _, s := range in {
		s = strings.TrimRight(s, ".,")
		if s != "" && !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
}
