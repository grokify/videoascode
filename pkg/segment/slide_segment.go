package segment

import (
	"fmt"

	"github.com/grokify/videoascode/pkg/transcript"
)

// SlideSegment represents a static slide from a presentation
type SlideSegment struct {
	index       int
	title       string
	imagePath   string
	transcripts map[string]transcript.LanguageContent
}

// NewSlideSegment creates a new slide segment
func NewSlideSegment(index int, title string, transcripts map[string]transcript.LanguageContent) *SlideSegment {
	return &SlideSegment{
		index:       index,
		title:       title,
		transcripts: transcripts,
	}
}

// NewSlideSegmentFromTranscript creates a slide segment from a transcript.Slide
func NewSlideSegmentFromTranscript(slide transcript.Slide) *SlideSegment {
	return &SlideSegment{
		index:       slide.Index,
		title:       slide.Title,
		transcripts: slide.Transcripts,
	}
}

// GetID returns the segment identifier
func (s *SlideSegment) GetID() string {
	return fmt.Sprintf("segment_%03d", s.index)
}

// GetIndex returns the segment position
func (s *SlideSegment) GetIndex() int {
	return s.index
}

// GetSourceType returns SourceTypeSlide
func (s *SlideSegment) GetSourceType() SourceType {
	return SourceTypeSlide
}

// GetTitle returns the slide title
func (s *SlideSegment) GetTitle() string {
	return s.title
}

// GetVoiceovers returns voiceovers for a specific language
// Slides have a single voiceover combining all segments
func (s *SlideSegment) GetVoiceovers(language string) []Voiceover {
	content, ok := s.transcripts[language]
	if !ok {
		return nil
	}

	// Combine all text segments into a single voiceover
	var fullText string
	var totalPause int
	for i, seg := range content.Segments {
		if i > 0 {
			fullText += " "
		}
		fullText += seg.Text
		totalPause += seg.Pause
	}

	if fullText == "" {
		return nil
	}

	return []Voiceover{
		{
			Index:    0,
			Text:     fullText,
			Language: language,
			Voice:    content.Voice,
			Pause:    totalPause,
		},
	}
}

// GetLanguages returns all available language codes
func (s *SlideSegment) GetLanguages() []string {
	languages := make([]string, 0, len(s.transcripts))
	for lang := range s.transcripts {
		languages = append(languages, lang)
	}
	return languages
}

// GetTranscripts returns the raw transcript data
func (s *SlideSegment) GetTranscripts() map[string]transcript.LanguageContent {
	return s.transcripts
}

// SetImagePath sets the path to the rendered slide image
func (s *SlideSegment) SetImagePath(path string) {
	s.imagePath = path
}

// GetImagePath returns the path to the rendered slide image
func (s *SlideSegment) GetImagePath() string {
	return s.imagePath
}

// GetFullText returns the complete voiceover text for a language
func (s *SlideSegment) GetFullText(language string) string {
	voiceovers := s.GetVoiceovers(language)
	if len(voiceovers) == 0 {
		return ""
	}
	return voiceovers[0].Text
}

// GetTotalPauseDuration returns the total pause duration for a language
func (s *SlideSegment) GetTotalPauseDuration(language string) int {
	content, ok := s.transcripts[language]
	if !ok {
		return 0
	}
	return content.GetTotalPauseDuration()
}
