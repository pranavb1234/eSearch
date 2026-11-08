package index

import (
	"sort"
)

// Posting represents a term's occurrence within a specific document.
type Posting struct {
	DocID         uint32   // Unique document identifier (monotonically increasing)
	TermFrequency uint32   // Number of times term appears in this document
	Positions     []uint32 // Token position offsets within this document
}

// PostingsList maintains an ordered collection of postings for a specific term.
// Postings are strictly sorted in ascending order of DocID.
type PostingsList struct {
	postings      []Posting
	totalTermFreq uint64
}

// NewPostingsList instantiates an empty postings list.
func NewPostingsList() *PostingsList {
	return &PostingsList{
		postings: make([]Posting, 0, 8),
	}
}

// Add appends or updates an occurrence of a term in the specified docID at the given token position.
// It maintains strictly sorted DocID order.
func (pl *PostingsList) Add(docID uint32, position uint32) {
	pl.totalTermFreq++

	n := len(pl.postings)
	// Fast path 1: List is empty
	if n == 0 {
		pl.postings = append(pl.postings, Posting{
			DocID:         docID,
			TermFrequency: 1,
			Positions:     []uint32{position},
		})
		return
	}

	// Fast path 2: Appending to the current document (same docID)
	last := &pl.postings[n-1]
	if last.DocID == docID {
		last.TermFrequency++
		last.Positions = append(last.Positions, position)
		return
	}

	// Fast path 3: New document with monotonic ascending DocID
	if docID > last.DocID {
		pl.postings = append(pl.postings, Posting{
			DocID:         docID,
			TermFrequency: 1,
			Positions:     []uint32{position},
		})
		return
	}

	// Slower fallback path: Out-of-order insertion (binary search insertion point)
	idx := sort.Search(n, func(i int) bool {
		return pl.postings[i].DocID >= docID
	})

	if idx < n && pl.postings[idx].DocID == docID {
		// Existing posting found in the middle
		target := &pl.postings[idx]
		target.TermFrequency++
		// Insert position in sorted order
		posIdx := sort.Search(len(target.Positions), func(i int) bool {
			return target.Positions[i] >= position
		})
		target.Positions = append(target.Positions, 0)
		copy(target.Positions[posIdx+1:], target.Positions[posIdx:])
		target.Positions[posIdx] = position
		return
	}

	// Insert new posting at sorted index
	pl.postings = append(pl.postings, Posting{})
	copy(pl.postings[idx+1:], pl.postings[idx:])
	pl.postings[idx] = Posting{
		DocID:         docID,
		TermFrequency: 1,
		Positions:     []uint32{position},
	}
}

// Get performs binary search to find a posting by DocID.
func (pl *PostingsList) Get(docID uint32) (*Posting, bool) {
	n := len(pl.postings)
	if n == 0 {
		return nil, false
	}

	idx := sort.Search(n, func(i int) bool {
		return pl.postings[i].DocID >= docID
	})

	if idx < n && pl.postings[idx].DocID == docID {
		return &pl.postings[idx], true
	}
	return nil, false
}

// Len returns the Document Frequency (DF) — number of distinct documents containing this term.
func (pl *PostingsList) Len() int {
	return len(pl.postings)
}

// TotalTermFreq returns the Total Term Frequency (TTF) — total occurrences across all documents.
func (pl *PostingsList) TotalTermFreq() uint64 {
	return pl.totalTermFreq
}

// Postings returns the underlying slice of sorted postings.
func (pl *PostingsList) Postings() []Posting {
	return pl.postings
}

// Iterator returns a new cursor for traversing the postings list.
func (pl *PostingsList) Iterator() *PostingsIterator {
	return &PostingsIterator{
		list:  pl,
		index: -1,
	}
}

// PostingsIterator provides forward cursor traversal with fast forward seeking.
type PostingsIterator struct {
	list  *PostingsList
	index int
}

// Next advances the iterator to the next posting. Returns false when exhausted.
func (it *PostingsIterator) Next() bool {
	it.index++
	return it.index < len(it.list.postings)
}

// DocID returns the DocID of the current posting.
func (it *PostingsIterator) DocID() uint32 {
	if it.index < 0 || it.index >= len(it.list.postings) {
		return 0
	}
	return it.list.postings[it.index].DocID
}

// Posting returns the current Posting record.
func (it *PostingsIterator) Posting() Posting {
	if it.index < 0 || it.index >= len(it.list.postings) {
		return Posting{}
	}
	return it.list.postings[it.index]
}

// Seek advances the cursor to the first posting with DocID >= targetDocID.
// Returns true if a posting with DocID >= targetDocID exists, false if exhausted.
// It uses binary search from the current position forward for O(log N) jump complexity.
func (it *PostingsIterator) Seek(targetDocID uint32) bool {
	n := len(it.list.postings)
	if n == 0 {
		return false
	}

	start := it.index
	if start < 0 {
		start = 0
	}

	// Already at or past target
	if it.index >= 0 && it.index < n && it.list.postings[it.index].DocID >= targetDocID {
		return true
	}

	// Binary search in the remainder of the slice [start, n)
	remaining := it.list.postings[start:]
	idx := sort.Search(len(remaining), func(i int) bool {
		return remaining[i].DocID >= targetDocID
	})

	targetIdx := start + idx
	if targetIdx < n {
		it.index = targetIdx
		return true
	}

	it.index = n
	return false
}

// Reset rewinds the iterator to before the first element.
func (it *PostingsIterator) Reset() {
	it.index = -1
}

