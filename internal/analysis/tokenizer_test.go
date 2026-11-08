package analysis

import (
	"reflect"
	"testing"
)

func TestStandardTokenizer(t *testing.T) {
	tok := NewStandardTokenizer()

	if tok.Name() != "standard" {
		t.Errorf("expected 'standard', got %q", tok.Name())
	}

	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "standard sentence with punctuation",
			input: "Hello, World! SearchForge v1.0.",
			expected: []Token{
				{Term: "Hello", Position: 0, StartOffset: 0, EndOffset: 5},
				{Term: "World", Position: 1, StartOffset: 7, EndOffset: 12},
				{Term: "SearchForge", Position: 2, StartOffset: 14, EndOffset: 25},
				{Term: "v1", Position: 3, StartOffset: 26, EndOffset: 28},
				{Term: "0", Position: 4, StartOffset: 29, EndOffset: 30},
			},
		},
		{
			name:  "unicode multi-byte characters",
			// "Distributed" (11B) + " " (1B) + "系统" (6B) + " " (1B) + "🚀" (4B) + " " (1B) + "Engine" (6B)
			input: "Distributed 系统 🚀 Engine",
			expected: []Token{
				{Term: "Distributed", Position: 0, StartOffset: 0, EndOffset: 11},
				{Term: "系统", Position: 1, StartOffset: 12, EndOffset: 18},
				{Term: "Engine", Position: 2, StartOffset: 24, EndOffset: 30},
			},
		},
		{
			name:     "consecutive punctuation and symbols",
			input:    "---!@#$%^&*()_+==--",
			expected: []Token{},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []Token{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := tok.Tokenize(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(tokens) == 0 && len(tc.expected) == 0 {
				return
			}

			if !reflect.DeepEqual(tokens, tc.expected) {
				t.Errorf("\nexpected: %+v\ngot:      %+v", tc.expected, tokens)
			}
		})
	}
}

func TestWhitespaceTokenizer(t *testing.T) {
	tok := NewWhitespaceTokenizer()

	if tok.Name() != "whitespace" {
		t.Errorf("expected 'whitespace', got %q", tok.Name())
	}

	input := "error-code:500 \t status=\"CRITICAL\" \n retry=true"
	tokens, err := tok.Tokenize(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []Token{
		{Term: "error-code:500", Position: 0, StartOffset: 0, EndOffset: 14},
		{Term: "status=\"CRITICAL\"", Position: 1, StartOffset: 17, EndOffset: 34},
		{Term: "retry=true", Position: 2, StartOffset: 37, EndOffset: 47},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, tokens)
	}
}

func TestKeywordTokenizer(t *testing.T) {
	tok := NewKeywordTokenizer()

	if tok.Name() != "keyword" {
		t.Errorf("expected 'keyword', got %q", tok.Name())
	}

	input := "custom/composite/path/id#42"
	tokens, err := tok.Tokenize(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}

	expected := Token{
		Term:        input,
		Position:    0,
		StartOffset: 0,
		EndOffset:   len(input),
	}

	if tokens[0] != expected {
		t.Errorf("expected %+v, got %+v", expected, tokens[0])
	}

	// Empty string
	emptyTokens, err := tok.Tokenize("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(emptyTokens) != 0 {
		t.Errorf("expected 0 tokens for empty input, got %d", len(emptyTokens))
	}
}

func BenchmarkStandardTokenizer(b *testing.B) {
	tok := NewStandardTokenizer()
	text := "SearchForge is a distributed search and indexing engine inspired by Elasticsearch, written in Go from scratch."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := tok.Tokenize(text)
		if err != nil {
			b.Fatal(err)
		}
	}
}
