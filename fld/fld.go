/*
fld selects fields to print from tabular input.

  # Prints the first two fields in input.
  fld 1 2 < myfile

  # Prints the first two fields in input, and the last.
  # If using negative indexes, you need "--" to prevent flag handling code from
  # attempting to use -1 etc. as flags.
  fld -- 1 2 -1 < myfile

  # Same thing with alternate syntax.
  fld -k 1,2,-1

Credit to Mark-Jason Dominus for the idea.
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gaal/shstat/internal"
)

var (
	inDelim  = flag.String("ifs", `\s+`, "input field delimiter (regexp)")
	outDelim = flag.String("ofs", " ", "output field separator (string)")
	keyspec  = flag.String("k", "", "field indices")
)

func fld(in io.Reader, w io.Writer, ifs *regexp.Regexp, ofs string, idx []int) error {
	p := internal.NewParter(ifs, idx)
	ofsb := []byte(ofs)

	out := bufio.NewWriter(w)
	s := bufio.NewScanner(in)
	for s.Scan() {
		parts := p.Fields(s.Bytes())
		if _, err := out.Write(bytes.Join(parts, ofsb)); err != nil {
			return err
		}
		if _, err := out.Write([]byte("\n")); err != nil {
			return err
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	if err := out.Flush(); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s prints selected input columns.\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	keys := flag.Args()
	if *keyspec != "" {
		if len(keys) > 0 {
			fmt.Fprintln(os.Stderr, "usage: fld KEY... or fld -k=KEYS, but not both")
			os.Exit(1)
		}
		keys = strings.Split(*keyspec, ",")
	}

	ifs := regexp.MustCompile(*inDelim)
	*outDelim = strings.Replace(*outDelim, `\t`, "\t", -1) // silent magic for now, figure it out later.
	idx, err := internal.AtoiList(keys)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := fld(os.Stdin, os.Stdout, ifs, *outDelim, idx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
