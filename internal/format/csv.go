// Package format reads and writes cleaner.Table values in the file
// formats supported by miniclean.
package format

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/CarbonRadian/miniclean/internal/cleaner"
)

// ReadCSV parses CSV data into a table. The first record is treated as
// the header. Rows shorter than the header are padded with empty cells;
// rows longer than the header are an error.
func ReadCSV(r io.Reader) (*cleaner.Table, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // allow ragged rows; we validate below

	records, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("parse csv: empty input")
	}

	header := records[0]
	rows := make([][]string, 0, len(records)-1)
	for i, rec := range records[1:] {
		if len(rec) > len(header) {
			return nil, fmt.Errorf("parse csv: row %d has %d fields, header has %d", i+2, len(rec), len(header))
		}
		for len(rec) < len(header) {
			rec = append(rec, "")
		}
		rows = append(rows, rec)
	}
	return &cleaner.Table{Header: header, Rows: rows}, nil
}

// WriteCSV writes the table as CSV with the header as the first record.
func WriteCSV(w io.Writer, t *cleaner.Table) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(t.Header); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}
	if err := cw.WriteAll(t.Rows); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}
	return nil
}
