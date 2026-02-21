package source

import (
	"fmt"

	"github.com/grokify/videoascode/pkg/config"
	"github.com/grokify/videoascode/pkg/segment"
	"github.com/grokify/videoascode/pkg/transcript"
)

// ConfigSource loads content from a unified config file (YAML/JSON).
// It supports mixed slide and browser segments.
type ConfigSource struct {
	config   *config.VideoConfig
	filePath string
}

// NewConfigSource creates a source from a config
func NewConfigSource(cfg *config.VideoConfig) *ConfigSource {
	return &ConfigSource{
		config: cfg,
	}
}

// NewConfigSourceFromFile loads a config from a YAML/JSON file
func NewConfigSourceFromFile(path string) (*ConfigSource, error) {
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &ConfigSource{
		config:   cfg,
		filePath: path,
	}, nil
}

// Load parses the config and returns segments
func (s *ConfigSource) Load() ([]segment.Segment, error) {
	if s.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	segments := make([]segment.Segment, len(s.config.Segments))
	for i, segCfg := range s.config.Segments {
		switch segCfg.Type {
		case config.SourceTypeSlide:
			segments[i] = s.createSlideSegment(i, segCfg)
		case config.SourceTypeBrowser:
			segments[i] = s.createBrowserSegment(i, segCfg)
		default:
			return nil, fmt.Errorf("segment %d: unknown type %s", i, segCfg.Type)
		}
	}

	return segments, nil
}

// createSlideSegment creates a SlideSegment from a SegmentConfig
func (s *ConfigSource) createSlideSegment(index int, cfg config.SegmentConfig) *segment.SlideSegment {
	// Convert config transcripts to transcript.LanguageContent
	transcripts := make(map[string]transcript.LanguageContent)
	for lang, st := range cfg.Transcripts {
		var lc transcript.LanguageContent
		lc.Voice = st.Voice

		if st.Text != "" {
			// Single text block
			lc.Segments = []transcript.Segment{{Text: st.Text}}
		} else {
			// Multiple segments
			lc.Segments = st.Segments
		}

		transcripts[lang] = lc
	}

	return segment.NewSlideSegment(index, cfg.Name, transcripts)
}

// createBrowserSegment creates a BrowserSegment from a SegmentConfig
func (s *ConfigSource) createBrowserSegment(index int, cfg config.SegmentConfig) *segment.BrowserSegment {
	seg := segment.NewBrowserSegment(index, cfg.Name, cfg.URL, cfg.Steps)

	// Convert config transcripts to transcript.LanguageContent
	transcripts := make(map[string]transcript.LanguageContent)
	for lang, st := range cfg.Transcripts {
		var lc transcript.LanguageContent
		lc.Voice = st.Voice

		if st.Text != "" {
			lc.Segments = []transcript.Segment{{Text: st.Text}}
		} else {
			lc.Segments = st.Segments
		}

		transcripts[lang] = lc
	}

	if len(transcripts) > 0 {
		seg.SetTranscripts(transcripts)
	}

	return seg
}

// GetMetadata returns presentation-level metadata
func (s *ConfigSource) GetMetadata() Metadata {
	if s.config == nil {
		return Metadata{}
	}

	languages := s.config.Languages
	if len(languages) == 0 {
		languages = []string{s.config.DefaultLanguage}
	}

	return Metadata{
		Title:           s.config.Title,
		Description:     s.config.Description,
		DefaultLanguage: s.config.DefaultLanguage,
		DefaultVoice:    s.config.DefaultVoice,
		Languages:       languages,
		Custom:          s.config.Metadata,
	}
}

// GetConfig returns the underlying config
func (s *ConfigSource) GetConfig() *config.VideoConfig {
	return s.config
}

// GetFilePath returns the source file path
func (s *ConfigSource) GetFilePath() string {
	return s.filePath
}
