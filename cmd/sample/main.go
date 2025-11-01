package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Jumpaku/goxe/cmd/sample/dir"
	"github.com/samber/lo"
)

func main() {
	sepFlag := flag.String("F", "", "field separator (default: runs of whitespace)")
	printNR := flag.Bool("n", false, "print record number (NR)")
	printNF := flag.Bool("N", false, "print number of fields (NF)")
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		if err := processFile(os.Stdin, *sepFlag, *printNR, *printNF); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	for _, fname := range files {
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open %s: %v\n", fname, err)
			continue
		}
		if err := processFile(f, *sepFlag, *printNR, *printNF); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", fname, err)
		}
		f.Close()
	}

	dir.SampleFunc()
	lo.Async(func() int { return 1 })
	sample()
}

func processFile(r io.Reader, sep string, showNR, showNF bool) error {
	sc := bufio.NewScanner(r)
	nr := 0
	for sc.Scan() {
		nr++
		line := sc.Text()
		var fields []string
		if sep == "" {
			fields = strings.Fields(line)
		} else {
			fields = splitBySep(line, sep)
		}
		var b strings.Builder
		if showNR {
			b.WriteString(fmt.Sprintf("%d ", nr))
		}
		if showNF {
			b.WriteString(fmt.Sprintf("%d ", len(fields)))
		}
		if b.Len() > 0 {
			b.WriteString(":\t")
		}
		if len(fields) == 0 {
			b.WriteString(line)
		} else {
			b.WriteString(strings.Join(fields, " "))
		}
		fmt.Println(b.String())
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}

func splitBySep(s, sep string) []string {
	if sep == "\\t" {
		sep = "\t"
	}
	if sep == "" {
		return strings.Fields(s)
	}
	// simple split; keep empty fields
	return strings.Split(s, sep)
}
