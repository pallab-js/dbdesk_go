package store

import (
	"path/filepath"
	"testing"

	"dbdesk/internal/model"
)

func TestStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Re-opening must be safe and idempotent.
	st2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	st2.Close()

	conn, err := st.UpsertConnection("/tmp/example.db", false)
	if err != nil {
		t.Fatal(err)
	}
	if conn.ID == 0 || conn.Name != "example.db" {
		t.Errorf("connection = %+v", conn)
	}
	conn2, err := st.UpsertConnection("/tmp/example.db", true)
	if err != nil {
		t.Fatal(err)
	}
	if conn2.ID != conn.ID || !conn2.ReadOnly {
		t.Errorf("upsert did not update: %+v", conn2)
	}
	list, err := st.Connections(10)
	if err != nil || len(list) != 1 {
		t.Fatalf("connections = %v, %v", list, err)
	}

	if err := st.AddHistory(conn.ID, "SELECT 1", "select", 12, 1, true, ""); err != nil {
		t.Fatal(err)
	}
	if err := st.AddHistory(conn.ID, "bad", "select", 3, 0, false, "syntax error"); err != nil {
		t.Fatal(err)
	}
	history, err := st.History(10)
	if err != nil || len(history) != 2 {
		t.Fatalf("history = %v, %v", history, err)
	}
	if history[0].SQL != "bad" || history[0].Success {
		t.Errorf("history order/content = %+v", history)
	}

	settings := model.DefaultSettings()
	settings.RowLimit = 2500
	settings.HistoryEnabled = false
	if err := st.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}
	got, err := st.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if got.RowLimit != 2500 || got.HistoryEnabled {
		t.Errorf("settings = %+v", got)
	}

	q, err := st.SaveQuery("Customers", "SELECT * FROM customers")
	if err != nil {
		t.Fatal(err)
	}
	queries, err := st.SavedQueries()
	if err != nil || len(queries) != 1 {
		t.Fatalf("saved = %v, %v", queries, err)
	}
	if err := st.DeleteSavedQuery(q.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.ClearHistory(); err != nil {
		t.Fatal(err)
	}
	if err := st.ForgetConnection(conn.ID); err != nil {
		t.Fatal(err)
	}
}
