package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"dbdesk/internal/model"
)

func fixture(t *testing.T) *DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	stmts := []string{
		`PRAGMA foreign_keys=ON`,
		`CREATE TABLE customers (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT UNIQUE, note TEXT, photo BLOB)`,
		`CREATE TABLE orders (id INTEGER PRIMARY KEY, customer_id INTEGER NOT NULL REFERENCES customers(id), total REAL, status TEXT DEFAULT 'new')`,
		`CREATE TABLE composite (a TEXT, b INTEGER, val TEXT, PRIMARY KEY(a, b))`,
		`CREATE TABLE norowid (key TEXT PRIMARY KEY, val INTEGER) WITHOUT ROWID`,
		`CREATE TABLE logs (msg TEXT)`,
		`CREATE VIEW v_orders AS SELECT o.id, c.name FROM orders o JOIN customers c ON c.id = o.customer_id`,
		`CREATE INDEX idx_orders_customer ON orders(customer_id)`,
		`INSERT INTO customers(id, name, email, note) VALUES (1, 'Ada', 'ada@example.com', 'first')`,
		`INSERT INTO customers(id, name, email, note) VALUES (2, 'Grace', 'grace@example.com', NULL)`,
		`INSERT INTO customers(id, name, email, note, photo) VALUES (3, 'Linus', 'linus@example.com', 'third', X'00FF')`,
		`INSERT INTO orders(id, customer_id, total, status) VALUES (1, 1, 10.5, 'paid')`,
		`INSERT INTO orders(id, customer_id, total, status) VALUES (2, 2, 20.0, 'new')`,
		`INSERT INTO composite(a, b, val) VALUES ('x', 1, 'one')`,
		`INSERT INTO norowid(key, val) VALUES ('k', 7)`,
		`INSERT INTO logs(msg) VALUES ('hello')`,
	}
	for _, stmt := range stmts {
		if _, err := raw.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	raw.Close()
	db, err := Open(path, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSplitStatements(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"SELECT 1", []string{"SELECT 1"}},
		{"SELECT 1;", []string{"SELECT 1"}},
		{"SELECT ';'; SELECT 2", []string{"SELECT ';'", "SELECT 2"}},
		{`SELECT "a;b" FROM t; -- trailing`, []string{`SELECT "a;b" FROM t`, "-- trailing"}},
		{"SELECT 1 /* ; */; SELECT 2", []string{"SELECT 1 /* ; */", "SELECT 2"}},
		{"SELECT 'it''s; fine'", []string{"SELECT 'it''s; fine'"}},
		{"SELECT [a;b] FROM t", []string{"SELECT [a;b] FROM t"}},
	}
	for _, c := range cases {
		got := SplitStatements(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("SplitStatements(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestFirstKeyword(t *testing.T) {
	cases := map[string]string{
		"  select 1":       "SELECT",
		"-- hello\nINSERT": "INSERT",
		"/* x */ WITH q AS (SELECT 1) SELECT * FROM q": "WITH",
		"": "",
	}
	for in, want := range cases {
		if got := FirstKeyword(in); got != want {
			t.Errorf("FirstKeyword(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSchema(t *testing.T) {
	db := fixture(t)
	s := db.Schema()
	if len(s.Tables) != 6 {
		t.Fatalf("got %d tables, want 6", len(s.Tables))
	}
	byName := map[string]model.Table{}
	for _, tbl := range s.Tables {
		byName[tbl.Name] = tbl
	}
	logs := byName["logs"]
	if logs.RowKey != "rowid" || !logs.Editable {
		t.Errorf("logs rowkey/editable = %q/%v", logs.RowKey, logs.Editable)
	}
	if len(pkColumns(byName["composite"])) != 2 {
		t.Errorf("composite PK = %v", pkColumns(byName["composite"]))
	}
	if !byName["customers"].Editable {
		t.Error("customers should be editable")
	}
	if byName["v_orders"].Editable {
		t.Error("view must not be editable")
	}
	if len(byName["orders"].FKs) != 1 || byName["orders"].FKs[0].Table != "customers" {
		t.Errorf("orders FKs = %#v", byName["orders"].FKs)
	}
	if len(byName["orders"].Indexes) == 0 {
		t.Error("expected order index")
	}
}

func TestExecute(t *testing.T) {
	db := fixture(t)
	ctx := context.Background()

	res, err := db.Execute(ctx, "SELECT id, name, NULL AS missing, photo FROM customers ORDER BY id", 100, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 3 || len(res.Columns) != 4 {
		t.Fatalf("rows/cols = %d/%d", res.RowCount, len(res.Columns))
	}
	if res.Rows[1][1] != "Grace" || res.Rows[1][2] != nil {
		t.Errorf("unexpected row: %#v", res.Rows[1])
	}
	if res.Rows[2][3] != "<blob 2 bytes>" {
		t.Errorf("blob display = %#v", res.Rows[2][3])
	}

	res, err = db.Execute(ctx, "SELECT * FROM customers", 2, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Truncated || res.RowCount != 2 {
		t.Errorf("truncation = %v rows=%d", res.Truncated, res.RowCount)
	}

	if _, err := db.Execute(ctx, "SELECT 1; SELECT 2", 10, time.Second); err == nil {
		t.Error("expected multi-statement error")
	} else if !strings.Contains(err.Error(), "one statement") {
		t.Errorf("unexpected error: %v", err)
	}

	res, err = db.Execute(ctx, "UPDATE customers SET name='Ada L' WHERE id=1", 10, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if res.Affected != 1 || res.IsRead {
		t.Errorf("update result = %+v", res)
	}
}

func TestExecuteTimeout(t *testing.T) {
	db := fixture(t)
	ctx := context.Background()
	start := time.Now()
	_, err := db.Execute(ctx, `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM c LIMIT 1000000000) SELECT count(*) FROM c`, 10, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("timeout took too long: %v", time.Since(start))
	}
}

func TestReadOnly(t *testing.T) {
	db := fixture(t)
	path := db.Path()
	db.Close()
	ro, err := Open(path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	if _, err := ro.Execute(context.Background(), "DELETE FROM customers", 10, time.Second); err == nil {
		t.Fatal("expected read-only error")
	}
	if _, err := ro.InsertRow(context.Background(), model.RowInsertRequest{Table: "customers", Values: []model.KeyValue{{Name: "name", Value: "x"}}}); err == nil {
		t.Fatal("expected read-only insert error")
	}
}

func TestPreviewAndCRUD(t *testing.T) {
	db := fixture(t)
	ctx := context.Background()

	page, err := db.PreviewTable(ctx, model.TablePageRequest{Table: "customers", Page: 1, PageSize: 2, SortColumn: "id", SortDir: "desc"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || len(page.Rows) != 2 || !page.HasNext {
		t.Fatalf("page = %+v rows=%d", page, len(page.Rows))
	}
	if page.Rows[0][0] != int64(3) {
		t.Errorf("expected descending first id 3, got %#v", page.Rows[0][0])
	}
	if !reflect.DeepEqual(page.KeyIndexes, []int{0}) || !reflect.DeepEqual(page.KeyNames, []string{"id"}) {
		t.Errorf("keys = %v %v", page.KeyIndexes, page.KeyNames)
	}

	ins, err := db.InsertRow(ctx, model.RowInsertRequest{Table: "customers", Values: []model.KeyValue{
		{Name: "name", Value: "Margaret"},
		{Name: "email", Value: "margaret@example.com"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if ins.LastInsertID != 4 {
		t.Errorf("last insert id = %d", ins.LastInsertID)
	}

	upd, err := db.UpdateRow(ctx, model.RowUpdateRequest{
		Table:  "customers",
		Keys:   []model.KeyValue{{Name: "id", Value: string("4")}},
		Values: []model.KeyValue{{Name: "name", Value: "Margaret H"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if upd.Affected != 1 {
		t.Errorf("update affected = %d", upd.Affected)
	}

	del, err := db.DeleteRow(ctx, model.RowDeleteRequest{Table: "customers", Keys: []model.KeyValue{{Name: "id", Value: float64(4)}}})
	if err != nil || del.Affected != 1 {
		t.Fatalf("delete: %v %+v", err, del)
	}

	// Composite primary key.
	if _, err := db.UpdateRow(ctx, model.RowUpdateRequest{
		Table:  "composite",
		Keys:   []model.KeyValue{{Name: "a", Value: "x"}, {Name: "b", Value: "1"}},
		Values: []model.KeyValue{{Name: "val", Value: "updated"}},
	}); err != nil {
		t.Fatal(err)
	}

	// Rowid table.
	logs, err := db.PreviewTable(ctx, model.TablePageRequest{Table: "logs", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if !logs.Editable || logs.KeyNames[0] != "rowid" {
		t.Fatalf("logs page = %+v", logs)
	}
	rowid := logs.Rows[0][logs.KeyIndexes[0]]
	if _, err := db.UpdateRow(ctx, model.RowUpdateRequest{
		Table:  "logs",
		Keys:   []model.KeyValue{{Name: "rowid", Value: rowid}},
		Values: []model.KeyValue{{Name: "msg", Value: "changed"}},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCRUDValidation(t *testing.T) {
	db := fixture(t)
	ctx := context.Background()

	if _, err := db.PreviewTable(ctx, model.TablePageRequest{Table: `customers"; DROP TABLE customers; --`, Page: 1, PageSize: 10}); err == nil {
		t.Error("expected unknown table error")
	}
	if _, err := db.PreviewTable(ctx, model.TablePageRequest{Table: "customers", SortColumn: `name"; DROP TABLE customers; --`}); err == nil {
		t.Error("expected unknown sort column error")
	}
	if _, err := db.InsertRow(ctx, model.RowInsertRequest{Table: "customers", Values: []model.KeyValue{{Name: "id", Value: "not-an-int"}}}); err == nil {
		t.Error("expected integer validation error")
	}
	if _, err := db.InsertRow(ctx, model.RowInsertRequest{Table: "v_orders", Values: []model.KeyValue{{Name: "name", Value: "x"}}}); err == nil {
		t.Error("expected view write rejection")
	}
	if _, err := db.UpdateRow(ctx, model.RowUpdateRequest{
		Table: "customers", Keys: []model.KeyValue{{Name: "id", Value: "1"}},
	}); err == nil {
		t.Error("expected no-values error")
	}
}

func TestCSVExport(t *testing.T) {
	db := fixture(t)
	if _, err := db.Execute(context.Background(), `SELECT '=1+1' AS formula, 'a,b' AS comma, NULL AS empty, 'line
break' AS newline`, 10, time.Second); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := db.WriteLastCSV(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `'=1+1,`) {
		t.Errorf("formula not guarded:\n%s", got)
	}
	if !strings.Contains(got, `"a,b"`) {
		t.Errorf("comma not quoted:\n%s", got)
	}
	if !strings.Contains(got, `"line`+"\n"+`break"`) {
		t.Errorf("newline not quoted:\n%s", got)
	}
}

func TestOpenMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	if _, err := Open(path, false); err == nil {
		t.Fatal("expected error for missing database")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("opening a missing a database should not create it")
	}
}
