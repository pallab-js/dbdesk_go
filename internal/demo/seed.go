package demo

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const seed = 20261101

var (
	firstNames = []string{"Avery", "Jordan", "Riley", "Morgan", "Casey", "Quinn", "Rowan", "Sage", "Drew", "Emerson", "Finley", "Harper", "Jamie", "Kai", "Lane", "Marlow", "Noel", "Parker", "Reese", "Skyler"}
	lastNames  = []string{"Adams", "Baker", "Chen", "Diaz", "Evans", "Fischer", "Garcia", "Hughes", "Ibrahim", "Jensen", "Kim", "Lopez", "Miller", "Nguyen", "Owens", "Patel", "Quinn", "Rossi", "Silva", "Turner"}
	cities     = []string{"Guwahati", "Shillong", "Imphal", "Aizawl", "Agartala", "Kohima", "Itanagar", "Dibrugarh", "Silchar", "Jorhat", "Tinsukia", "Dimapur"}
	countries  = []string{"India"}
	categories = []string{"Generation", "Transmission", "Distribution", "Compression", "Metering", "Storage", "Maintenance", "Safety", "Billing", "SCADA", "Field Ops", "Compliance"}
	statuses   = []string{"completed", "completed", "completed", "paid", "paid", "pending", "cancelled", "refunded"}
	methods    = []string{"card", "card", "card", "bank_transfer", "paypal", "wallet"}
	comments   = []string{"Operates within spec.", "Reliable in cold weather.", "Installed on schedule.", "Meets safety standards.", "Documentation could be clearer.", "Solid build quality.", "Needs follow-up inspection.", "Decent for routine use.", "Excellent performance.", "Replaced aging unit successfully.", ""}
)

// Seed creates the demo database if it does not already exist.
func Seed(path string) error {
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	_ = os.Remove(path)

	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := createSchema(db); err != nil {
		removeDB(path)
		return err
	}
	if err := fill(db); err != nil {
		removeDB(path)
		return err
	}
	if _, err := db.Exec("ANALYZE"); err != nil {
		return err
	}
	return nil
}

func removeDB(path string) {
	_ = os.Remove(path)
	_ = os.Remove(path + "-wal")
	_ = os.Remove(path + "-shm")
}

