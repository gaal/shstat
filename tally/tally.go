/*
tally computes sums of columns.

  # Shows the size of files in a directory.
  $ ls -l | grep ^- | tally 5
  59260

  # Same thing when input is already assumed to be a single column.
  $ ls -l | grep ^- | fld 5 | tally
  59260

  # Tallies two columns, printing them in argument order and omitting
  # other columns. Negative indexes are allowed and count from end.
  $ cat mytable
  moose 36 42
  elk 98 19

  $ tally 2 3 < mytable
  134 61

Currently only integer data is supported.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gaal/shstat/internal"
)

var (
	inDelim  = flag.String("ifs", `\s+`, "input field delimiter (regexp)")
	outDelim = flag.String("ofs", " ", "output field separator (string)")
	quiet    = flag.Bool("q", false, "silence warnings on bad input")
)

func tally(in io.Reader, w io.Writer, ifs *regexp.Regexp, ofs string, idx []int) error {
	sums := make([]int64, len(idx))
	p := internal.NewParter(ifs, idx)
	s := bufio.NewScanner(in)
	var nlines int
	for s.Scan() {
		var bad bool
		nlines++
		parts := p.Fields(s.Bytes())
		for i, v := range parts {
			n, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				bad = true
				continue
			}
			sums[i] += n
		}
		if bad && !*quiet {
			fmt.Fprintf(os.Stderr, "bad input: line %d\n", nlines)
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	for i, v := range sums {
		var sep string
		if i > 0 {
			sep = ofs
		}
		if _, err := fmt.Fprintf(w, "%s%d", sep, v); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, ""); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s sums input by columns (or the first input column by default).\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	ifs := regexp.MustCompile(*inDelim)
	*outDelim = strings.Replace(*outDelim, `\t`, "\t", -1) // silent magic for now, figure it out later.
	idx, err := internal.AtoiList(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(idx) == 0 {
		idx = []int{1}
	}
	if err := tally(os.Stdin, os.Stdout, ifs, *outDelim, idx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
