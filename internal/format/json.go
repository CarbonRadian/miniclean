package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/CarbonRadian/miniclean/internal/cleaner"
)

// ReadJSON parses a JSON array of flat objects into a table. Columns
// appear in first-seen key order; objects missing a key get an empty
// cell. Scalar values are stringified; null becomes the empty string.
// Nested objects and arrays are rejected.
func ReadJSON(r io.Reader) (*cleaner.Table, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	if err := expectDelim(dec, '['); err != nil {
		return nil, fmt.Errorf("parse json: expected array of objects: %w", err)
	}

	var header []string
	index := make(map[string]int)
	var rows [][]string

	for dec.More() {
		if err := expectDelim(dec, '{'); err != nil {
			return nil, fmt.Errorf("parse json: record %d: expected object: %w", len(rows)+1, err)
		}
		row := make([]string, len(header))
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return nil, fmt.Errorf("parse json: record %d: %w", len(rows)+1, err)
			}
			key := keyTok.(string)

			var val any
			if err := dec.Decode(&val); err != nil {
				return nil, fmt.Errorf("parse json: record %d, field %q: %w", len(rows)+1, key, err)
			}
			cell, err := stringify(val)
			if err != nil {
				return nil, fmt.Errorf("parse json: record %d, field %q: %w", len(rows)+1, key, err)
			}

			col, known := index[key]
			if !known {
				col = len(header)
				index[key] = col
				header = append(header, key)
				row = append(row, "")
			}
			row[col] = cell
		}
		if _, err := dec.Token(); err != nil { // consume '}'
			return nil, fmt.Errorf("parse json: record %d: %w", len(rows)+1, err)
		}
		rows = append(rows, row)
	}
	if _, err := dec.Token(); err != nil { // consume ']'
		return nil, fmt.Errorf("parse json: %w", err)
	}

	// Earlier rows may be shorter than a header that grew later.
	for i, row := range rows {
		for len(row) < len(header) {
			row = append(row, "")
		}
		rows[i] = row
	}
	return &cleaner.Table{Header: header, Rows: rows}, nil
}

func expectDelim(dec *json.Decoder, want json.Delim) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != want {
		return fmt.Errorf("got %v, want %v", tok, want)
	}
	return nil
}

func stringify(val any) (string, error) {
	switch v := val.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("nested objects and arrays are not supported")
	}
}

// WriteJSON writes the table as a pretty-printed JSON array of objects.
// Keys follow header order and all values are emitted as strings.
// (Objects are emitted by hand because encoding/json sorts map keys.)
func WriteJSON(w io.Writer, t *cleaner.Table) error {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, row := range t.Rows {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("\n  {")
		for col, key := range t.Header {
			if col > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("\n    ")
			writeJSONString(&buf, key)
			buf.WriteString(": ")
			writeJSONString(&buf, row[col])
		}
		buf.WriteString("\n  }")
	}
	if len(t.Rows) > 0 {
		buf.WriteString("\n")
	}
	buf.WriteString("]\n")

	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}

func writeJSONString(buf *bytes.Buffer, s string) {
	b, _ := json.Marshal(s) // marshaling a string cannot fail
	buf.Write(b)
}
