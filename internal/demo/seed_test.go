package demo

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestSeed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.db")
	if err := Seed(path); err != nil {
		t.Fatal(err)
	}
	// Seeding twice must be a no-op.
	if err := Seed(path); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	want := map[string]int64{
		"categories": 12,
		"suppliers":  30,
		"products":   400,
		"customers":  2000,
		"orders":     12000,
		"reviews":    5000,
	}
	for table, n := range want {
		var got int64
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		if got != n {
			t.Errorf("%s rows = %d, want %d", table, got, n)
		}
	}
	var extra int64
	if err := db.QueryRow("SELECT COUNT(*) FROM order_items").Scan(&extra); err != nil {
		t.Fatal(err)
	}
	if extra < 12000 || extra > 36000 {
		t.Errorf("order_items = %d, outside expected range", extra)
	}
	var violations int
	rows, err := db.Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		violations++
	}
	rows.Close()
	if violations != 0 {
		t.Errorf("foreign key violations = %d", violations)
	}
	var viewCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM order_summary").Scan(&viewCount); err != nil {
		t.Fatal(err)
	}
	if viewCount != 12000 {
		t.Errorf("order_summary rows = %d", viewCount)
	}
}
