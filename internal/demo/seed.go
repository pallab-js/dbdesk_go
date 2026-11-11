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
	firstNames = []string{"Rajesh", "Amit", "Sanjay", "Vikram", "Deepak", "Manoj", "Rakesh", "Sunil", "Anil", "Pankaj", "Nitin", "Rohit", "Sachin", "Arun", "Ajay", "Vinod", "Mahesh", "Dinesh", "Harish", "Ganesh"}
	lastNames  = []string{"Sharma", "Singh", "Gupta", "Verma", "Agarwal", "Jain", "Kumar", "Yadav", "Tiwari", "Mishra", "Pandey", "Chauhan", "Rathore", "Thakur", "Patel", "Das", "Borah", "Gogoi", "Hazarika", "Saikia"}
	cities     = []string{"Guwahati", "Dibrugarh", "Jorhat", "Tinsukia", "Silchar", "Nagaon", "Tezpur", "Bongaigaon", "Diphu", "Karimganj", "Hailakandi", "Goalpara", "Dhubri", "Kokrajhar", "Udalguri", "Sonitpur", "Lakhimpur", "Dhemaji", "Sivasagar", "Golaghat"}
	states     = []string{"Assam", "Arunachal Pradesh", "Meghalaya", "Nagaland", "Manipur", "Mizoram", "Tripura", "Sikkim"}
	products   = []struct {
		name     string
		unit     string
		minPrice float64
		maxPrice float64
	}{
		{"20mm Aggregates", "MT", 850, 1200},
		{"10mm Aggregates", "MT", 900, 1250},
		{"6mm Aggregates", "MT", 950, 1300},
		{"Stone Dust", "MT", 400, 650},
		{"GSB (Graded Stone Base)", "MT", 600, 900},
		{"WMM (Wet Mix Macadam)", "MT", 700, 1000},
		{"40mm Aggregates", "MT", 750, 1100},
		{"M-Sand (Manufactured Sand)", "MT", 800, 1150},
		{"Crusher Run", "MT", 500, 800},
		{"Boulder Stone", "MT", 350, 600},
	}
	customerTypes = []string{"PWD Contractor", "NHAI Contractor", "Private Builder", "Railway Contractor", "Irrigation Dept", "Urban Dev Authority", "Road Construction Co", "Infrastructure Ltd"}
	expenseCategories = []string{"Diesel", "Explosives", "Drill Bits", "Crusher Wear Parts", "Screen Mesh", "Lubricants", "Electricity", "Labor Wages", "Equipment Hire", "Transport", "Maintenance", "Royalty", "Pollution Control", "Office Expenses"}
	equipmentTypes = []string{"Primary Jaw Crusher", "Secondary Cone Crusher", "VSI Crusher", "Vibrating Screen", "Feeder", "Conveyor Belt", "Loader", "Excavator", "Dump Truck", "Water Tanker", "Generator", "Compressor"}
	truckTypes = []string{"10 Wheeler", "12 Wheeler", "16 Wheeler", "25 Tonner", "31 Tonner"}
)

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
CREATE TABLE mines (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	location TEXT NOT NULL,
	state TEXT NOT NULL,
	area_hectares REAL NOT NULL,
	lease_no TEXT NOT NULL,
	lease_expiry TEXT NOT NULL,
	active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL
);
CREATE TABLE plants (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	mine_id INTEGER NOT NULL REFERENCES mines(id),
	location TEXT NOT NULL,
	capacity_tph INTEGER NOT NULL,
	commissioned_at TEXT NOT NULL,
	active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL
);
CREATE TABLE products (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	unit TEXT NOT NULL DEFAULT 'MT',
	min_price REAL NOT NULL,
	max_price REAL NOT NULL,
	active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL
);
CREATE TABLE customers (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	contact_person TEXT,
	phone TEXT,
	email TEXT,
	address TEXT,
	city TEXT NOT NULL,
	state TEXT NOT NULL,
	gstin TEXT,
	active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL
);
CREATE TABLE equipment (
	id INTEGER PRIMARY KEY,
	plant_id INTEGER NOT NULL REFERENCES plants(id),
	type TEXT NOT NULL,
	make_model TEXT NOT NULL,
	capacity TEXT,
	installed_at TEXT NOT NULL,
	last_service_at TEXT,
	next_service_due TEXT,
	status TEXT NOT NULL DEFAULT 'running',
	created_at TEXT NOT NULL
);
CREATE TABLE production (
	id INTEGER PRIMARY KEY,
	plant_id INTEGER NOT NULL REFERENCES plants(id),
	product_id INTEGER NOT NULL REFERENCES products(id),
	date TEXT NOT NULL,
	quantity_mt REAL NOT NULL,
	operating_hours REAL NOT NULL,
	downtime_hours REAL NOT NULL DEFAULT 0,
	downtime_reason TEXT,
	shift TEXT NOT NULL,
	created_at TEXT NOT NULL,
	UNIQUE(plant_id, product_id, date, shift)
);
CREATE TABLE stock (
	id INTEGER PRIMARY KEY,
	plant_id INTEGER NOT NULL REFERENCES plants(id),
	product_id INTEGER NOT NULL REFERENCES products(id),
	quantity_mt REAL NOT NULL DEFAULT 0,
	reorder_level_mt REAL NOT NULL DEFAULT 100,
	updated_at TEXT NOT NULL,
	UNIQUE(plant_id, product_id)
);
CREATE TABLE sales_orders (
	id INTEGER PRIMARY KEY,
	customer_id INTEGER NOT NULL REFERENCES customers(id),
	product_id INTEGER NOT NULL REFERENCES products(id),
	plant_id INTEGER NOT NULL REFERENCES plants(id),
	order_no TEXT NOT NULL UNIQUE,
	quantity_mt REAL NOT NULL,
	rate_per_mt REAL NOT NULL,
	total_amount REAL NOT NULL,
	order_date TEXT NOT NULL,
	delivery_date TEXT,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at TEXT NOT NULL
);
CREATE TABLE dispatches (
	id INTEGER PRIMARY KEY,
	sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id),
	truck_no TEXT NOT NULL,
	truck_type TEXT NOT NULL,
	driver_name TEXT,
	driver_phone TEXT,
	weighbridge_in REAL NOT NULL,
	weighbridge_out REAL NOT NULL,
	net_weight_mt REAL NOT NULL,
	dispatch_date TEXT NOT NULL,
	dispatch_time TEXT NOT NULL,
	destination TEXT,
	eway_bill_no TEXT,
	status TEXT NOT NULL DEFAULT 'dispatched',
	created_at TEXT NOT NULL
);
CREATE TABLE expenses (
	id INTEGER PRIMARY KEY,
	plant_id INTEGER NOT NULL REFERENCES plants(id),
	category TEXT NOT NULL,
	description TEXT,
	amount REAL NOT NULL,
	expense_date TEXT NOT NULL,
	bill_no TEXT,
	vendor TEXT,
	payment_mode TEXT NOT NULL DEFAULT 'cash',
	status TEXT NOT NULL DEFAULT 'paid',
	created_at TEXT NOT NULL
);
CREATE TABLE royalty_payments (
	id INTEGER PRIMARY KEY,
	mine_id INTEGER NOT NULL REFERENCES mines(id),
	period_from TEXT NOT NULL,
	period_to TEXT NOT NULL,
	quantity_mt REAL NOT NULL,
	rate_per_mt REAL NOT NULL,
	amount REAL NOT NULL,
	paid_at TEXT NOT NULL,
	challan_no TEXT NOT NULL,
	created_at TEXT NOT NULL
);
CREATE VIEW daily_production_summary AS
	SELECT p.date, pl.name AS plant, pr.name AS product,
	       SUM(p.quantity_mt) AS total_qty_mt,
	       SUM(p.operating_hours) AS total_hours,
	       SUM(p.downtime_hours) AS total_downtime
	FROM production p
	JOIN plants pl ON pl.id = p.plant_id
	JOIN products pr ON pr.id = p.product_id
	GROUP BY p.date, pl.id, pr.id;
