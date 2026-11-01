package dashboard

import (
	"context"
	"database/sql"
	"math"
	"path/filepath"
	"testing"

	"dbdesk/internal/sqlite"
	"dbdesk/internal/store"
)

func TestBuild(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "data.db")
	raw, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		`CREATE TABLE a (id INTEGER PRIMARY KEY, name TEXT)`,
		`CREATE TABLE b (id INTEGER PRIMARY KEY, value REAL, note BLOB)`,
		`INSERT INTO a(name) VALUES ('one'), ('two')`,
		`INSERT INTO b(value) VALUES (1.5)`,
	} {
		if _, err := raw.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	raw.Close()

	db, err := sqlite.Open(dbPath, false)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	st, err := store.Open(filepath.Join(dir, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	conn, err := st.UpsertConnection(dbPath, false)
	if err != nil {
		t.Fatal(err)
	}

	_ = st.AddHistory(conn.ID, "SELECT 1", "select", 10, 1, true, "")
	_ = st.AddHistory(conn.ID, "SELECT 2", "select", 30, 1, true, "")
	_ = st.AddHistory(conn.ID, "bad", "select", 20, 0, false, "error")

	data, err := Build(context.Background(), db, st, conn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if data.Objects.Tables != 2 {
		t.Errorf("tables = %d", data.Objects.Tables)
	}
	if len(data.Tables) != 2 || data.Tables[0].Name != "a" {
		t.Errorf("table stats = %+v", data.Tables)
	}
	if data.Totals.Queries != 3 || data.Totals.Errors != 1 {
		t.Errorf("totals = %+v", data.Totals)
	}
	if math.Abs(data.Totals.ErrorRate-33.333) > 0.1 {
		t.Errorf("error rate = %f", data.Totals.ErrorRate)
	}
	if data.Totals.P50 != 20 || data.Totals.P95 != 30 {
		t.Errorf("percentiles = %+v", data.Totals)
	}
	var success, failures int
	for _, b := range data.Activity {
		success += b.Success
		failures += b.Error
	}
	if success != 2 || failures != 1 {
		t.Errorf("activity = %d/%d", success, failures)
	}
	if len(data.Recent) != 3 || data.Recent[0].SQL != "bad" {
		t.Errorf("recent = %+v", data.Recent)
	}
}
