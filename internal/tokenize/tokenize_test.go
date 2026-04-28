package tokenize

import (
	"testing"
)

func TestNormalize_basic(t *testing.T) {
	tokens := Normalize("Hello World foo")
	if len(tokens) == 0 {
		t.Fatal("expected tokens, got none")
	}
	found := false
	for _, tok := range tokens {
		if tok == "foo" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'foo' in tokens")
	}
}

func TestNormalize_stripsMarkdown(t *testing.T) {
	tokens := Normalize("## Header\n**bold** `code` [link](url)")
	for _, tok := range tokens {
		if tok == "##" || tok == "**" || tok == "`" {
			t.Errorf("markdown artifact %q should be stripped", tok)
		}
	}
}

func TestNormalize_stripsStopwords(t *testing.T) {
	tokens := Normalize("the quick brown fox")
	for _, tok := range tokens {
		if tok == "the" {
			t.Error("stopword 'the' should be removed")
		}
	}
}

func TestNormalize_stripsPunctuation(t *testing.T) {
	tokens := Normalize("hello, world!")
	for _, tok := range tokens {
		if tok == "hello," || tok == "world!" {
			t.Errorf("punctuation should be stripped from %q", tok)
		}
	}
}

func TestNormalize_empty(t *testing.T) {
	if tokens := Normalize(""); len(tokens) != 0 {
		t.Errorf("expected empty slice for empty input, got %v", tokens)
	}
}

func TestNormalize_shortTokensFiltered(t *testing.T) {
	tokens := Normalize("a ab abc abcd")
	for _, tok := range tokens {
		if len(tok) <= 2 {
			t.Errorf("token %q with len<=2 should be filtered", tok)
		}
	}
}

func TestSet_uniqueness(t *testing.T) {
	s := Set("foo foo bar baz baz")
	if len(s) != 3 {
		t.Errorf("expected 3 unique tokens, got %d", len(s))
	}
}

func TestCounts_frequency(t *testing.T) {
	c := Counts("foo foo bar")
	if c["foo"] != 2 {
		t.Errorf("expected foo count=2, got %d", c["foo"])
	}
	if c["bar"] != 1 {
		t.Errorf("expected bar count=1, got %d", c["bar"])
	}
}
