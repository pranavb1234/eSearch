package document

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// FieldType represents the supported data types for document fields.
type FieldType uint8

const (
	// TypeText indicates a full-text field that undergoes tokenization and analysis.
	TypeText FieldType = iota + 1
	// TypeKeyword indicates an exact-match string field not broken down by an analyzer.
	TypeKeyword
	// TypeNumeric indicates an integer or floating-point numerical field.
	TypeNumeric
	// TypeTimestamp indicates a temporal field stored with UTC representation.
	TypeTimestamp
	// TypeBoolean indicates a boolean true/false field.
	TypeBoolean
)

// String returns the human-readable representation of FieldType.
func (t FieldType) String() string {
	switch t {
	case TypeText:
		return "text"
	case TypeKeyword:
		return "keyword"
	case TypeNumeric:
		return "numeric"
	case TypeTimestamp:
		return "timestamp"
	case TypeBoolean:
		return "boolean"
	default:
		return "unknown"
	}
}

var (
	// validFieldNameRegex ensures field names contain only alphanumeric characters, underscores, and dots.
	validFieldNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`)

	ErrInvalidFieldName  = errors.New("document: invalid field name")
	ErrInvalidFieldType  = errors.New("document: invalid field type")
	ErrInvalidFieldValue = errors.New("document: field value does not match declared type")
)

// Field represents a single named, typed attribute of a Document.
type Field struct {
	Name    string
	Type    FieldType
	Value   any
	Stored  bool // If true, stored on disk for retrieval in search hits
	Indexed bool // If true, indexed into inverted index for querying
}

// NewTextField creates a full-text field.
func NewTextField(name string, value string, stored bool) (Field, error) {
	f := Field{
		Name:    name,
		Type:    TypeText,
		Value:   value,
		Stored:  stored,
		Indexed: true,
	}
	if err := f.Validate(); err != nil {
		return Field{}, err
	}
	return f, nil
}

// NewKeywordField creates an unanalyzed exact-match string field.
func NewKeywordField(name string, value string, stored bool) (Field, error) {
	f := Field{
		Name:    name,
		Type:    TypeKeyword,
		Value:   value,
		Stored:  stored,
		Indexed: true,
	}
	if err := f.Validate(); err != nil {
		return Field{}, err
	}
	return f, nil
}

// NewNumericField creates a numeric field (accepts any Go numeric primitive).
func NewNumericField(name string, value float64, stored bool) (Field, error) {
	f := Field{
		Name:    name,
		Type:    TypeNumeric,
		Value:   value,
		Stored:  stored,
		Indexed: true,
	}
	if err := f.Validate(); err != nil {
		return Field{}, err
	}
	return f, nil
}

// NewTimestampField creates a timestamp field converted to UTC.
func NewTimestampField(name string, value time.Time, stored bool) (Field, error) {
	f := Field{
		Name:    name,
		Type:    TypeTimestamp,
		Value:   value.UTC(),
		Stored:  stored,
		Indexed: true,
	}
	if err := f.Validate(); err != nil {
		return Field{}, err
	}
	return f, nil
}

// NewBooleanField creates a boolean field.
func NewBooleanField(name string, value bool, stored bool) (Field, error) {
	f := Field{
		Name:    name,
		Type:    TypeBoolean,
		Value:   value,
		Stored:  stored,
		Indexed: true,
	}
	if err := f.Validate(); err != nil {
		return Field{}, err
	}
	return f, nil
}

// Validate checks whether the field satisfies naming rules and value-type conformance.
func (f Field) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("%w: field name cannot be empty", ErrInvalidFieldName)
	}
	if len(f.Name) > 255 {
		return fmt.Errorf("%w: field name exceeds max length of 255 characters", ErrInvalidFieldName)
	}
	if !validFieldNameRegex.MatchString(f.Name) {
		return fmt.Errorf("%w: %q (allowed: alphanumeric, underscore, dot)", ErrInvalidFieldName, f.Name)
	}

	switch f.Type {
	case TypeText, TypeKeyword:
		if _, ok := f.Value.(string); !ok {
			return fmt.Errorf("%w: expected string for %s, got %T", ErrInvalidFieldValue, f.Type, f.Value)
		}
	case TypeNumeric:
		switch f.Value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid numeric types
		default:
			return fmt.Errorf("%w: expected numeric value for %s, got %T", ErrInvalidFieldValue, f.Type, f.Value)
		}
	case TypeTimestamp:
		if _, ok := f.Value.(time.Time); !ok {
			return fmt.Errorf("%w: expected time.Time for %s, got %T", ErrInvalidFieldValue, f.Type, f.Value)
		}
	case TypeBoolean:
		if _, ok := f.Value.(bool); !ok {
			return fmt.Errorf("%w: expected bool for %s, got %T", ErrInvalidFieldValue, f.Type, f.Value)
		}
	default:
		return fmt.Errorf("%w: unknown field type id %d", ErrInvalidFieldType, f.Type)
	}

	return nil
}

