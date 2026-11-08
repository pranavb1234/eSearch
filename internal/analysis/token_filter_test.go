package analysis

import (
	"reflect"
	"testing"
)

func TestStopwordFilter(t *testing.T) {
	filter := NewStopwordFilter(nil)

	if filter.Name() != "stopword" {
		t.Errorf("expected 'stopword', got %q", filter.Name())
	}

	tokens := []Token{
		{Term: "the", Position: 0, StartOffset: 0, EndOffset: 3},
		{Term: "distributed", Position: 1, StartOffset: 4, EndOffset: 15},
		{Term: "and", Position: 2, StartOffset: 16, EndOffset: 19},
		{Term: "fault", Position: 3, StartOffset: 20, EndOffset: 25},
		{Term: "tolerant", Position: 4, StartOffset: 26, EndOffset: 34},
		{Term: "system", Position: 5, StartOffset: 35, EndOffset: 41},
	}

	filtered, err := filter.Filter(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "the" (pos 0) and "and" (pos 2) are dropped; remaining retain original positions
	expected := []Token{
		{Term: "distributed", Position: 1, StartOffset: 4, EndOffset: 15},
		{Term: "fault", Position: 3, StartOffset: 20, EndOffset: 25},
		{Term: "tolerant", Position: 4, StartOffset: 26, EndOffset: 34},
		{Term: "system", Position: 5, StartOffset: 35, EndOffset: 41},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, filtered)
	}
}

func TestPorterStemmer(t *testing.T) {
	stemmer := NewStemmerFilter()

	if stemmer.Name() != "porter_stemmer" {
		t.Errorf("expected 'porter_stemmer', got %q", stemmer.Name())
	}

	testCases := []struct {
		word     string
		expected string
	}{
		{"connections", "connect"},
		{"connecting", "connect"},
		{"connection", "connect"},
		{"distributed", "distribut"},
		{"distributing", "distribut"},
		{"engineering", "engin"},
		{"engineer", "engin"},
		{"running", "run"},
		{"runs", "run"},
		{"cars", "car"},
		{"go", "go"}, // <= 2 runes preserved
	}

	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			got := Stem(tc.word)
			if got != tc.expected {
				t.Errorf("Stem(%q) = %q, expected %q", tc.word, got, tc.expected)
			}
		})
	}
}

func TestLengthFilter(t *testing.T) {
	filter := NewLengthFilter(3, 10)

	tokens := []Token{
		{Term: "a", Position: 0},
		{Term: "in", Position: 1},
		{Term: "valid", Position: 2},
		{Term: "acceptable", Position: 3},
		{Term: "supercalifragilistic", Position: 4}, // > 10 chars
	}

	filtered, err := filter.Filter(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []Token{
		{Term: "valid", Position: 2},
		{Term: "acceptable", Position: 3},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, filtered)
	}
}

func TestTokenFilterPipeline(t *testing.T) {
	stopwords := NewStopwordFilter(nil)
	stemmer := NewStemmerFilter()
	length := NewLengthFilter(3, 15)

	pipeline := NewTokenFilterPipeline(stopwords, stemmer, length)

	tokens := []Token{
		{Term: "the", Position: 0},
		{Term: "distributed", Position: 1},
		{Term: "databases", Position: 2},
		{Term: "of", Position: 3},
		{Term: "tomorrow", Position: 4},
	}

	filtered, err := pipeline.Filter(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "the" and "of" dropped by stopwords.
	// "distributed" -> "distribut"
	// "databases" -> "databas"
	// "tomorrow" -> "tomorrow"
	expected := []Token{
		{Term: "distribut", Position: 1},
		{Term: "databas", Position: 2},
		{Term: "tomorrow", Position: 4},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, filtered)
	}
}

func BenchmarkStopwordFilter(b *testing.B) {
	filter := NewStopwordFilter(nil)
	tokens := []Token{
		{Term: "the", Position: 0},
		{Term: "distributed", Position: 1},
		{Term: "search", Position: 2},
		{Term: "engine", Position: 3},
		{Term: "in", Position: 4},
		{Term: "go", Position: 5},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = filter.Filter(tokens)
	}
}

func BenchmarkStemmerFilter(b *testing.B) {
	filter := NewStemmerFilter()
	tokens := []Token{
		{Term: "distributed", Position: 0},
		{Term: "connections", Position: 1},
		{Term: "engineering", Position: 2},
		{Term: "running", Position: 3},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = filter.Filter(tokens)
	}
}

