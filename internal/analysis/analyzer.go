package analysis

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNilTokenizer      = errors.New("analyzer: tokenizer cannot be nil")
	ErrAnalyzerNotFound  = errors.New("analyzer: analyzer not found in registry")
	ErrDuplicateAnalyzer = errors.New("analyzer: duplicate analyzer name")
)

// Analyzer encapsulates the end-to-end text analysis process.
// It transforms raw text into an indexed slice of tokens through CharFilters, Tokenizer, and TokenFilters.
type Analyzer interface {
	Name() string
	Analyze(text string) ([]Token, error)
}

// CustomAnalyzer provides a configurable composite implementation of Analyzer.
type CustomAnalyzer struct {
	name         string
	charFilters  []CharFilter
	tokenizer    Tokenizer
	tokenFilters []TokenFilter
}

// NewCustomAnalyzer constructs a custom analyzer pipeline.
func NewCustomAnalyzer(name string, tokenizer Tokenizer, charFilters []CharFilter, tokenFilters []TokenFilter) (*CustomAnalyzer, error) {
	if tokenizer == nil {
		return nil, ErrNilTokenizer
	}
	if name == "" {
		return nil, errors.New("analyzer: name cannot be empty")
	}

	return &CustomAnalyzer{
		name:         name,
		charFilters:  charFilters,
		tokenizer:    tokenizer,
		tokenFilters: tokenFilters,
	}, nil
}

// Name returns the analyzer's identifier.
func (a *CustomAnalyzer) Name() string {
	return a.name
}

// Analyze runs the complete analysis pipeline on input text.
func (a *CustomAnalyzer) Analyze(text string) ([]Token, error) {
	// 1. Character Filtering
	currentText := text
	for _, cf := range a.charFilters {
		currentText = cf.Filter(currentText)
	}

	// 2. Tokenization
	tokens, err := a.tokenizer.Tokenize(currentText)
	if err != nil {
		return nil, fmt.Errorf("analyzer %q tokenization failed: %w", a.name, err)
	}

	// 3. Token Filtering
	currentTokens := tokens
	for _, tf := range a.tokenFilters {
		currentTokens, err = tf.Filter(currentTokens)
		if err != nil {
			return nil, fmt.Errorf("analyzer %q filter %q failed: %w", a.name, tf.Name(), err)
		}
	}

	return currentTokens, nil
}

// NewStandardAnalyzer creates the default general-purpose full-text analyzer.
// Pipeline: LowercaseCharFilter -> StandardTokenizer -> StopwordFilter -> StemmerFilter
func NewStandardAnalyzer() *CustomAnalyzer {
	analyzer, _ := NewCustomAnalyzer(
		"standard",
		NewStandardTokenizer(),
		[]CharFilter{NewLowercaseCharFilter()},
		[]TokenFilter{
			NewStopwordFilter(nil),
			NewStemmerFilter(),
		},
	)
	return analyzer
}

// NewSimpleAnalyzer creates an analyzer with lowercasing and standard tokenization (no stemming or stopwords).
// Pipeline: LowercaseCharFilter -> StandardTokenizer
func NewSimpleAnalyzer() *CustomAnalyzer {
	analyzer, _ := NewCustomAnalyzer(
		"simple",
		NewStandardTokenizer(),
		[]CharFilter{NewLowercaseCharFilter()},
		nil,
	)
	return analyzer
}

// NewWhitespaceAnalyzer creates an analyzer that splits strictly on whitespace with lowercasing.
// Pipeline: LowercaseCharFilter -> WhitespaceTokenizer
func NewWhitespaceAnalyzer() *CustomAnalyzer {
	analyzer, _ := NewCustomAnalyzer(
		"whitespace",
		NewWhitespaceTokenizer(),
		[]CharFilter{NewLowercaseCharFilter()},
		nil,
	)
	return analyzer
}

// NewKeywordAnalyzer creates an analyzer that treats the entire input as a single unanalyzed token.
// Pipeline: KeywordTokenizer (no char filters or token filters)
func NewKeywordAnalyzer() *CustomAnalyzer {
	analyzer, _ := NewCustomAnalyzer(
		"keyword",
		NewKeywordTokenizer(),
		nil,
		nil,
	)
	return analyzer
}

// Registry manages thread-safe registration and resolution of named Analyzers.
type Registry struct {
	mu        sync.RWMutex
	analyzers map[string]Analyzer
}

// NewRegistry creates a new analyzer registry.
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]Analyzer),
	}
}

// NewDefaultRegistry creates a registry pre-loaded with standard built-in analyzers.
func NewDefaultRegistry() *Registry {
	reg := NewRegistry()
	_ = reg.Register(NewStandardAnalyzer())
	_ = reg.Register(NewSimpleAnalyzer())
	_ = reg.Register(NewWhitespaceAnalyzer())
	_ = reg.Register(NewKeywordAnalyzer())
	return reg
}

// Register adds an analyzer to the registry. Returns error if name already exists.
func (r *Registry) Register(a Analyzer) error {
	if a == nil {
		return errors.New("analyzer: cannot register nil analyzer")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := a.Name()
	if _, exists := r.analyzers[name]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateAnalyzer, name)
	}

	r.analyzers[name] = a
	return nil
}

// Get retrieves an analyzer by name.
func (r *Registry) Get(name string) (Analyzer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, found := r.analyzers[name]
	return a, found
}

