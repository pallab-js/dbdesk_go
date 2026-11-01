package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"dbdesk/internal/model"
)

// Execute runs exactly one SQL statement with a row cap and timeout. It returns
// a partial result together with the error when execution fails.
func (d *DB) Execute(ctx context.Context, sqlText string, maxRows int, timeout time.Duration) (res model.QueryResult, err error) {
	start := time.Now()
	defer func() { res.DurationMs = time.Since(start).Milliseconds() }()

	if maxRows < 1 {
		maxRows = 1000
	}
	if maxRows > 10000 {
		maxRows = 10000
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if timeout > 120*time.Second {
		timeout = 120 * time.Second
	}

	statements := SplitStatements(sqlText)
	if len(statements) == 0 {
		return res, fmt.Errorf("enter a SQL statement")
	}
	if len(statements) > 1 {
		return res, fmt.Errorf("run one statement at a time (found %d)", len(statements))
	}
	stmt := statements[0]
	keyword := FirstKeyword(stmt)
	res.Kind = strings.ToLower(keyword)
	res.Statement = stmt
	res.IsRead = IsReadStatement(keyword)

	if d.readOnly && !res.IsRead {
		return res, fmt.Errorf("this connection is read-only")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if res.IsRead || strings.Contains(strings.ToUpper(stmt), "RETURNING") {
		rows, queryErr := d.sql.QueryContext(ctx, stmt)
		if queryErr != nil {
			return res, cleanError(queryErr)
		}
		defer rows.Close()
		res.Columns, res.Rows, res.RowCount, res.Truncated, queryErr = scanRows(rows, maxRows)
		if queryErr != nil {
			return res, cleanError(queryErr)
		}
	} else {
		result, execErr := d.sql.ExecContext(ctx, stmt)
		if execErr != nil {
			return res, cleanError(execErr)
		}
		res.Affected, _ = result.RowsAffected()
		res.LastInsertID, _ = result.LastInsertId()
	}

	d.InvalidateCounts()
	d.lastMu.Lock()
	copied := res
	copied.Rows = append([][]any(nil), res.Rows...)
	d.last = &copied
	d.lastMu.Unlock()
	return res, nil
}

func scanRows(rows *sql.Rows, limit int) ([]model.ColumnMeta, [][]any, int, bool, error) {
	names, err := rows.Columns()
	if err != nil {
		return nil, nil, 0, false, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, 0, false, err
	}
	cols := make([]model.ColumnMeta, len(names))
	for i, name := range names {
		typ := ""
		if i < len(types) {
			typ = types[i].DatabaseTypeName()
		}
		cols[i] = model.ColumnMeta{Name: name, Type: typ}
	}

	out := make([][]any, 0, 64)
	truncated := false
	for rows.Next() {
		if len(out) >= limit {
			truncated = true
			break
		}
		dest := make([]any, len(names))
		ptrs := make([]any, len(names))
		for i := range dest {
			ptrs[i] = &dest[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return cols, out, len(out), truncated, err
		}
		for i, v := range dest {
			dest[i] = displayValue(v)
		}
		out = append(out, dest)
	}
	if err := rows.Err(); err != nil {
		return cols, out, len(out), truncated, err
	}
	return cols, out, len(out), truncated, nil
}

func displayValue(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case int64:
		if x > 1<<53 || x < -(1<<53) {
			return strconv.FormatInt(x, 10)
		}
		return x
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return fmt.Sprintf("%v", x)
		}
		return x
	case []byte:
		return fmt.Sprintf("<blob %d bytes>", len(x))
	case string:
		if len(x) > maxCellChars {
			return x[:maxCellChars] + "… (truncated)"
		}
		return x
	default:
		return fmt.Sprint(x)
	}
}

func cleanError(err error) error {
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "SQL logic error: ")
	return fmt.Errorf("%s", msg)
}
