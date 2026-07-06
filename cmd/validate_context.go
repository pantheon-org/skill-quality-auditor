package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// validate context — real JSON-schema validation of .context/**/*.md frontmatter.
//
// The shell validator (validate-context-frontmatter.sh) re-implements only a
// subset of the schema (required / enum / pattern / per-type rules) and never
// enforces additionalProperties:false, so a typo'd top-level key (valeu: HIGH)
// passes silently. This command compiles the real JSON Schemas and validates
// each file against them, catching unknown/typo'd keys. It runs alongside the
// shell validator (which keeps the per-type requiredness logic a single shared
// schema cannot model). See .context/plans/governance-enum-and-schema-hardening.

const (
	relContextFrontmatterSchema = ".context/plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json"
	relRemediationPlanSchema    = ".context/plugins/pantheon-org/context-mgmt/context-file/assets/schemas/remediation-plan.schema.json"
)

var validateContextCmd = &cobra.Command{
	Use:   "context [paths...]",
	Short: "Validate .context/**/*.md frontmatter against the JSON schemas (catches typo'd/unknown keys)",
	Long: `Validate .context frontmatter with a real JSON-schema validator.

Enforces the schemas' additionalProperties:false — a typo'd or unknown top-level
key (e.g. valeu: HIGH) is rejected, which the shell validator cannot catch.
Generated remediation plans (*-remediation-plan-*.md) validate against the
dedicated remediation-plan schema; all other files against context-frontmatter.

With no paths, walks .context/**/*.md (excluding .context/plugins/**). Exit 1 on
any violation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRootFlag, _ := cmd.Flags().GetString("repo-root")
		repoRoot, err := resolveRepoRoot(repoRootFlag)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}
		v, err := newContextSchemaValidator(repoRoot)
		if err != nil {
			return err
		}

		files := args
		if len(files) == 0 {
			files = v.walk()
		}
		for _, f := range files {
			path := f
			if !filepath.IsAbs(path) {
				path = filepath.Join(repoRoot, path)
			}
			v.checkFile(path)
		}
		if v.errors > 0 {
			return fmt.Errorf("context frontmatter schema validation failed (%d error(s))", v.errors)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ context frontmatter schema validation passed (%d file(s))\n", v.checked)
		return nil
	},
}

type contextSchemaValidator struct {
	repoRoot    string
	frontmatter *jsonschema.Schema
	remediation *jsonschema.Schema
	errors      int
	checked     int
}

func newContextSchemaValidator(repoRoot string) (*contextSchemaValidator, error) {
	fm, err := compileSchema(filepath.Join(repoRoot, relContextFrontmatterSchema))
	if err != nil {
		return nil, fmt.Errorf("compile context-frontmatter schema: %w", err)
	}
	rem, err := compileSchema(filepath.Join(repoRoot, relRemediationPlanSchema))
	if err != nil {
		return nil, fmt.Errorf("compile remediation-plan schema: %w", err)
	}
	return &contextSchemaValidator{repoRoot: repoRoot, frontmatter: fm, remediation: rem}, nil
}

func compileSchema(path string) (*jsonschema.Schema, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	url := "file://" + path
	if err := c.AddResource(url, doc); err != nil {
		return nil, err
	}
	return c.Compile(url)
}

// walk returns all .context/**/*.md paths except those under .context/plugins/.
func (v *contextSchemaValidator) walk() []string {
	var out []string
	contextDir := filepath.Join(v.repoRoot, ".context")
	pluginsDir := filepath.Join(contextDir, "plugins")
	_ = filepath.Walk(contextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasPrefix(path, pluginsDir) {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			out = append(out, path)
		}
		return nil
	})
	return out
}

func (v *contextSchemaValidator) errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", a...)
	v.errors++
}

func (v *contextSchemaValidator) checkFile(path string) {
	inst, ok := v.frontmatterInstance(path)
	if !ok {
		return // no frontmatter — not this validator's concern (shell gate reports it)
	}
	v.checked++
	// Route by content signature, not filename: a generator-produced remediation
	// plan carries a top-level `skill_name` key (and the other structured summary
	// keys). Hand-authored plans that merely have "remediation-plan" in their name
	// (e.g. se-principles-remediation-plan) have ordinary plan frontmatter and must
	// validate against context-frontmatter, not the generator schema.
	schema := v.frontmatter
	if m, ok := inst.(map[string]any); ok {
		if _, hasSkillName := m["skill_name"]; hasSkillName {
			schema = v.remediation
		}
	}
	if err := schema.Validate(inst); err != nil {
		rel, _ := filepath.Rel(v.repoRoot, path)
		v.errorf("%s: %s", rel, summarizeSchemaError(err))
	}
}

// frontmatterInstance extracts the YAML frontmatter block and returns it as a
// JSON-canonical value ready for schema validation. ok is false when the file
// has no frontmatter.
func (v *contextSchemaValidator) frontmatterInstance(path string) (any, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		v.errorf("%s: cannot read: %v", path, err)
		return nil, false
	}
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return nil, false
	}
	rest := text[4:]
	end := strings.Index(rest, "---\n")
	if end < 0 {
		if strings.HasSuffix(rest, "---") {
			end = len(rest) - 3
		} else {
			v.errorf("%s: unclosed frontmatter", path)
			return nil, false
		}
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(rest[:end]), &doc); err != nil {
		v.errorf("%s: invalid YAML frontmatter: %v", path, err)
		return nil, false
	}
	return yamlNodeToJSON(&doc), true
}

// yamlNodeToJSON converts a parsed YAML node tree to JSON-canonical Go values
// (map[string]any / []any / float64 / bool / nil / string) for schema
// validation. Crucially it keeps timestamp-shaped scalars (e.g. dates like
// 2026-04-29) as STRINGS rather than letting yaml coerce them to time.Time —
// otherwise a valid `date: 2026-04-29` would round-trip to an RFC3339 timestamp
// and spuriously fail the YYYY-MM-DD pattern. Only explicitly int/float/bool/null
// tagged scalars are converted to non-string types.
func yamlNodeToJSON(n *yaml.Node) any {
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return nil
		}
		return yamlNodeToJSON(n.Content[0])
	case yaml.MappingNode:
		m := make(map[string]any, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			m[n.Content[i].Value] = yamlNodeToJSON(n.Content[i+1])
		}
		return m
	case yaml.SequenceNode:
		s := make([]any, 0, len(n.Content))
		for _, c := range n.Content {
			s = append(s, yamlNodeToJSON(c))
		}
		return s
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!int":
			if f, err := strconv.ParseFloat(n.Value, 64); err == nil {
				return f
			}
		case "!!float":
			if f, err := strconv.ParseFloat(n.Value, 64); err == nil {
				return f
			}
		case "!!bool":
			if b, err := strconv.ParseBool(n.Value); err == nil {
				return b
			}
		case "!!null":
			return nil
		}
		return n.Value // strings, timestamps, and anything else stay as text
	default:
		return nil
	}
}

// summarizeSchemaError renders a jsonschema validation error as a single line.
// The library's own Error() gives a readable tree (schema URL + failing
// keyword + instance location); we just collapse the whitespace so each file's
// failure fits on one line in the report.
func summarizeSchemaError(err error) string {
	fields := strings.Fields(err.Error())
	return strings.Join(fields, " ")
}

func init() {
	validateContextCmd.Flags().StringP("repo-root", "r", "", "repo root (auto-detected if empty)")
	validateCmd.AddCommand(validateContextCmd)
}
