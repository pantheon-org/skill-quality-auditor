package duplication

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkillEntry holds the key, path, and loaded content of a single skill.
type SkillEntry struct {
	Key     string // domain/skill-name relative to skills dir
	Path    string // absolute path to SKILL.md
	Content string
}

// Inventory walks skillsDir and returns one SkillEntry per SKILL.md found.
// All file content is loaded eagerly; for repos with >500 skills, consider lazy loading.
func Inventory(skillsDir string) ([]SkillEntry, error) {
	var entries []SkillEntry
	err := filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		rel, relErr := filepath.Rel(skillsDir, path)
		if relErr != nil {
			return relErr
		}
		key := strings.TrimSuffix(rel, string(filepath.Separator)+"SKILL.md")
		entries = append(entries, SkillEntry{
			Key:     key,
			Path:    path,
			Content: string(data),
		})
		return nil
	})
	return entries, err
}
