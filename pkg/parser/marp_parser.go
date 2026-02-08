package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Slide represents a single slide with content and voiceover
type Slide struct {
	Index        int
	Content      string
	Voiceover    string
	PauseDurations []int // milliseconds
	TotalPauseDuration int
}

// Presentation represents the parsed Marp presentation
type Presentation struct {
	Frontmatter string
	Slides      []Slide
}

// ParseMarpFile parses a Marp markdown file and extracts slides with voiceovers
func ParseMarpFile(content string) (*Presentation, error) {
	// Split frontmatter from slides
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid Marp file format: missing frontmatter")
	}

	presentation := &Presentation{
		Frontmatter: "---" + parts[1] + "---",
		Slides:      []Slide{},
	}

	// Split slides by ---
	slideContent := parts[2]
	slideTexts := strings.Split(slideContent, "\n---\n")

	for idx, slideText := range slideTexts {
		slide := parseSlide(idx, slideText)
		presentation.Slides = append(presentation.Slides, slide)
	}

	return presentation, nil
}

// parseSlide extracts voiceover from HTML comments and slide content
func parseSlide(index int, text string) Slide {
	slide := Slide{
		Index:          index,
		PauseDurations: []int{},
	}

	// Extract HTML comments (voiceover)
	commentRegex := regexp.MustCompile(`(?s)<!--\s*(.*?)\s*-->`)
	comments := commentRegex.FindAllStringSubmatch(text, -1)

	var voiceoverParts []string
	for _, comment := range comments {
		if len(comment) > 1 {
			voiceoverParts = append(voiceoverParts, comment[1])
		}
	}

	// Join all voiceover comments
	rawVoiceover := strings.Join(voiceoverParts, "\n")

	// Extract pause directives like [PAUSE:1000]
	pauseRegex := regexp.MustCompile(`\[PAUSE:(\d+)\]`)
	pauseMatches := pauseRegex.FindAllStringSubmatch(rawVoiceover, -1)

	for _, match := range pauseMatches {
		if len(match) > 1 {
			duration, err := strconv.Atoi(match[1])
			if err != nil {
				continue // Skip invalid pause directives
			}
			slide.PauseDurations = append(slide.PauseDurations, duration)
			slide.TotalPauseDuration += duration
		}
	}

	// Remove pause directives from voiceover text
	cleanVoiceover := pauseRegex.ReplaceAllString(rawVoiceover, "")
	slide.Voiceover = strings.TrimSpace(cleanVoiceover)

	// Remove comments from slide content
	slide.Content = commentRegex.ReplaceAllString(text, "")
	slide.Content = strings.TrimSpace(slide.Content)

	return slide
}
