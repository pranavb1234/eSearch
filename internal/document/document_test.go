package document

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewDocumentValid(t *testing.T) {
	doc, err := NewDocument("doc-1001")
	if err != nil {
		t.Fatalf("unexpected error creating document: %v", err)
	}

	if doc.ID != "doc-1001" {
		t.Errorf("expected ID 'doc-1001', got %q", doc.ID)
	}

	if doc.FieldCount() != 0 {
		t.Errorf("expected 0 fields, got %d", doc.FieldCount())
	}
}

func TestNewDocumentInvalidID(t *testing.T) {
	_, err := NewDocument("")
	if !errors.Is(err, ErrEmptyDocID) {
		t.Errorf("expected ErrEmptyDocID, got %v", err)
	}

	oversizedID := strings.Repeat("a", 513)
	_, err = NewDocument(oversizedID)
	if !errors.Is(err, ErrDocIDTooLong) {
		t.Errorf("expected ErrDocIDTooLong, got %v", err)
	}
}

func TestAddAllFieldTypes(t *testing.T) {
	doc, err := NewDocument("doc-valid")
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	textField, err := NewTextField("title", "Distributed Systems in Go", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keywordField, err := NewKeywordField("category", "engineering", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	numericField, err := NewNumericField("price", 49.99, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	timestampField, err := NewTimestampField("published_at", time.Now(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	booleanField, err := NewBooleanField("is_active", true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := []Field{textField, keywordField, numericField, timestampField, booleanField}
	for _, f := range fields {
		if err := doc.AddField(f); err != nil {
			t.Fatalf("failed to add field %s: %v", f.Name, err)
		}
	}

	if doc.FieldCount() != 5 {
		t.Fatalf("expected 5 fields, got %d", doc.FieldCount())
	}

	if err := doc.Validate(); err != nil {
		t.Fatalf("expected document to be valid, got: %v", err)
	}

	// Verify retrieval
	title, found := doc.GetField("title")
	if !found {
		t.Fatal("expected 'title' field to be found")
	}
	if title.Value != "Distributed Systems in Go" {
		t.Errorf("expected title value mismatch, got %v", title.Value)
	}

	// Verify non-existent field
	_, found = doc.GetField("non_existent")
	if found {
		t.Fatal("expected non_existent field to not be found")
	}
}

func TestDuplicateFieldRejection(t *testing.T) {
	doc, _ := NewDocument("doc-dup")

	f1, err := NewTextField("tag", "first", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := doc.AddField(f1); err != nil {
		t.Fatalf("unexpected error adding field: %v", err)
	}

	f2, err := NewTextField("tag", "second", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = doc.AddField(f2)
	if !errors.Is(err, ErrDuplicateField) {
		t.Fatalf("expected ErrDuplicateField, got %v", err)
	}
}

func TestFieldValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		field       Field
		expectedErr error
	}{
		{
			name: "empty field name",
			field: Field{
				Name:  "",
				Type:  TypeText,
				Value: "hello",
			},
			expectedErr: ErrInvalidFieldName,
		},
		{
			name: "invalid characters in name",
			field: Field{
				Name:  "invalid name with spaces",
				Type:  TypeText,
				Value: "hello",
			},
			expectedErr: ErrInvalidFieldName,
		},
		{
			name: "oversized field name",
			field: Field{
				Name:  strings.Repeat("x", 256),
				Type:  TypeText,
				Value: "hello",
			},
			expectedErr: ErrInvalidFieldName,
		},
		{
			name: "type mismatch text expects string",
			field: Field{
				Name:  "title",
				Type:  TypeText,
				Value: 12345,
			},
			expectedErr: ErrInvalidFieldValue,
		},
		{
			name: "type mismatch numeric expects number",
			field: Field{
				Name:  "count",
				Type:  TypeNumeric,
				Value: "not a number",
			},
			expectedErr: ErrInvalidFieldValue,
		},
		{
			name: "type mismatch timestamp expects time.Time",
			field: Field{
				Name:  "created_at",
				Type:  TypeTimestamp,
				Value: "2026-01-01",
			},
			expectedErr: ErrInvalidFieldValue,
		},
		{
			name: "type mismatch boolean expects bool",
			field: Field{
				Name:  "active",
				Type:  TypeBoolean,
				Value: "true",
			},
			expectedErr: ErrInvalidFieldValue,
		},
		{
			name: "unknown field type",
			field: Field{
				Name:  "unknown_field",
				Type:  FieldType(99),
				Value: "val",
			},
			expectedErr: ErrInvalidFieldType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.field.Validate()
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestDocumentValidateEmptyFields(t *testing.T) {
	doc, _ := NewDocument("doc-empty-fields")
	err := doc.Validate()
	if !errors.Is(err, ErrNoFields) {
		t.Fatalf("expected ErrNoFields, got %v", err)
	}
}

func TestFieldsSliceImmutability(t *testing.T) {
	doc, _ := NewDocument("doc-immut")
	f, _ := NewTextField("title", "original", true)
	_ = doc.AddField(f)

	fields := doc.Fields()
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}

	// Mutate returned slice
	fields[0].Name = "mutated"

	// Verify document internal fields were not changed
	internalField, found := doc.GetField("title")
	if !found {
		t.Fatal("expected original field 'title' to still exist")
	}
	if internalField.Name != "title" {
		t.Errorf("expected field name to remain 'title', got %q", internalField.Name)
	}
}

func TestFieldTypeString(t *testing.T) {
	cases := []struct {
		ft       FieldType
		expected string
	}{
		{TypeText, "text"},
		{TypeKeyword, "keyword"},
		{TypeNumeric, "numeric"},
		{TypeTimestamp, "timestamp"},
		{TypeBoolean, "boolean"},
		{FieldType(99), "unknown"},
	}

	for _, c := range cases {
		if c.ft.String() != c.expected {
			t.Errorf("expected %q, got %q", c.expected, c.ft.String())
		}
	}
}

