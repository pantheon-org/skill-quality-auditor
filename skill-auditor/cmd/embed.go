package cmd

import (
	"embed"
	_ "embed"
)

//go:embed assets/SKILL.md
var embeddedSkill []byte

//go:embed assets/references
var embeddedRefs embed.FS
