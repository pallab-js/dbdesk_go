package sqlite

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"dbdesk/internal/model"
)

func (d *DB) InsertRow(ctx context.Context, req model.RowInsertRequest) (model.MutationResult, error) {
	if d.readOnly {
		return model.MutationResult{}, fmt.Errorf("this connection is read-only")
	}
	t, ok := d.table(req.Table)
	if !ok {
		return model.MutationResult{}, fmt.Errorf("unknown table %q", req.Table)
	}
	if !t.Editable {
		return model.MutationResult{}, fmt.Errorf("%s is read-only", t.Name)
	}

	var cols, marks []string
	args := make([]any, 0, len(req.Values))
	seen := map[string]bool{}
	for _, kv := range req.Values {
		col, ok := columnByName(t, kv.Name)
		if !ok {
			return model.MutationResult{}, fmt.Errorf("unknown column %q", kv.Name)
		}
		if col.Hidden != 0 {
			return model.MutationResult{}, fmt.Errorf("%s is generated or hidden", col.Name)
		}
		if seen[col.Name] {
			return model.MutationResult{}, fmt.Errorf("duplicate column %q", col.Name)
		}
		seen[col.Name] = true
		v, err := coerceValue(col, kv.Value)
		if err != nil {
			return model.MutationResult{}, err
		}
		cols = append(cols, quoteIdent(col.Name))
		marks = append(marks, "?")
		args = append(args, v)
	}

	query := "INSERT INTO " + quoteIdent(t.Name) + " DEFAULT VALUES"
	if len(cols) > 0 {
		query = "INSERT INTO " + quoteIdent(t.Name) + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(marks, ", ") + ")"
	}
	res, err := d.sql.ExecContext(ctx, query, args...)
	if err != nil {
		return model.MutationResult{}, cleanError(err)
	}
	d.InvalidateCounts()
	out := model.MutationResult{}
	out.Affected, _ = res.RowsAffected()
	out.LastInsertID, _ = res.LastInsertId()
	return out, nil
}

func (d *DB) UpdateRow(ctx context.Context, req model.RowUpdateRequest) (model.MutationResult, error) {
	if d.readOnly {
		return model.MutationResult{}, fmt.Errorf("this connection is read-only")
	}
	t, ok := d.table(req.Table)
	if !ok {
		return model.MutationResult{}, fmt.Errorf("unknown table %q", req.Table)
	}
	if !t.Editable {
		return model.MutationResult{}, fmt.Errorf("%s is read-only", t.Name)
	}
	if len(req.Values) == 0 {
		return model.MutationResult{}, fmt.Errorf("no changed values to save")
	}

	args := make([]any, 0, len(req.Values)+len(req.Keys))
	var sets []string
	seen := map[string]bool{}
	for _, kv := range req.Values {
		col, ok := columnByName(t, kv.Name)
		if !ok {
			return model.MutationResult{}, fmt.Errorf("unknown column %q", kv.Name)
		}
		if col.Hidden != 0 {
			return model.MutationResult{}, fmt.Errorf("%s is generated or hidden", col.Name)
		}
		if seen[col.Name] {
			return model.MutationResult{}, fmt.Errorf("duplicate column %q", col.Name)
		}
		seen[col.Name] = true
		v, err := coerceValue(col, kv.Value)
		if err != nil {
			return model.MutationResult{}, err
		}
		sets = append(sets, quoteIdent(col.Name)+" = ?")
		args = append(args, v)
	}

	where, whereArgs, err := d.keyWhere(t, req.Keys)
	if err != nil {
		return model.MutationResult{}, err
	}
	args = append(args, whereArgs...)
	query := "UPDATE " + quoteIdent(t.Name) + " SET " + strings.Join(sets, ", ") + " WHERE " + where
	res, err := d.sql.ExecContext(ctx, query, args...)
	if err != nil {
		return model.MutationResult{}, cleanError(err)
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		return model.MutationResult{}, fmt.Errorf("expected to update 1 row, updated %d", affected)
	}
	d.InvalidateCounts()
	return model.MutationResult{Affected: affected}, nil
}

func (d *DB) DeleteRow(ctx context.Context, req model.RowDeleteRequest) (model.MutationResult, error) {
	if d.readOnly {
		return model.MutationResult{}, fmt.Errorf("this connection is read-only")
	}
	t, ok := d.table(req.Table)
	if !ok {
		return model.MutationResult{}, fmt.Errorf("unknown table %q", req.Table)
	}
	if !t.Editable {
		return model.MutationResult{}, fmt.Errorf("%s is read-only", t.Name)
	}
	where, args, err := d.keyWhere(t, req.Keys)
	if err != nil {
		return model.MutationResult{}, err
	}
	res, err := d.sql.ExecContext(ctx, "DELETE FROM "+quoteIdent(t.Name)+" WHERE "+where, args...)
	if err != nil {
		return model.MutationResult{}, cleanError(err)
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		return model.MutationResult{}, fmt.Errorf("expected to delete 1 row, deleted %d", affected)
	}
	d.InvalidateCounts()
	return model.MutationResult{Affected: affected}, nil
}

func (d *DB) keyWhere(t model.Table, keys []model.KeyValue) (string, []any, error) {
	required := pkColumns(t)
	if t.RowKey != "" {
		required = []string{t.RowKey}
	}
	if len(required) == 0 {
		return "", nil, fmt.Errorf("%s has no primary key or rowid", t.Name)
	}
	if len(keys) != len(required) {
		return "", nil, fmt.Errorf("row key must include %s", strings.Join(required, ", "))
	}
	byName := make(map[string]model.KeyValue, len(keys))
	for _, kv := range keys {
		byName[kv.Name] = kv
	}
	var parts []string
	args := make([]any, 0, len(required))
	for _, name := range required {
		kv, ok := byName[name]
		if !ok {
			return "", nil, fmt.Errorf("missing row key %q", name)
		}
		col, isCol := columnByName(t, name)
		var (
			v   any
			err error
		)
		if isCol {
			v, err = coerceValue(col, kv.Value)
		} else {
			v, err = coerceRowID(kv.Value)
		}
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, quoteIdent(name)+" = ?")
		args = append(args, v)
	}
	return strings.Join(parts, " AND "), args, nil
}

func coerceValue(col model.Column, v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	decl := strings.ToUpper(col.DeclType)
	isInt := strings.Contains(decl, "INT")
	isReal := strings.Contains(decl, "REAL") || strings.Contains(decl, "FLOA") || strings.Contains(decl, "DOUB")
	isBlob := strings.Contains(decl, "BLOB")

	switch x := v.(type) {
	case float64:
		if isInt && x == float64(int64(x)) {
			return int64(x), nil
		}
		if isInt || isReal {
			return x, nil
		}
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	case int64:
		return x, nil
	case bool:
		if x {
			return int64(1), nil
		}
		return int64(0), nil
	case string:
		if isBlob {
			return nil, fmt.Errorf("%s is a BLOB column and cannot be edited here", col.Name)
		}
		if isInt {
			n, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%s must be an integer or NULL", col.Name)
			}
			return n, nil
		}
		if isReal {
			f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
			if err != nil {
				return nil, fmt.Errorf("%s must be a number or NULL", col.Name)
			}
			return f, nil
		}
		return x, nil
	default:
		return fmt.Sprint(x), nil
	}
}

func coerceRowID(v any) (any, error) {
	switch x := v.(type) {
	case float64:
		if x != float64(int64(x)) {
			return nil, fmt.Errorf("invalid rowid")
		}
		return int64(x), nil
	case int64:
		return x, nil
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid rowid")
		}
		return n, nil
	default:
		return v, nil
	}
}
