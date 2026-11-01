package model

// ConnectionInfo is a saved database profile plus live connection state.
type ConnectionInfo struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	ReadOnly     bool   `json:"readOnly"`
	LastOpenedAt int64  `json:"lastOpenedAt"`
}

// Settings are user preferences persisted in the app database.
type Settings struct {
	RowLimit       int  `json:"rowLimit"`
	TimeoutSeconds int  `json:"timeoutSeconds"`
	HistoryEnabled bool `json:"historyEnabled"`
	RetentionDays  int  `json:"retentionDays"`
	AutoRefresh    bool `json:"autoRefresh"`
}

func DefaultSettings() Settings {
	return Settings{
		RowLimit:       1000,
		TimeoutSeconds: 30,
		HistoryEnabled: true,
		RetentionDays:  30,
		AutoRefresh:    true,
	}
}

type AppState struct {
	Connection  *ConnectionInfo  `json:"connection"`
	Recent      []ConnectionInfo `json:"recent"`
	Settings    Settings         `json:"settings"`
	AppDataDir  string           `json:"appDataDir"`
	DemoPath    string           `json:"demoPath"`
	Version     string           `json:"version"`
	Initialized bool             `json:"initialized"`
}

type Column struct {
	Name     string `json:"name"`
	DeclType string `json:"declType"`
	NotNull  bool   `json:"notNull"`
	Default  string `json:"default"`
	PK       int    `json:"pk"`
	Hidden   int    `json:"hidden"`
}

type ForeignKey struct {
	From     string `json:"from"`
	Table    string `json:"table"`
	To       string `json:"to"`
	OnUpdate string `json:"onUpdate"`
	OnDelete string `json:"onDelete"`
}

type IndexInfo struct {
	Name    string   `json:"name"`
	Unique  bool     `json:"unique"`
	Columns []string `json:"columns"`
	Partial bool     `json:"partial"`
}

type Table struct {
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	Columns  []Column     `json:"columns"`
	FKs      []ForeignKey `json:"fks"`
	Indexes  []IndexInfo  `json:"indexes"`
	DDL      string       `json:"ddl"`
	RowKey   string       `json:"rowKey"`
	Editable bool         `json:"editable"`
}

type Schema struct {
	Tables []Table `json:"tables"`
}

type ColumnMeta struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Hidden bool   `json:"hidden"`
}

type TablePageRequest struct {
	Table      string `json:"table"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	SortColumn string `json:"sortColumn"`
	SortDir    string `json:"sortDir"`
}

type TablePage struct {
	Table      string       `json:"table"`
	Columns    []ColumnMeta `json:"columns"`
	Rows       [][]any      `json:"rows"`
	KeyIndexes []int        `json:"keyIndexes"`
	KeyNames   []string     `json:"keyNames"`
	Total      int64        `json:"total"`
	CountError string       `json:"countError"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	HasNext    bool         `json:"hasNext"`
	Editable   bool         `json:"editable"`
}

type QueryResult struct {
	Columns      []ColumnMeta `json:"columns"`
	Rows         [][]any      `json:"rows"`
	RowCount     int          `json:"rowCount"`
	Truncated    bool         `json:"truncated"`
	DurationMs   int64        `json:"durationMs"`
	Affected     int64        `json:"affected"`
	Kind         string       `json:"kind"`
	Statement    string       `json:"statement"`
	IsRead       bool         `json:"isRead"`
	LastInsertID int64        `json:"lastInsertId"`
}

type KeyValue struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type RowInsertRequest struct {
	Table  string     `json:"table"`
	Values []KeyValue `json:"values"`
}

type RowUpdateRequest struct {
	Table  string     `json:"table"`
	Keys   []KeyValue `json:"keys"`
	Values []KeyValue `json:"values"`
}

type RowDeleteRequest struct {
	Table string     `json:"table"`
	Keys  []KeyValue `json:"keys"`
}

type MutationResult struct {
	Affected     int64 `json:"affected"`
	LastInsertID int64 `json:"lastInsertId"`
}

type DatabaseStats struct {
	Path        string `json:"path"`
	SizeBytes   int64  `json:"sizeBytes"`
	PageSize    int64  `json:"pageSize"`
	PageCount   int64  `json:"pageCount"`
	FreePages   int64  `json:"freePages"`
	JournalMode string `json:"journalMode"`
	Encoding    string `json:"encoding"`
	ReadOnly    bool   `json:"readOnly"`
}

type ObjectCounts struct {
	Tables   int `json:"tables"`
	Views    int `json:"views"`
	Indexes  int `json:"indexes"`
	Triggers int `json:"triggers"`
}

type TableStat struct {
	Name   string `json:"name"`
	Rows   int64  `json:"rows"`
	Approx bool   `json:"approx"`
	Error  string `json:"error"`
}

type TypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ActivityBucket struct {
	Start   int64  `json:"start"`
	Label   string `json:"label"`
	Success int    `json:"success"`
	Error   int    `json:"error"`
}

type LatencyPoint struct {
	Start int64   `json:"start"`
	Label string  `json:"label"`
	P50   float64 `json:"p50"`
	P95   float64 `json:"p95"`
}

type TelemetryTotals struct {
	Queries   int     `json:"queries"`
	Errors    int     `json:"errors"`
	ErrorRate float64 `json:"errorRate"`
	P50       float64 `json:"p50"`
	P95       float64 `json:"p95"`
	Today     int     `json:"today"`
}

type HistoryEntry struct {
	ID         int64  `json:"id"`
	SQL        string `json:"sql"`
	Kind       string `json:"kind"`
	DurationMs int64  `json:"durationMs"`
	RowCount   int    `json:"rowCount"`
	Success    bool   `json:"success"`
	Error      string `json:"error"`
	CreatedAt  int64  `json:"createdAt"`
}

type SavedQuery struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SQL       string `json:"sql"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type DashboardData struct {
	GeneratedAt int64            `json:"generatedAt"`
	Database    DatabaseStats    `json:"database"`
	Objects     ObjectCounts     `json:"objects"`
	Tables      []TableStat      `json:"tables"`
	ColumnTypes []TypeCount      `json:"columnTypes"`
	Activity    []ActivityBucket `json:"activity"`
	Latency     []LatencyPoint   `json:"latency"`
	Totals      TelemetryTotals  `json:"totals"`
	Recent      []HistoryEntry   `json:"recent"`
	Partial     bool             `json:"partial"`
}
