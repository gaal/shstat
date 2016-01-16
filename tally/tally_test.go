package main

import (
	"bytes"
	"regexp"
	"testing"
)

func TestFld(t *testing.T) {
	in := `1 10
2 20
4 30
8 40
16 50
32 60
64 70
128 80
256 90`
	want := "511 450 450 511\n"
	spec := []int{1, 2, -1, -2}
	re := regexp.MustCompile(" +")
	ofs := " "
	have := &bytes.Buffer{}
	if err := tally(bytes.NewBufferString(in), have, re, ofs, spec); err != nil {
		t.Fatalf("unexpected error=%v", err)
	}
	if have.String() != want {
		t.Errorf("fld returned wrong results.\nhave=%q,\nwant=%q", have.String(), want)
	}
}
