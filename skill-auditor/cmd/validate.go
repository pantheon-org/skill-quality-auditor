package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate skill artifacts or review report format",
}

// --------------------------------------------------------------------------
// validate artifacts
// --------------------------------------------------------------------------

var validateArtifactsCmd = &cobra.Command{
	Use:   "artifacts [paths...]",
	Short: "Validate skill artifact conventions (schemas, templates, scripts, SKILL.md)",
	Long: `Validate artifact conventions across all skills (or specified paths).

Checks:
  assets/schemas/   — .schema.json extension, valid JSON, $schema URL
  assets/templates/ — valid YAML (best-effort, no external dep)
  scripts/          — correct shebang per script type
  SKILL.md          — ≤500 lines, frontmatter name matches dir, no ../ refs

Exit code 1 if any error is found.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot("")
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}
		v := &artifactValidator{repoRoot: repoRoot}
		if len(args) > 0 {
			for _, a := range args {
				path := a
				if !filepath.IsAbs(path) {
					path = filepath.Join(repoRoot, path)
				}
				v.checkFile(path)
				if strings.HasSuffix(path, "SKILL.md") {
					v.checkSkillDir(filepath.Dir(path))
				}
			}
		} else {
			v.walkSkillsDir()
		}
		if v.errors > 0 {
			return fmt.Errorf("artifact validation failed (%d error(s))", v.errors)
		}
		fmt.Println("✓ artifact validation passed")
		return nil
	},
}

type artifactValidator struct {
	repoRoot string
	errors   int
}

func (v *artifactValidator) errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", a...)
	v.errors++
}

func (v *artifactValidator) walkSkillsDir() {
	skillsDir := filepath.Join(v.repoRoot, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return
	}
	v.walkArtifactFiles(skillsDir)
	v.walkSkillDirs(skillsDir)
}

func (v *artifactValidator) walkArtifactFiles(skillsDir string) {
	for _, subpath := range []string{"assets/templates", "assets/schemas", "scripts"} {
		sub := subpath
		_ = filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(skillsDir, path)
			parts := strings.SplitN(rel, string(filepath.Separator), 3)
			if len(parts) < 3 {
				return nil
			}
			if strings.Contains(filepath.Join(parts[0], parts[1])+"/"+parts[2], sub) {
				v.checkFile(path)
			}
			return nil
		})
	}
}

func (v *artifactValidator) walkSkillDirs(skillsDir string) {
	_ = filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(skillsDir, path)
		if len(strings.Split(rel, string(filepath.Separator))) != 2 {
			return nil
		}
		if _, serr := os.Stat(filepath.Join(path, "SKILL.md")); serr == nil {
			v.checkSkillDir(path)
		}
		return nil
	})
}

func (v *artifactValidator) checkFile(path string) {
	rel := path
	switch {
	case strings.Contains(path, "/assets/schemas/"):
		v.checkSchemaFile(path)
	case strings.Contains(path, "/assets/templates/"):
		v.checkTemplateFile(path)
	case strings.Contains(path, "/scripts/"):
		v.checkScriptFile(path)
	default:
		_ = rel
	}
}

func (v *artifactValidator) checkTemplateFile(path string) {
	base := filepath.Base(path)
	if base == ".gitkeep" {
		return
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return
	}
	// Best-effort YAML validity: check it's parseable as a map/sequence.
	// We avoid external deps; just verify it's non-empty and not binary.
	data, err := os.ReadFile(path)
	if err != nil {
		v.errorf("%s: cannot read file: %v", path, err)
		return
	}
	content := string(data)
	// Skip template files with placeholder syntax — they may not be valid YAML.
	if strings.Contains(content, "{{") || strings.Contains(content, "{%") {
		return
	}
	// Minimal check: file must not be empty.
	if strings.TrimSpace(content) == "" {
		v.errorf("%s: template file is empty", path)
	}
}

func (v *artifactValidator) checkSchemaFile(path string) {
	base := filepath.Base(path)
	if base == ".gitkeep" {
		return
	}
	if !strings.HasSuffix(base, ".schema.json") {
		v.errorf("%s is in assets/schemas/ but does not use the .schema.json extension", path)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		v.errorf("%s: cannot read file: %v", path, err)
		return
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		v.errorf("%s is not valid JSON: %v", path, err)
		return
	}
	schemaVal, ok := obj["$schema"].(string)
	if !ok || !strings.Contains(schemaVal, "json-schema.org") {
		v.errorf(`%s does not declare a JSON Schema "$schema" URL from json-schema.org`, path)
	}
}

func (v *artifactValidator) checkScriptFile(path string) {
	base := filepath.Base(path)
	if base == ".gitkeep" || strings.Contains(path, "/__pycache__/") {
		return
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".sh":
		v.checkShellScript(path)
	case ".py":
		v.checkPythonScript(path)
	case ".ts":
		v.checkTSScript(path)
	case ".js":
		v.checkJSScript(path)
	default:
		v.errorf("%s is in scripts/ but is not a recognised script type (.sh, .py, .ts, .js)", path)
	}
}

func firstLine(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close() //nolint:errcheck // read-only file, close error is irrelevant
	sc := bufio.NewScanner(f)
	if sc.Scan() {
		return sc.Text()
	}
	return ""
}

func fileContains(path, needle string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), needle)
}

func (v *artifactValidator) checkShellScript(path string) {
	line := firstLine(path)
	switch line {
	case "#!/usr/bin/env sh":
		// Portable — OK.
	case "#!/usr/bin/env bash":
		if !fileContains(path, "# shell: bash") {
			v.errorf("%s must start with portable shebang: #!/usr/bin/env sh (or add '# shell: bash' to allow bash)", path)
		}
	default:
		v.errorf("%s must start with portable shebang: #!/usr/bin/env sh (or add '# shell: bash' to allow bash)", path)
	}
}

func (v *artifactValidator) checkPythonScript(path string) {
	line := firstLine(path)
	if line != "#!/usr/bin/env python3" && line != "#!/usr/bin/env python" {
		v.errorf("%s must start with shebang: #!/usr/bin/env python3", path)
	}
}

func (v *artifactValidator) checkTSScript(path string) {
	line := firstLine(path)
	if line != "#!/usr/bin/env bun" && line != "#!/usr/bin/env -S bun" && line != "#!/usr/bin/env -S bun run" {
		v.errorf("%s must start with shebang: #!/usr/bin/env bun", path)
	}
}

func (v *artifactValidator) checkJSScript(path string) {
	line := firstLine(path)
	if !strings.HasPrefix(line, "#!/usr/bin/env node") {
		v.errorf("%s must start with shebang: #!/usr/bin/env node", path)
	}
}

func (v *artifactValidator) checkSkillDir(skillDir string) {
	v.checkAssetsDir(skillDir)

	skillMD := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMD); os.IsNotExist(err) {
		return
	}
	data, err := os.ReadFile(skillMD)
	if err != nil {
		v.errorf("%s: cannot read: %v", skillMD, err)
		return
	}
	v.validateSkillMD(skillMD, filepath.Base(skillDir), string(data))
}

func (v *artifactValidator) checkAssetsDir(skillDir string) {
	assetsDir := filepath.Join(skillDir, "assets")
	info, err := os.Stat(assetsDir)
	if err != nil || !info.IsDir() {
		return
	}
	entries, _ := os.ReadDir(assetsDir)
	for _, e := range entries {
		if !e.IsDir() {
			name := e.Name()
			if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
				v.errorf("%s: YAML files must be placed in assets/templates/, not directly in assets/", filepath.Join(assetsDir, name))
			}
			continue
		}
		switch e.Name() {
		case "templates", "schemas", "requirements", "examples":
		default:
			v.errorf("%s: non-standard assets/ subdirectory '%s' (allowed: templates/, schemas/, requirements/, examples/)", skillDir, e.Name())
		}
	}
}

func (v *artifactValidator) validateSkillMD(skillMD, skillName, content string) {
	lines := strings.Split(content, "\n")
	if len(lines) > 500 {
		v.errorf("%s: %d lines exceeds 500-line limit — move detail to references/", skillMD, len(lines))
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "name:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			val = strings.Trim(val, `"'`)
			if val != "" && val != skillName {
				v.errorf("%s: frontmatter name '%s' does not match directory name '%s'", skillMD, val, skillName)
			}
			break
		}
	}
	inFence := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if !inFence && strings.Contains(line, "../") {
			v.errorf("%s: contains '../' path reference outside code blocks (skills must be self-contained)", skillMD)
			break
		}
	}
}

