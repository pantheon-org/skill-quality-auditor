package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveOutputFormat_DefaultJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("json", "j", false, "")
	cmd.Flags().BoolP("markdown", "m", false, "")

	got, err := resolveOutputFormat(cmd, OutputFormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != OutputFormatJSON {
		t.Errorf("expected OutputFormatJSON, got %v", got)
	}
}

func TestResolveOutputFormat_DefaultMarkdown(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("json", "j", false, "")
	cmd.Flags().BoolP("markdown", "m", false, "")

	got, err := resolveOutputFormat(cmd, OutputFormatMarkdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != OutputFormatMarkdown {
		t.Errorf("expected OutputFormatMarkdown, got %v", got)
	}
}

func TestResolveOutputFormat_JSONFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("json", "j", false, "")
	cmd.Flags().BoolP("markdown", "m", false, "")
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}

	got, err := resolveOutputFormat(cmd, OutputFormatMarkdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != OutputFormatJSON {
		t.Errorf("expected OutputFormatJSON, got %v", got)
	}
}

func TestResolveOutputFormat_MarkdownFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("json", "j", false, "")
	cmd.Flags().BoolP("markdown", "m", false, "")
	if err := cmd.Flags().Set("markdown", "true"); err != nil {
		t.Fatal(err)
	}

	got, err := resolveOutputFormat(cmd, OutputFormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != OutputFormatMarkdown {
		t.Errorf("expected OutputFormatMarkdown, got %v", got)
	}
}

func TestResolveOutputFormat_MutualExclusion(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("json", "j", false, "")
	cmd.Flags().BoolP("markdown", "m", false, "")
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("markdown", "true"); err != nil {
		t.Fatal(err)
	}

	_, err := resolveOutputFormat(cmd, OutputFormatJSON)
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags, got nil")
	}
	want := "--json and --markdown are mutually exclusive"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}