CREATE VIEW stock_status AS
	SELECT pl.name AS plant, pr.name AS product, s.quantity_mt, s.reorder_level_mt,
	       CASE WHEN s.quantity_mt <= s.reorder_level_mt THEN 'LOW' ELSE 'OK' END AS status
	FROM stock s
	JOIN plants pl ON pl.id = s.plant_id
	JOIN products pr ON pr.id = s.product_id;
CREATE VIEW sales_summary AS
	SELECT so.order_date, c.name AS customer, c.type AS customer_type,
	       pr.name AS product, pl.name AS plant,
	       so.quantity_mt, so.rate_per_mt, so.total_amount, so.status
	FROM sales_orders so
	JOIN customers c ON c.id = so.customer_id
	JOIN products pr ON pr.id = so.product_id
	JOIN plants pl ON pl.id = so.plant_id;
CREATE VIEW dispatch_summary AS
	SELECT d.dispatch_date, so.order_no, c.name AS customer,
	       pr.name AS product, d.truck_no, d.net_weight_mt, d.eway_bill_no
	FROM dispatches d
	JOIN sales_orders so ON so.id = d.sales_order_id
	JOIN customers c ON c.id = so.customer_id
	JOIN products pr ON pr.id = so.product_id;
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

	for i, p := range products {
		created := now.Add(-time.Duration(365+rng.Intn(365)) * 24 * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO products(id, name, unit, min_price, max_price, active, created_at)
			VALUES(?, ?, ?, ?, ?, 1, ?)`,
			i+1, p.name, p.unit, p.minPrice, p.maxPrice, created); err != nil {
			return err
		}
	}

	mines := []struct {
		name        string
		location    string
		state       string
		area        float64
		leaseNo     string
		leaseExpiry string
	}{
		{"Dhansiri Stone Quarry", "Numaligarh", "Assam", 12.5, "ML/ASM/2020/045", "2030-03-31"},
		{"Borgang Granite Mine", "Borgang", "Assam", 8.2, "ML/ASM/2019/012", "2029-12-31"},
		{"Dirang Dolomite Mine", "Dirang", "Arunachal Pradesh", 15.0, "ML/ARP/2021/008", "2031-06-30"},
		{"Nongstoin Limestone Quarry", "Nongstoin", "Meghalaya", 10.5, "ML/MEG/2018/023", "2028-09-30"},
		{"Pfutsero Basalt Mine", "Pfutsero", "Nagaland", 6.8, "ML/NAG/2022/003", "2032-01-31"},
	}
	for i, m := range mines {
		created := now.Add(-time.Duration(730+rng.Intn(730)) * 24 * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO mines(id, name, location, state, area_hectares, lease_no, lease_expiry, active, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, 1, ?)`,
			i+1, m.name, m.location, m.state, m.area, m.leaseNo, m.leaseExpiry, created); err != nil {
			return err
		}
	}

	plants := []struct {
		name        string
		mineID      int
		location    string
		capacityTPH int
		commissioned string
	}{
		{"Numaligarh Crushing Plant - 1", 1, "Numaligarh, Golaghat", 200, "2020-06-15"},
		{"Numaligarh Crushing Plant - 2", 1, "Numaligarh, Golaghat", 150, "2021-11-20"},
		{"Borgang Stone Crusher", 2, "Borgang, Sonitpur", 180, "2019-08-10"},
		{"Dirang Dolomite Plant", 3, "Dirang, West Kameng", 120, "2021-04-05"},
		{"Nongstoin Limestone Crusher", 4, "Nongstoin, West Khasi Hills", 160, "2018-12-01"},
		{"Pfutsero Basalt Crusher", 5, "Pfutsero, Phek", 100, "2022-03-10"},
	}
	for i, pl := range plants {
		created := now.Add(-time.Duration(365+rng.Intn(730)) * 24 * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO plants(id, name, mine_id, location, capacity_tph, commissioned_at, active, created_at)
			VALUES(?, ?, ?, ?, ?, ?, 1, ?)`,
			i+1, pl.name, pl.mineID, pl.location, pl.capacityTPH, pl.commissioned, created); err != nil {
			return err
		}
	}

	for i := 1; i <= 6; i++ {
		for j := 1; j <= 10; j++ {
			reorder := 50.0 + float64(rng.Intn(150))
			qty := reorder + float64(rng.Intn(500))
			updated := now.Add(-time.Duration(rng.Intn(7)) * 24 * time.Hour).Format(time.RFC3339)
			if err := exec(`INSERT INTO stock(plant_id, product_id, quantity_mt, reorder_level_mt, updated_at)
				VALUES(?, ?, ?, ?, ?)`,
				i, j, qty, reorder, updated); err != nil {
				return err
			}
		}
	}

	equipData := []struct {
		plantID  int
		eqType   string
		makeModel string
		capacity  string
		installed string
		status    string
	}{
		{1, "Primary Jaw Crusher", "Metso C125", "200 TPH", "2020-06-15", "running"},
		{1, "Secondary Cone Crusher", "Metso HP300", "180 TPH", "2020-06-15", "running"},
		{1, "VSI Crusher", "Metso Barmac B7150", "120 TPH", "2021-02-10", "running"},
		{1, "Vibrating Screen", "Metso CVB2060", "4 Deck", "2020-06-15", "running"},
		{1, "Feeder", "Metso VF610", "250 TPH", "2020-06-15", "running"},
		{1, "Loader", "CAT 966M", "4.2 m3", "2020-05-20", "running"},
		{1, "Excavator", "CAT 320D", "0.9 m3", "2020-05-20", "running"},
		{1, "Dump Truck", "Tata 2518", "25T", "2020-06-01", "running"},
		{2, "Primary Jaw Crusher", "Terex J1175", "150 TPH", "2021-11-20", "running"},
		{2, "Secondary Cone Crusher", "Terex TC1150", "130 TPH", "2021-11-20", "running"},
		{2, "Vibrating Screen", "Terex 6x20", "3 Deck", "2021-11-20", "maintenance"},
		{2, "Loader", "JCB 455ZX", "3.5 m3", "2021-10-15", "running"},
		{2, "Excavator", "JCB 220X", "1.0 m3", "2021-10-15", "running"},
		{3, "Primary Jaw Crusher", "Propel AVJ-811", "180 TPH", "2019-08-10", "running"},
		{3, "Secondary Cone Crusher", "Propel TCC-300", "160 TPH", "2019-08-10", "running"},
		{3, "VSI Crusher", "Propel VSI-500", "100 TPH", "2020-03-15", "running"},
		{3, "Vibrating Screen", "Propel 6x16", "4 Deck", "2019-08-10", "running"},
		{3, "Loader", "L&T 980", "4.0 m3", "2019-07-20", "running"},
		{4, "Primary Jaw Crusher", "ThyssenKrupp KB 54-75", "120 TPH", "2021-04-05", "running"},
		{4, "Secondary Cone Crusher", "ThyssenKrupp Kubria", "100 TPH", "2021-04-05", "running"},
		{4, "Vibrating Screen", "ThyssenKrupp 2-Deck", "2 Deck", "2021-04-05", "running"},
		{5, "Primary Jaw Crusher", "Metso C106", "160 TPH", "2018-12-01", "running"},
		{5, "Secondary Cone Crusher", "Metso HP200", "140 TPH", "2018-12-01", "running"},
		{5, "VSI Crusher", "Metro VSI-400", "90 TPH", "2019-06-10", "running"},
		{5, "Vibrating Screen", "Metso 3-Deck", "3 Deck", "2018-12-01", "running"},
		{6, "Primary Jaw Crusher", "Terex J1160", "100 TPH", "2022-03-10", "running"},
		{6, "Secondary Cone Crusher", "Terex TC1000", "90 TPH", "2022-03-10", "running"},
		{6, "Vibrating Screen", "Terex 5x16", "3 Deck", "2022-03-10", "running"},
	}
	for i, e := range equipData {
		lastService := now.Add(-time.Duration(30+rng.Intn(60)) * 24 * time.Hour).Format(time.RFC3339)
		nextService := now.Add(time.Duration(30+rng.Intn(60)) * 24 * time.Hour).Format(time.RFC3339)
		created := now.Add(-time.Duration(365+rng.Intn(730)) * 24 * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO equipment(id, plant_id, type, make_model, capacity, installed_at, last_service_at, next_service_due, status, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			i+1, e.plantID, e.eqType, e.makeModel, e.capacity, e.installed, lastService, nextService, e.status, created); err != nil {
			return err
		}
	}

	customerNames := []string{
		"Brahmaputra Construction Co", "North East Infra Ltd", "Assam PWD Division 1", "NHAI Project Unit Guwahati",
		"Deka Builders", "Baruah Construction", "Gogoi Infrastructure", "Hazarika & Sons",
		"Patel Roadways", "Singh Contractors", "Sharma Construction Co", "Verma Builders",
		"Arunachal PWD", "Meghalaya Urban Dev", "Nagaland PWD", "Manipur Irrigation",
		"Railway Construction NFR", "Brahmaputra Bridge Corp", "Assam Gas Cracker Project", "Numaligarh Refinery Expansion",
		"Guwahati Metro Rail", "Silchar Airport Expansion", "Tezpur University Campus", "Dibrugarh Medical College",
		"Jorhat Engineering College", "Nagaon District Roads", "Goalpara River Embankment", "Kokrajhar BTC Roads",
		"Private Builder - Sharma", "Private Builder - Singh", "Private Builder - Gupta",
	}
	for i := 0; i < 30; i++ {
		name := customerNames[i]
		cType := customerTypes[rng.Intn(len(customerTypes))]
		contact := firstNames[rng.Intn(len(firstNames))] + " " + lastNames[rng.Intn(len(lastNames))]
		city := cities[rng.Intn(len(cities))]
		state := states[rng.Intn(len(states))]
		phone := fmt.Sprintf("9%d%d%d%d%d%d%d%d%d", rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10))
		email := fmt.Sprintf("%s@%s.in", strings.ToLower(strings.ReplaceAll(name, " ", "")), strings.ToLower(strings.ReplaceAll(city, " ", "")))
		gstin := fmt.Sprintf("18%s%s%d%d%d%d%d%d%d%d", strings.ToUpper(name[:3]), strings.ToUpper(city[:2]), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10))
		created := now.Add(-time.Duration(30+rng.Intn(365)) * 24 * time.Hour).Format(time.RFC3339)
		if err := exec(`INSERT INTO customers(id, name, type, contact_person, phone, email, address, city, state, gstin, active, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?)`,
			i+1, name, cType, contact, phone, email, fmt.Sprintf("%s, %s", city, state), city, state, gstin, created); err != nil {
			return err
		}
	}

	shifts := []string{"A", "B", "C"}
	startDate := now.Add(-365 * 24 * time.Hour)
	for plantID := 1; plantID <= 6; plantID++ {
		for productID := 1; productID <= 10; productID++ {
			for d := 0; d < 365; d++ {
				date := startDate.Add(time.Duration(d) * 24 * time.Hour)
				if rng.Intn(10) == 0 {
					continue
				}
				for _, shift := range shifts {
					if rng.Intn(5) == 0 {
						continue
					}
					baseCapacity := float64(plants[plantID-1].capacityTPH)
					efficiency := 0.65 + rng.Float64()*0.25
					operatingHours := 6.0 + rng.Float64()*2.0
					qty := baseCapacity * efficiency * operatingHours * (0.8 + rng.Float64()*0.4)
					downtime := rng.Float64() * 2.0
					var reason any
					if downtime > 1.0 {
						reasons := []string{"Screen mesh change", "Crusher jaw plate change", "Belt tracking", "Motor overload", "Power fluctuation", "Raw material shortage", "Preventive maintenance"}
						reason = reasons[rng.Intn(len(reasons))]
					}
					created := date.Add(time.Duration(rng.Intn(24)) * time.Hour).Format(time.RFC3339)
					if err := exec(`INSERT INTO production(plant_id, product_id, date, quantity_mt, operating_hours, downtime_hours, downtime_reason, shift, created_at)
						VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
						plantID, productID, date.Format("2006-01-02"), qty, operatingHours, downtime, reason, shift, created); err != nil {
						return err
					}
				}
			}
		}
	}

	orderID := 0
	dispatchID := 0
	
	for i := 1; i <= 500; i++ {
		customerID := 1 + rng.Intn(30)
		productID := 1 + rng.Intn(10)
		plantID := 1 + rng.Intn(6)
		qty := 50.0 + float64(rng.Intn(500))
		product := products[productID-1]
		rate := product.minPrice + rng.Float64()*(product.maxPrice-product.minPrice)
		total := qty * rate
		orderDate := now.Add(-time.Duration(rng.Intn(180)) * 24 * time.Hour)
		deliveryDate := orderDate.Add(time.Duration(1+rng.Intn(14)) * 24 * time.Hour)
		statuses := []string{"pending", "confirmed", "partial", "completed", "cancelled"}
		status := statuses[rng.Intn(len(statuses))]
		orderNo := fmt.Sprintf("SO-%d-%05d", orderDate.Year(), i)
		created := orderDate.Format(time.RFC3339)
		if err := exec(`INSERT INTO sales_orders(id, customer_id, product_id, plant_id, order_no, quantity_mt, rate_per_mt, total_amount, order_date, delivery_date, status, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			i, customerID, productID, plantID, orderNo, qty, rate, total, orderDate.Format("2006-01-02"), deliveryDate.Format("2006-01-02"), status, created); err != nil {
			return err
		}
		orderID = i

		if status != "pending" && status != "cancelled" {
			dispatchesCount := 1 + rng.Intn(5)
			dispatchedQty := 0.0
			for d := 0; d < dispatchesCount; d++ {
				dispatchID++
				truckType := truckTypes[rng.Intn(len(truckTypes))]
				truckNo := fmt.Sprintf("AS-%02d-%s-%04d", 1+rng.Intn(30), []string{"B", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M", "N", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}[rng.Intn(23)], 1000+rng.Intn(9000))
				driverName := firstNames[rng.Intn(len(firstNames))] + " " + lastNames[rng.Intn(len(lastNames))]
				driverPhone := fmt.Sprintf("9%d%d%d%d%d%d%d%d%d", rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10))
				grossWeight := 25.0 + rng.Float64()*20.0
				tareWeight := 8.0 + rng.Float64()*4.0
				netWeight := grossWeight - tareWeight
				dispatchDate := orderDate.Add(time.Duration(1+rng.Intn(10)) * 24 * time.Hour)
				dispatchTime := fmt.Sprintf("%02d:%02d", 6+rng.Intn(14), rng.Intn(60))
				destination := cities[rng.Intn(len(cities))]
				ewayBill := fmt.Sprintf("%d%d%d%d%d%d%d%d%d%d%d%d", rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10), rng.Intn(10))
				dispatchedQty += netWeight
				if err := exec(`INSERT INTO dispatches(id, sales_order_id, truck_no, truck_type, driver_name, driver_phone, weighbridge_in, weighbridge_out, net_weight_mt, dispatch_date, dispatch_time, destination, eway_bill_no, status, created_at)
					VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'dispatched', ?)`,
					dispatchID, orderID, truckNo, truckType, driverName, driverPhone, grossWeight, tareWeight, netWeight, dispatchDate.Format("2006-01-02"), dispatchTime, destination, ewayBill, dispatchDate.Format(time.RFC3339)); err != nil {
					return err
				}
			}
			if dispatchedQty >= qty*0.95 {
				if err := exec(`UPDATE sales_orders SET status = 'completed' WHERE id = ?`, orderID); err != nil {
					return err
				}
			} else if dispatchedQty > 0 {
				if err := exec(`UPDATE sales_orders SET status = 'partial' WHERE id = ?`, orderID); err != nil {
					return err
				}
			}
		}
	}

	expenseCategoriesList := []string{"Diesel", "Explosives", "Drill Bits", "Crusher Wear Parts", "Screen Mesh", "Lubricants", "Electricity", "Labor Wages", "Equipment Hire", "Transport", "Maintenance", "Royalty", "Pollution Control", "Office Expenses"}
	vendors := []string{"Indian Oil", "Bharat Petroleum", "Hindustan Petroleum", "Solar Explosives", "IDE India", "Sandvik Mining", "Metso India", "Terex India", "Propel Industries", "ThyssenKrupp", "L&T Construction", "JCB India", "Tata Motors", "Ashok Leyland", "Local Vendor"}
	paymentModes := []string{"cash", "bank_transfer", "upi", "cheque"}
	for i := 1; i <= 1200; i++ {
		plantID := 1 + rng.Intn(6)
		category := expenseCategoriesList[rng.Intn(len(expenseCategoriesList))]
		vendor := vendors[rng.Intn(len(vendors))]
		amount := 5000.0 + rng.Float64()*500000.0
		expenseDate := now.Add(-time.Duration(rng.Intn(180)) * 24 * time.Hour)
		billNo := fmt.Sprintf("BILL-%d-%06d", expenseDate.Year(), i)
		paymentMode := paymentModes[rng.Intn(len(paymentModes))]
		created := expenseDate.Format(time.RFC3339)
		if err := exec(`INSERT INTO expenses(id, plant_id, category, description, amount, expense_date, bill_no, vendor, payment_mode, status, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, 'paid', ?)`,
			i, plantID, category, fmt.Sprintf("%s - %s", category, vendor), amount, expenseDate.Format("2006-01-02"), billNo, vendor, paymentMode, created); err != nil {
			return err
		}
		
	}

	for mineID := 1; mineID <= 5; mineID++ {
		for period := 0; period < 12; period++ {
			periodStart := now.Add(-time.Duration(period+1) * 30 * 24 * time.Hour).AddDate(0, 0, 1)
			periodStart = time.Date(periodStart.Year(), periodStart.Month(), 1, 0, 0, 0, 0, time.UTC)
			periodEnd := periodStart.AddDate(0, 1, -1)
			qty := 5000.0 + rng.Float64()*50000.0
			rate := 45.0 + rng.Float64()*35.0
			amount := qty * rate
			paidAt := periodEnd.Add(time.Duration(rng.Intn(10)) * 24 * time.Hour)
			challanNo := fmt.Sprintf("ROY-%d-%02d-%04d", paidAt.Year(), int(paidAt.Month()), mineID)
			created := paidAt.Format(time.RFC3339)
			if err := exec(`INSERT INTO royalty_payments(id, mine_id, period_from, period_to, quantity_mt, rate_per_mt, amount, paid_at, challan_no, created_at)
				VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				(mineID-1)*12+period+1, mineID, periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"), qty, rate, amount, paidAt.Format("2006-01-02"), challanNo, created); err != nil {
				return err
			}
		}
	}

	indexes := []string{
		`CREATE INDEX idx_production_plant_date ON production(plant_id, date)`,
		`CREATE INDEX idx_production_product_date ON production(product_id, date)`,
		`CREATE INDEX idx_stock_plant_product ON stock(plant_id, product_id)`,
		`CREATE INDEX idx_sales_orders_customer ON sales_orders(customer_id)`,
		`CREATE INDEX idx_sales_orders_date ON sales_orders(order_date)`,
		`CREATE INDEX idx_sales_orders_status ON sales_orders(status)`,
		`CREATE INDEX idx_dispatches_order ON dispatches(sales_order_id)`,
		`CREATE INDEX idx_dispatches_date ON dispatches(dispatch_date)`,
		`CREATE INDEX idx_expenses_plant_date ON expenses(plant_id, expense_date)`,
		`CREATE INDEX idx_expenses_category ON expenses(category)`,
		`CREATE INDEX idx_royalty_mine_period ON royalty_payments(mine_id, period_from)`,
		`CREATE INDEX idx_equipment_plant ON equipment(plant_id)`,
		`CREATE INDEX idx_equipment_status ON equipment(status)`,
	}
	for _, stmt := range indexes {
		if err := exec(stmt); err != nil {
			return err
		}
	}
	return tx.Commit()
}