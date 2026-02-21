package source

import (
	"fmt"

	"github.com/grokify/videoascode/pkg/segment"
	"github.com/grokify/videoascode/pkg/transcript"
)

// TranscriptSource loads content from a transcript JSON file.
// It supports both slide and browser segments via the sourceType field.
type TranscriptSource struct {
	transcript *transcript.Transcript
	filePath   string
}

// NewTranscriptSource creates a source from a transcript
func NewTranscriptSource(t *transcript.Transcript) *TranscriptSource {
	return &TranscriptSource{
		transcript: t,
	}
}

// NewTranscriptSourceFromFile loads a transcript from a JSON file
func NewTranscriptSourceFromFile(path string) (*TranscriptSource, error) {
	t, err := transcript.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load transcript: %w", err)
	}
	return &TranscriptSource{
		transcript: t,
		filePath:   path,
	}, nil
}

// Load parses the transcript and returns segments
func (s *TranscriptSource) Load() ([]segment.Segment, error) {
	if s.transcript == nil {
		return nil, fmt.Errorf("transcript not loaded")
	}

	segments := make([]segment.Segment, len(s.transcript.Slides))
	for i, slide := range s.transcript.Slides {
		switch slide.GetEffectiveSourceType() {
		case transcript.SourceTypeSlide:
			segments[i] = segment.NewSlideSegmentFromTranscript(slide)
		case transcript.SourceTypeBrowser:
			segments[i] = segment.NewBrowserSegmentFromTranscript(slide)
		default:
			// Default to slide for backward compatibility
			segments[i] = segment.NewSlideSegmentFromTranscript(slide)
		}
	}

	return segments, nil
}

// GetMetadata returns presentation-level metadata
func (s *TranscriptSource) GetMetadata() Metadata {
	if s.transcript == nil {
		return Metadata{}
	}

	m := s.transcript.Metadata

	// Collect all available languages from slides
	languageSet := make(map[string]bool)
	languageSet[m.DefaultLanguage] = true
	for _, slide := range s.transcript.Slides {
		for lang := range slide.Transcripts {
			languageSet[lang] = true
		}
	}

	languages := make([]string, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	return Metadata{
		Title:           m.Title,
		Description:     m.Description,
		DefaultLanguage: m.DefaultLanguage,
		DefaultVoice:    m.DefaultVoice,
		Languages:       languages,
		Tags:            m.Tags,
		Custom:          m.Custom,
	}
}

// GetTranscript returns the underlying transcript
func (s *TranscriptSource) GetTranscript() *transcript.Transcript {
	return s.transcript
}

// GetFilePath returns the source file path
func (s *TranscriptSource) GetFilePath() string {
	return s.filePath
}
