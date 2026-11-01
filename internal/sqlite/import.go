package sqlite

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"dbdesk/internal/model"
)

func (d *DB) ImportData(ctx context.Context, req model.ImportRequest) (model.ImportResult, error) {
	if d.readOnly {
		return model.ImportResult{}, fmt.Errorf("cannot import into read-only connection")
	}
	t, ok := d.table(req.Table)
	if !ok {
		return model.ImportResult{}, fmt.Errorf("unknown table %q", req.Table)
	}
	if !t.Editable {
		return model.ImportResult{}, fmt.Errorf("%s is not editable", t.Name)
	}

	if req.Truncate {
		if _, err := d.sql.ExecContext(ctx, "DELETE FROM "+quoteIdent(t.Name)); err != nil {
			return model.ImportResult{}, cleanError(err)
		}
		d.InvalidateCounts()
	}

	var rows [][]string
	var err error
	if req.Format == "json" {
		rows, err = parseJSONImport(req.Data)
	} else {
		rows, err = parseCSVImport(req.Data, req.HasHeader)
	}
	if err != nil {
		return model.ImportResult{}, err
	}

	res := model.ImportResult{}
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return model.ImportResult{}, err
	}
	defer tx.Rollback()

	// Build column mapping
	colNames := make([]string, 0, len(t.Columns))
	for _, c := range t.Columns {
		if c.Hidden == 0 {
			colNames = append(colNames, c.Name)
		}
	}
	if len(req.ColumnMap) > 0 {
		mapped := make([]string, 0, len(req.ColumnMap))
		for _, km := range req.ColumnMap {
			if km.Value == nil {
				continue
			}
			if s, ok := km.Value.(string); ok && s != "" {
				mapped = append(mapped, s)
			}
		}
		if len(mapped) > 0 {
			colNames = mapped
		}
	}

	placeholders := strings.Repeat("?,", len(colNames))
	if len(placeholders) > 0 {
		placeholders = placeholders[:len(placeholders)-1]
	}
	colsSQL := make([]string, len(colNames))
	for i, c := range colNames {
		colsSQL[i] = quoteIdent(c)
	}
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO "+quoteIdent(t.Name)+" ("+strings.Join(colsSQL, ", ")+") VALUES ("+placeholders+")")
	if err != nil {
		return model.ImportResult{}, cleanError(err)
	}
	defer stmt.Close()

	for _, row := range rows {
		args := make([]any, 0, len(colNames))
		for i := 0; i < len(colNames); i++ {
			if i < len(row) {
				s := row[i]
				if s == "" {
					args = append(args, nil)
				} else {
					args = append(args, s)
				}
			} else {
				args = append(args, nil)
			}
		}
		_, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			res.Errors++
			if res.Details == "" {
				res.Details = err.Error()
			}
			continue
		}
		res.Inserted++
	}
	if err := tx.Commit(); err != nil {
		return model.ImportResult{}, err
	}
	d.InvalidateCounts()
	return res, nil
}

func parseCSVImport(data string, hasHeader bool) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	if hasHeader && len(records) > 0 {
		records = records[1:]
	}
	return records, nil
}

func parseJSONImport(data string) ([][]string, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return nil, err
	}
	switch v := raw.(type) {
	case []map[string]interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		// Get all keys in order from first row
		rows := make([][]string, 0, len(v))
		for _, obj := range v {
			row := make([]string, 0, len(obj))
			for _, val := range obj {
				row = append(row, stringifyVal(val))
			}
			rows = append(rows, row)
		}
		return rows, nil
	case [][]interface{}:
		rows := make([][]string, 0, len(v))
		for _, arr := range v {
			row := make([]string, len(arr))
			for i, val := range arr {
				row[i] = stringifyVal(val)
			}
			rows = append(rows, row)
		}
		return rows, nil
	default:
		return nil, fmt.Errorf("unsupported JSON format")
	}
}

func stringifyVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "1"
		}
		return "0"
	default:
		b, _ := json.Marshal(x)
		return string(bytes.Trim(b, `"`))
	}
}
