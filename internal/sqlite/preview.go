package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dbdesk/internal/model"
)

// PreviewTable reads one page of a table using PK (or rowid) keys for edits.
func (d *DB) PreviewTable(ctx context.Context, req model.TablePageRequest) (model.TablePage, error) {
	t, ok := d.table(req.Table)
	if !ok {
		return model.TablePage{}, fmt.Errorf("unknown table %q", req.Table)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 100
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	cols := visibleColumns(t)
	selects := make([]string, 0, len(cols)+1)
	selects = append(selects, quoteIdent(t.Name)+".*")
	if t.RowKey != "" {
		alias := "__rowid_key__"
		for _, c := range cols {
			if c.Name == alias {
				alias = "__rowid_key_2__"
				break
			}
		}
		selects = append(selects, quoteIdent(t.RowKey)+" AS "+quoteIdent(alias))
	}

	order := ""
	if req.SortColumn != "" {
		if _, ok := columnByName(t, req.SortColumn); !ok {
			return model.TablePage{}, fmt.Errorf("unknown sort column %q", req.SortColumn)
		}
		dir := "ASC"
		if strings.EqualFold(req.SortDir, "desc") {
			dir = "DESC"
		}
		order = " ORDER BY " + quoteIdent(req.SortColumn) + " " + dir
	}

	offset := (req.Page - 1) * req.PageSize
	query := "SELECT " + strings.Join(selects, ", ") + " FROM " + quoteIdent(t.Name) + order + " LIMIT ? OFFSET ?"
	rows, err := d.sql.QueryContext(ctx, query, req.PageSize, offset)
	if err != nil {
		return model.TablePage{}, cleanError(err)
	}
	defer rows.Close()

	columnMeta, data, _, _, err := scanRows(rows, req.PageSize)
	rows.Close()
	if err != nil {
		return model.TablePage{}, cleanError(err)
	}

	keyIndexes := make([]int, 0, 2)
	keyNames := pkColumns(t)
	if t.RowKey != "" {
		keyNames = []string{t.RowKey}
		keyIndexes = append(keyIndexes, len(columnMeta)-1)
	} else {
		for _, name := range keyNames {
			for i, col := range columnMeta {
				if col.Name == name {
					keyIndexes = append(keyIndexes, i)
					break
				}
			}
		}
	}
	if t.RowKey != "" && len(columnMeta) > 0 {
		columnMeta[len(columnMeta)-1].Hidden = true
	}

	total := int64(-1)
	countErr := ""
	countCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	n, err := d.CountTable(countCtx, t.Name, time.Minute)
	cancel()
	if err != nil {
		countErr = cleanError(err).Error()
	} else {
		total = n
	}

	hasNext := len(data) == req.PageSize
	if total >= 0 {
		hasNext = int64(offset+len(data)) < total
	}

	// Keep a copy for CSV export with hidden key columns removed.
	exportIdx := make([]int, 0, len(columnMeta))
	exportCols := make([]model.ColumnMeta, 0, len(columnMeta))
	for i, c := range columnMeta {
		if c.Hidden {
			continue
		}
		exportIdx = append(exportIdx, i)
		exportCols = append(exportCols, c)
	}
	exportRows := make([][]any, 0, len(data))
	for _, row := range data {
		scoped := make([]any, 0, len(exportIdx))
		for _, i := range exportIdx {
			scoped = append(scoped, row[i])
		}
		exportRows = append(exportRows, scoped)
	}
	d.lastMu.Lock()
	d.last = &model.QueryResult{Columns: exportCols, Rows: exportRows, RowCount: len(exportRows), Kind: "select", IsRead: true}
	d.lastMu.Unlock()

	return model.TablePage{
		Table:      t.Name,
		Columns:    columnMeta,
		Rows:       data,
		KeyIndexes: keyIndexes,
		KeyNames:   keyNames,
		Total:      total,
		CountError: countErr,
		Page:       req.Page,
		PageSize:   req.PageSize,
		HasNext:    hasNext,
		Editable:   t.Editable && !d.readOnly,
	}, nil
}
