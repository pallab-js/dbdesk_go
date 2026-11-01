package sqlite

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"dbdesk/internal/model"
)

func (d *DB) HasExport() bool {
	d.lastMu.RLock()
	defer d.lastMu.RUnlock()
	return d.last != nil
}

// WriteLastCSV exports the result currently shown in the UI. Cells that a
// spreadsheet would treat as formulas are prefixed so exports are safe to open.
func (d *DB) WriteLastCSV(w io.Writer) error {
	d.lastMu.RLock()
	last := d.last
	d.lastMu.RUnlock()
	if last == nil {
		return fmt.Errorf("nothing to export")
	}
	cw := csv.NewWriter(w)
	header := make([]string, len(last.Columns))
	for i, c := range last.Columns {
		header[i] = c.Name
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	record := make([]string, len(last.Columns))
	for _, row := range last.Rows {
		for i := range record {
			if i < len(row) {
				record[i] = csvCell(row[i])
			} else {
				record[i] = ""
			}
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func csvCell(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return csvSafe(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return csvSafe(fmt.Sprint(x))
	}
}

func csvSafe(s string) string {
	t := strings.TrimLeft(s, " \t\r\n")
	if t != "" && strings.ContainsRune("=+-@", rune(t[0])) {
		return "'" + s
	}
	if strings.HasPrefix(s, "\t") || strings.HasPrefix(s, "\r") {
		return "'" + s
	}
	return s
}

// ExportSnapshot returns a copy for tests and callers that need the raw rows.
func (d *DB) ExportSnapshot() ([]model.ColumnMeta, [][]any, bool) {
	d.lastMu.RLock()
	defer d.lastMu.RUnlock()
	if d.last == nil {
		return nil, nil, false
	}
	rows := make([][]any, len(d.last.Rows))
	copy(rows, d.last.Rows)
	cols := append([]model.ColumnMeta(nil), d.last.Columns...)
	return cols, rows, true
}
