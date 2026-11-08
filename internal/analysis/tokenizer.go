package analysis

import (
	"unicode"
)

// StandardTokenizer splits text into tokens on word boundaries (letters and numbers).
// Punctuation, symbols, and whitespace are discarded as delimiters.
type StandardTokenizer struct{}

// NewStandardTokenizer instantiates a new StandardTokenizer.
func NewStandardTokenizer() *StandardTokenizer {
	return &StandardTokenizer{}
}

// Name returns the identifier of this tokenizer.
func (t *StandardTokenizer) Name() string {
	return "standard"
}

// Tokenize processes the input string into a slice of positional Tokens.
func (t *StandardTokenizer) Tokenize(input string) ([]Token, error) {
	if len(input) == 0 {
		return []Token{}, nil
	}

	// Heuristic pre-allocation: average word length ~ 5-6 bytes
	estimatedCapacity := len(input) / 6
	if estimatedCapacity < 4 {
		estimatedCapacity = 4
	}
	tokens := make([]Token, 0, estimatedCapacity)

	inToken := false
	startOffset := 0
	position := 0

	for offset, r := range input {
		isWordChar := unicode.IsLetter(r) || unicode.IsDigit(r)

		if isWordChar {
			if !inToken {
				startOffset = offset
				inToken = true
			}
		} else {
			if inToken {
				tokens = append(tokens, Token{
					Term:        input[startOffset:offset],
					Position:    position,
					StartOffset: startOffset,
					EndOffset:   offset,
				})
				position++
				inToken = false
			}
		}
	}

	// Flush any trailing token at the end of the input
	if inToken {
		tokens = append(tokens, Token{
			Term:        input[startOffset:],
			Position:    position,
			StartOffset: startOffset,
			EndOffset:   len(input),
		})
	}

	return tokens, nil
}

// WhitespaceTokenizer splits text strictly on whitespace characters, preserving punctuation.
type WhitespaceTokenizer struct{}

// NewWhitespaceTokenizer instantiates a WhitespaceTokenizer.
func NewWhitespaceTokenizer() *WhitespaceTokenizer {
	return &WhitespaceTokenizer{}
}

// Name returns the identifier of this tokenizer.
func (t *WhitespaceTokenizer) Name() string {
	return "whitespace"
}

// Tokenize splits on Unicode whitespace runes.
func (t *WhitespaceTokenizer) Tokenize(input string) ([]Token, error) {
	if len(input) == 0 {
		return []Token{}, nil
	}

	tokens := make([]Token, 0, len(input)/8+2)
	inToken := false
	startOffset := 0
	position := 0

	for offset, r := range input {
		isSpace := unicode.IsSpace(r)

		if !isSpace {
			if !inToken {
				startOffset = offset
				inToken = true
			}
		} else {
			if inToken {
				tokens = append(tokens, Token{
					Term:        input[startOffset:offset],
					Position:    position,
					StartOffset: startOffset,
					EndOffset:   offset,
				})
				position++
				inToken = false
			}
		}
	}

	if inToken {
		tokens = append(tokens, Token{
			Term:        input[startOffset:],
			Position:    position,
			StartOffset: startOffset,
			EndOffset:   len(input),
		})
	}

	return tokens, nil
}

// KeywordTokenizer treats the entire input string as a single unanalyzed token.
type KeywordTokenizer struct{}

// NewKeywordTokenizer instantiates a KeywordTokenizer.
func NewKeywordTokenizer() *KeywordTokenizer {
	return &KeywordTokenizer{}
}

// Name returns the identifier of this tokenizer.
func (t *KeywordTokenizer) Name() string {
	return "keyword"
}

// Tokenize returns the input as a single token.
func (t *KeywordTokenizer) Tokenize(input string) ([]Token, error) {
	if len(input) == 0 {
		return []Token{}, nil
	}

	return []Token{
		{
			Term:        input,
			Position:    0,
			StartOffset: 0,
			EndOffset:   len(input),
		},
	}, nil
}

