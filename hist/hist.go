/*
hist computes a histogram on its input.

  # Show a histogram of words in input (uses naive tokenization)
  $ hist -words < corpus.txt
     203 of        +
     771 a         ++++
    4025 the       +++++++++++++++++++++++
         ...

  # Same, on input that happens to have a word per line. Use log scale in graph.
  $ hist -scale=log < corpus_tokenized.txt
     203 of        +++++++++++++++
     771 a         ++++++++++++++++++
    4025 the       +++++++++++++++++++++++
         ...

  # Show a histogram of the 2nd field, using the 3rd field as weight.
  $ cat mydata
  orange vest 42
  blue vest 5
  white jumpsuit 2

  $ hist -k 2 -w 3 -graph=false < mydata
       2 jumpsuit
      47 vest
         ...

  # You can set -ofs=, for CSV output, or \t for TSV.
  $ hist -k -w 3 -graph -ofs=\\t

Output order is by increasing counts. To change order, pipe through sort and
possibly use its -n and -k flags.

Currently only integer data is supported.
*/
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gaal/shstat/internal"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	inDelim  = flag.String("ifs", `\s+`, "input field delimiter (regexp)")
	outDelim = flag.String("ofs", ``, `output delimiter. {empty=auto formatting; \t=tab; other values taken literally}`)

	words = flag.Bool("words", false, "tokenize input by unicode.IsSpace. Excludes -k and -w")

	keyspec    = flag.String("k", "", "input key fields. Comma separated, or empty to use entire line")
	weightspec = flag.Int("w", 0, "weight column. Zero to use implicit weight 1 for all inputs. Negative values are allowed and count backwards from last column")

	graph = flag.Bool("graph", true, "graph output")
	scale = flag.String("scale", "linear", "graph scale {log, linear}")

	width   = flag.Int("width", 0, "terminal width (autodetect by default, fallback to 80)")
	snippet = flag.Bool("snippet", false, "snippet long keys")
)

const (
	countAvail   = 15
	defaultWidth = 80
)

var termWidth = func() int {
	if *width > 0 {
		return *width
	}
	if *outDelim != "" { // get consistent output with CSV etc.
		return defaultWidth
	}
	if w, _, err := terminal.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return defaultWidth
}

type histogrammer struct {
	keys      []int
	weightCol int
	words     bool

	termWidth int
	ifs       *regexp.Regexp
	ofs       string
	gt        gType
	snip      bool

	gscale float64
	maxVal int64
	maxKey int

	hfmt   string
	kavail int
	gavail int
}

type gType int

const (
	gNone gType = iota
	gLinear
	gLog
)

type keyCount struct {
	key string
	cnt int64

	// display fields: may be padded, snippeted etc.
	dCnt, dKey, dGraph string
}

// byCountKey sorts keyCounts lexically by counts, then keys.
type byCountKey []keyCount

func (a byCountKey) Len() int      { return len(a) }
func (a byCountKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCountKey) Less(i, j int) bool {
	if a[i].cnt < a[j].cnt {
		return true
	}
	if a[i].cnt > a[j].cnt {
		return false
	}
	return a[i].key < a[j].key
}

func abs(v int64) int64 {
	if v >= 0 {
		return v
	}
	return -v
}

// snip returns a snippet of s at most width runes long, along with a bool
// reporting whether snippeting has occurred.
func snip(s string, width int) (string, bool) {
	const snipMark = "…"
	for i, w := 0, 0; i < len(s); i += w {
		if width <= 1 {
			return s[:i] + snipMark, true
		}
		width--
		_, rw := utf8.DecodeRuneInString(s[i:])
		w = rw
	}
	return s, false
}

func (h histogrammer) gv(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Sprint(v)
	}
	c := "+"
	if v < 0 {
		c = "-"
	}
	v = math.Abs(math.Trunc(v + math.Copysign(0.5, v))) // Abs(Round(v))
	return strings.Repeat(c, int(v))
}

// hlinefmt prepares a format string for records in a histogram, as well as max
// available key and graph width, according to the given terminal and display
// options.
func hlinefmt(tw int, graph bool, ofs string) (hfmt string, kavail int, gavail int) {
	sep := strings.Replace(ofs, "%", "%%", -1)
	cfmt := func(avail int) string { return "%d" }
	kfmt := func(avail int) string { return "%s" }
	if ofs == "" { // autoformatting
		sep = " "
		cfmt = func(avail int) string { return "%" + strconv.Itoa(avail) + "d" }
		kfmt = func(avail int) string { return "%-" + strconv.Itoa(avail) + "s" }
	}
	if graph {
		kavail = tw/2 - countAvail - 1
		gavail = tw - countAvail - kavail - 3
		hfmt = strings.Join([]string{cfmt(countAvail), kfmt(kavail), "%s"}, sep)
	} else {
		kavail = tw - countAvail - 2
		hfmt = strings.Join([]string{cfmt(countAvail), kfmt(kavail)}, sep)
	}
	return
}

