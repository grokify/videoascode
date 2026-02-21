package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads a video configuration from a file
// Supports JSON and YAML formats based on file extension
func LoadFromFile(path string) (*VideoConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	var config VideoConfig
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, &config); err != nil {
			if err := yaml.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config (tried JSON and YAML): %w", err)
			}
		}
	}

	// Apply defaults
	if config.Version == "" {
		config.Version = "1.0"
	}
	if config.Resolution.Width == 0 {
		config.Resolution = Resolution1080p
	}
	if config.FrameRate == 0 {
		config.FrameRate = 30
	}
	if config.DefaultLanguage == "" {
		config.DefaultLanguage = "en-US"
	}

	return &config, nil
}

// SaveToFile saves a video configuration to a file
// Format is determined by file extension
func (c *VideoConfig) SaveToFile(path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c)
	default:
		data, err = json.MarshalIndent(c, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *VideoConfig) Validate() error {
	if c.Title == "" {
		return fmt.Errorf("title is required")
	}

	if len(c.Segments) == 0 {
		return fmt.Errorf("at least one segment is required")
	}

	for i, seg := range c.Segments {
		if err := seg.Validate(); err != nil {
			return fmt.Errorf("segment %d: %w", i, err)
		}
	}

	if c.Resolution.Width <= 0 || c.Resolution.Height <= 0 {
		return fmt.Errorf("invalid resolution: %dx%d", c.Resolution.Width, c.Resolution.Height)
	}

	if c.FrameRate <= 0 {
		return fmt.Errorf("frame rate must be positive")
	}

	return nil
}

// Validate checks if a segment configuration is valid
func (s *SegmentConfig) Validate() error {
	switch s.Type {
	case SourceTypeSlide:
		if s.Source == "" {
			return fmt.Errorf("slide segment requires source file")
		}
	case SourceTypeBrowser:
		if s.URL == "" {
			return fmt.Errorf("browser segment requires URL")
		}
		for i, step := range s.Steps {
			if err := step.Validate(); err != nil {
				return fmt.Errorf("step %d: %w", i, err)
			}
		}
	default:
		return fmt.Errorf("unknown segment type: %s", s.Type)
	}
	return nil
}

// LoadStepsFromFile loads browser steps from a separate file
func LoadStepsFromFile(path string) ([]SegmentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read steps file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	var segments []SegmentConfig
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &segments); err != nil {
			return nil, fmt.Errorf("failed to parse JSON steps: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &segments); err != nil {
			return nil, fmt.Errorf("failed to parse YAML steps: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return segments, nil
}

// MergeConfigs merges multiple configurations
// Later configs override earlier ones
func MergeConfigs(configs ...*VideoConfig) *VideoConfig {
	if len(configs) == 0 {
		return &VideoConfig{}
	}

	result := *configs[0]

	for _, cfg := range configs[1:] {
		if cfg.Title != "" {
			result.Title = cfg.Title
		}
		if cfg.Description != "" {
			result.Description = cfg.Description
		}
		if cfg.DefaultLanguage != "" {
			result.DefaultLanguage = cfg.DefaultLanguage
		}
		if cfg.DefaultVoice.VoiceID != "" {
			result.DefaultVoice = cfg.DefaultVoice
		}
		if cfg.Resolution.Width > 0 {
			result.Resolution = cfg.Resolution
		}
		if cfg.FrameRate > 0 {
			result.FrameRate = cfg.FrameRate
		}
		if cfg.OutputDir != "" {
			result.OutputDir = cfg.OutputDir
		}
		if len(cfg.Languages) > 0 {
			result.Languages = cfg.Languages
		}
		if cfg.TransitionDuration > 0 {
			result.TransitionDuration = cfg.TransitionDuration
		}
		if len(cfg.Segments) > 0 {
			result.Segments = append(result.Segments, cfg.Segments...)
		}
		for k, v := range cfg.Metadata {
			if result.Metadata == nil {
				result.Metadata = make(map[string]string)
			}
			result.Metadata[k] = v
		}
	}

	return &result
}
