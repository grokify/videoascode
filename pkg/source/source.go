// Package source provides interfaces and implementations for loading
// content from various sources (Marp markdown, transcript JSON, config YAML).
package source

import (
	"github.com/grokify/videoascode/pkg/segment"
	"github.com/grokify/videoascode/pkg/transcript"
)

// ContentSource represents a source of content that can be loaded into segments.
type ContentSource interface {
	// Load parses the source and returns segments
	Load() ([]segment.Segment, error)

	// GetMetadata returns presentation-level metadata
	GetMetadata() Metadata
}

// Metadata contains presentation-level settings
type Metadata struct {
	// Title is the presentation title
	Title string

	// Description is an optional description
	Description string

	// DefaultLanguage is the BCP-47 language code
	DefaultLanguage string

	// DefaultVoice is the default TTS voice configuration
	DefaultVoice transcript.VoiceConfig

	// Languages is the list of languages available
	Languages []string

	// Tags for organization
	Tags []string

	// Custom metadata
	Custom map[string]string
}

// RenderableSource extends ContentSource for sources that can render visual content.
type RenderableSource interface {
	ContentSource

	// Render generates visual content (images, HTML) for the segments.
	// outputDir is where rendered files should be written.
	// Returns paths to rendered files indexed by segment ID.
	Render(outputDir string) (map[string]string, error)
}
