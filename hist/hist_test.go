package main

import (
	"bytes"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func norm(s string) string {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	lines := strings.SplitAfter(s, "\n")
	for i, s := range lines {
		lines[i] = s
	}
	sort.Strings(lines)
	return strings.Join(lines, "")
}

func TestHlinefmt(t *testing.T) {
	for _, d := range []struct {
		tw    int
		graph bool
		ofs   string

		wantHfmt   string
		wantKavail int
		wantGavail int
	}{
		{tw: 40,
			wantHfmt: "%15d %-23s", wantKavail: 23},
		{tw: 40, ofs: ",",
			wantHfmt: "%d,%s", wantKavail: 23},
		{tw: 60, graph: true,
			wantHfmt: "%15d %-14s %s", wantKavail: 14, wantGavail: 28},
		{tw: 60, graph: true, ofs: ",",
			wantHfmt: "%d,%s,%s", wantKavail: 14, wantGavail: 28},
	} {
		hfmt, kavail, gavail := hlinefmt(d.tw, d.graph, d.ofs)
		if hfmt != d.wantHfmt {
			t.Errorf("hlinefmt(%d, %v, %q) hfmt=%q, want=%q", d.tw, d.graph, d.ofs, hfmt, d.wantHfmt)
		}
		if kavail != d.wantKavail {
			t.Errorf("hlinefmt(%d, %v, %q) kavail=%d, want=%d", d.tw, d.graph, d.ofs, kavail, d.wantKavail)
		}
		if gavail != d.wantGavail {
			t.Errorf("hlinefmt(%d, %v, %q) gavail=%d, want=%d", d.tw, d.graph, d.ofs, gavail, d.wantGavail)
		}
	}
}

func TestHist(t *testing.T) {
	const in = `a b 2
a c 5`
	ifs := regexp.MustCompile(" +")
	for _, d := range []struct {
		keys      []int
		weightCol int
		want      []keyCount
	}{
		{nil, 0,
			[]keyCount{kc("a b 2", 1), kc("a c 5", 1)}},
		{[]int{1, 2}, -1,
			[]keyCount{kc("a b", 2), kc("a c", 5)}},
		{[]int{1}, 0,
			[]keyCount{kc("a", 2)}},
		{[]int{1}, -1,
			[]keyCount{kc("a", 7)}},
	} {
		h := &histogrammer{
			keys:      d.keys,
			weightCol: d.weightCol,
			ifs:       ifs,
		}
		have, err := h.hist(bytes.NewBufferString(in))
		if err != nil {
			t.Fatalf("h.hist returned unexpected error=%v", err)
		}
		if !reflect.DeepEqual(have, d.want) {
			t.Errorf("hist(..., %v, %d) returned bad results.\nhave=%v\nwant=%v", d.keys, d.weightCol, have, d.want)
		}
	}
}

func kc(k string, c int64) keyCount { return keyCount{key: k, cnt: c} }

func TestWords(t *testing.T) {
	const in = `... What
conquest brings  he  home? What  tributaries 
follow ...  `
	want := []keyCount{
		kc("What", 2),
		kc("conquest", 1),
		kc("brings", 1),
		kc("he", 1),
		kc("home?", 1),
		kc("tributaries", 1),
		kc("follow", 1),
		kc("...", 2),
	}
	sort.Sort(byCountKey(want))

	h := &histogrammer{
		words: true,
	}
	have, err := h.hist(bytes.NewBufferString(in))
	if err != nil {
		t.Fatalf("h.hist returned unexpected error=%v", err)
	}

	if !reflect.DeepEqual(have, want) {
		t.Errorf("hist(words:true):\nhave=%v\nwant=%v", have, want)
	}
}

func TestGraph(t *testing.T) {
	const in = `- -10
0 0
a 1
b 10
c 100
z_long_key_that_is_snippetted 100`
	for _, d := range []struct {
		scale gType
		want  string
	}{
		{
			scale: gNone,
			want: strings.Join([]string{
				"            -10 -",
				"              0 0",
				"              1 a",
				"             10 b",
				"            100 c",
				"            100 z_long_key_that_is_sni…",
				""}, "\n"),
		},
		{
			scale: gLinear,
			want: strings.Join([]string{
				"            -10 -    --",
				"              0 0",
				"              1 a",
				"             10 b    ++",
				"            100 c    ++++++++++++++++++",
				"            100 z_l… ++++++++++++++++++",
				""}, "\n"),
		},
		{
			scale: gLog,
			want: strings.Join([]string{
				"            -10 -    NaN",
				"              0 0    -Inf",
				"              1 a",
				"             10 b    +++++++++",
				"            100 c    ++++++++++++++++++",
				"            100 z_l… ++++++++++++++++++",
				""}, "\n"),
		},
	} {
		h := &histogrammer{
			keys:      []int{1},
			weightCol: 2,
			ifs:       regexp.MustCompile(" +"),
			termWidth: 40,
			snip:      true,
			gt:        d.scale,
		}
		data, err := h.hist(bytes.NewBufferString(in))
		if err != nil {
			t.Fatalf("h.hist returned unexpected error=%v", err)
		}
		haveRaw := &bytes.Buffer{}
		if err = h.printHist(haveRaw, data); err != nil {
			t.Fatalf("h.printHist returned unexpected error=%v", err)
		}

		have := haveRaw.String()
		if have != d.want {
			t.Errorf("hist (scale=%v) returned bad results.\nhave=%q\nwant=%q", d.scale, have, d.want)
		}
	}
}

func TestTermWidth(t *testing.T) {
	w := termWidth()
	t.Logf("termwidth=%d", w)
	if w > 1<<14 {
		t.Errorf("termWidth()=%d, want < 16k", w)
	}
	if w <= 0 {
		t.Errorf("termWidth()=%d, want > 0", w)
	}
}
