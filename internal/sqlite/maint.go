package sqlite

import (
	"context"
	"fmt"
	"time"

	"dbdesk/internal/model"
)

// RunMaintenance executes VACUUM, ANALYZE, REINDEX, or integrity_check.
func (d *DB) RunMaintenance(ctx context.Context, action string) (model.MaintenanceResult, error) {
	start := time.Now()
	res := model.MaintenanceResult{Action: action}
	if d.readOnly {
		res.Success = false
		res.Message = "maintenance not allowed on read-only connection"
		res.DurationMs = time.Since(start).Milliseconds()
		return res, nil
	}
	var err error
	switch action {
	case "vacuum":
		_, err = d.sql.ExecContext(ctx, "VACUUM")
	case "analyze":
		_, err = d.sql.ExecContext(ctx, "ANALYZE")
	case "reindex":
		_, err = d.sql.ExecContext(ctx, "REINDEX")
	case "integrity_check":
		var result string
		err = d.sql.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result)
		if err == nil {
			res.Message = result
			if result != "ok" {
				res.Success = false
			} else {
				res.Success = true
			}
			res.DurationMs = time.Since(start).Milliseconds()
			return res, nil
		}
	default:
		res.Success = false
		res.Message = "unknown maintenance action"
		res.DurationMs = time.Since(start).Milliseconds()
		return res, nil
	}
	if err != nil {
		res.Success = false
		res.Message = cleanError(err).Error()
	} else {
		res.Success = true
		res.Message = action + " completed"
	}
	res.DurationMs = time.Since(start).Milliseconds()
	return res, nil
}

// GetPragmas returns common pragma values.
func (d *DB) GetPragmas(ctx context.Context) ([]model.PragmaInfo, error) {
	names := []string{"journal_mode", "synchronous", "foreign_keys", "cache_size", "temp_store", "auto_vacuum", "page_size", "user_version"}
	pragmas := make([]model.PragmaInfo, 0, len(names))
	for _, name := range names {
		var v interface{}
		err := d.sql.QueryRowContext(ctx, "PRAGMA "+name).Scan(&v)
		if err != nil {
			continue
		}
		pragmas = append(pragmas, model.PragmaInfo{Name: name, Value: stringify(v)})
	}
	return pragmas, nil
}

// SetPragma sets a pragma value (read-only pragmas not attempted).
func (d *DB) SetPragma(ctx context.Context, name, value string) error {
	if d.readOnly {
		return fmt.Errorf("cannot change pragmas on read-only connection")
	}
	allowed := map[string]bool{
		"journal_mode": true, "synchronous": true, "foreign_keys": true, "cache_size": true, "temp_store": true, "auto_vacuum": true,
	}
	if !allowed[name] {
		return fmt.Errorf("pragma %s not editable", name)
	}
	_, err := d.sql.ExecContext(ctx, "PRAGMA "+name+" = "+value)
	return cleanError(err)
}

func stringify(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case int64:
		return fmtInt64(x)
	default:
		return fmt.Sprint(x)
	}
}

func fmtInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [32]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return sign + string(buf[i:])
}
