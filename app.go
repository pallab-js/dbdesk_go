package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"dbdesk/internal/config"
	"dbdesk/internal/dashboard"
	"dbdesk/internal/demo"
	"dbdesk/internal/model"
	"dbdesk/internal/sqlite"
	"dbdesk/internal/store"
)

const appVersion = "0.1.0"

// App is the Wails-bound facade. All exported methods become frontend bindings.
type App struct {
	ctx      context.Context
	paths    config.Paths
	store    *store.Store
	db       *sqlite.DB
	conn     *model.ConnectionInfo
	settings model.Settings
	logger   *log.Logger
	logFile  *os.File
	initErr  error

	mu sync.RWMutex
}

func NewApp() *App {
	return &App{settings: model.DefaultSettings()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	paths, err := config.Ensure()
	if err != nil {
		a.initErr = err
		return
	}
	a.paths = paths
	if f, err := os.OpenFile(paths.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		a.logFile = f
		a.logger = log.New(f, "", log.LstdFlags)
	} else {
		a.logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	st, err := store.Open(paths.DBPath)
	if err != nil {
		a.initErr = err
		a.logf("open app store: %v", err)
		return
	}
	a.store = st
	settings, err := st.Settings()
	if err != nil {
		a.logf("read settings: %v", err)
		settings = model.DefaultSettings()
	}
	a.settings = normalizeSettings(settings)
	_ = st.SaveSettings(a.settings)
	_ = st.PruneHistory(a.settings.RetentionDays, 20000)
	a.logf("started %s, data=%s", appVersion, paths.DataDir)
}

func (a *App) shutdown(context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db != nil {
		_ = a.db.Close()
		a.db = nil
	}
	if a.store != nil {
		_ = a.store.Close()
		a.store = nil
	}
	if a.logFile != nil {
		_ = a.logFile.Close()
	}
}

func (a *App) logf(format string, args ...any) {
	if a.logger != nil {
		a.logger.Printf(format, args...)
	}
}

func (a *App) ready() error {
	if a.initErr != nil {
		return fmt.Errorf("app initialization failed: %w", a.initErr)
	}
	if a.store == nil {
		return errors.New("app store is not ready")
	}
	return nil
}

func (a *App) emit(event string) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event)
	}
}

// GetState returns the connection, recent profiles and preferences.
func (a *App) GetState() (model.AppState, error) {
	if err := a.ready(); err != nil {
		return model.AppState{}, err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	recent, err := a.store.Connections(20)
	if err != nil {
		return model.AppState{}, err
	}
	out := model.AppState{
		Recent:      recent,
		Settings:    a.settings,
		AppDataDir:  a.paths.DataDir,
		DemoPath:    a.paths.DemoPath,
		Version:     appVersion,
		Initialized: true,
	}
	if a.conn != nil {
		c := *a.conn
		out.Connection = &c
	}
	return out, nil
}

func (a *App) PickDatabaseFile() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	defaultDir, _ := os.UserHomeDir()
	if a.conn != nil {
		defaultDir = filepath.Dir(a.conn.Path)
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Open SQLite database",
		DefaultDirectory: defaultDir,
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite databases", Pattern: "*.db;*.sqlite;*.sqlite3;*.db3"},
			{DisplayName: "All files", Pattern: "*"},
		},
	})
}

// OpenDatabase opens a database and remembers it in the recent list.
func (a *App) OpenDatabase(path string, readOnly bool) (model.ConnectionInfo, error) {
	if err := a.ready(); err != nil {
		return model.ConnectionInfo{}, err
	}
	abs, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return model.ConnectionInfo{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return model.ConnectionInfo{}, fmt.Errorf("cannot read %s: %w", filepath.Base(abs), err)
	}
	if info.IsDir() {
		return model.ConnectionInfo{}, errors.New("choose a database file, not a folder")
	}

	db, err := sqlite.Open(abs, readOnly)
	if err != nil {
		return model.ConnectionInfo{}, err
	}
	conn, err := a.store.UpsertConnection(abs, readOnly)
	if err != nil {
		_ = db.Close()
		return model.ConnectionInfo{}, err
	}

	a.mu.Lock()
	if a.db != nil {
		_ = a.db.Close()
	}
	a.db = db
	a.conn = &conn
	a.mu.Unlock()
	a.emit("db:changed")
	a.logf("opened %s (read-only=%v)", abs, readOnly)
	return conn, nil
}

