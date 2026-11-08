package index

import (
	"reflect"
	"testing"
)

func TestPostingsListAddMonotonic(t *testing.T) {
	pl := NewPostingsList()

	// Document 1 contains term at positions 0, 5, 12
	pl.Add(1, 0)
	pl.Add(1, 5)
	pl.Add(1, 12)

	// Document 3 contains term at position 2
	pl.Add(3, 2)

	// Document 10 contains term at positions 1, 4
	pl.Add(10, 1)
	pl.Add(10, 4)

	if pl.Len() != 3 {
		t.Fatalf("expected 3 documents (DF), got %d", pl.Len())
	}
	if pl.TotalTermFreq() != 6 {
		t.Fatalf("expected 6 total term occurrences (TTF), got %d", pl.TotalTermFreq())
	}

	expected := []Posting{
		{DocID: 1, TermFrequency: 3, Positions: []uint32{0, 5, 12}},
		{DocID: 3, TermFrequency: 1, Positions: []uint32{2}},
		{DocID: 10, TermFrequency: 2, Positions: []uint32{1, 4}},
	}

	if !reflect.DeepEqual(pl.Postings(), expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, pl.Postings())
	}
}

func TestPostingsListAddOutOfOrder(t *testing.T) {
	pl := NewPostingsList()

	// Insert DocIDs in non-monotonic order: 10, 2, 5, 2
	pl.Add(10, 1)
	pl.Add(2, 3)
	pl.Add(5, 0)
	pl.Add(2, 7) // Additional position in doc 2

	if pl.Len() != 3 {
		t.Fatalf("expected 3 documents, got %d", pl.Len())
	}
	if pl.TotalTermFreq() != 4 {
		t.Fatalf("expected 4 total term frequency, got %d", pl.TotalTermFreq())
	}

	// Must be sorted in ascending order of DocID: 2, 5, 10
	expected := []Posting{
		{DocID: 2, TermFrequency: 2, Positions: []uint32{3, 7}},
		{DocID: 5, TermFrequency: 1, Positions: []uint32{0}},
		{DocID: 10, TermFrequency: 1, Positions: []uint32{1}},
	}

	if !reflect.DeepEqual(pl.Postings(), expected) {
		t.Errorf("\nexpected: %+v\ngot:      %+v", expected, pl.Postings())
	}
}

func TestPostingsListGet(t *testing.T) {
	pl := NewPostingsList()
	pl.Add(5, 1)
	pl.Add(12, 4)
	pl.Add(20, 9)

	// Existing DocID
	p, found := pl.Get(12)
	if !found {
		t.Fatal("expected docID 12 to be found")
	}
	if p.DocID != 12 || p.TermFrequency != 1 || p.Positions[0] != 4 {
		t.Errorf("unexpected posting: %+v", p)
	}

	// Non-existent DocIDs
	for _, id := range []uint32{0, 6, 21} {
		_, found := pl.Get(id)
		if found {
			t.Errorf("expected docID %d to not be found", id)
		}
	}

	// Empty list lookup
	emptyList := NewPostingsList()
	if _, found := emptyList.Get(1); found {
		t.Error("expected not found on empty list")
	}
}

func TestPostingsIteratorTraversal(t *testing.T) {
	pl := NewPostingsList()
	docIDs := []uint32{1, 4, 8, 15, 20}
	for _, id := range docIDs {
		pl.Add(id, 0)
	}

	it := pl.Iterator()

	// Prior to calling Next(), index is -1
	if it.DocID() != 0 {
		t.Errorf("expected 0 before Next(), got %d", it.DocID())
	}

	visited := make([]uint32, 0, len(docIDs))
	for it.Next() {
		visited = append(visited, it.DocID())
	}

	if !reflect.DeepEqual(visited, docIDs) {
		t.Errorf("expected %+v, got %+v", docIDs, visited)
	}

	// Calling Next() after exhaustion returns false
	if it.Next() {
		t.Error("expected false after exhaustion")
	}

	// Reset iterator
	it.Reset()
	if !it.Next() || it.DocID() != 1 {
		t.Error("expected iterator reset to start at first element")
	}
}

func TestPostingsIteratorSeek(t *testing.T) {
	pl := NewPostingsList()
	docIDs := []uint32{10, 20, 30, 40, 50, 60, 70}
	for _, id := range docIDs {
		pl.Add(id, 0)
	}

	it := pl.Iterator()

	// 1. Seek exact match from beginning
	if !it.Seek(30) || it.DocID() != 30 {
		t.Errorf("expected seek(30) to land on 30, got %d", it.DocID())
	}

	// 2. Seek non-exact value (landing on next highest)
	if !it.Seek(45) || it.DocID() != 50 {
		t.Errorf("expected seek(45) to land on 50, got %d", it.DocID())
	}

	// 3. Seek to already current position
	if !it.Seek(50) || it.DocID() != 50 {
		t.Errorf("expected seek(50) to remain on 50, got %d", it.DocID())
	}

	// 4. Seek past the end
	if it.Seek(100) {
		t.Errorf("expected seek(100) to return false, got true (docID %d)", it.DocID())
	}

	// Empty iterator seek
	emptyIt := NewPostingsList().Iterator()
	if emptyIt.Seek(10) {
		t.Error("expected seek on empty list to return false")
	}
}

func BenchmarkPostingsListAddMonotonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pl := NewPostingsList()
		for docID := uint32(0); docID < 1000; docID++ {
			pl.Add(docID, 1)
			pl.Add(docID, 4)
		}
	}
}

func BenchmarkPostingsIteratorNext(b *testing.B) {
	pl := NewPostingsList()
	for docID := uint32(0); docID < 10000; docID++ {
		pl.Add(docID, 0)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		it := pl.Iterator()
		for it.Next() {
			_ = it.DocID()
		}
	}
}

func BenchmarkPostingsIteratorSeek(b *testing.B) {
	pl := NewPostingsList()
	for docID := uint32(0); docID < 10000; docID++ {
		pl.Add(docID*2, 0)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		it := pl.Iterator()
		for target := uint32(0); target < 20000; target += 100 {
			it.Seek(target)
		}
	}
}

