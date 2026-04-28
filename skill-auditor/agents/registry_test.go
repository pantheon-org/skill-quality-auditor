package agents

import (
	"strings"
	"testing"
)

func TestByID_found(t *testing.T) {
	a, ok := ByID("claude-code")
	if !ok {
		t.Fatal("expected claude-code to be found")
	}
	if a.ID != "claude-code" {
		t.Errorf("got ID %q, want claude-code", a.ID)
	}
	if a.DisplayName == "" {
		t.Error("DisplayName should not be empty")
	}
}

func TestByID_notFound(t *testing.T) {
	_, ok := ByID("nonexistent-agent")
	if ok {
		t.Error("expected not found for unknown agent")
	}
}

func TestByID_allRegistryIDs(t *testing.T) {
	for _, a := range Registry {
		found, ok := ByID(a.ID)
		if !ok {
			t.Errorf("ByID(%q) returned not found", a.ID)
		}
		if found.ID != a.ID {
			t.Errorf("ByID(%q) returned ID %q", a.ID, found.ID)
		}
	}
}

func TestSkillDir_global(t *testing.T) {
	a, _ := ByID("claude-code")
	dir := a.SkillDir("/home/user", true)
	if !strings.HasPrefix(dir, "/home/user") {
		t.Errorf("global skill dir should be under homeDir, got %q", dir)
	}
	if !strings.Contains(dir, a.GlobalPath) {
		t.Errorf("global skill dir should contain GlobalPath %q, got %q", a.GlobalPath, dir)
	}
}

func TestSkillDir_project(t *testing.T) {
	a, _ := ByID("claude-code")
	dir := a.SkillDir("/home/user", false)
	if dir != a.ProjectPath {
		t.Errorf("project skill dir should equal ProjectPath %q, got %q", a.ProjectPath, dir)
	}
}

func TestHarnessDirs_notEmpty(t *testing.T) {
	dirs := HarnessDirs()
	if len(dirs) == 0 {
		t.Fatal("HarnessDirs should return at least one entry")
	}
}

func TestHarnessDirs_allHaveDotPrefix(t *testing.T) {
	for _, d := range HarnessDirs() {
		if !strings.HasPrefix(d, ".") {
			t.Errorf("harness dir %q should start with '.'", d)
		}
	}
}

func TestHarnessDirs_unique(t *testing.T) {
	dirs := HarnessDirs()
	seen := map[string]bool{}
	for _, d := range dirs {
		if seen[d] {
			t.Errorf("duplicate harness dir %q", d)
		}
		seen[d] = true
	}
}

func TestHarnessDirs_containsKnownDirs(t *testing.T) {
	dirs := HarnessDirs()
	set := map[string]bool{}
	for _, d := range dirs {
		set[d] = true
	}
	for _, want := range []string{".claude", ".cursor", ".windsurf"} {
		if !set[want] {
			t.Errorf("expected %q in HarnessDirs", want)
		}
	}
}

func TestDisplayNames_length(t *testing.T) {
	names := DisplayNames()
	if len(names) != len(Registry) {
		t.Errorf("DisplayNames length %d != Registry length %d", len(names), len(Registry))
	}
}

func TestDisplayNames_allLowercase(t *testing.T) {
	for _, n := range DisplayNames() {
		if n != strings.ToLower(n) {
			t.Errorf("DisplayName %q is not lowercase", n)
		}
	}
}

func TestDisplayNames_notEmpty(t *testing.T) {
	for i, n := range DisplayNames() {
		if n == "" {
			t.Errorf("DisplayNames[%d] is empty", i)
		}
	}
}