func (a *App) CloseDatabase() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db != nil {
		_ = a.db.Close()
		a.db = nil
		a.conn = nil
	}
	a.emit("db:changed")
	return nil
}

func (a *App) CreateDemoDatabase() (model.ConnectionInfo, error) {
	if err := a.ready(); err != nil {
		return model.ConnectionInfo{}, err
	}
	if err := demo.Seed(a.paths.DemoPath); err != nil {
		return model.ConnectionInfo{}, fmt.Errorf("create demo database: %w", err)
	}
	return a.OpenDatabase(a.paths.DemoPath, false)
}

func (a *App) ResetDemoDatabase() (model.ConnectionInfo, error) {
	if err := a.ready(); err != nil {
		return model.ConnectionInfo{}, err
	}
	a.mu.Lock()
	if a.db != nil && a.conn != nil && a.conn.Path == a.paths.DemoPath {
		_ = a.db.Close()
		a.db = nil
		a.conn = nil
	}
	a.mu.Unlock()
	for _, p := range []string{a.paths.DemoPath, a.paths.DemoPath + "-wal", a.paths.DemoPath + "-shm"} {
		_ = os.Remove(p)
	}
	return a.CreateDemoDatabase()
}

func (a *App) RecentConnections() ([]model.ConnectionInfo, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.Connections(20)
}

func (a *App) ForgetConnection(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	a.mu.Lock()
	if a.conn != nil && a.conn.ID == id {
		if a.db != nil {
			_ = a.db.Close()
		}
		a.db = nil
		a.conn = nil
	}
	a.mu.Unlock()
	if err := a.store.ForgetConnection(id); err != nil {
		return err
	}
	a.emit("db:changed")
	return nil
}

func (a *App) Schema() (model.Schema, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.Schema{}, errors.New("open a database first")
	}
	return a.db.Schema(), nil
}

func (a *App) RefreshSchema() (model.Schema, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.Schema{}, errors.New("open a database first")
	}
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	return a.db.RefreshSchema(ctx)
}

func (a *App) PreviewTable(req model.TablePageRequest) (model.TablePage, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.TablePage{}, errors.New("open a database first")
	}
	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.settings.TimeoutSeconds)*time.Second)
	defer cancel()
	return a.db.PreviewTable(ctx, req)
}

func (a *App) RunQuery(sqlText string) (model.QueryResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.QueryResult{}, errors.New("open a database first")
	}
	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.settings.TimeoutSeconds)*time.Second)
	defer cancel()
	res, err := a.db.Execute(ctx, sqlText, a.settings.RowLimit, time.Duration(a.settings.TimeoutSeconds)*time.Second)

	if a.settings.HistoryEnabled {
		historySQL := sqlText
		if len(historySQL) > 10000 {
			historySQL = historySQL[:10000] + "…"
		}
		errText := ""
		if err != nil {
			errText = err.Error()
		}
		connID := int64(0)
		if a.conn != nil {
			connID = a.conn.ID
		}
		if logErr := a.store.AddHistory(connID, historySQL, res.Kind, res.DurationMs, res.RowCount, err == nil, errText); logErr != nil {
			a.logf("store query history: %v", logErr)
		}
	}
	if err == nil && !res.IsRead {
		if kind := res.Kind; kind == "create" || kind == "drop" || kind == "alter" {
			if _, refreshErr := a.db.RefreshSchema(ctx); refreshErr != nil {
				a.logf("refresh schema after DDL: %v", refreshErr)
			}
		}
		a.emit("db:changed")
	}
	if err != nil {
		return model.QueryResult{}, err
	}
	return res, nil
}

// ExplainQuery runs EXPLAIN QUERY PLAN for the given SQL.
func (a *App) ExplainQuery(sqlText string) (model.QueryResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.QueryResult{}, errors.New("open a database first")
	}
	trimmed := strings.TrimSpace(sqlText)
	if trimmed == "" {
		return model.QueryResult{}, errors.New("write a query first")
	}
	stmt := trimmed
	upper := strings.ToUpper(stmt)
	if !strings.HasPrefix(upper, "EXPLAIN") {
		stmt = "EXPLAIN QUERY PLAN " + stmt
	}
	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.settings.TimeoutSeconds)*time.Second)
	defer cancel()
	res, err := a.db.Execute(ctx, stmt, a.settings.RowLimit, time.Duration(a.settings.TimeoutSeconds)*time.Second)

	if a.settings.HistoryEnabled {
		historySQL := stmt
		if len(historySQL) > 10000 {
			historySQL = historySQL[:10000] + "…"
		}
		errText := ""
		if err != nil {
			errText = err.Error()
		}
		connID := int64(0)
		if a.conn != nil {
			connID = a.conn.ID
		}
		if logErr := a.store.AddHistory(connID, historySQL, res.Kind, res.DurationMs, res.RowCount, err == nil, errText); logErr != nil {
			a.logf("store query history: %v", logErr)
		}
	}
	if err != nil {
		return model.QueryResult{}, err
	}
	return res, nil
}

