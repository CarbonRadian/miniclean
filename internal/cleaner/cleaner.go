// Package cleaner implements the core data-cleaning engine of miniclean.
//
// Data is represented as a Table (a header plus rows of string cells),
// and cleaning is expressed as a pipeline of Rules applied in order.
package cleaner

import (
	"fmt"
	"strings"
)

// Table is an in-memory representation of tabular data.
// All cells are strings; type interpretation is left to consumers.
type Table struct {
	Header []string
	Rows   [][]string
}

// Rule transforms a table in place.
type Rule func(*Table)

// Apply runs the given rules against the table in order.
func Apply(t *Table, rules ...Rule) {
	for _, rule := range rules {
		rule(t)
	}
}

// TrimSpace removes leading and trailing whitespace from every header
// and cell value.
func TrimSpace(t *Table) {
	for i, h := range t.Header {
		t.Header[i] = strings.TrimSpace(h)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			row[i] = strings.TrimSpace(cell)
		}
	}
}

// DropEmptyRows removes rows in which every cell is empty.
func DropEmptyRows(t *Table) {
	kept := t.Rows[:0]
	for _, row := range t.Rows {
		if !rowEmpty(row) {
			kept = append(kept, row)
		}
	}
	t.Rows = kept
}

func rowEmpty(row []string) bool {
	for _, cell := range row {
		if cell != "" {
			return false
		}
	}
	return true
}

// DedupRows removes exact duplicate rows, keeping the first occurrence.
func DedupRows(t *Table) {
	seen := make(map[string]struct{}, len(t.Rows))
	kept := t.Rows[:0]
	for _, row := range t.Rows {
		key := strings.Join(row, "\x1f") // unit separator: unlikely in data
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		kept = append(kept, row)
	}
	t.Rows = kept
}

// NormalizeHeaders lowercases headers and replaces runs of spaces and
// dashes with a single underscore, producing snake_case-style names.
// If normalization makes two or more headers collide, later ones get a
// numeric suffix (_2, _3, ...) so headers stay unique.
func NormalizeHeaders(t *Table) {
	used := make(map[string]struct{}, len(t.Header))
	for i, h := range t.Header {
		base := normalizeHeader(h)
		name := base
		for n := 2; ; n++ {
			if _, taken := used[name]; !taken {
				break
			}
			name = fmt.Sprintf("%s_%d", base, n)
		}
		used[name] = struct{}{}
		t.Header[i] = name
	}
}

func normalizeHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	fields := strings.FieldsFunc(h, func(r rune) bool {
		return r == ' ' || r == '-' || r == '\t'
	})
	return strings.Join(fields, "_")
}

// nullTokens are values commonly used to mean "no data".
// Matched case-insensitively after trimming.
var nullTokens = map[string]struct{}{
	"na":   {},
	"n/a":  {},
	"null": {},
	"none": {},
	"nil":  {},
	"-":    {},
}

// NormalizeNulls replaces null-ish placeholder values (NA, N/A, null,
// none, nil, -) with the empty string.
func NormalizeNulls(t *Table) {
	for _, row := range t.Rows {
		for i, cell := range row {
			token := strings.ToLower(strings.TrimSpace(cell))
			if _, isNull := nullTokens[token]; isNull {
				row[i] = ""
			}
		}
	}
}
