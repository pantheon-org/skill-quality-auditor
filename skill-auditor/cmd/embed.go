package cmd

import (
	"embed"
	_ "embed"
)

//go:embed assets/tile.json
var embeddedTile []byte

//go:embed assets/SKILL.md
var embeddedSkill []byte

//go:embed assets/references
var embeddedRefs embed.FS

//go:embed assets/schemas
var embeddedSchemas embed.FS //nolint:unused // reserved for schema validation

//go:embed assets/templates
var embeddedTemplates embed.FS //nolint:unused // reserved for plan generation

//go:embed assets/requirements
var embeddedRequirements embed.FS
