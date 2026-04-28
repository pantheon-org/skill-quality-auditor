package cmd

import (
	"embed"
	_ "embed"
)

//go:embed assets/SKILL.md
var embeddedSkill []byte

//go:embed assets/references
var embeddedRefs embed.FS

//go:embed assets/schemas
var embeddedSchemas embed.FS //nolint:unused // reserved for validate command

//go:embed assets/templates
var embeddedTemplates embed.FS //nolint:unused // reserved for plan generation command
