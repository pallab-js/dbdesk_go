package dashboard

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"dbdesk/internal/model"
	"dbdesk/internal/sqlite"
	"dbdesk/internal/store"
)

// Build assembles the dashboard from cheap database metadata plus the app's
// own 24-hour query history. Slow counts stop when the caller's context ends.
func Build(ctx context.Context, db *sqlite.DB, st *store.Store, connectionID int64) (model.DashboardData, error) {
	now := time.Now()
	out := model.DashboardData{
		GeneratedAt: now.UnixMilli(),
		Tables:      []model.TableStat{},
		ColumnTypes: []model.TypeCount{},
		Activity:    make([]model.ActivityBucket, 24),
		Latency:     make([]model.LatencyPoint, 24),
		Recent:      []model.HistoryEntry{},
	}

	ds, objects, err := db.DatabaseStats(ctx)
	if err != nil {
		return out, fmt.Errorf("read database stats: %w", err)
	}
	out.Database = ds
	out.Objects = objects

	schema := db.Schema()
	typeTotals := map[string]int{}
	tableNames := make([]string, 0, len(schema.Tables))
	for _, t := range schema.Tables {
		for _, c := range t.Columns {
			typeTotals[normalizeType(c.DeclType)]++
		}
		if t.Type == "table" {
			tableNames = append(tableNames, t.Name)
		}
	}
	for name, n := range typeTotals {
		out.ColumnTypes = append(out.ColumnTypes, model.TypeCount{Type: name, Count: n})
	}
	sort.Slice(out.ColumnTypes, func(i, j int) bool {
		if out.ColumnTypes[i].Count == out.ColumnTypes[j].Count {
			return out.ColumnTypes[i].Type < out.ColumnTypes[j].Type
		}
		return out.ColumnTypes[i].Count > out.ColumnTypes[j].Count
	})

	for i, name := range tableNames {
		if i >= 100 {
			out.Partial = true
			break
		}
		if ctx.Err() != nil {
			out.Partial = true
			break
		}
		stat := model.TableStat{Name: name}
		countCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		n, err := db.CountTable(countCtx, name, 5*time.Minute)
		cancel()
		if err != nil {
			stat.Error = "count timed out"
			out.Partial = true
		} else {
			stat.Rows = n
		}
		out.Tables = append(out.Tables, stat)
	}
	sort.Slice(out.Tables, func(i, j int) bool {
		if out.Tables[i].Rows == out.Tables[j].Rows {
			return out.Tables[i].Name < out.Tables[j].Name
		}
		return out.Tables[i].Rows > out.Tables[j].Rows
	})
	_ = st.UpsertTableStats(connectionID, out.Tables)

	hourStart := now.Truncate(time.Hour).Add(-23 * time.Hour)
	history, err := st.HistorySince(hourStart.UnixMilli())
	if err != nil {
		return out, fmt.Errorf("read query history: %w", err)
	}

	activity := make([]model.ActivityBucket, 24)
	latencies := make([][]int64, 24)
	for i := range activity {
		start := hourStart.Add(time.Duration(i) * time.Hour)
		activity[i] = model.ActivityBucket{Start: start.UnixMilli(), Label: start.Format("15:04")}
	}
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var allDurations []int64
	out.Totals.Queries = len(history)
	for _, h := range history {
		if h.CreatedAt >= dayStart.UnixMilli() {
			out.Totals.Today++
		}
		if !h.Success {
			out.Totals.Errors++
		}
		allDurations = append(allDurations, h.DurationMs)
		idx := int(time.UnixMilli(h.CreatedAt).Sub(hourStart) / time.Hour)
		if idx < 0 || idx > 23 {
			continue
		}
		if h.Success {
			activity[idx].Success++
		} else {
			activity[idx].Error++
		}
		latencies[idx] = append(latencies[idx], h.DurationMs)
	}
	out.Activity = activity
	for i := range activity {
		out.Latency[i] = model.LatencyPoint{
			Start: activity[i].Start,
			Label: activity[i].Label,
			P50:   percentile(latencies[i], 50),
			P95:   percentile(latencies[i], 95),
		}
	}
	if out.Totals.Queries > 0 {
		out.Totals.ErrorRate = float64(out.Totals.Errors) / float64(out.Totals.Queries) * 100
		out.Totals.P50 = percentile(allDurations, 50)
		out.Totals.P95 = percentile(allDurations, 95)
	}

	recent := history
	if len(recent) > 12 {
		recent = recent[len(recent)-12:]
	}
	for i := len(recent) - 1; i >= 0; i-- {
		out.Recent = append(out.Recent, recent[i])
	}
	return out, nil
}

func normalizeType(decl string) string {
	t := strings.ToUpper(strings.TrimSpace(decl))
	switch {
	case t == "":
		return "ANY"
	case strings.Contains(t, "INT"):
		return "INTEGER"
	case strings.Contains(t, "CHAR"), strings.Contains(t, "CLOB"), strings.Contains(t, "TEXT"):
		return "TEXT"
	case strings.Contains(t, "BLOB"):
		return "BLOB"
	case strings.Contains(t, "REAL"), strings.Contains(t, "FLOA"), strings.Contains(t, "DOUB"):
		return "REAL"
	case strings.Contains(t, "NUM"), strings.Contains(t, "DEC"):
		return "NUMERIC"
	default:
		return "OTHER"
	}
}

func percentile(values []int64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(math.Ceil(float64(len(sorted))*p/100)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return float64(sorted[idx])
}
