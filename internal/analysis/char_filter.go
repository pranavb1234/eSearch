package analysis

import (
	"strings"
	"unicode"
)

// CharFilter defines the interface for character-level transformations applied to raw text
// before tokenization occurs.
type CharFilter interface {
	Name() string
	Filter(input string) string
}

// LowercaseCharFilter converts all characters in a string to lowercase.
// It includes an allocation-free fast path for ASCII text that is already lowercase.
type LowercaseCharFilter struct{}

// NewLowercaseCharFilter creates a new lowercase character filter.
func NewLowercaseCharFilter() *LowercaseCharFilter {
	return &LowercaseCharFilter{}
}

// Name returns the identifier of this filter.
func (f *LowercaseCharFilter) Name() string {
	return "lowercase"
}

// Filter transforms uppercase characters to lowercase.
func (f *LowercaseCharFilter) Filter(input string) string {
	hasUpper := false
	isASCII := true

	for i := 0; i < len(input); i++ {
		c := input[i]
		if c >= 0x80 {
			isASCII = false
			break
		}
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
	}

	// If non-ASCII runes are present, delegate to full Unicode lowercasing
	if !isASCII {
		return strings.ToLower(input)
	}

	// Fast-path: already pure lowercase ASCII - 0 allocations
	if !hasUpper {
		return input
	}

	// Single allocation ASCII transform
	var b strings.Builder
	b.Grow(len(input))
	for i := 0; i < len(input); i++ {
		c := input[i]
		if c >= 'A' && c <= 'Z' {
			b.WriteByte(c + ('a' - 'A'))
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// MappingCharFilter replaces configured target substrings with specified replacements.
type MappingCharFilter struct {
	replacements map[string]string
}

// NewMappingCharFilter creates a character filter with a key-value replacement map.
func NewMappingCharFilter(replacements map[string]string) *MappingCharFilter {
	cp := make(map[string]string, len(replacements))
	for k, v := range replacements {
		cp[k] = v
	}
	return &MappingCharFilter{replacements: cp}
}

// Name returns the identifier of this filter.
func (f *MappingCharFilter) Name() string {
	return "mapping"
}

// Filter replaces occurrences of mapped keys with their corresponding values.
func (f *MappingCharFilter) Filter(input string) string {
	if len(f.replacements) == 0 || input == "" {
		return input
	}
	res := input
	for oldSub, newSub := range f.replacements {
		if strings.Contains(res, oldSub) {
			res = strings.ReplaceAll(res, oldSub, newSub)
		}
	}
	return res
}

// WhitespaceNormalizerFilter collapses consecutive whitespace runes into a single ASCII space
// and strips leading and trailing whitespace.
type WhitespaceNormalizerFilter struct{}

// NewWhitespaceNormalizerFilter creates a whitespace normalizer.
func NewWhitespaceNormalizerFilter() *WhitespaceNormalizerFilter {
	return &WhitespaceNormalizerFilter{}
}

// Name returns the identifier of this filter.
func (f *WhitespaceNormalizerFilter) Name() string {
	return "whitespace_normalizer"
}

// Filter normalizes spaces, tabs, and newlines.
func (f *WhitespaceNormalizerFilter) Filter(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(trimmed))

	inWhitespace := false
	for _, r := range trimmed {
		if unicode.IsSpace(r) {
			if !inWhitespace {
				b.WriteByte(' ')
				inWhitespace = true
			}
		} else {
			b.WriteRune(r)
			inWhitespace = false
		}
	}
	return b.String()
}

// CharFilterPipeline chains multiple CharFilters sequentially.
type CharFilterPipeline struct {
	filters []CharFilter
}

// NewCharFilterPipeline constructs an ordered chain of CharFilters.
func NewCharFilterPipeline(filters ...CharFilter) *CharFilterPipeline {
	return &CharFilterPipeline{filters: filters}
}

// Name returns the pipeline identifier.
func (p *CharFilterPipeline) Name() string {
	return "pipeline"
}

// Filter passes the string through each filter in sequence.
func (p *CharFilterPipeline) Filter(input string) string {
	res := input
	for _, filter := range p.filters {
		res = filter.Filter(res)
	}
	return res
}
