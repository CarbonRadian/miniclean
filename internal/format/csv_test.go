package format

import (
	"reflect"
	"strings"
	"testing"

	"github.com/CarbonRadian/miniclean/internal/cleaner"
)

func TestReadCSV(t *testing.T) {
	in := "id,name,city\n1,alice,berlin\n2,bob,tokyo\n"
	tb, err := ReadCSV(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadCSV: %v", err)
	}

	wantHeader := []string{"id", "name", "city"}
	wantRows := [][]string{{"1", "alice", "berlin"}, {"2", "bob", "tokyo"}}
	if !reflect.DeepEqual(tb.Header, wantHeader) {
		t.Errorf("header = %q, want %q", tb.Header, wantHeader)
	}
	if !reflect.DeepEqual(tb.Rows, wantRows) {
		t.Errorf("rows = %q, want %q", tb.Rows, wantRows)
	}
}

func TestReadCSVPadsShortRows(t *testing.T) {
	in := "a,b,c\n1,2\n"
	tb, err := ReadCSV(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadCSV: %v", err)
	}
	want := [][]string{{"1", "2", ""}}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestReadCSVRejectsLongRows(t *testing.T) {
	in := "a,b\n1,2,3\n"
	if _, err := ReadCSV(strings.NewReader(in)); err == nil {
		t.Fatal("ReadCSV: expected error for row longer than header, got nil")
	}
}

func TestReadCSVRejectsEmptyInput(t *testing.T) {
	if _, err := ReadCSV(strings.NewReader("")); err == nil {
		t.Fatal("ReadCSV: expected error for empty input, got nil")
	}
}

func TestReadCSVQuotedFields(t *testing.T) {
	in := "id,note\n1,\"hello, world\"\n"
	tb, err := ReadCSV(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadCSV: %v", err)
	}
	want := [][]string{{"1", "hello, world"}}
	if !reflect.DeepEqual(tb.Rows, want) {
		t.Errorf("rows = %q, want %q", tb.Rows, want)
	}
}

func TestWriteCSV(t *testing.T) {
	tb := &cleaner.Table{
		Header: []string{"id", "note"},
		Rows:   [][]string{{"1", "plain"}, {"2", "with, comma"}},
	}
	var sb strings.Builder
	if err := WriteCSV(&sb, tb); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	want := "id,note\n1,plain\n2,\"with, comma\"\n"
	if sb.String() != want {
		t.Errorf("output = %q, want %q", sb.String(), want)
	}
}

func TestCSVRoundTrip(t *testing.T) {
	tb := &cleaner.Table{
		Header: []string{"a", "b"},
		Rows:   [][]string{{"x", "quoted \"stuff\""}, {"", "line\nbreak"}},
	}
	var sb strings.Builder
	if err := WriteCSV(&sb, tb); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	got, err := ReadCSV(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("ReadCSV: %v", err)
	}
	if !reflect.DeepEqual(got, tb) {
		t.Errorf("round trip = %+v, want %+v", got, tb)
	}
}
