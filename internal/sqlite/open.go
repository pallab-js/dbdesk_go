package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"dbdesk/internal/model"
)

const maxCellChars = 10000

type countEntry struct {
	rows int64
	at   time.Time
}

// DB is one open user database. It serializes access through a single
// connection so writes and reads never fight over SQLite locks.
type DB struct {
	sql      *sql.DB
	path     string
	readOnly bool

	schemaMu sync.RWMutex
	schema   model.Schema

	lastMu sync.RWMutex
	last   *model.QueryResult

	countMu sync.Mutex
	counts  map[string]countEntry
}

func Open(path string, readOnly bool) (*DB, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve database path: %w", err)
	}
	mode := "rw"
	if readOnly {
		mode = "ro"
	}
	dsn := "file:" + filepath.ToSlash(abs) + "?mode=" + mode
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot open %s: %w", filepath.Base(abs), err)
	}
	pragmas := []string{
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	if readOnly {
		pragmas = append(pragmas, "PRAGMA query_only=ON")
	}
	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			db.Close()
			return nil, fmt.Errorf("configure database: %w", err)
		}
	}

	schema, err := loadSchema(ctx, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read schema: %w", err)
	}
	return &DB{
		sql:      db,
		path:     abs,
		readOnly: readOnly,
		schema:   schema,
		counts:   map[string]countEntry{},
	}, nil
}

func (d *DB) Close() error { return d.sql.Close() }

func (d *DB) Path() string   { return d.path }
func (d *DB) ReadOnly() bool { return d.readOnly }

func (d *DB) Schema() model.Schema {
	d.schemaMu.RLock()
	defer d.schemaMu.RUnlock()
	out := make([]model.Table, len(d.schema.Tables))
	copy(out, d.schema.Tables)
	return model.Schema{Tables: out}
}

func (d *DB) RefreshSchema(ctx context.Context) (model.Schema, error) {
	s, err := loadSchema(ctx, d.sql)
	if err != nil {
		return model.Schema{}, err
	}
	d.schemaMu.Lock()
	d.schema = s
	d.schemaMu.Unlock()
	out := make([]model.Table, len(s.Tables))
	copy(out, s.Tables)
	return model.Schema{Tables: out}, nil
}

func (d *DB) InvalidateCounts() {
	d.countMu.Lock()
	d.counts = map[string]countEntry{}
	d.countMu.Unlock()
}

// DatabaseStats reads cheap database-level metadata.
func (d *DB) DatabaseStats(ctx context.Context) (model.DatabaseStats, model.ObjectCounts, error) {
	st := model.DatabaseStats{Path: d.path, ReadOnly: d.readOnly}
	if err := d.sql.QueryRowContext(ctx, "PRAGMA page_size").Scan(&st.PageSize); err != nil {
		return st, model.ObjectCounts{}, err
	}
	if err := d.sql.QueryRowContext(ctx, "PRAGMA page_count").Scan(&st.PageCount); err != nil {
		return st, model.ObjectCounts{}, err
	}
	if err := d.sql.QueryRowContext(ctx, "PRAGMA freelist_count").Scan(&st.FreePages); err != nil {
		return st, model.ObjectCounts{}, err
	}
	if err := d.sql.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&st.JournalMode); err != nil {
		return st, model.ObjectCounts{}, err
	}
	if err := d.sql.QueryRowContext(ctx, "PRAGMA encoding").Scan(&st.Encoding); err != nil {
		return st, model.ObjectCounts{}, err
	}
	st.SizeBytes = st.PageSize * st.PageCount

	var objects model.ObjectCounts
	rows, err := d.sql.QueryContext(ctx, `
		SELECT type, COUNT(*) FROM sqlite_schema
		WHERE name NOT LIKE 'sqlite_%' GROUP BY type`)
	if err != nil {
		return st, objects, err
	}
	defer rows.Close()
	for rows.Next() {
		var kind string
		var n int
		if err := rows.Scan(&kind, &n); err != nil {
			return st, objects, err
		}
		switch kind {
		case "table":
			objects.Tables = n
		case "view":
			objects.Views = n
		case "index":
			objects.Indexes = n
		case "trigger":
			objects.Triggers = n
		}
	}
	return st, objects, rows.Err()
}

// CountTable returns a cached exact row count.
func (d *DB) CountTable(ctx context.Context, table string, ttl time.Duration) (int64, error) {
	if _, ok := d.table(table); !ok {
		return 0, fmt.Errorf("unknown table %q", table)
	}
	d.countMu.Lock()
	entry, ok := d.counts[table]
	d.countMu.Unlock()
	if ok && time.Since(entry.at) < ttl {
		return entry.rows, nil
	}
	var n int64
	if err := d.sql.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+quoteIdent(table)).Scan(&n); err != nil {
		return 0, err
	}
	d.countMu.Lock()
	d.counts[table] = countEntry{rows: n, at: time.Now()}
	d.countMu.Unlock()
	return n, nil
}

func (d *DB) table(name string) (model.Table, bool) {
	d.schemaMu.RLock()
	defer d.schemaMu.RUnlock()
	for _, t := range d.schema.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return model.Table{}, false
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func pkColumns(t model.Table) []string {
	var out []string
	for _, c := range t.Columns {
		if c.PK > 0 {
			out = append(out, c.Name)
		}
	}
	return out
}

func visibleColumns(t model.Table) []model.Column {
	out := make([]model.Column, 0, len(t.Columns))
	for _, c := range t.Columns {
		if c.Hidden == 1 {
			continue
		}
		out = append(out, c)
	}
	return out
}

func columnByName(t model.Table, name string) (model.Column, bool) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return model.Column{}, false
}
