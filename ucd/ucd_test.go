package ucd

import (
	"reflect"
	"testing"
)

func TestLookup(t *testing.T) {
	var runes []rune
	for i := rune(0); i < 512; i++ {
		runes = append(runes, i)
	}
	for i := rune(10000); i < 10100; i++ {
		runes = append(runes, i)
	}
	runes = append(runes, []rune{
		0xFFFD,
		0x2FA1D,
		0xE0001,
		0x10FFFD,
	}...)

	for _, r := range runes {
		rec := Lookup(r)
		if rec.Rune() != r {
			t.Errorf("Lookup(%04X).Rune()=%04X, want=%04X", r, rec.Rune(), r)
			t.FailNow()
		}
	}
}

func TestLookupFail(t *testing.T) {
	runes := []rune{
		0x5FE,
		0x4E01, // We don't support UCD ranges.
		0x10FFFE,
		0x10FFFF,
	}

	for _, r := range runes {
		rec := Lookup(r)
		if rec.Rune() != 0 {
			t.Errorf("Lookup(%04X).Rune()=%04X, want=%04X", r, rec.Rune(), 0)
			t.FailNow()
		}
	}
}

func TestScanner(t *testing.T) {
	for _, d := range []struct {
		start rune
		want  []rune
	}{
		{0x0000, []rune{0x0000, 0x0001, 0x0002}},
		{0x10000, []rune{0x10000, 0x10001, 0x10002}},
		{0x100000, []rune{0x100000, 0x10FFFD}},
		{0x10FFFD, []rune{0x10FFFD}},
	} {
		var have []rune
		s := NewScanner(d.start)
		for i := 0; !s.Done() && i < 3; s.Next() {
			have = append(have, s.Record().Rune())
			i++
		}
		if !reflect.DeepEqual(have, d.want) {
			t.Log("Note want/have are in decimal")
			t.Errorf("Scan(%04X)=%v..., want %v...", d.start, have, d.want)
		}
	}
}
