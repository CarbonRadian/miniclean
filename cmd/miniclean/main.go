// Command miniclean cleans structured data in CSV and JSON files.
//
// Usage:
//
//	miniclean [flags] [input-file]
//
// Reads the input file (or stdin), applies the selected cleaning rules,
// and writes the result to the output file (or stdout). With no rule
// flags, all rules are applied. Formats are inferred from file
// extensions and can be overridden with -from and -to, which also
// enables CSV <-> JSON conversion.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CarbonRadian/miniclean/internal/cleaner"
	"github.com/CarbonRadian/miniclean/internal/format"
)

const version = "0.1.0"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "miniclean: %v\n", err)
		os.Exit(1)
	}
}

type config struct {
	inPath  string // empty means stdin
	outPath string // empty means stdout
	from    string // "csv" or "json"
	to      string // "csv" or "json"
	rules   []cleaner.Rule
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	cfg, done, err := parseArgs(args, stdout)
	if err != nil || done {
		return err
	}

	in := stdin
	if cfg.inPath != "" {
		f, err := os.Open(cfg.inPath)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	table, err := read(in, cfg.from)
	if err != nil {
		return err
	}
	cleaner.Apply(table, cfg.rules...)

	out := stdout
	if cfg.outPath != "" {
		f, err := os.Create(cfg.outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	return write(out, table, cfg.to)
}

// parseArgs parses flags into a config. done is true when the
// invocation was fully handled by a flag such as -version.
func parseArgs(args []string, stdout io.Writer) (cfg config, done bool, err error) {
	fs := flag.NewFlagSet("miniclean", flag.ContinueOnError)
	fs.SetOutput(stdout)
	fs.Usage = func() {
		fmt.Fprintln(stdout, "usage: miniclean [flags] [input-file]")
		fmt.Fprintln(stdout, "\nCleans CSV/JSON data. With no rule flags, all rules are applied.")
		fmt.Fprintln(stdout, "\nFlags:")
		fs.PrintDefaults()
	}

	var (
		showVersion = fs.Bool("version", false, "print version and exit")
		output      = fs.String("o", "", "output `file` (default stdout)")
		from        = fs.String("from", "", "input format: csv or json (default: inferred from extension, csv on stdin)")
		to          = fs.String("to", "", "output format: csv or json (default: same as input format)")

		trim       = fs.Bool("trim", false, "trim whitespace from headers and cells")
		nulls      = fs.Bool("normalize-nulls", false, "replace null-ish values (NA, N/A, null, none, nil, -) with empty")
		dropEmpty  = fs.Bool("drop-empty", false, "drop rows whose cells are all empty")
		dedup      = fs.Bool("dedup", false, "drop exact duplicate rows, keeping the first")
		normHeader = fs.Bool("normalize-headers", false, "lowercase headers and convert to snake_case")
	)
	if err := fs.Parse(args); err != nil {
		return cfg, false, err
	}
	if *showVersion {
		fmt.Fprintln(stdout, "miniclean", version)
		return cfg, true, nil
	}
	if fs.NArg() > 1 {
		return cfg, false, fmt.Errorf("expected at most one input file, got %d", fs.NArg())
	}
	cfg.inPath = fs.Arg(0)
	cfg.outPath = *output

	if cfg.from, err = resolveFormat(*from, cfg.inPath, "csv"); err != nil {
		return cfg, false, fmt.Errorf("input: %w", err)
	}
	if cfg.to, err = resolveFormat(*to, cfg.outPath, cfg.from); err != nil {
		return cfg, false, fmt.Errorf("output: %w", err)
	}

	// Rule order matters: trim and null normalization expose rows that
	// the structural rules can then drop.
	ruleFlags := []struct {
		on   bool
		rule cleaner.Rule
	}{
		{*trim, cleaner.TrimSpace},
		{*nulls, cleaner.NormalizeNulls},
		{*dropEmpty, cleaner.DropEmptyRows},
		{*dedup, cleaner.DedupRows},
		{*normHeader, cleaner.NormalizeHeaders},
	}
	anySelected := false
	for _, rf := range ruleFlags {
		if rf.on {
			anySelected = true
			cfg.rules = append(cfg.rules, rf.rule)
		}
	}
	if !anySelected {
		for _, rf := range ruleFlags {
			cfg.rules = append(cfg.rules, rf.rule)
		}
	}
	return cfg, false, nil
}

// resolveFormat picks a format from an explicit flag value, then the
// path's extension, then the fallback.
func resolveFormat(explicit, path, fallback string) (string, error) {
	name := explicit
	if name == "" {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".csv":
			name = "csv"
		case ".json":
			name = "json"
		default:
			name = fallback
		}
	}
	switch name {
	case "csv", "json":
		return name, nil
	default:
		return "", fmt.Errorf("unsupported format %q (want csv or json)", name)
	}
}

func read(r io.Reader, formatName string) (*cleaner.Table, error) {
	if formatName == "json" {
		return format.ReadJSON(r)
	}
	return format.ReadCSV(r)
}

func write(w io.Writer, t *cleaner.Table, formatName string) error {
	if formatName == "json" {
		return format.WriteJSON(w, t)
	}
	return format.WriteCSV(w, t)
}
