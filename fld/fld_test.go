package main

import (
	"bytes"
	"regexp"
	"testing"
)

func TestFld(t *testing.T) {
	in :=
		`one_two_three_four_five
one__two__three__four__five
one_two
one

one_two_three`
	want :=
		`one five four
one five four
one  one
one  
  
one  two
`
	re := regexp.MustCompile("_+")
	ofs := " "
	have := &bytes.Buffer{}
	if err := fld(bytes.NewBufferString(in), have, re, ofs, []int{1, 5, -2}); err != nil {
		t.Fatalf("unexpected error=%v", err)
	}
	if have.String() != want {
		t.Errorf("fld returned wrong results.\nhave=%q,\nwant=%q", have.String(), want)
	}
}
