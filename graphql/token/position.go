// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package token

import (
	"fmt"
	"sort"
	"sync"
)

// Position describes an arbitrary source position
// including the document, line, and column location.
// A Position is valid if the line number is > 0.
//
type Position struct {
	Filename string // filename, if any
	Offset   int    // offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (byte count)
}

// IsValid reports whether the position is valid.
func (pos *Position) IsValid() bool { return pos.Line > 0 }

// String returns a string in one of several forms:
//
//	file:line:column    valid position with document name
//	file:line           valid position with document name but no column (column == 0)
//	line:column         valid position without document name
//	line                valid position without document name and no column (column == 0)
//	file                invalid position with document name
//	-                   invalid position without document name
//
func (pos Position) String() string {
	s := pos.Filename
	if pos.IsValid() {
		if s != "" {
			s += ":"
		}
		s += fmt.Sprintf("%d", pos.Line)
		if pos.Column != 0 {
			s += fmt.Sprintf(":%d", pos.Column)
		}
	}
	if s == "" {
		s = "-"
	}
	return s
}

// Pos is a compact encoding of a source position within a document set.
// It can be converted into a Position for a more convenient, but much
// larger, representation.
//
// The Pos value for a given document is a number in the range [base, base+size],
// where base and size are specified when adding the document to the document set
// via AddDoc.
//
// To create the Pos value for a specific source offset (measured in bytes),
// first add the respective document to the current document set using DocSet.AddDoc
// and then call Doc.Pos(offset) for that document. Given a Pos value p
// for a specific document set dset, the corresponding Position value is
// obtained by calling dset.Position(p).
//
// Pos values can be compared directly with the usual comparison operators:
// If two Pos values p and q are in the same document, comparing p and q is
// equivalent to comparing the respective source document offsets. If p and q
// are in different documents, p < q is true if the document implied by p was added
// to the respective document set before the document implied by q.
//
type Pos int

// The zero value for Pos is NoPos; there is no document and line information
// associated with it, and NoPos.IsValid() is false. NoPos is always
// smaller than any other Pos value. The corresponding Position value
// for NoPos is the zero value for Position.
//
const NoPos Pos = 0

// IsValid reports whether the position is valid.
func (p Pos) IsValid() bool {
	return p != NoPos
}

// A Doc is a handle for a GraphQL document belonging to a DocSet.
// A Doc has a name, size, and line offset table.
//
type Doc struct {
	set  *DocSet
	name string // document name provided to AddDoc
	base int    // Pos value range for this file is [base, base+size]
	size int    // document size as provided to AddDoc

	// lines and infos are protected by mutex
	mutex sync.Mutex
	lines []int // lines contains the offset of the first character for each line (the first entry is always 0)
	infos []lineInfo
}

// Name returns the document name of document d as registered with AddDoc.
func (d *Doc) Name() string {
	return d.name
}

// Base returns the base offset of document d as registered with AddDoc.
func (d *Doc) Base() int {
	return d.base
}

// Size returns the size of document d as registered with AddDoc.
func (d *Doc) Size() int {
	return d.size
}

// LineCount returns the number of lines in document d.
func (d *Doc) LineCount() int {
	d.mutex.Lock()
	n := len(d.lines)
	d.mutex.Unlock()
	return n
}

// AddLine adds the line offset for a new line.
// The line offset must be larger than the offset for the previous line
// and smaller than the file size; otherwise the line offset is ignored.
//
func (d *Doc) AddLine(offset int) {
	d.mutex.Lock()
	if i := len(d.lines); (i == 0 || d.lines[i-1] < offset) && offset < d.size {
		d.lines = append(d.lines, offset)
	}
	d.mutex.Unlock()
}

