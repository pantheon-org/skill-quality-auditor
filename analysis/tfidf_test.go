package analysis

import (
	"math"
	"testing"
)

func TestTermFrequency_EmptyString(t *testing.T) {
	tf := TermFrequency("")
	if len(tf) != 0 {
		t.Errorf("expected empty map, got %v", tf)
	}
}

func TestTermFrequency_MarkdownStripping(t *testing.T) {
	text := "# Heading\n**bold** _italic_ `code` [link](http://example.com)\n- list item\n1. numbered"
	tf := TermFrequency(text)
	for term := range tf {
		if term == "#" || term == "**bold**" || term == "_italic_" || term == "`code`" {
			t.Errorf("markdown formatting not stripped, found term: %q", term)
		}
	}
	if _, ok := tf["heading"]; !ok {
		t.Error("expected 'heading' after stripping markdown")
	}
}

func TestTermFrequency_StopwordsRemoved(t *testing.T) {
	text := "the quick brown fox and the lazy dog"
	tf := TermFrequency(text)
	stopwords := []string{"the", "and"}
	for _, sw := range stopwords {
		if _, ok := tf[sw]; ok {
			t.Errorf("stopword %q should have been removed", sw)
		}
	}
}

func TestTermFrequency_ShortTokensRemoved(t *testing.T) {
	text := "go is a language"
	tf := TermFrequency(text)
	for term := range tf {
		if len(term) <= 2 {
			t.Errorf("token with len<=2 should be removed, found: %q", term)
		}
	}
}

func TestTermFrequency_CountAccuracy(t *testing.T) {
	text := "keyword keyword keyword other"
	tf := TermFrequency(text)
	if tf["keyword"] != 3 {
		t.Errorf("expected keyword count=3, got %d", tf["keyword"])
	}
	if tf["other"] != 1 {
		t.Errorf("expected other count=1, got %d", tf["other"])
	}
}

func TestTermFrequency_Lowercase(t *testing.T) {
	text := "Hello HELLO hello"
	tf := TermFrequency(text)
	if tf["hello"] != 3 {
		t.Errorf("expected hello count=3 (case-insensitive), got %d", tf["hello"])
	}
}

func TestTermFrequency_PunctuationStripped(t *testing.T) {
	text := "word, word. word! word?"
	tf := TermFrequency(text)
	if tf["word"] != 4 {
		t.Errorf("expected word count=4 after punctuation strip, got %d", tf["word"])
	}
}

func TestTermFrequency_HRLinesStripped(t *testing.T) {
	text := "before\n---\nafter"
	tf := TermFrequency(text)
	if _, ok := tf["---"]; ok {
		t.Error("HR line '---' should be stripped")
	}
	if _, ok := tf["before"]; !ok {
		t.Error("expected 'before' to be present")
	}
	if _, ok := tf["after"]; !ok {
		t.Error("expected 'after' to be present")
	}
}

func TestExtractKeywords_EmptyCorpus_IDFIsOne(t *testing.T) {
	content := "machine learning model training"
	keywords := ExtractKeywords(content, nil, 10)
	for _, kw := range keywords {
		if math.Abs(kw.IDF-1.0) > 1e-9 {
			t.Errorf("expected IDF=1.0 for empty corpus, got %f for term %q", kw.IDF, kw.Term)
		}
		if math.Abs(kw.Score-kw.TF*1.0) > 1e-9 {
			t.Errorf("expected Score=TF*1.0, got Score=%f TF=%f for term %q", kw.Score, kw.TF, kw.Term)
		}
	}
}

func TestExtractKeywords_EmptyCorpusSlice_IDFIsOne(t *testing.T) {
	content := "machine learning model training"
	keywords := ExtractKeywords(content, []map[string]bool{}, 10)
	for _, kw := range keywords {
		if math.Abs(kw.IDF-1.0) > 1e-9 {
			t.Errorf("expected IDF=1.0 for empty corpus slice, got %f", kw.IDF)
		}
	}
}

func TestExtractKeywords_LimitRespected(t *testing.T) {
	content := "alpha beta gamma delta epsilon zeta eta theta iota kappa"
	keywords := ExtractKeywords(content, nil, 3)
	if len(keywords) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(keywords))
	}
}

