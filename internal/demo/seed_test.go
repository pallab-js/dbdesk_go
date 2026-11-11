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
		"mines":             5,
		"plants":            6,
		"products":          10,
		"customers":         30,
		"equipment":         28,
		"production":        0, // varies due to randomness
		"stock":             60,
		"sales_orders":      500,
		"dispatches":        0, // varies due to randomness
		"expenses":          1200,
		"royalty_payments":  60,
	}
	for table, n := range want {
		var got int64
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		if n > 0 && got != n {
			t.Errorf("%s rows = %d, want %d", table, got, n)
		}
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
	if err := db.QueryRow("SELECT COUNT(*) FROM daily_production_summary").Scan(&viewCount); err != nil {
		t.Fatal(err)
	}
	if viewCount == 0 {
		t.Errorf("daily_production_summary rows = 0, expected > 0")
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM stock_status").Scan(&viewCount); err != nil {
		t.Fatal(err)
	}
	if viewCount != 60 {
		t.Errorf("stock_status rows = %d, want 60", viewCount)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM sales_summary").Scan(&viewCount); err != nil {
		t.Fatal(err)
	}
	if viewCount != 500 {
		t.Errorf("sales_summary rows = %d, want 500", viewCount)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM dispatch_summary").Scan(&viewCount); err != nil {
		t.Fatal(err)
	}
	if viewCount == 0 {
		t.Errorf("dispatch_summary rows = 0, expected > 0")
	}
}