func createSchema(db *sql.DB) error {
	for _, stmt := range strings.Split(ddl, ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

const ddl = `
CREATE TABLE categories (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE
);
CREATE TABLE suppliers (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	country TEXT NOT NULL,
	email TEXT NOT NULL
);
CREATE TABLE products (
	id INTEGER PRIMARY KEY,
	category_id INTEGER NOT NULL REFERENCES categories(id),
	supplier_id INTEGER NOT NULL REFERENCES suppliers(id),
	name TEXT NOT NULL,
	sku TEXT NOT NULL UNIQUE,
	price REAL NOT NULL,
	stock INTEGER NOT NULL DEFAULT 0,
	active INTEGER NOT NULL DEFAULT 1,
	thumbnail BLOB,
	created_at TEXT NOT NULL
);
CREATE TABLE customers (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	city TEXT,
	country TEXT NOT NULL,
	created_at TEXT NOT NULL
);
CREATE TABLE orders (
	id INTEGER PRIMARY KEY,
	customer_id INTEGER NOT NULL REFERENCES customers(id),
	status TEXT NOT NULL,
	total REAL NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL
);
CREATE TABLE order_items (
	id INTEGER PRIMARY KEY,
	order_id INTEGER NOT NULL REFERENCES orders(id),
	product_id INTEGER NOT NULL REFERENCES products(id),
	quantity INTEGER NOT NULL,
	unit_price REAL NOT NULL
);
CREATE TABLE payments (
	id INTEGER PRIMARY KEY,
	order_id INTEGER NOT NULL REFERENCES orders(id),
	method TEXT NOT NULL,
	amount REAL NOT NULL,
	paid_at TEXT NOT NULL
);
CREATE TABLE reviews (
	id INTEGER PRIMARY KEY,
	product_id INTEGER NOT NULL REFERENCES products(id),
	customer_id INTEGER NOT NULL REFERENCES customers(id),
	rating INTEGER NOT NULL,
	comment TEXT,
	created_at TEXT NOT NULL
);
CREATE VIEW order_summary AS
	SELECT o.id, c.name AS customer, o.status, o.total, o.created_at,
	       COUNT(oi.id) AS item_count
	FROM orders o
	JOIN customers c ON c.id = o.customer_id
	LEFT JOIN order_items oi ON oi.order_id = o.id
	GROUP BY o.id;
`

func fill(db *sql.DB) error {
	rng := rand.New(rand.NewSource(seed))
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	exec := func(query string, args ...any) error {
		_, err := tx.Exec(query, args...)
		return err
	}
	now := time.Now()

	for i, name := range categories {
		if err := exec(`INSERT INTO categories(id, name) VALUES(?, ?)`, i+1, name); err != nil {
			return err
		}
	}
	for i := 0; i < 30; i++ {
		if err := exec(`INSERT INTO suppliers(id, name, country, email) VALUES(?, ?, ?, ?)`,
			i+1, fmt.Sprintf("GasGrid Supplier %02d", i+1), countries[rng.Intn(len(countries))],
			fmt.Sprintf("supplier%02d@example.com", i+1)); err != nil {
			return err
		}
	}

	productPrices := make([]float64, 401)
	for i := 1; i <= 400; i++ {
		price := float64(5+rng.Intn(49500)) / 100
		productPrices[i] = price
		var thumb any
		if rng.Intn(4) == 0 {
			thumb = []byte{0x89, 'P', 'N', 'G', byte(rng.Intn(255)), byte(rng.Intn(255))}
		}
		created := now.Add(-time.Duration(30+rng.Intn(300)) * 24 * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO products(id, category_id, supplier_id, name, sku, price, stock, active, thumbnail, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			i, 1+rng.Intn(12), 1+rng.Intn(30), fmt.Sprintf("Asset %03d", i),
			fmt.Sprintf("SKU-%04d", i), price, rng.Intn(500), rng.Intn(6) != 0, thumb, created); err != nil {
			return err
		}
	}

	for i := 1; i <= 2000; i++ {
		name := firstNames[rng.Intn(len(firstNames))] + " " + lastNames[rng.Intn(len(lastNames))]
		var city any
		if rng.Intn(20) != 0 {
			city = cities[rng.Intn(len(cities))]
		}
		created := now.Add(-time.Duration(rng.Intn(60)) * 24 * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO customers(id, name, email, city, country, created_at) VALUES(?, ?, ?, ?, ?, ?)`,
			i, name, fmt.Sprintf("customer%04d@example.com", i), city,
			countries[rng.Intn(len(countries))], created); err != nil {
			return err
		}
	}

	itemID := 0
	paymentID := 0
	for i := 1; i <= 12000; i++ {
		customerID := 1 + rng.Intn(2000)
		status := statuses[rng.Intn(len(statuses))]
		createdAt := now.Add(-time.Duration(rng.Intn(30*24)) * time.Hour)
		created := createdAt.Format(time.RFC3339)
		if err := exec(`INSERT INTO orders(id, customer_id, status, total, created_at) VALUES(?, ?, ?, 0, ?)`,
			i, customerID, status, created); err != nil {
			return err
		}
		total := 0.0
		for n := 1 + rng.Intn(3); n > 0; n-- {
			productID := 1 + rng.Intn(400)
			qty := 1 + rng.Intn(5)
			price := productPrices[productID]
			total += float64(qty) * price
			itemID++
			if err := exec(`INSERT INTO order_items(id, order_id, product_id, quantity, unit_price) VALUES(?, ?, ?, ?, ?)`,
				itemID, i, productID, qty, price); err != nil {
				return err
			}
		}
		if err := exec(`UPDATE orders SET total = ? WHERE id = ?`, total, i); err != nil {
			return err
		}
		if status == "paid" || status == "completed" {
			paymentID++
			if err := exec(`INSERT INTO payments(id, order_id, method, amount, paid_at) VALUES(?, ?, ?, ?, ?)`,
				paymentID, i, methods[rng.Intn(len(methods))], total, created); err != nil {
				return err
			}
		}
	}

	for i := 1; i <= 5000; i++ {
		var comment any
		c := comments[rng.Intn(len(comments))]
		if c != "" {
			comment = c
		}
		created := now.Add(-time.Duration(rng.Intn(30*24)) * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO reviews(id, product_id, customer_id, rating, comment, created_at) VALUES(?, ?, ?, ?, ?, ?)`,
			i, 1+rng.Intn(400), 1+rng.Intn(2000), 1+rng.Intn(5), comment, created); err != nil {
			return err
		}
	}

	indexes := []string{
		`CREATE INDEX idx_orders_customer ON orders(customer_id)`,
		`CREATE INDEX idx_orders_created ON orders(created_at)`,
		`CREATE INDEX idx_order_items_order ON order_items(order_id)`,
		`CREATE INDEX idx_order_items_product ON order_items(product_id)`,
		`CREATE INDEX idx_products_category ON products(category_id)`,
		`CREATE INDEX idx_reviews_product ON reviews(product_id)`,
		`CREATE INDEX idx_payments_order ON payments(order_id)`,
	}
	for _, stmt := range indexes {
		if err := exec(stmt); err != nil {
			return err
		}
	}
	return tx.Commit()
}
