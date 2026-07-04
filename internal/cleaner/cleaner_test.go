package cleaner

import (
	"reflect"
	"testing"
)

func table(header []string, rows ...[]string) *Table {
	return &Table{Header: header, Rows: rows}
}

func TestTrimSpace(t *testing.T) {
	tb := table(
		[]string{" id ", "name\t"},
		[]string{" 1", "  alice  "},
		[]string{"2 ", "bob"},
	)
	TrimSpace(tb)

	wantHeader := []string{"id", "name"}
	wantRows := [][]string{{"1", "alice"}, {"2", "bob"}}
	if !reflect.DeepEqual(tb.Header, wantHeader) {
		t.Errorf("header = %q, want %q", tb.Header, wantHeader)
	}
	if !reflect.DeepEqual(tb.Rows, wantRows) {
		t.Errorf("rows = %q, want %q", tb.Rows, wantRows)
	}
}

func TestDropEmptyRows(t *testing.T) {
	tb := table(
		[]string{"a", "b"},
		[]string{"", ""},
		[]string{"1", "2"},
		[]string{"", ""},
		[]string{"", "x"},
	)
	DropEmptyRows(tb)

	want := [][]string{{"1", "2"}, {"", "x"}}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestDropEmptyRowsAllEmpty(t *testing.T) {
	tb := table([]string{"a"}, []string{""}, []string{""})
	DropEmptyRows(tb)
	if len(tb.Rows) != 0 {
		t.Errorf("rows = %q, want none", tb.Rows)
	}
}

func TestDedupRows(t *testing.T) {
	tb := table(
		[]string{"a", "b"},
		[]string{"1", "2"},
		[]string{"3", "4"},
		[]string{"1", "2"},
		[]string{"1", "2"},
	)
	DedupRows(tb)

	want := [][]string{{"1", "2"}, {"3", "4"}}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestDedupRowsKeepsFirstOccurrence(t *testing.T) {
	tb := table(
		[]string{"a"},
		[]string{"first"},
		[]string{"second"},
		[]string{"first"},
	)
	DedupRows(tb)

	want := [][]string{{"first"}, {"second"}}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestNormalizeHeaders(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Name", "name"},
		{"First Name", "first_name"},
		{"created-at", "created_at"},
		{"  Total   Amount  ", "total_amount"},
		{"already_snake", "already_snake"},
		{"Mixed-Case Header", "mixed_case_header"},
	}
	for _, c := range cases {
		tb := table([]string{c.in})
		NormalizeHeaders(tb)
		if got := tb.Header[0]; got != c.want {
			t.Errorf("NormalizeHeaders(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeNulls(t *testing.T) {
	tb := table(
		[]string{"a", "b", "c"},
		[]string{"NA", "n/a", "keep"},
		[]string{"NULL", "None", "-"},
		[]string{"nil", " NA ", "0"},
	)
	NormalizeNulls(tb)

	want := [][]string{
		{"", "", "keep"},
		{"", "", ""},
		{"", "", "0"},
	}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestApplyRunsRulesInOrder(t *testing.T) {
	tb := table(
		[]string{" ID ", "Name"},
		[]string{" 1 ", "alice"},
		[]string{"1", "alice"},
		[]string{"", ""},
	)
	// Trim first so the duplicate and empty rows are detected post-trim.
	Apply(tb, TrimSpace, DropEmptyRows, DedupRows, NormalizeHeaders)

	wantHeader := []string{"id", "name"}
	wantRows := [][]string{{"1", "alice"}}
	if !reflect.DeepEqual(tb.Header, wantHeader) {
		t.Errorf("header = %q, want %q", tb.Header, wantHeader)
	}
	if !reflect.DeepEqual(tb.Rows, wantRows) {
		t.Errorf("rows = %q, want %q", tb.Rows, wantRows)
	}
}
