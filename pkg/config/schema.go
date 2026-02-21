package config

import (
	_ "embed"
)

// ConfigSchema is the JSON Schema for VideoConfig.
// This can be used by AI assistants to understand how to write browser demo scripts.
//
//go:embed config.schema.json
var ConfigSchema string

// SchemaVersion is the version of the config schema.
const SchemaVersion = "1.0"
