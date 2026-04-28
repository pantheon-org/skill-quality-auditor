package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

// ---- trendArrow ----

func TestTrendArrow_positive(t *testing.T) {
	if got := trendArrow(5); got != "↑" {
		t.Errorf("expected ↑, got %q", got)
	}
}

func TestTrendArrow_negative(t *testing.T) {
	if got := trendArrow(-3); got != "↓" {
		t.Errorf("expected ↓, got %q", got)
	}
}

func TestTrendArrow_zero(t *testing.T) {
	if got := trendArrow(0); got != "—" {
		t.Errorf("expected —, got %q", got)
	}
}

// ---- buildTrendEntry ----

func TestBuildTrendEntry_insufficientPaths(t *testing.T) {
	_, ok := buildTrendEntry("skill", []string{"/only/one"})
	if ok {
		t.Error("expected ok=false for fewer than 2 paths")
	}
}

func TestBuildTrendEntry_unreadableOld(t *testing.T) {
	_, ok := buildTrendEntry("skill", []string{"/no/such/old.json", "/no/such/new.json"})
	if ok {
		t.Error("expected ok=false when audit files are missing")
	}
}

func TestBuildTrendEntry_validPair(t *testing.T) {
	dir := t.TempDir()
	writeAudit(t, dir, "2026-01-01", 80, "B")
	writeAudit(t, dir, "2026-02-01", 100, "A")

	paths := []string{
		filepath.Join(dir, "2026-01-01", "audit.json"),
		filepath.Join(dir, "2026-02-01", "audit.json"),
	}
	entry, ok := buildTrendEntry("domain/skill", paths)
	if !ok {
		t.Fatal("expected ok=true for valid pair")
	}
	if entry.OldScore != 80 || entry.NewScore != 100 {
		t.Errorf("scores: old=%d new=%d", entry.OldScore, entry.NewScore)
	}
	if entry.Delta != 20 {
		t.Errorf("expected delta=20, got %d", entry.Delta)
	}
	if entry.Trend != "↑" {
		t.Errorf("expected ↑, got %q", entry.Trend)
	}
}

func TestBuildTrendEntry_decline(t *testing.T) {
	dir := t.TempDir()
	writeAudit(t, dir, "2026-01-01", 100, "A")
	writeAudit(t, dir, "2026-02-01", 90, "A-")

	paths := []string{
		filepath.Join(dir, "2026-01-01", "audit.json"),
		filepath.Join(dir, "2026-02-01", "audit.json"),
	}
	entry, ok := buildTrendEntry("domain/skill", paths)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if entry.Trend != "↓" {
		t.Errorf("expected ↓, got %q", entry.Trend)
	}
}

func TestBuildTrendEntry_stable(t *testing.T) {
	dir := t.TempDir()
	writeAudit(t, dir, "2026-01-01", 90, "A-")
	writeAudit(t, dir, "2026-02-01", 90, "A-")

	paths := []string{
		filepath.Join(dir, "2026-01-01", "audit.json"),
		filepath.Join(dir, "2026-02-01", "audit.json"),
	}
	entry, ok := buildTrendEntry("domain/skill", paths)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if entry.Trend != "—" {
		t.Errorf("expected —, got %q", entry.Trend)
	}
}

// ---- groupAuditsBySkill ----

func TestGroupAuditsBySkill_groupsCorrectly(t *testing.T) {
	dir := t.TempDir()
	writeAudit(t, dir, "2026-01-01", 80, "B")
	writeAudit(t, dir, "2026-02-01", 90, "A-")

	// groupAuditsBySkill expects auditsRoot/domain/skill/date/audit.json
	auditsRoot := t.TempDir()
	skillDir := filepath.Join(auditsRoot, "domain", "my-skill")
	copyAudits(t, dir, skillDir)

	groups, err := groupAuditsBySkill(auditsRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups["domain/my-skill"]) != 2 {
		t.Errorf("expected 2 audits for domain/my-skill, got %d", len(groups["domain/my-skill"]))
	}
}

func TestGroupAuditsBySkill_nonexistentDir(t *testing.T) {
	_, err := groupAuditsBySkill("/nonexistent/audits")
	if err == nil {
		t.Error("expected error for nonexistent audits dir")
	}
}

// ---- collectTrends ----

func TestCollectTrends_returnsEntries(t *testing.T) {
	auditsRoot := t.TempDir()
	skillDir := filepath.Join(auditsRoot, "domain", "skill-a")
	writeAudit(t, skillDir, "2026-01-01", 70, "C")
	writeAudit(t, skillDir, "2026-02-01", 80, "B")

	entries, err := collectTrends(auditsRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 trend entry, got %d", len(entries))
	}
}

func TestCollectTrends_skipsSkillWithOneAudit(t *testing.T) {
	auditsRoot := t.TempDir()
	skillDir := filepath.Join(auditsRoot, "domain", "solo")
	writeAudit(t, skillDir, "2026-01-01", 70, "C")

	entries, err := collectTrends(auditsRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for skill with only one audit, got %d", len(entries))
	}
}

// ---- printTrendTable ----

func TestPrintTrendTable_doesNotPanic(t *testing.T) {
	entries := []TrendEntry{
		{Skill: "domain/skill", OldDate: "2026-01-01", NewDate: "2026-02-01", OldScore: 80, NewScore: 100, OldGrade: "B", NewGrade: "A", Delta: 20, Trend: "↑"},
		{Skill: "x", OldDate: "2026-01-01", NewDate: "2026-02-01", OldScore: 90, NewScore: 90, OldGrade: "A-", NewGrade: "A-", Delta: 0, Trend: "—"},
	}
	printTrendTable(entries)
}

// ---- helpers ----

// writeAudit creates <base>/<date>/audit.json with a minimal scorer.Result.
func writeAudit(t *testing.T, base, date string, total int, grade string) {
	t.Helper()
	dir := filepath.Join(base, date)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	r := scorer.Result{Date: date, Total: total, Grade: grade}
	data, _ := json.Marshal(r)
	if err := os.WriteFile(filepath.Join(dir, "audit.json"), data, 0o644); err != nil {
		t.Fatalf("write audit: %v", err)
	}
}

// copyAudits copies date/<audit.json> subdirs from src into dst.
func copyAudits(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		srcFile := filepath.Join(src, e.Name(), "audit.json")
		dstDir := filepath.Join(dst, e.Name())
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		data, err := os.ReadFile(srcFile)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dstDir, "audit.json"), data, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
}