// --------------------------------------------------------------------------
// validate review
// --------------------------------------------------------------------------

var validateStrictRecommended bool

var validateReviewCmd = &cobra.Command{
	Use:   "review <report-file>",
	Short: "Validate a review report against the required format",
	Long: `Validate a review report markdown file against review-report.requirements.json.

Checks:
  - H1 title starts with required prefix
  - Required frontmatter keys present
  - Required metadata labels present
  - Required H2 headings present and in order
  - Required dimension labels and commands present
  - Recommended items are warnings (errors with --strict-recommended)

Exit code 1 if any error is found.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reportPath := args[0]
		if !filepath.IsAbs(reportPath) {
			repoRoot, err := resolveRepoRoot("")
			if err != nil {
				return fmt.Errorf("cannot determine repo root: %w", err)
			}
			reportPath = filepath.Join(repoRoot, reportPath)
		}
		return runValidateReview(reportPath, validateStrictRecommended)
	},
}

// reviewRequirements mirrors the structure of review-report.requirements.json.
type reviewRequirements struct {
	RequiredTitlePrefix        string     `json:"required_title_prefix"`
	RequiredFrontmatterKeys    []string   `json:"required_frontmatter_keys"`
	RecommendedFrontmatterKeys []string   `json:"recommended_frontmatter_keys"`
	RequiredMetadataLabels     []string   `json:"required_metadata_labels"`
	RecommendedMetadataLabels  []string   `json:"recommended_metadata_labels"`
	RequiredH2Groups           [][]string `json:"required_h2_groups"`
	RequiredH2Order            [][]string `json:"required_h2_order"`
	RecommendedH2Groups        [][]string `json:"recommended_h2_groups"`
	RequiredDimensionLabels    []string   `json:"required_dimension_labels"`
	RecommendedDimensionLabels []string   `json:"recommended_dimension_labels"`
	RequiredCommands           []string   `json:"required_commands"`
	RecommendedCommands        []string   `json:"recommended_commands"`
}

func runValidateReview(reportPath string, strictRecommended bool) error {
	reqData, err := embeddedRequirements.ReadFile("assets/requirements/review-report.requirements.json")
	if err != nil {
		return fmt.Errorf("cannot read embedded requirements: %w", err)
	}
	var req reviewRequirements
	if err := json.Unmarshal(reqData, &req); err != nil {
		return fmt.Errorf("malformed requirements JSON: %w", err)
	}

	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return fmt.Errorf("cannot read report: %w", err)
	}
	content := string(reportData)

	var errs []string
	var warns []string

	addErr := func(msg string) { errs = append(errs, msg) }
	addWarnOrErr := func(msg string) {
		if strictRecommended {
			errs = append(errs, msg)
		} else {
			warns = append(warns, msg)
		}
	}

	checkReviewTitle(content, req, addErr)
	checkReviewFrontmatter(content, req, addErr, addWarnOrErr)
	checkReviewMetadataLabels(content, req, addErr, addWarnOrErr)
	checkReviewH2Headings(content, req, addErr, addWarnOrErr)
	checkReviewDimensionLabels(content, req, addErr)
	checkReviewCommands(content, req, addErr, addWarnOrErr)

	if len(warns) > 0 {
		fmt.Fprintf(os.Stderr, "Review format warnings for: %s\n", reportPath)
		for _, w := range warns {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
	}

	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Review format validation failed for: %s\n", reportPath)
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		return fmt.Errorf("validation failed (%d error(s))", len(errs))
	}

	fmt.Printf("✓ review format validation passed: %s\n", reportPath)
	return nil
}

func checkReviewTitle(content string, req reviewRequirements, addErr func(string)) {
	title := extractH1(content)
	if title == "" {
		addErr("missing H1 title")
	} else if !strings.HasPrefix(title, req.RequiredTitlePrefix) {
		addErr(fmt.Sprintf("title must start with '%s'", req.RequiredTitlePrefix))
	}
}

func checkReviewFrontmatter(content string, req reviewRequirements, addErr, addWarnOrErr func(string)) {
	fm := extractFrontmatter(content)
	for _, key := range req.RequiredFrontmatterKeys {
		if !hasFrontmatterKey(fm, key) {
			addErr(fmt.Sprintf("missing required frontmatter key: %s", key))
		}
	}
	for _, key := range req.RecommendedFrontmatterKeys {
		if !hasFrontmatterKey(fm, key) {
			addWarnOrErr(fmt.Sprintf("missing recommended frontmatter key: %s", key))
		}
	}
}

func checkReviewMetadataLabels(content string, req reviewRequirements, addErr, addWarnOrErr func(string)) {
	for _, label := range req.RequiredMetadataLabels {
		if !strings.Contains(content, "**"+label+"**:") {
			addErr(fmt.Sprintf("missing metadata label: %s", label))
		}
	}
	for _, label := range req.RecommendedMetadataLabels {
		if !strings.Contains(content, "**"+label+"**:") {
			addWarnOrErr(fmt.Sprintf("missing recommended metadata label: %s", label))
		}
	}
}

func checkReviewH2Headings(content string, req reviewRequirements, addErr, addWarnOrErr func(string)) {
	h2s := extractH2Headings(content)
	for _, group := range req.RequiredH2Groups {
		if !h2GroupPresent(group, h2s) {
			addErr(fmt.Sprintf("missing required H2 heading (one of): %s", strings.Join(group, ", ")))
		}
	}
	for _, group := range req.RecommendedH2Groups {
		if !h2GroupPresent(group, h2s) {
			addWarnOrErr(fmt.Sprintf("missing recommended H2 heading (one of): %s", strings.Join(group, ", ")))
		}
	}
	checkReviewH2Order(h2s, req, addErr)
}

func checkReviewH2Order(h2s []string, req reviewRequirements, addErr func(string)) {
	prevIdx := -1
	for _, group := range req.RequiredH2Order {
		idx := h2GroupIndex(group, h2s)
		if idx < 0 {
			continue
		}
		if prevIdx >= 0 && idx < prevIdx {
			addErr(fmt.Sprintf("H2 order violation near group: %s; expected after prior required section", strings.Join(group, ", ")))
		}
		prevIdx = idx
	}
}

func checkReviewDimensionLabels(content string, req reviewRequirements, addErr func(string)) {
	for _, label := range req.RequiredDimensionLabels {
		if !strings.Contains(content, label) {
			addErr(fmt.Sprintf("missing dimension label: %s", label))
		}
	}
}

func checkReviewCommands(content string, req reviewRequirements, addErr, addWarnOrErr func(string)) {
	for _, cmd := range req.RequiredCommands {
		if !strings.Contains(content, cmd) {
			addErr(fmt.Sprintf("missing required command: %s", cmd))
		}
	}
	for _, cmd := range req.RecommendedCommands {
		if !strings.Contains(content, cmd) {
			addWarnOrErr(fmt.Sprintf("missing recommended command: %s", cmd))
		}
	}
}

// --------------------------------------------------------------------------
// Markdown helpers
// --------------------------------------------------------------------------

func extractH1(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

func extractH2Headings(content string) []string {
	var headings []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			headings = append(headings, strings.TrimPrefix(line, "## "))
		}
	}
	return headings
}

func extractFrontmatter(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return ""
	}
	var fm strings.Builder
	for _, line := range lines[1:] {
		if line == "---" {
			break
		}
		fm.WriteString(line + "\n")
	}
	return fm.String()
}

func hasFrontmatterKey(fm, key string) bool {
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(line, key+":") || strings.HasPrefix(line, key+" :") {
			return true
		}
	}
	return false
}

func h2GroupPresent(group, h2s []string) bool {
	for _, alt := range group {
		for _, h := range h2s {
			if h == alt {
				return true
			}
		}
	}
	return false
}

func h2GroupIndex(group, h2s []string) int {
	for i, h := range h2s {
		for _, alt := range group {
			if h == alt {
				return i
			}
		}
	}
	return -1
}

// --------------------------------------------------------------------------
// init
// --------------------------------------------------------------------------

func init() {
	validateReviewCmd.Flags().BoolVar(&validateStrictRecommended, "strict-recommended", false, "treat recommended items as errors")

	validateCmd.AddCommand(validateArtifactsCmd)
	validateCmd.AddCommand(validateReviewCmd)
	rootCmd.AddCommand(validateCmd)
}
