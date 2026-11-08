package analysis

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestStandardAnalyzer(t *testing.T) {
	analyzer := NewStandardAnalyzer()

	if analyzer.Name() != "standard" {
		t.Errorf("expected name 'standard', got %q", analyzer.Name())
	}

	input := "The Distributed Systems are Engineering Tomorrow's Databases!"
	tokens, err := analyzer.Analyze(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. Lowercased: "the distributed systems are engineering tomorrow's databases!"
	// 2. Tokenized: "the"(0), "distributed"(1), "systems"(2), "are"(3), "engineering"(4), "tomorrow"(5), "s"(6), "databases"(7)
	// 3. Stopwords: "the"(0), "are"(3), "s"(6) dropped.
	// 4. Stemmed:
	//    - "distributed" -> "distribut" (pos 1)
	//    - "systems" -> "system" (pos 2)
	//    - "engineering" -> "engin" (pos 4)
	//    - "tomorrow" -> "tomorrow" (pos 5)
	//    - "databases" -> "databas" (pos 7)
	expected := []Token{
		{Term: "distribut", Position: 1, StartOffset: 4, EndOffset: 15},
		{Term: "system", Position: 2, StartOffset: 16, EndOffset: 23},
		{Term: "engin", Position: 4, StartOffset: 28, EndOffset: 39},
		{Term: "tomorrow", Position: 5, StartOffset: 40, EndOffset: 48},
		{Term: "databas", Position: 7, StartOffset: 51, EndOffset: 60},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, tokens)
	}
}

func TestSimpleAnalyzer(t *testing.T) {
	analyzer := NewSimpleAnalyzer()

	input := "The Quick Brown Fox!"
	tokens, err := analyzer.Analyze(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Simple analyzer lowercases and tokenizes without stopword removal or stemming
	expected := []Token{
		{Term: "the", Position: 0, StartOffset: 0, EndOffset: 3},
		{Term: "quick", Position: 1, StartOffset: 4, EndOffset: 9},
		{Term: "brown", Position: 2, StartOffset: 10, EndOffset: 15},
		{Term: "fox", Position: 3, StartOffset: 16, EndOffset: 19},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, tokens)
	}
}

func TestWhitespaceAnalyzer(t *testing.T) {
	analyzer := NewWhitespaceAnalyzer()

	input := "SearchForge:v1.0 is=running status=\"OK\""
	tokens, err := analyzer.Analyze(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []Token{
		{Term: "searchforge:v1.0", Position: 0, StartOffset: 0, EndOffset: 16},
		{Term: "is=running", Position: 1, StartOffset: 17, EndOffset: 27},
		{Term: "status=\"ok\"", Position: 2, StartOffset: 28, EndOffset: 39},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, tokens)
	}
}

func TestKeywordAnalyzer(t *testing.T) {
	analyzer := NewKeywordAnalyzer()

	input := "  Exact_Match / Partition / 99  "
	tokens, err := analyzer.Analyze(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}

	if tokens[0].Term != input {
		t.Errorf("expected %q, got %q", input, tokens[0].Term)
	}
}

func TestCustomAnalyzerValidation(t *testing.T) {
	_, err := NewCustomAnalyzer("custom", nil, nil, nil)
	if !errors.Is(err, ErrNilTokenizer) {
		t.Errorf("expected ErrNilTokenizer, got %v", err)
	}

	_, err = NewCustomAnalyzer("", NewStandardTokenizer(), nil, nil)
	if err == nil {
		t.Error("expected error for empty analyzer name, got nil")
	}
}

func TestRegistry(t *testing.T) {
	reg := NewDefaultRegistry()

	// Verify pre-loaded analyzers
	for _, name := range []string{"standard", "simple", "whitespace", "keyword"} {
		a, found := reg.Get(name)
		if !found {
			t.Fatalf("expected analyzer %q to be in default registry", name)
		}
		if a.Name() != name {
			t.Errorf("expected name %q, got %q", name, a.Name())
		}
	}

	// Non-existent analyzer
	_, found := reg.Get("non_existent")
	if found {
		t.Fatal("expected non_existent analyzer to not be found")
	}

	// Register duplicate
	err := reg.Register(NewStandardAnalyzer())
	if !errors.Is(err, ErrDuplicateAnalyzer) {
		t.Fatalf("expected ErrDuplicateAnalyzer, got %v", err)
	}

	// Register nil
	err = reg.Register(nil)
	if err == nil {
		t.Fatal("expected error registering nil analyzer, got nil")
	}

	// Register new custom analyzer
	custom, err := NewCustomAnalyzer("custom_test", NewKeywordTokenizer(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := reg.Register(custom); err != nil {
		t.Fatalf("failed to register custom analyzer: %v", err)
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	reg := NewDefaultRegistry()
	var wg sync.WaitGroup

	// Concurrently read existing analyzers and register new ones
	for i := 0; i < 50; i++ {
		wg.Add(2)
		idx := i

		go func() {
			defer wg.Done()
			_, _ = reg.Get("standard")
		}()

		go func() {
			defer wg.Done()
			custom, _ := NewCustomAnalyzer(
				string(rune('a'+idx)),
				NewKeywordTokenizer(),
				nil,
				nil,
			)
			_ = reg.Register(custom)
		}()
	}

	wg.Wait()
}

func BenchmarkStandardAnalyzer(b *testing.B) {
	analyzer := NewStandardAnalyzer()
	text := "SearchForge is an Elasticsearch-inspired distributed search and indexing engine written in Go from scratch."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := analyzer.Analyze(text)
		if err != nil {
			b.Fatal(err)
		}
	}
}