func (a *App) InsertRow(req model.RowInsertRequest) (model.MutationResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.MutationResult{}, errors.New("open a database first")
	}
	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.settings.TimeoutSeconds)*time.Second)
	defer cancel()
	res, err := a.db.InsertRow(ctx, req)
	if err == nil {
		a.emit("db:changed")
	}
	return res, err
}

func (a *App) UpdateRow(req model.RowUpdateRequest) (model.MutationResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.MutationResult{}, errors.New("open a database first")
	}
	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.settings.TimeoutSeconds)*time.Second)
	defer cancel()
	res, err := a.db.UpdateRow(ctx, req)
	if err == nil {
		a.emit("db:changed")
	}
	return res, err
}

func (a *App) DeleteRow(req model.RowDeleteRequest) (model.MutationResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.MutationResult{}, errors.New("open a database first")
	}
	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.settings.TimeoutSeconds)*time.Second)
	defer cancel()
	res, err := a.db.DeleteRow(ctx, req)
	if err == nil {
		a.emit("db:changed")
	}
	return res, err
}

func (a *App) ExportLastResultCSV() (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return "", errors.New("open a database first")
	}
	if !a.db.HasExport() {
		return "", errors.New("run a query or open a table first")
	}
	name := "result"
	if a.conn != nil {
		name = strings.TrimSuffix(a.conn.Name, filepath.Ext(a.conn.Name))
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export results as CSV",
		DefaultFilename: fmt.Sprintf("%s-%s.csv", name, time.Now().Format("20060102-150405")),
		Filters:         []runtime.FileFilter{{DisplayName: "CSV", Pattern: "*.csv"}},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	if err := a.db.WriteLastCSV(f); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) DashboardStats(refresh bool) (model.DashboardData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return model.DashboardData{}, errors.New("open a database first")
	}
	if refresh {
		a.db.InvalidateCounts()
	}
	ctx, cancel := context.WithTimeout(a.ctx, 8*time.Second)
	defer cancel()
	connID := int64(0)
	if a.conn != nil {
		connID = a.conn.ID
	}
	return dashboard.Build(ctx, a.db, a.store, connID)
}

func (a *App) History(limit int) ([]model.HistoryEntry, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.History(limit)
}

func (a *App) ClearHistory() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.ClearHistory()
}

func (a *App) SavedQueries() ([]model.SavedQuery, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.SavedQueries()
}

func (a *App) SaveQuery(name, sqlText string) (model.SavedQuery, error) {
	if err := a.ready(); err != nil {
		return model.SavedQuery{}, err
	}
	name = strings.TrimSpace(name)
	sqlText = strings.TrimSpace(sqlText)
	if name == "" {
		return model.SavedQuery{}, errors.New("give the query a name")
	}
	if sqlText == "" {
		return model.SavedQuery{}, errors.New("the query is empty")
	}
	return a.store.SaveQuery(name, sqlText)
}

func (a *App) DeleteSavedQuery(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteSavedQuery(id)
}

func (a *App) GetSettings() (model.Settings, error) {
	if err := a.ready(); err != nil {
		return model.Settings{}, err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.settings, nil
}

func (a *App) SaveSettings(st model.Settings) (model.Settings, error) {
	if err := a.ready(); err != nil {
		return model.Settings{}, err
	}
	st = normalizeSettings(st)
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.store.SaveSettings(st); err != nil {
		return model.Settings{}, err
	}
	a.settings = st
	a.emit("settings:changed")
	return st, nil
}

func normalizeSettings(st model.Settings) model.Settings {
	if st.RowLimit < 100 || st.RowLimit > 10000 {
		st.RowLimit = 1000
	}
	if st.TimeoutSeconds < 5 || st.TimeoutSeconds > 120 {
		st.TimeoutSeconds = 30
	}
	if st.RetentionDays < 1 || st.RetentionDays > 365 {
		st.RetentionDays = 30
	}
	return st
}