func (h histogrammer) hline(kc keyCount) string {
	if h.snip {
		kc.key, _ = snip(kc.key, h.kavail)
	}
	if h.gt == gNone {
		return strings.TrimRight(fmt.Sprintf(h.hfmt, kc.cnt, kc.key), " ")
	}
	var g float64
	switch h.gt {
	case gLinear:
		g = float64(kc.cnt) / h.gscale
	case gLog:
		g = math.Log2(float64(kc.cnt)) / h.gscale
	}
	return strings.TrimRight(fmt.Sprintf(h.hfmt, kc.cnt, kc.key, h.gv(g)), " ")
}

func (h *histogrammer) hist(in io.Reader) ([]keyCount, error) {
	h.hfmt, h.kavail, h.gavail = hlinefmt(h.termWidth, h.gt != gNone, h.ofs)

	key := func(line []byte) []byte { return line }
	if len(h.keys) > 0 {
		kp := internal.NewParter(h.ifs, h.keys)
		key = func(line []byte) []byte {
			parts := kp.Fields(line)
			// TODO: is there a better way to rejoin parted keys than hardcode
			// space? ofs is wrong, but ifs is a regexp.
			return bytes.Join(parts, []byte(" "))
		}
	}

	weight := func(line []byte) (int64, error) { return 1, nil }
	if h.weightCol != 0 {
		wp := internal.NewParter(h.ifs, []int{h.weightCol})
		weight = func(line []byte) (int64, error) {
			parts := wp.Fields(line)
			if len(parts) == 0 {
				return 0, errors.New("short line")
			}
			w, err := strconv.ParseInt(string(parts[0]), 10, 64)
			if err != nil {
				return 0, err
			}
			return w, nil
		}
	}

	d := make(map[string]int64)
	var nlines int
	s := bufio.NewScanner(in)
	if h.words {
		s.Split(bufio.ScanWords)
	}
	for s.Scan() {
		nlines++

		line := s.Bytes()
		w, err := weight(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %d", err, nlines)
			continue
		}
		k := string(key(line))
		d[k] += w
		if len(k) > h.maxKey {
			h.maxKey = len(k)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	var kc []keyCount
	for k, v := range d {
		kc = append(kc, keyCount{key: k, cnt: v})
	}
	sort.Sort(byCountKey(kc))
	if len(kc) > 0 {
		h.maxVal = abs(kc[0].cnt)
		if abs(kc[len(kc)-1].cnt) > h.maxVal {
			h.maxVal = abs(kc[len(kc)-1].cnt)
		}
	}
	h.gscale = float64(h.maxVal)
	if h.gt == gLog {
		h.gscale = math.Log2(float64(h.maxVal))
	}
	h.gscale /= float64(h.gavail)

	return kc, nil
}

func (h histogrammer) printHist(out io.Writer, kc []keyCount) error {
	for _, kv := range kc {
		if _, err := fmt.Fprintln(out, h.hline(kv)); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	internal.SetUsage(
		`hist computes a histogram on its input.

  # Show a histogram of words in input (uses naive tokenization)
  $ hist -words < corpus.txt
     203 of        +
     771 a         ++++
    4025 the       +++++++++++++++++++++++
         ...

  # Same, on input that happens to have a word per line. Use log scale in graph.
  $ hist -scale=log < corpus_tokenized.txt
     203 of        +++++++++++++++
     771 a         ++++++++++++++++++
    4025 the       +++++++++++++++++++++++
         ...

  # Show a histogram of the 2nd field, using the 3rd field as weight.
  $ cat mydata
  orange vest 42
  blue vest 5
  white jumpsuit 2

  $ hist -k 2 -w 3 -graph=false < mydata
       2 jumpsuit
      47 vest
         ...

  # You can set -ofs=, for CSV output, or \t for TSV.
  $ hist -k -w 3 -graph -ofs=\\t

Output order is by increasing counts. To change order, pipe through sort and
possibly use its -n and -k flags.

Currently only integer data is supported.`)
	flag.Parse()

	*outDelim = strings.Replace(*outDelim, `\t`, "\t", -1)
	*inDelim = strings.Replace(*inDelim, `\t`, "\t", -1)
	ifs := regexp.MustCompile(*inDelim)
	keysstr := strings.Split(*keyspec, ",")
	if len(keysstr) == 1 && keysstr[0] == "" {
		keysstr = nil
	}
	keys, err := internal.AtoiList(keysstr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *words && (len(keys) > 0 || *weightspec != 0) {
		fmt.Fprintln(os.Stderr, "--words cannot be used with -k or -w")
		os.Exit(1)
	}

	var gtype gType
	switch *scale {
	case "none":
		gtype = gNone
	case "log":
		gtype = gLog
	case "linear":
		gtype = gLinear
	default:
		fmt.Fprintf(os.Stderr, "bad -scale=%q\n", *scale)
		os.Exit(1)
	}
	tw := termWidth()
	if tw < 20 {
		tw = 20
	}
	h := &histogrammer{
		keys:      keys,
		weightCol: *weightspec,
		words:     *words,
		gt:        gtype,
		ifs:       ifs,
		ofs:       *outDelim,
		termWidth: tw,
		snip:      *snippet,
	}
	kc, err := h.hist(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = h.printHist(os.Stdout, kc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
