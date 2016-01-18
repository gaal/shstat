/*
uni assists in identifying and enumerating Unicode characters.

  # Look up a code point. Note this is always in hexadecimal.
  $ uni 1d11e

  # Look up character by value.
  $ uni λ
  $ echo λ | uni

  # Find characters by name.
  $ uni tagalog
  $ uni camel
  $ uni math lamda

Credit to Larry Wall for the idea and original implementation.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/gaal/shstat/ucd"
)

func lookup(r rune) {
	if rec := ucd.Lookup(r); rec != nil {
		fmt.Printf("U+%04X [%c] %s\n", r, r, rec.Name())
	}
}

func scan(re *regexp.Regexp) {
	s := &ucd.Scanner{}
	for ; !s.Done(); s.Next() {
		rec := s.Record()
		r := rec.Rune()
		nam := rec.Name()
		if re.Match(nam) {
			fmt.Printf("U+%04X [%c] %s\n", r, r, nam)
		}
	}
}

func main() {
	flag.Parse()
	nargs := len(flag.Args())

	if nargs == 1 {
		if r, err := strconv.ParseInt(flag.Arg(0), 16, 32); err == nil {
			lookup(rune(r))
			os.Exit(0)
		}
	}
	switch nargs {
	case 0:
		// Read one character from stdin.
		rd := bufio.NewReader(os.Stdin)
		r, _, err := rd.ReadRune()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		lookup(r)
	case 1:
		r, size := utf8.DecodeRuneInString(flag.Arg(0))
		if len(flag.Arg(0)) == size {
			lookup(r)
			os.Exit(0)
		}
		fallthrough
	default:
		exp := "(?i)"
		for i, v := range flag.Args() {
			if i > 0 {
				exp += ".*"
			}
			exp += regexp.QuoteMeta(v)
		}
		re := regexp.MustCompile(exp)
		scan(re)
	}
}
