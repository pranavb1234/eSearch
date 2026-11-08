package document

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyDocID     = errors.New("document: document ID cannot be empty")
	ErrDocIDTooLong   = errors.New("document: document ID exceeds maximum length of 512 bytes")
	ErrDuplicateField = errors.New("document: duplicate field name")
	ErrNoFields       = errors.New("document: document must contain at least one field")
)

// Document represents an ingestible unit in SearchForge.
type Document struct {
	ID     string
	fields []Field
	index  map[string]int // maps field name to index in fields slice for O(1) lookup
}

// NewDocument instantiates a new empty document with the given external ID.
func NewDocument(id string) (*Document, error) {
	if id == "" {
		return nil, ErrEmptyDocID
	}
	if len(id) > 512 {
		return nil, ErrDocIDTooLong
	}
	return &Document{
		ID:     id,
		fields: make([]Field, 0, 8),
		index:  make(map[string]int),
	}, nil
}

// AddField adds a validated field to the document. Returns error if field name already exists or field is invalid.
func (d *Document) AddField(f Field) error {
	if err := f.Validate(); err != nil {
		return err
	}
	if _, exists := d.index[f.Name]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateField, f.Name)
	}

	d.fields = append(d.fields, f)
	d.index[f.Name] = len(d.fields) - 1
	return nil
}

// GetField retrieves a field by name.
func (d *Document) GetField(name string) (Field, bool) {
	idx, ok := d.index[name]
	if !ok {
		return Field{}, false
	}
	return d.fields[idx], true
}

// Fields returns a copy of the document's fields to prevent external mutation.
func (d *Document) Fields() []Field {
	result := make([]Field, len(d.fields))
	copy(result, d.fields)
	return result
}

// FieldCount returns the number of fields in the document.
func (d *Document) FieldCount() int {
	return len(d.fields)
}

// Validate ensures the document conforms to all internal constraints.
func (d *Document) Validate() error {
	if d.ID == "" {
		return ErrEmptyDocID
	}
	if len(d.ID) > 512 {
		return ErrDocIDTooLong
	}
	if len(d.fields) == 0 {
		return ErrNoFields
	}

	seen := make(map[string]struct{}, len(d.fields))
	for _, f := range d.fields {
		if err := f.Validate(); err != nil {
			return err
		}
		if _, exists := seen[f.Name]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateField, f.Name)
		}
		seen[f.Name] = struct{}{}
	}

	return nil
}

