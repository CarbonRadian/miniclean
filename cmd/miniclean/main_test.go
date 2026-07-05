package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args []string, stdin string) string {
	t.Helper()
	var out strings.Builder
	if err := run(args, strings.NewReader(stdin), &out); err != nil {
		t.Fatalf("run(%v): %v", args, err)
	}
	return out.String()
}

func TestDefaultAppliesAllRules(t *testing.T) {
	in := " ID , Full Name \n1, alice \n1,alice\nNA,\n,\n"
	got := runCLI(t, nil, in)

	want := "id,full_name\n1,alice\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestSelectedRulesOnly(t *testing.T) {
	// Only trim: duplicates and empties must survive.
	in := "a,b\n 1 ,2\n1,2\n,\n"
	got := runCLI(t, []string{"-trim"}, in)

	want := "a,b\n1,2\n1,2\n,\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestCSVToJSONConversion(t *testing.T) {
	in := "id,name\n1,alice\n"
	got := runCLI(t, []string{"-to", "json"}, in)

	if !strings.Contains(got, `"id": "1"`) || !strings.Contains(got, `"name": "alice"`) {
		t.Errorf("unexpected json output:\n%s", got)
	}
}

func TestJSONToCSVConversion(t *testing.T) {
	in := `[{"id": 1, "name": " bob "}]`
	got := runCLI(t, []string{"-from", "json", "-to", "csv"}, in)

	want := "id,name\n1,bob\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestFileInputAndOutputWithFormatInference(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.csv")
	outPath := filepath.Join(dir, "out.json")
	if err := os.WriteFile(inPath, []byte("id,name\n1, alice \n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runCLI(t, []string{"-o", outPath, inPath}, "")

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"name": "alice"`) {
		t.Errorf("unexpected output file contents:\n%s", data)
	}
}

func TestUnsupportedFormat(t *testing.T) {
	var out strings.Builder
	err := run([]string{"-from", "xml"}, strings.NewReader(""), &out)
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("err = %v, want unsupported format error", err)
	}
}

func TestTooManyArgs(t *testing.T) {
	var out strings.Builder
	err := run([]string{"a.csv", "b.csv"}, strings.NewReader(""), &out)
	if err == nil || !strings.Contains(err.Error(), "at most one input file") {
		t.Errorf("err = %v, want too-many-args error", err)
	}
}

func TestVersionFlag(t *testing.T) {
	got := runCLI(t, []string{"-version"}, "")
	if !strings.HasPrefix(got, "miniclean ") {
		t.Errorf("output = %q, want version string", got)
	}
}
