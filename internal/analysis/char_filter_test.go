package analysis

import (
	"testing"
)

func TestLowercaseCharFilter(t *testing.T) {
	filter := NewLowercaseCharFilter()

	if filter.Name() != "lowercase" {
		t.Errorf("expected name 'lowercase', got %q", filter.Name())
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already lowercase ASCII",
			input:    "distributed search engine",
			expected: "distributed search engine",
		},
		{
			name:     "mixed case ASCII",
			input:    "SearchForge Distributed Systems",
			expected: "searchforge distributed systems",
		},
		{
			name:     "all uppercase ASCII",
			input:    "RAFT CONSENSUS",
			expected: "raft consensus",
		},
		{
			name:     "unicode accents",
			input:    "MÜNCHEN ÉLÉPHANT",
			expected: "münchen éléphant",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filter.Filter(tc.input)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestMappingCharFilter(t *testing.T) {
	replacements := map[string]string{
		"&":  " and ",
		"@":  " at ",
		"//": " ",
	}
	filter := NewMappingCharFilter(replacements)

	if filter.Name() != "mapping" {
		t.Errorf("expected name 'mapping', got %q", filter.Name())
	}

	input := "Go & Rust @ SearchForge // 2026"
	expected := "Go  and  Rust  at  SearchForge   2026"
	got := filter.Filter(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	// Empty mapping edge case
	emptyFilter := NewMappingCharFilter(nil)
	if emptyFilter.Filter("test") != "test" {
		t.Errorf("expected unmodified string with empty mapping")
	}
}

func TestWhitespaceNormalizerFilter(t *testing.T) {
	filter := NewWhitespaceNormalizerFilter()

	if filter.Name() != "whitespace_normalizer" {
		t.Errorf("expected name 'whitespace_normalizer', got %q", filter.Name())
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mixed tabs newlines and spaces",
			input:    "  hello \t\t distributed\n\n\r systems   ",
			expected: "hello distributed systems",
		},
		{
			name:     "already normalized",
			input:    "clean string",
			expected: "clean string",
		},
		{
			name:     "only whitespace",
			input:    " \t \n \r  ",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filter.Filter(tc.input)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestCharFilterPipeline(t *testing.T) {
	mapping := NewMappingCharFilter(map[string]string{
		"&": " and ",
	})
	whitespace := NewWhitespaceNormalizerFilter()
	lowercase := NewLowercaseCharFilter()

	pipeline := NewCharFilterPipeline(mapping, whitespace, lowercase)

	if pipeline.Name() != "pipeline" {
		t.Errorf("expected name 'pipeline', got %q", pipeline.Name())
	}

	input := "   DISTRIBUTED & FAULT-TOLERANT   SYSTEMS \n\n"
	expected := "distributed and fault-tolerant systems"

	got := pipeline.Filter(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func BenchmarkLowercaseFilterCleanASCII(b *testing.B) {
	filter := NewLowercaseCharFilter()
	input := "distributed search and indexing engine written in go"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = filter.Filter(input)
	}
}

func BenchmarkLowercaseFilterMixedASCII(b *testing.B) {
	filter := NewLowercaseCharFilter()
	input := "SearchForge Distributed Systems Engine Written In Go"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = filter.Filter(input)
	}
}

