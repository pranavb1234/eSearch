package analysis

import (
	"strings"
)

// DefaultEnglishStopwords provides the standard Lucene/Elasticsearch English stopword set.
var DefaultEnglishStopwords = map[string]struct{}{
	"a": {}, "about": {}, "above": {}, "after": {}, "again": {}, "against": {}, "all": {},
	"am": {}, "an": {}, "and": {}, "any": {}, "are": {}, "as": {}, "at": {},
	"be": {}, "because": {}, "been": {}, "before": {}, "being": {}, "below": {}, "between": {},
	"both": {}, "but": {}, "by": {},
	"could": {},
	"did": {}, "do": {}, "does": {}, "doing": {}, "down": {}, "during": {},
	"each": {},
	"few": {}, "for": {}, "from": {}, "further": {},
	"had": {}, "has": {}, "have": {}, "having": {}, "he": {}, "her": {}, "here": {},
	"hers": {}, "herself": {}, "him": {}, "himself": {}, "his": {}, "how": {},
	"i": {}, "if": {}, "in": {}, "into": {}, "is": {}, "it": {}, "its": {}, "itself": {},
	"me": {}, "more": {}, "most": {}, "my": {}, "myself": {},
	"no": {}, "nor": {}, "not": {},
	"of": {}, "off": {}, "on": {}, "once": {}, "only": {}, "or": {}, "other": {}, "ought": {},
	"our": {}, "ours": {}, "ourselves": {}, "out": {}, "over": {}, "own": {},
	"s": {}, "same": {}, "she": {}, "should": {}, "so": {}, "some": {}, "such": {},
	"t": {}, "than": {}, "that": {}, "the": {}, "their": {}, "theirs": {}, "them": {}, "themselves": {},
	"then": {}, "there": {}, "these": {}, "they": {}, "this": {}, "those": {}, "through": {},
	"to": {}, "too": {},
	"under": {}, "until": {}, "up": {},
	"very": {},
	"was": {}, "we": {}, "were": {}, "what": {}, "when": {}, "where": {}, "which": {},
	"while": {}, "who": {}, "whom": {}, "why": {}, "with": {}, "would": {},
	"you": {}, "your": {}, "yours": {}, "yourself": {}, "yourselves": {},
}

// TokenFilter defines the interface for transforming, pruning, or enriching a slice of Tokens.
type TokenFilter interface {
	Name() string
	Filter(tokens []Token) ([]Token, error)
}

// StopwordFilter drops tokens whose terms exist in a designated stopword set.
// It preserves original token positions so phrase proximity remains accurate.
type StopwordFilter struct {
	stopwords map[string]struct{}
}

// NewStopwordFilter creates a StopwordFilter with the provided stopword set.
// If customSet is nil, DefaultEnglishStopwords is used.
func NewStopwordFilter(customSet map[string]struct{}) *StopwordFilter {
	if customSet == nil {
		customSet = DefaultEnglishStopwords
	}
	cp := make(map[string]struct{}, len(customSet))
	for k := range customSet {
		cp[strings.ToLower(k)] = struct{}{}
	}
	return &StopwordFilter{stopwords: cp}
}

// Name returns the identifier of this filter.
func (f *StopwordFilter) Name() string {
	return "stopword"
}

// Filter filters out stopwords from the token list while preserving position indices.
func (f *StopwordFilter) Filter(tokens []Token) ([]Token, error) {
	if len(tokens) == 0 {
		return tokens, nil
	}

	result := make([]Token, 0, len(tokens))
	for _, tok := range tokens {
		lower := strings.ToLower(tok.Term)
		if _, isStop := f.stopwords[lower]; !isStop {
			result = append(result, tok)
		}
	}
	return result, nil
}

// LengthFilter drops tokens whose term length falls outside [min, max].
type LengthFilter struct {
	min int
	max int
}

// NewLengthFilter instantiates a LengthFilter with min and max bounds.
func NewLengthFilter(min, max int) *LengthFilter {
	return &LengthFilter{min: min, max: max}
}

// Name returns the identifier of this filter.
func (f *LengthFilter) Name() string {
	return "length"
}

// Filter drops tokens shorter than min or longer than max.
func (f *LengthFilter) Filter(tokens []Token) ([]Token, error) {
	if len(tokens) == 0 {
		return tokens, nil
	}

	result := make([]Token, 0, len(tokens))
	for _, tok := range tokens {
		l := len([]rune(tok.Term))
		if l >= f.min && l <= f.max {
			result = append(result, tok)
		}
	}
	return result, nil
}

// TokenFilterPipeline chains multiple TokenFilters in sequential order.
type TokenFilterPipeline struct {
	filters []TokenFilter
}

// NewTokenFilterPipeline instantiates an ordered pipeline of TokenFilters.
func NewTokenFilterPipeline(filters ...TokenFilter) *TokenFilterPipeline {
	return &TokenFilterPipeline{filters: filters}
}

// Name returns the pipeline identifier.
func (p *TokenFilterPipeline) Name() string {
	return "token_filter_pipeline"
}

// Filter executes each filter in the chain sequentially.
func (p *TokenFilterPipeline) Filter(tokens []Token) ([]Token, error) {
	current := tokens
	var err error
	for _, f := range p.filters {
		current, err = f.Filter(current)
		if err != nil {
			return nil, err
		}
	}
	return current, nil
}
