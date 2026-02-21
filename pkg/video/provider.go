package video

import (
	"context"

	"github.com/grokify/marp2video/pkg/segment"
)

// VideoProvider creates videos from segments.
// Different implementations handle different segment types.
type VideoProvider interface {
	// CreateVideo generates a video for a segment.
	// audioPath is the path to the audio file to combine with the video.
	// outputPath is where the video should be written.
	// Returns the actual duration of the generated video in milliseconds.
	CreateVideo(ctx context.Context, seg segment.Segment, audioPath string, outputPath string) (int, error)

	// SupportsSegmentType returns true if this provider can handle the given segment type.
	SupportsSegmentType(sourceType segment.SourceType) bool
}

// ProviderRegistry manages multiple video providers and routes segments to the appropriate one.
type ProviderRegistry struct {
	providers []VideoProvider
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make([]VideoProvider, 0),
	}
}

// Register adds a provider to the registry.
func (r *ProviderRegistry) Register(p VideoProvider) {
	r.providers = append(r.providers, p)
}

// GetProvider returns a provider that can handle the given segment type.
func (r *ProviderRegistry) GetProvider(sourceType segment.SourceType) VideoProvider {
	for _, p := range r.providers {
		if p.SupportsSegmentType(sourceType) {
			return p
		}
	}
	return nil
}

// CreateVideo routes to the appropriate provider based on segment type.
func (r *ProviderRegistry) CreateVideo(ctx context.Context, seg segment.Segment, audioPath string, outputPath string) (int, error) {
	provider := r.GetProvider(seg.GetSourceType())
	if provider == nil {
		return 0, &UnsupportedSegmentError{SourceType: seg.GetSourceType()}
	}
	return provider.CreateVideo(ctx, seg, audioPath, outputPath)
}

// UnsupportedSegmentError is returned when no provider supports a segment type.
type UnsupportedSegmentError struct {
	SourceType segment.SourceType
}

func (e *UnsupportedSegmentError) Error() string {
	return "no video provider supports segment type: " + string(e.SourceType)
}
