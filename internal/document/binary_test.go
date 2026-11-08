package document

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func createSampleDocument(t *testing.T) *Document {
	t.Helper()
	doc, err := NewDocument("doc-sample-42")
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	f1, err := NewTextField("title", "Distributed Systems with Go", true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	f2, err := NewKeywordField("category", "engineering", true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	f3, err := NewNumericField("score", 98.75, true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	f4, err := NewTimestampField("timestamp", time.Date(2026, 11, 8, 12, 0, 0, 123456000, time.UTC), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	f5, err := NewBooleanField("is_replicated", true, true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, f := range []Field{f1, f2, f3, f4, f5} {
		if err := doc.AddField(f); err != nil {
			t.Fatalf("failed adding field %s: %v", f.Name, err)
		}
	}

	return doc
}

func TestBinaryRoundTrip(t *testing.T) {
	doc := createSampleDocument(t)

	data, err := EncodeDocument(doc)
	if err != nil {
		t.Fatalf("failed to encode document: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty byte slice")
	}

	decoded, err := DecodeDocument(data)
	if err != nil {
		t.Fatalf("failed to decode document: %v", err)
	}

	if decoded.ID != doc.ID {
		t.Errorf("expected ID %q, got %q", doc.ID, decoded.ID)
	}

	if decoded.FieldCount() != doc.FieldCount() {
		t.Fatalf("expected %d fields, got %d", doc.FieldCount(), decoded.FieldCount())
	}

	// Verify each field
	for _, orig := range doc.Fields() {
		dec, found := decoded.GetField(orig.Name)
		if !found {
			t.Fatalf("missing field %q in decoded doc", orig.Name)
		}

		if dec.Type != orig.Type {
			t.Errorf("field %s: expected type %v, got %v", orig.Name, orig.Type, dec.Type)
		}
		if dec.Stored != orig.Stored {
			t.Errorf("field %s: expected stored %v, got %v", orig.Name, orig.Stored, dec.Stored)
		}
		if dec.Indexed != orig.Indexed {
			t.Errorf("field %s: expected indexed %v, got %v", orig.Name, orig.Indexed, dec.Indexed)
		}

		switch orig.Type {
		case TypeTimestamp:
			origTime := orig.Value.(time.Time)
			decTime := dec.Value.(time.Time)
			if !origTime.Equal(decTime) {
				t.Errorf("time mismatch for %s: orig %v, got %v", orig.Name, origTime, decTime)
			}
		default:
			if dec.Value != orig.Value {
				t.Errorf("value mismatch for %s: orig %v (%T), got %v (%T)", orig.Name, orig.Value, orig.Value, dec.Value, dec.Value)
			}
		}
	}
}

func TestCorruptMagicByte(t *testing.T) {
	doc := createSampleDocument(t)
	data, err := EncodeDocument(doc)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Corrupt the magic byte
	data[0] = 0xFF

	_, err = DecodeDocument(data)
	if !errors.Is(err, ErrCorruptData) {
		t.Fatalf("expected ErrCorruptData, got %v", err)
	}
}

func TestUnsupportedVersion(t *testing.T) {
	doc := createSampleDocument(t)
	data, err := EncodeDocument(doc)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Corrupt version byte to 99
	data[1] = 99

	_, err = DecodeDocument(data)
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("expected ErrUnsupportedVersion, got %v", err)
	}
}

func TestTruncatedPayloads(t *testing.T) {
	doc := createSampleDocument(t)
	data, err := EncodeDocument(doc)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Truncate at every single byte offset from 0 to len-1
	for i := 0; i < len(data); i++ {
		truncated := bytes.Clone(data[:i])
		_, err := DecodeDocument(truncated)
		if err == nil {
			t.Fatalf("expected error for truncated slice of length %d, got nil", i)
		}
	}
}

func TestEncodeNilDocument(t *testing.T) {
	_, err := EncodeDocument(nil)
	if err == nil {
		t.Fatal("expected error encoding nil document, got nil")
	}
}

func BenchmarkDocumentEncoding(b *testing.B) {
	doc, _ := NewDocument("bench-doc-1")
	f1, _ := NewTextField("title", "Distributed Systems with SearchForge", true)
	f2, _ := NewKeywordField("author", "engineer", true)
	f3, _ := NewNumericField("views", 102450.0, true)
	_ = doc.AddField(f1)
	_ = doc.AddField(f2)
	_ = doc.AddField(f3)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := EncodeDocument(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDocumentDecoding(b *testing.B) {
	doc, _ := NewDocument("bench-doc-1")
	f1, _ := NewTextField("title", "Distributed Systems with SearchForge", true)
	f2, _ := NewKeywordField("author", "engineer", true)
	f3, _ := NewNumericField("views", 102450.0, true)
	_ = doc.AddField(f1)
	_ = doc.AddField(f2)
	_ = doc.AddField(f3)

	data, err := EncodeDocument(doc)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DecodeDocument(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

