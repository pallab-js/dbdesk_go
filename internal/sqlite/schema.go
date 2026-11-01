package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"dbdesk/internal/model"
)

func loadSchema(ctx context.Context, db *sql.DB) (model.Schema, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name, type, sql FROM sqlite_schema
		WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%'
		ORDER BY CASE type WHEN 'table' THEN 0 ELSE 1 END, name`)
	if err != nil {
		return model.Schema{}, err
	}
	defer rows.Close()

	type entry struct {
		name, kind, ddl string
	}
	var entries []entry
	for rows.Next() {
		var name, kind string
		var ddl sql.NullString
		if err := rows.Scan(&name, &kind, &ddl); err != nil {
			return model.Schema{}, err
		}
		entries = append(entries, entry{name: name, kind: kind, ddl: ddl.String})
	}
	if err := rows.Err(); err != nil {
		return model.Schema{}, err
	}
	rows.Close()

	schema := model.Schema{Tables: []model.Table{}}
	for _, e := range entries {
		name, kind, ddl := e.name, e.kind, sql.NullString{String: e.ddl, Valid: e.ddl != ""}
		t := model.Table{Name: name, Type: kind, DDL: ddl.String, Editable: false}

		cols, err := tableColumns(ctx, db, name)
		if err != nil {
			return model.Schema{}, fmt.Errorf("columns for %s: %w", name, err)
		}
		t.Columns = cols
		t.FKs, err = tableFKs(ctx, db, name)
		if err != nil {
			return model.Schema{}, fmt.Errorf("foreign keys for %s: %w", name, err)
		}
		t.Indexes, err = tableIndexes(ctx, db, name)
		if err != nil {
			return model.Schema{}, fmt.Errorf("indexes for %s: %w", name, err)
		}

		if kind == "table" && !strings.Contains(strings.ToUpper(ddl.String), "VIRTUAL TABLE") {
			if len(pkColumns(t)) == 0 {
				t.RowKey = findRowID(ctx, db, name)
			}
			t.Editable = len(pkColumns(t)) > 0 || t.RowKey != ""
		}
		schema.Tables = append(schema.Tables, t)
	}
	if err := rows.Err(); err != nil {
		return model.Schema{}, err
	}
	return schema, nil
}

func tableColumns(ctx context.Context, db *sql.DB, table string) ([]model.Column, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA table_xinfo("+quoteIdent(table)+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Column{}
	for rows.Next() {
		var (
			cid     int
			name    string
			decl    string
			notNull int
			dflt    sql.NullString
			pk      int
			hidden  int
		)
		if err := rows.Scan(&cid, &name, &decl, &notNull, &dflt, &pk, &hidden); err != nil {
			return nil, err
		}
		out = append(out, model.Column{
			Name:     name,
			DeclType: decl,
			NotNull:  notNull != 0,
			Default:  dflt.String,
			PK:       pk,
			Hidden:   hidden,
		})
	}
	return out, rows.Err()
}

func tableFKs(ctx context.Context, db *sql.DB, table string) ([]model.ForeignKey, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_list("+quoteIdent(table)+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ForeignKey{}
	for rows.Next() {
		var (
			id, seq            int
			refTable, from, to string
			onUpdate, onDelete string
			match              string
		)
		if err := rows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}
		out = append(out, model.ForeignKey{From: from, Table: refTable, To: to, OnUpdate: onUpdate, OnDelete: onDelete})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].From < out[j].From })
	return out, rows.Err()
}

func tableIndexes(ctx context.Context, db *sql.DB, table string) ([]model.IndexInfo, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA index_list("+quoteIdent(table)+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.IndexInfo{}
	for rows.Next() {
		var (
			seq     int
			name    string
			unique  int
			origin  string
			partial int
		)
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, err
		}
		out = append(out, model.IndexInfo{Name: name, Unique: unique != 0, Partial: partial != 0})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	for i := range out {
		var err error
		out[i].Columns, err = indexColumns(ctx, db, out[i].Name)
		if err != nil {
			return nil, err
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func indexColumns(ctx context.Context, db *sql.DB, index string) ([]string, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA index_info("+quoteIdent(index)+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := []string{}
	for rows.Next() {
		var seqno, cid int
		var name sql.NullString
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			return nil, err
		}
		if name.Valid {
			cols = append(cols, name.String)
		} else {
			cols = append(cols, "(expression)")
		}
	}
	return cols, rows.Err()
}

// findRowID returns the alias that exposes the implicit rowid, trying all three
// names in case one is shadowed by a real column.
func findRowID(ctx context.Context, db *sql.DB, table string) string {
	for _, name := range []string{"rowid", "_rowid_", "oid"} {
		var v any
		err := db.QueryRowContext(ctx, "SELECT "+name+" FROM "+quoteIdent(table)+" LIMIT 1").Scan(&v)
		if err == nil || err == sql.ErrNoRows {
			return name
		}
	}
	return ""
}

// GetObjectDDL returns the DDL for a schema object.
func (d *DB) GetObjectDDL(name string) (string, error) {
	var ddl string
	err := d.sql.QueryRow("SELECT sql FROM sqlite_master WHERE name = ?", name).Scan(&ddl)
	if err != nil {
		return "", err
	}
	return ddl, nil
}
