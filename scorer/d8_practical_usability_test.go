package scorer

import (
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

func TestD8_ManyCodeBlocks(t *testing.T) {
	// baseline 5 + 4 (>5 blocks via library) + 4 (run cmd ./) = 13; capped at 15
	b := &validatorBridge{Content: &types.ContentReport{
		CodeBlockCount: 6,
		CodeLanguages:  []string{"bash", "typescript"},
	}}
	content := "---\ndescription: x\n---\nRun ./script.sh to start."
	score, _ := scoreD8(content, b)
	// 5 + 4 (>5) + 2 (languages) + 4 (./) = 15
	if score != 15 {
		t.Errorf("want 15, got %d", score)
	}
}

func TestD8_FewCodeBlocks_Fallback(t *testing.T) {
	// nilBridge → fallback path; no code blocks → baseline 5; no run cmd → 5
	content := "---\ndescription: x\n---\n# Skill\nNo code blocks here."
	if score, _ := scoreD8(content, nilBridge()); score != 5 {
		t.Errorf("want 5, got %d", score)
	}
}

func TestD8_MediumCodeBlocks_Fallback(t *testing.T) {
	// nilBridge → fallback; 3 blocks → +2; no run cmd → 5+2=7
	content := "---\ndescription: x\n---\n" +
		"```\nfoo\n```\n```\nbar\n```\n```\nbaz\n```\n"
	if score, _ := scoreD8(content, nilBridge()); score != 7 {
		t.Errorf("want 7, got %d", score)
	}
}

func TestD8_LibraryCodeBlockCount(t *testing.T) {
	cases := []struct {
		count int
		want  int
	}{
		{0, 5},
		{1, 6}, // 5 + 1
		{3, 7}, // 5 + 2
		{6, 9}, // 5 + 4
	}
	for _, tc := range cases {
		b := &validatorBridge{Content: &types.ContentReport{CodeBlockCount: tc.count}}
		content := "---\ndescription: x\n---\n# Skill"
		score, _ := scoreD8(content, b)
		if score != tc.want {
			t.Errorf("CodeBlockCount=%d: want %d, got %d", tc.count, tc.want, score)
		}
	}
}

func TestD8_LibraryLanguageTags(t *testing.T) {
	// library path: 1 block + languages → 5+1+2 = 8
	b := &validatorBridge{Content: &types.ContentReport{
		CodeBlockCount: 1,
		CodeLanguages:  []string{"bash"},
	}}
	content := "---\ndescription: x\n---\n```bash\necho hi\n```\n"
	if score, _ := scoreD8(content, b); score != 8 {
		t.Errorf("want 8, got %d", score)
	}
}

func TestD8_ScoreCappedAt15(t *testing.T) {
	// 5 + 4 (>5 blocks) + 2 (languages) + 4 (run cmd) = 15 already hit by TestD8_ManyCodeBlocks.
	// Test with high block count AND many run commands to confirm cap holds.
	b := &validatorBridge{Content: &types.ContentReport{
		CodeBlockCount: 10,
		CodeLanguages:  []string{"bash", "typescript", "go"},
	}}
	content := "---\ndescription: x\n---\nRun ./script.sh and npm run build and go run ./main.go"
	score, _ := scoreD8(content, b)
	if score != 15 {
		t.Errorf("want 15 (capped), got %d", score)
	}
}

func TestD8_OutcomeLinkage_AllLinked(t *testing.T) {
	content := "---\ndescription: x\n---\n" +
		"Run the migration:\n" +
		"```bash\n" +
		"go run ./cmd/migrate up\n" +
		"```\n" +
		"# verify: migration table should show 3 rows\n"
	if score := scoreOutcomeLinkage(content); score != 3 {
		t.Errorf("want 3, got %d", score)
	}
}

func TestD8_OutcomeLinkage_NoneLinked(t *testing.T) {
	content := "---\ndescription: x\n---\n" +
		"```bash\n" +
		"echo hello\n" +
		"```\n" +
		"```python\n" +
		"print(42)\n" +
		"```\n"
	if score := scoreOutcomeLinkage(content); score != 0 {
		t.Errorf("want 0, got %d", score)
	}
}

func TestD8_OutcomeLinkage_FalsePositive_ProsePhrases(t *testing.T) {
	content := "---\ndescription: x\n---\n" +
		"This command creates a new project. The tool produces a scaffold that writes to disk.\n" +
		"Run `init` to get started.\n"
	if score := scoreOutcomeLinkage(content); score != 0 {
		t.Errorf("want 0 (no fenced code blocks), got %d", score)
	}
}

func TestD8_OutcomeLinkage_FalsePositive_DistantProse(t *testing.T) {
	content := "---\ndescription: x\n---\n" +
		"## Background\n" +
		"The deploy pipeline confirms that all checks pass before merging.\n" +
		"You should see green indicators in the dashboard.\n" +
		"\n" +
		"```bash\n" +
		"git push origin main\n" +
		"```\n"
	if score := scoreOutcomeLinkage(content); score != 0 {
		t.Errorf("want 0 (outcome language before block, not after), got %d", score)
	}
}

func TestD8_OutcomeLinkage_Partial(t *testing.T) {
	content := "---\ndescription: x\n---\n" +
		"```bash\n" +
		"go run ./cmd/migrate up\n" +
		"```\n" +
		"# verify: migration table has 3 rows\n" +
		"\n" +
		"```python\n" +
		"print(42)\n" +
		"```\n"
	if score := scoreOutcomeLinkage(content); score != 2 {
		t.Errorf("want 2 (1/2 linked), got %d", score)
	}
}

func TestD8_RunCommands(t *testing.T) {
	cases := []struct {
		name    string
		snippet string
	}{
		{"./", "Run ./build.sh to start."},
		{"npm run", "npm run build"},
		{"yarn", "yarn install"},
		{"pnpm run", "pnpm run test"},
		{"bun run", "bun run dev"},
		{"make", "make install"},
		{"python", "python main.py"},
		{"go run", "go run ./cmd/main.go"},
	}
	for _, tc := range cases {
		content := "---\ndescription: x\n---\n# Skill\n" + tc.snippet
		score, _ := scoreD8(content, nilBridge())
		// 5 + 4 (run cmd) = 9
		if score != 9 {
			t.Errorf("runner=%s: want 9, got %d", tc.name, score)
		}
	}
}
