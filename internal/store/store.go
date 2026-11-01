package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"dbdesk/internal/model"
)

const schemaVersion = 1

const schemaSQL = `
CREATE TABLE IF NOT EXISTS connections (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	read_only INTEGER NOT NULL DEFAULT 0,
	last_opened_at INTEGER NOT NULL,
	created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS query_history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	connection_id INTEGER,
	sql TEXT NOT NULL DEFAULT '',
	kind TEXT NOT NULL DEFAULT '',
	duration_ms INTEGER NOT NULL DEFAULT 0,
	row_count INTEGER NOT NULL DEFAULT 0,
	success INTEGER NOT NULL DEFAULT 0,
	error TEXT NOT NULL DEFAULT '',
	created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_history_created ON query_history(created_at);

CREATE TABLE IF NOT EXISTS saved_queries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	connection_id INTEGER,
	name TEXT NOT NULL,
	sql TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS table_stats (
	connection_id INTEGER NOT NULL,
	table_name TEXT NOT NULL,
	row_count INTEGER NOT NULL DEFAULT 0,
	captured_at INTEGER NOT NULL,
	PRIMARY KEY (connection_id, table_name)
);
`

type Store struct {
	db   *sql.DB
	path string
}

// Open opens (and migrates) the application database. A copy of the previous
// database is kept as app.db.bak whenever a real migration is about to run.
func Open(path string) (*Store, error) {
	if err := backupBeforeMigration(path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open app database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable app database WAL: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db, path: path}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func currentVersion(path string) (int, error) {
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var v int
	if err := db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		return 0, err
	}
	return v, nil
}

func backupBeforeMigration(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) || (err == nil && info.Size() == 0) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect app database: %w", err)
	}
	v, err := currentVersion(path)
	if err != nil {
		return fmt.Errorf("read app database version: %w", err)
	}
	if v == 0 || v >= schemaVersion {
		return nil
	}
	src, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("backup app database: %w", err)
	}
	defer src.Close()
	dst, err := os.Create(path + ".bak")
	if err != nil {
		return fmt.Errorf("backup app database: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return fmt.Errorf("backup app database: %w", err)
	}
	return dst.Close()
}

