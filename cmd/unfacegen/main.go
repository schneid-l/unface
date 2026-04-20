package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

var version = "v0.1.0"

func usage() {
	fmt.Fprintf(os.Stderr, `unfacegen %s — generate Unfacer methods for user types.

Usage:
    unfacegen -type=T1,T2,... [-output=<file>] [-mode=dispatch|walker]
              [-strict] [-tags=<buildtags>] [-matchmode=exact|fold|insensitive]

Flags:
`, version)
	flag.PrintDefaults()
}

func main() {
	var (
		typeList  = flag.String("type", "", "comma-separated list of type names to generate Unfacer methods for (required)")
		output    = flag.String("output", "", "output file path (default: <first_type>_unface.go in current package)")
		dir       = flag.String("dir", ".", "directory of the package to inspect")
		mode      = flag.String("mode", "dispatch", "generation mode: dispatch (type-switch into Un*er methods) or walker (inline struct walker)")
		strict    = flag.Bool("strict", false, "emit unface.Strict.Unface in fallback paths instead of unface.Unface")
		tags      = flag.String("tags", "", "optional build constraint tags to prepend as //go:build <tags>")
		matchMode = flag.String("matchmode", "exact", "walker-mode match mode: exact|fold|insensitive (only exact is honored; the others are accepted for CLI symmetry)")
	)
	flag.Usage = usage
	flag.Parse()

	if strings.TrimSpace(*typeList) == "" {
		usage()
		os.Exit(2)
	}
	types := strings.Split(*typeList, ",")
	for i, t := range types {
		types[i] = strings.TrimSpace(t)
	}

	req := Request{
		Dir:       *dir,
		Types:     types,
		Mode:      *mode,
		Strict:    *strict,
		Tags:      *tags,
		MatchMode: *matchMode,
	}
	src, err := Generate(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unfacegen: %v\n", err)
		os.Exit(1)
	}
	formatted, err := format.Source(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unfacegen: format: %v\n", err)
		fmt.Fprintln(os.Stderr, string(src))
		os.Exit(1)
	}

	out := *output
	if out == "" {
		out = filepath.Join(*dir, strings.ToLower(types[0])+"_unface.go")
	}
	if err := os.WriteFile(out, formatted, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "unfacegen: write %s: %v\n", out, err)
		os.Exit(1)
	}
	fmt.Printf("unfacegen: wrote %s\n", out)
}
