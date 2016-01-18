/*
Package ucd provides access to the Unicode Character Databse.

This package is intended for access to names and other information
possibly not provided by package unicode in the standard library. It
is not optimized for speed.
*/
package ucd

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

// Record represents an entry in the UCD.
type Record []byte

// Rune returns the rune identifying this record.
func (r Record) Rune() rune {
	parts := bytes.SplitN(r, []byte(";"), 2)
	if r, err := strconv.ParseInt(string(parts[0]), 16, 32); err == nil {
		return rune(r)
	}
	return 0
}

// Name returns the Name field of a character record.
func (r Record) Name() []byte {
	parts := bytes.SplitN(r, []byte(";"), 3)
	return parts[1]
}

// If needed, the other accessors are simple enough to add.
// See http://www.unicode.org/reports/tr44/ .

// Lookup searches for r in the UCD. It returns the matching Record or nil if not found.
//
// Searches take logarithmic time and are not currently cached.
func Lookup(r rune) Record {
	rec, _ := search(r)
	return rec
}

var (
	searchRE = regexp.MustCompile(`\n(([\dA-F]+);([^;]*);.*)\n`)
)

func compare(a, b []byte) int {
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return bytes.Compare(a, b)
}

// search searches for r in the UCD.
// It returns a Record and position. (Position always points to the newline
// at the start of a record.) If not found, returns nil and -1.
func search(r rune) (Record, int) {
	key := []byte(fmt.Sprintf("%04X", r))
	n := sort.Search(maxRaw, func(i int) bool {
		wind := rawUCD[i : i+pad]
		m := searchRE.FindSubmatch(wind)
		if m == nil {
			return true
		}
		return compare(m[2], key) >= 0
	})
	wind := rawUCD[n : n+pad]
	if m := searchRE.FindSubmatchIndex(wind); m != nil {
		// With the sort.Search API, we got a guess at n, but
		// still need to compare it to our desired key.
		if bytes.Equal(rawUCD[n+m[4]:n+m[5]], key) {
			return rawUCD[n+m[2] : n+m[3]], n + m[0]
		}
	}
	return nil, -1
}

// Scanner is an iterator over the UCD. A zero Scanner is ready to use
// and scans from the first character.
type Scanner struct {
	pos  int // Always points to newline at raw record start.
	done bool
}

// NewScanner returns a Scanner that starts scanning from the given rune.
func NewScanner(start rune) *Scanner {
	rec, pos := search(start)
	if rec == nil {
		return &Scanner{done: true}
	}
	return &Scanner{pos: pos}
}

// Done reports whether scanning has ended. Once this returns true, Record
// must not be called.
func (s Scanner) Done() bool { return s.done }

// Next advances the iterator.
func (s *Scanner) Next() {
	i := bytes.Index(rawUCD[s.pos+1:], []byte("\n"))
	if i < 0 {
		s.done = true
	}
	s.pos += i + 1
	if s.pos >= maxRaw-1 {
		s.done = true
	}
}

// Record returns the current UCD record. It is a fatal error to call this
// after Done() becomes true.
func (s Scanner) Record() Record {
	if s.done {
		panic("Record called after Done()=true")
	}
	end := bytes.Index(rawUCD[s.pos+1:], []byte("\n"))
	return rawUCD[s.pos+1 : s.pos+1+end]
}