func migrate(db *sql.DB) error {
	var v int
	if err := db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		return fmt.Errorf("read app database version: %w", err)
	}
	if v >= schemaVersion {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("start migration: %w", err)
	}
	defer tx.Rollback()
	for _, stmt := range strings.Split(schemaSQL, ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migrate app database: %w", err)
		}
	}
	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersion)); err != nil {
		return fmt.Errorf("set app database version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}
	return nil
}

func (s *Store) UpsertConnection(path string, readOnly bool) (model.ConnectionInfo, error) {
	now := time.Now().UnixMilli()
	name := filepath.Base(path)
	_, err := s.db.Exec(`
		INSERT INTO connections(path, name, read_only, last_opened_at, created_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			name = excluded.name,
			read_only = excluded.read_only,
			last_opened_at = excluded.last_opened_at`,
		path, name, boolInt(readOnly), now, now)
	if err != nil {
		return model.ConnectionInfo{}, fmt.Errorf("save connection: %w", err)
	}
	row := s.db.QueryRow(`SELECT id, path, name, read_only, last_opened_at FROM connections WHERE path = ?`, path)
	return scanConnection(row)
}

func (s *Store) Connections(limit int) ([]model.ConnectionInfo, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(`SELECT id, path, name, read_only, last_opened_at FROM connections ORDER BY last_opened_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.ConnectionInfo, 0, limit)
	for rows.Next() {
		c, err := scanConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ForgetConnection(id int64) error {
	_, err := s.db.Exec(`DELETE FROM connections WHERE id = ?`, id)
	return err
}

type scanner interface{ Scan(dest ...any) error }

func scanConnection(row scanner) (model.ConnectionInfo, error) {
	var c model.ConnectionInfo
	var readOnly int
	if err := row.Scan(&c.ID, &c.Path, &c.Name, &readOnly, &c.LastOpenedAt); err != nil {
		return model.ConnectionInfo{}, err
	}
	c.ReadOnly = readOnly != 0
	return c, nil
}

func (s *Store) AddHistory(connectionID int64, sqlText, kind string, durationMs int64, rowCount int, success bool, errText string) error {
	_, err := s.db.Exec(`
		INSERT INTO query_history(connection_id, sql, kind, duration_ms, row_count, success, error, created_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		connectionID, sqlText, kind, durationMs, rowCount, boolInt(success), errText, time.Now().UnixMilli())
	return err
}

func (s *Store) History(limit int) ([]model.HistoryEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	rows, err := s.db.Query(`
		SELECT id, sql, kind, duration_ms, row_count, success, error, created_at
		FROM query_history ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanHistory(rows)
}

func (s *Store) HistorySince(sinceMs int64) ([]model.HistoryEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, sql, kind, duration_ms, row_count, success, error, created_at
		FROM query_history WHERE created_at >= ? ORDER BY id ASC LIMIT 50000`, sinceMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanHistory(rows)
}

func scanHistory(rows *sql.Rows) ([]model.HistoryEntry, error) {
	out := make([]model.HistoryEntry, 0, 64)
	for rows.Next() {
		var h model.HistoryEntry
		var success int
		if err := rows.Scan(&h.ID, &h.SQL, &h.Kind, &h.DurationMs, &h.RowCount, &success, &h.Error, &h.CreatedAt); err != nil {
			return nil, err
		}
		h.Success = success != 0
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Store) ClearHistory() error {
	_, err := s.db.Exec(`DELETE FROM query_history`)
	return err
}

func (s *Store) PruneHistory(retentionDays, maxRows int) error {
	if retentionDays < 1 {
		retentionDays = 30
	}
	if maxRows < 1000 {
		maxRows = 20000
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays).UnixMilli()
	if _, err := s.db.Exec(`DELETE FROM query_history WHERE created_at < ?`, cutoff); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM query_history WHERE id NOT IN (SELECT id FROM query_history ORDER BY id DESC LIMIT ?)`, maxRows)
	return err
}

func (s *Store) SavedQueries() ([]model.SavedQuery, error) {
	rows, err := s.db.Query(`SELECT id, name, sql, created_at, updated_at FROM saved_queries ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.SavedQuery, 0, 16)
	for rows.Next() {
		var q model.SavedQuery
		if err := rows.Scan(&q.ID, &q.Name, &q.SQL, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (s *Store) SaveQuery(name, sqlText string) (model.SavedQuery, error) {
	now := time.Now().UnixMilli()
	res, err := s.db.Exec(`INSERT INTO saved_queries(name, sql, created_at, updated_at) VALUES(?, ?, ?, ?)`, name, sqlText, now, now)
	if err != nil {
		return model.SavedQuery{}, err
	}
	id, _ := res.LastInsertId()
	return model.SavedQuery{ID: id, Name: name, SQL: sqlText, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *Store) DeleteSavedQuery(id int64) error {
	_, err := s.db.Exec(`DELETE FROM saved_queries WHERE id = ?`, id)
	return err
}

func (s *Store) Settings() (model.Settings, error) {
	st := model.DefaultSettings()
	var raw string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = 'app'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return model.DefaultSettings(), nil
	}
	return st, nil
}

func (s *Store) SaveSettings(st model.Settings) error {
	raw, err := json.Marshal(st)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO settings(key, value) VALUES('app', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, string(raw))
	return err
}

func (s *Store) UpsertTableStats(connectionID int64, stats []model.TableStat) error {
	now := time.Now().UnixMilli()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, t := range stats {
		if t.Error != "" {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO table_stats(connection_id, table_name, row_count, captured_at)
			VALUES(?, ?, ?, ?)
			ON CONFLICT(connection_id, table_name) DO UPDATE SET
				row_count = excluded.row_count, captured_at = excluded.captured_at`,
			connectionID, t.Name, t.Rows, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
