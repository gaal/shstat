package internal

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// AtoiList runs strconv.Atoi on every element in nums. If this fails it
// returns the first error encountered.
func AtoiList(nums []string) ([]int, error) {
	var res []int
	for _, v := range nums {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("not an integer: %q", v)
		}
		res = append(res, n)
	}
	return res, nil
}

// splitb is very similar to regexp.Split(s, -1) but returns [][]byte.
func splitb(re *regexp.Regexp, b []byte) [][]byte {
	matches := re.FindAllIndex(b, -1)
	outs := make([][]byte, 0, len(matches))

	var beg, end int
	for _, m := range matches {
		end = m[0]
		if m[1] != 0 {
			outs = append(outs, b[beg:end])
			beg = m[1]
		}
	}
	if end != len(b) {
		outs = append(outs, b[beg:])
	}
	return outs
}

// Parter parts lines into fields specified by a regexp, and returns certain
// parts of the split using 1-based indexing. Negative values are allowed and
// count from the input end.
type Parter struct {
	re  *regexp.Regexp
	idx []int
}

// NewParter creates a new Parter using ifs and the given index list.
func NewParter(ifs *regexp.Regexp, idx []int) *Parter {
	return &Parter{re: ifs, idx: append([]int(nil), idx...)}
}

// Fields returns the fields in line matching the Parter spec.
func (p Parter) Fields(line []byte) [][]byte {
	parts := splitb(p.re, line)
	// No idx means all fields (simple way to change delim / normalize its width)
	if len(p.idx) == 0 {
		return parts
	}

	var out [][]byte
	for _, v := range p.idx {
		if v < 0 {
			v = len(parts) + v
		} else if v > 0 {
			v--
		}
		if v >= 0 && v < len(parts) {
			out = append(out, parts[v])
		} else {
			out = append(out, nil)
		}
	}
	return out
}

// SetUsage updates flag.Usage with help text. It should be called before flag.Parse.
//
// TODO: our convention is to repeat the package doc string here. Is there a way to
// extract that from the binary?
func SetUsage(s string) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, s)
		flag.PrintDefaults()
	}
}
