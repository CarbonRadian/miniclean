package format

import (
	"reflect"
	"strings"
	"testing"

	"github.com/CarbonRadian/miniclean/internal/cleaner"
)

func TestReadJSON(t *testing.T) {
	in := `[
		{"id": 1, "name": "alice", "active": true},
		{"id": 2, "name": "bob", "active": false}
	]`
	tb, err := ReadJSON(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}

	wantHeader := []string{"id", "name", "active"}
	wantRows := [][]string{{"1", "alice", "true"}, {"2", "bob", "false"}}
	if !reflect.DeepEqual(tb.Header, wantHeader) {
		t.Errorf("header = %q, want %q", tb.Header, wantHeader)
	}
	if !reflect.DeepEqual(tb.Rows, wantRows) {
		t.Errorf("rows = %q, want %q", tb.Rows, wantRows)
	}
}

func TestReadJSONHandlesMissingAndNewKeys(t *testing.T) {
	in := `[
		{"a": "1"},
		{"a": "2", "b": "x"},
		{"b": "y"}
	]`
	tb, err := ReadJSON(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}

	wantHeader := []string{"a", "b"}
	wantRows := [][]string{{"1", ""}, {"2", "x"}, {"", "y"}}
	if !reflect.DeepEqual(tb.Header, wantHeader) {
		t.Errorf("header = %q, want %q", tb.Header, wantHeader)
	}
	if !reflect.DeepEqual(tb.Rows, wantRows) {
		t.Errorf("rows = %q, want %q", tb.Rows, wantRows)
	}
}

func TestReadJSONNullBecomesEmpty(t *testing.T) {
	in := `[{"a": null, "b": "ok"}]`
	tb, err := ReadJSON(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	want := [][]string{{"", "ok"}}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestReadJSONPreservesNumberFormat(t *testing.T) {
	in := `[{"n": 1.50, "big": 12345678901234567890}]`
	tb, err := ReadJSON(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	want := [][]string{{"1.50", "12345678901234567890"}}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestReadJSONRejectsNestedValues(t *testing.T) {
	cases := []string{
		`[{"a": {"nested": 1}}]`,
		`[{"a": [1, 2]}]`,
	}
	for _, in := range cases {
		if _, err := ReadJSON(strings.NewReader(in)); err == nil {
			t.Errorf("ReadJSON(%s): expected error, got nil", in)
		}
	}
}

func TestReadJSONRejectsNonArray(t *testing.T) {
	if _, err := ReadJSON(strings.NewReader(`{"a": 1}`)); err == nil {
		t.Fatal("ReadJSON: expected error for top-level object, got nil")
	}
}

func TestWriteJSONPreservesKeyOrder(t *testing.T) {
	tb := &cleaner.Table{
		Header: []string{"zeta", "alpha"},
		Rows:   [][]string{{"1", "2"}},
	}
	var sb strings.Builder
	if err := WriteJSON(&sb, tb); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	out := sb.String()
	if strings.Index(out, `"zeta"`) > strings.Index(out, `"alpha"`) {
		t.Errorf("keys not in header order:\n%s", out)
	}
}

func TestWriteJSONEmptyTable(t *testing.T) {
	tb := &cleaner.Table{Header: []string{"a"}}
	var sb strings.Builder
	if err := WriteJSON(&sb, tb); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if got := sb.String(); got != "[]\n" {
		t.Errorf("output = %q, want %q", got, "[]\n")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	tb := &cleaner.Table{
		Header: []string{"id", "note"},
		Rows:   [][]string{{"1", `has "quotes" and, commas`}, {"2", ""}},
	}
	var sb strings.Builder
	if err := WriteJSON(&sb, tb); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	got, err := ReadJSON(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if !reflect.DeepEqual(got, tb) {
		t.Errorf("round trip = %+v, want %+v", got, tb)
	}
}