func TestExtractKeywords_SortedByScoreDesc(t *testing.T) {
	content := "keyword keyword keyword rare"
	keywords := ExtractKeywords(content, nil, 10)
	for i := 1; i < len(keywords); i++ {
		if keywords[i].Score > keywords[i-1].Score {
			t.Errorf("results not sorted desc: keywords[%d].Score=%f > keywords[%d].Score=%f",
				i, keywords[i].Score, i-1, keywords[i-1].Score)
		}
	}
}

func TestExtractKeywords_NonEmptyCorpusIDF(t *testing.T) {
	corpus := []map[string]bool{
		{"machine": true, "learning": true},
		{"machine": true, "vision": true},
		{"natural": true, "language": true},
	}
	content := "machine learning natural"
	keywords := ExtractKeywords(content, corpus, 10)

	kwMap := make(map[string]KeywordScore)
	for _, kw := range keywords {
		kwMap[kw.Term] = kw
	}

	N := float64(len(corpus))
	// "machine" appears in 2 docs → IDF = log(3/3) = log(1) = 0
	expectedMachineIDF := math.Log(N / float64(1+2))
	if kw, ok := kwMap["machine"]; ok {
		if math.Abs(kw.IDF-expectedMachineIDF) > 1e-9 {
			t.Errorf("machine IDF: expected %f, got %f", expectedMachineIDF, kw.IDF)
		}
	}

	// "natural" appears in 1 doc → IDF = log(3/2)
	expectedNaturalIDF := math.Log(N / float64(1+1))
	if kw, ok := kwMap["natural"]; ok {
		if math.Abs(kw.IDF-expectedNaturalIDF) > 1e-9 {
			t.Errorf("natural IDF: expected %f, got %f", expectedNaturalIDF, kw.IDF)
		}
	}
}

func TestExtractKeywords_ScoreIsTFTimesIDF(t *testing.T) {
	content := "unique unique unique common"
	corpus := []map[string]bool{
		{"unique": true},
		{"common": true},
		{"common": true},
	}
	keywords := ExtractKeywords(content, corpus, 10)
	for _, kw := range keywords {
		expected := kw.TF * kw.IDF
		if math.Abs(kw.Score-expected) > 1e-9 {
			t.Errorf("Score should be TF*IDF: term=%q Score=%f TF=%f IDF=%f", kw.Term, kw.Score, kw.TF, kw.IDF)
		}
	}
}

func TestExtractKeywords_EmptyContent(t *testing.T) {
	keywords := ExtractKeywords("", nil, 10)
	if len(keywords) != 0 {
		t.Errorf("expected empty result for empty content, got %d results", len(keywords))
	}
}

func TestExtractKeywords_LimitHigherThanTerms(t *testing.T) {
	content := "alpha beta"
	keywords := ExtractKeywords(content, nil, 100)
	if len(keywords) > 2 {
		t.Errorf("expected at most 2 results, got %d", len(keywords))
	}
}

func TestExtractKeywords_TFCalculation(t *testing.T) {
	content := "alpha alpha beta"
	keywords := ExtractKeywords(content, nil, 10)
	kwMap := make(map[string]KeywordScore)
	for _, kw := range keywords {
		kwMap[kw.Term] = kw
	}
	// total terms = 3; alpha appears 2 times → TF = 2/3
	if kw, ok := kwMap["alpha"]; ok {
		expected := 2.0 / 3.0
		if math.Abs(kw.TF-expected) > 1e-9 {
			t.Errorf("alpha TF: expected %f, got %f", expected, kw.TF)
		}
	} else {
		t.Error("expected 'alpha' in results")
	}
}

func TestExtractKeywords_IDFNeverNegative(t *testing.T) {
	// Term appears in every corpus document → log(N/(1+df)) would be negative
	// without the math.Max clamp. Verify IDF is clamped to 0.
	corpus := []map[string]bool{
		{"everywhere": true, "alpha": true},
		{"everywhere": true, "beta": true},
		{"everywhere": true, "gamma": true},
	}
	keywords := ExtractKeywords("everywhere", corpus, 10)
	for _, kw := range keywords {
		if kw.IDF < 0 {
			t.Errorf("IDF should never be negative, got %f for term %q", kw.IDF, kw.Term)
		}
		if kw.Score < 0 {
			t.Errorf("Score should never be negative, got %f for term %q", kw.Score, kw.Term)
		}
	}
}

func TestExtractKeywords_ZeroLimit(t *testing.T) {
	content := "alpha beta gamma"
	keywords := ExtractKeywords(content, nil, 0)
	if len(keywords) != 0 {
		t.Errorf("expected empty result for limit=0, got %d results", len(keywords))
	}
}
