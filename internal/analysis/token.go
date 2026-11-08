package analysis

// Token represents an indexed term along with its positional and offset metadata.
type Token struct {
	Term        string // Normalized text of the token
	Position    int    // Sequential position index in the document/stream (used for phrase queries)
	StartOffset int    // Start byte offset in the source text (used for snippet highlighting)
	EndOffset   int    // End byte offset in the source text
}

// Tokenizer defines the interface for splitting character sequences into a slice of Tokens.
type Tokenizer interface {
	Name() string
	Tokenize(input string) ([]Token, error)
}