// MergeLine merges a line with the following line. It is akin to replacing
// the newline character at the end of the line with a space (to not change the
// remaining offsets). To obtain the line number, consult e.g. Position.Line.
// MergeLine will panic if given an invalid line number.
//
func (d *Doc) MergeLine(line int) {
	if line <= 0 {
		panic("illegal line number (line numbering starts at 1)")
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if line >= len(d.lines) {
		panic("illegal line number")
	}
	// To merge the line numbered <line> with the line numbered <line+1>,
	// we need to remove the entry in lines corresponding to the line
	// numbered <line+1>. The entry in lines corresponding to the line
	// numbered <line+1> is located at index <line>, since indices in lines
	// are 0-based and line numbers are 1-based.
	copy(d.lines[line:], d.lines[line+1:])
	d.lines = d.lines[:len(d.lines)-1]
}

// SetLines sets the document offsets for a document and reports whether it succeeded.
// The line offsets are the offsets of the first character of each line;
// for instance for the content "ab\nc\n" the line offsets are {0, 3}.
// An empty document has an empty line offset table.
// Each line offset must be larger than the offset for the previous line
// and smaller than the document size; otherwise SetLines fails and returns
// false.
// Callers must not mutate the provided slice after SetLines returns.
//
func (d *Doc) SetLines(lines []int) bool {
	// verify validity of lines table
	size := d.size
	for i, offset := range lines {
		if i > 0 && offset <= lines[i-1] || size <= offset {
			return false
		}
	}

	// set lines table
	d.mutex.Lock()
	d.lines = lines
	d.mutex.Unlock()
	return true
}

// SetLinesForContent sets the line offsets for the given document content.
// It ignores position-altering //line comments.
func (d *Doc) SetLinesForContent(content []byte) {
	var lines []int
	line := 0
	for offset, b := range content {
		if line >= 0 {
			lines = append(lines, line)
		}
		line = -1
		if b == '\n' {
			line = offset + 1
		}
	}

	// set lines table
	d.mutex.Lock()
	d.lines = lines
	d.mutex.Unlock()
}

// A lineInfo object describes alternative file, line, and column
// number information for a given file offset.
type lineInfo struct {
	// fields are exported to make them accessible to gob
	Offset       int
	Filename     string
	Line, Column int
}

// AddLineInfo is like AddLineColumnInfo with a column = 1 argument.
// It is here for backward-compatibility for code prior to Go 1.11.
//
func (d *Doc) AddLineInfo(offset int, filename string, line int) {
	d.AddLineColumnInfo(offset, filename, line, 1)
}

// AddLineColumnInfo adds alternative document, line, and column number
// information for a given document offset. The offset must be larger
// than the offset for the previously added alternative line info
// and smaller than the document size; otherwise the information is
// ignored.
//
// AddLineColumnInfo is typically used to register alternative position
// information for line directives such as //line filename:line:column.
//
func (d *Doc) AddLineColumnInfo(offset int, filename string, line, column int) {
	d.mutex.Lock()
	if i := len(d.infos); i == 0 || d.infos[i-1].Offset < offset && offset < d.size {
		d.infos = append(d.infos, lineInfo{offset, filename, line, column})
	}
	d.mutex.Unlock()
}

// Pos returns the Pos value for the given document offset;
// the offset must be <= d.Size().
// d.Pos(d.Offset(p)) == p.
//
func (d *Doc) Pos(offset int) Pos {
	if offset > d.size {
		panic("illegal file offset")
	}
	return Pos(d.base + offset)
}

// Offset returns the offset for the given document position p;
// p must be a valid Pos value in that document.
// d.Offset(d.Pos(offset)) == offset.
//
func (d *Doc) Offset(p Pos) int {
	if int(p) < d.base || int(p) > d.base+d.size {
		panic("illegal Pos value")
	}
	return int(p) - d.base
}

// Line returns the line number for the given document position p;
// p must be a Pos value in that document or NoPos.
//
func (d *Doc) Line(p Pos) int {
	return d.Position(p).Line
}

func searchLineInfos(a []lineInfo, x int) int {
	return sort.Search(len(a), func(i int) bool { return a[i].Offset > x }) - 1
}

// unpack returns the name and line and column number for a document offset.
// If adjusted is set, unpack will return the name and line information
// possibly adjusted by //line comments; otherwise those comments are ignored.
//
func (d *Doc) unpack(offset int, adjusted bool) (filename string, line, column int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	filename = d.name
	if i := searchInts(d.lines, offset); i >= 0 {
		line, column = i+1, offset-d.lines[i]+1
	}
	if adjusted && len(d.infos) > 0 {
		// few files have extra line infos
		if i := searchLineInfos(d.infos, offset); i >= 0 {
			alt := &d.infos[i]
			filename = alt.Filename
			if i := searchInts(d.lines, alt.Offset); i >= 0 {
				// i+1 is the line at which the alternative position was recorded
				d := line - (i + 1) // line distance from alternative position base
				line = alt.Line + d
				if alt.Column == 0 {
					// alternative column is unknown => relative column is unknown
					// (the current specification for line directives requires
					// this to apply until the next PosBase/line directive,
					// not just until the new newline)
					column = 0
				} else if d == 0 {
					// the alternative position base is on the current line
					// => column is relative to alternative column
					column = alt.Column + (offset - alt.Offset)
				}
			}
		}
	}
	return
}

func (d *Doc) position(p Pos, adjusted bool) (pos Position) {
	offset := int(p) - d.base
	pos.Offset = offset
	pos.Filename, pos.Line, pos.Column = d.unpack(offset, adjusted)
	return
}

// PositionFor returns the Position value for the given document position p.
// If adjusted is set, the position may be adjusted by position-altering
// //line comments; otherwise those comments are ignored.
// p must be a Pos value in d or NoPos.
//
func (d *Doc) PositionFor(p Pos, adjusted bool) (pos Position) {
	if p != NoPos {
		if int(p) < d.base || int(p) > d.base+d.size {
			panic("illegal Pos value")
		}
		pos = d.position(p, adjusted)
	}
	return
}

// Position returns the Position value for the given document position p.
// Calling d.Position(p) is equivalent to calling d.PositionFor(p, true).
//
func (d *Doc) Position(p Pos) (pos Position) {
	return d.PositionFor(p, true)
}

// A DocSet represents a set of GraphQL documents.
// Methods of file sets are synchronized; multiple goroutines
// may invoke them concurrently.
//
type DocSet struct {
	mutex sync.RWMutex // protects the document set
	base  int          // base offset for the next document
	docs  []*Doc       // list of documents in the order added to the set
	last  *Doc         // cache of last document looked up
}

// NewFileSet creates a new file set.
func NewDocSet() *DocSet {
	return &DocSet{
		base: 1, // 0 == NoPos
	}
}

// Base returns the minimum base offset that must be provided to
// AddDoc when adding the next document.
//
func (s *DocSet) Base() int {
	s.mutex.RLock()
	b := s.base
	s.mutex.RUnlock()
	return b
}

// AddDoc adds a new document with a given name, base offset, and document size
// to the document set s and returns the document. Multiple documents may have the same
// name. The base offset must not be smaller than the DocSet's Base(), and
// size must not be negative. As a special case, if a negative base is provided,
// the current value of the DocSet's Base() is used instead.
//
// Adding the document will set the document set's Base() value to base + size + 1
// as the minimum base value for the next document. The following relationship
// exists between a Pos value p for a given document offset offs:
//
//	int(p) = base + offs
//
// with offs in the range [0, size] and thus p in the range [base, base+size].
// For convenience, Doc.Pos may be used to create document-specific position
// values from a document offset.
//
func (s *DocSet) AddDoc(name string, base, size int) *Doc {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if base < 0 {
		base = s.base
	}
	if base < s.base || size < 0 {
		panic("illegal base or size")
	}
	// base >= s.base && size >= 0
	f := &Doc{set: s, name: name, base: base, size: size, lines: []int{0}}
	base += size + 1 // +1 because EOF also has a position
	if base < 0 {
		panic("token.Pos offset overflow (> 2G of source code in file set)")
	}
	// add the file to the file set
	s.base = base
	s.docs = append(s.docs, f)
	s.last = f
	return f
}

// Iterate calls f for the documents in the document set in the order they were added
// until f returns false.
//
func (s *DocSet) Iterate(f func(*Doc) bool) {
	for i := 0; ; i++ {
		var doc *Doc
		s.mutex.RLock()
		if i < len(s.docs) {
			doc = s.docs[i]
		}
		s.mutex.RUnlock()
		if doc == nil || !f(doc) {
			break
		}
	}
}

func searchDocs(a []*Doc, x int) int {
	return sort.Search(len(a), func(i int) bool { return a[i].base > x }) - 1
}

func (s *DocSet) doc(p Pos) *Doc {
	s.mutex.RLock()
	// common case: p is in last document
	if f := s.last; f != nil && f.base <= int(p) && int(p) <= f.base+f.size {
		s.mutex.RUnlock()
		return f
	}
	// p is not in last document - search all documents
	if i := searchDocs(s.docs, int(p)); i >= 0 {
		d := s.docs[i]
		// f.base <= int(p) by definition of searchDocs
		if int(p) <= d.base+d.size {
			s.mutex.RUnlock()
			s.mutex.Lock()
			s.last = d // race is ok - s.last is only a cache
			s.mutex.Unlock()
			return d
		}
	}
	s.mutex.RUnlock()
	return nil
}

// Doc returns the document that contains the position p.
// If no such document is found (for instance for p == NoPos),
// the result is nil.
//
func (s *DocSet) Doc(p Pos) (d *Doc) {
	if p != NoPos {
		d = s.doc(p)
	}
	return
}

// PositionFor converts a Pos p in the docset into a Position value.
// If adjusted is set, the position may be adjusted by position-altering
// //line comments; otherwise those comments are ignored.
// p must be a Pos value in s or NoPos.
//
func (s *DocSet) PositionFor(p Pos, adjusted bool) (pos Position) {
	if p != NoPos {
		if d := s.doc(p); d != nil {
			return d.position(p, adjusted)
		}
	}
	return
}

// Position converts a Pos p in the docset into a Position value.
// Calling s.Position(p) is equivalent to calling s.PositionFor(p, true).
//
func (s *DocSet) Position(p Pos) (pos Position) {
	return s.PositionFor(p, true)
}

func searchInts(a []int, x int) int {
	// This function body is a manually inlined version of:
	//
	//   return sort.Search(len(a), func(i int) bool { return a[i] > x }) - 1
	//
	// With better compiler optimizations, this may not be needed in the
	// future, but at the moment this change improves the go/printer
	// benchmark performance by ~30%. This has a direct impact on the
	// speed of gofmt and thus seems worthwhile (2011-04-29).
	// TODO: Remove this when compilers have caught up.
	i, j := 0, len(a)
	for i < j {
		h := i + (j-i)/2 // avoid overflow when computing h
		// i ≤ h < j
		if a[h] <= x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i - 1
}
