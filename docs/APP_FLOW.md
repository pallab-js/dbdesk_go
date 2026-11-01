# Application Flow

## 1. Startup

```
launch
  │
  ├─ resolve data dir ──► fail ──► fatal screen (retry)
  │
  ├─ open app.db + migrate
  ├─ load/normalise settings
  ├─ prune old history
  ├─ open log file
  │
  └─ GetState
        ├─ no connection ──► Welcome
        └─ connection ────► load schema ──► Dashboard
```

- The splash is shown until `GetState` resolves.
- Startup errors show the fatal panel with the message and a retry button.
- The app does not auto-open the last database in v1; recent entries are one click
  away on Welcome.

## 2. Open a database

### From a file

1. Click **Open database** (top bar or Welcome).
2. Native file picker opens.
3. Cancelling leaves state unchanged.
4. On selection: close previous connection, open with `rw` or `ro`, validate, save
   profile, emit `db:changed`.
5. UI reloads state and schema, switches to Dashboard, shows a success toast.

### From recent

1. Click a recent entry on Welcome or Settings.
2. The saved read-only flag is reused.
3. Same open path as above. Missing/permission-denied files produce an error toast
   and keep the current connection.

### Demo

1. Click **Demo** (top bar) or **Try the demo** (Welcome).
2. If the demo file is absent, seed it deterministically; otherwise open the
   existing file.
3. Switch to Dashboard and show a toast with the database name.

## 3. Browse a table

```
Sidebar table click / table selector
  │
  ├─ set active table, reset page/sort
  ├─ PreviewTable(page, pageSize, sort)
  │     ├─ success ──► DataGrid + total + editability badges
  │     └─ failure ──► error panel with retry
  │
  ├─ sort header ──► toggle asc/desc ──► page 1 ──► reload
  ├─ pager ─────────► previous/next ──► reload
  └─ page size ─────► 50…1000 ──► page 1 ──► reload
```

- The grid marks primary-key columns, shows types, and renders `NULL` distinctly.
- Hidden `rowid` keys never render but remain available to edit/delete.
- Views and read-only connections hide row actions.

## 4. Insert a row

1. Click **Insert row** (enabled only when editable).
2. `RowEditor` opens with blank fields; hints show type, nullability and defaults.
3. Blank untouched fields are omitted so SQLite applies defaults.
4. Tick **NULL** to send an explicit null.
5. Save → `InsertRow` → toast → close → reload page and dashboard.

Errors (constraint, type, read-only) appear inside the dialog; the dialog stays open
with the entered values.

## 5. Edit a row

1. Click the pencil icon on a row.
2. Dialog shows current values; key columns are disabled.
3. Only changed fields are sent.
4. Save with no changes shows “Nothing changed yet.”
5. Save → `UpdateRow` (key + changed values) → toast → reload.

## 6. Delete a row

1. Click the trash icon.
2. Confirmation names the table and key values.
3. Delete → `DeleteRow` → toast → reload.
4. Cancel or Escape closes without writing.

## 7. Run SQL

```
write/choose SQL
  │
  ├─ selection present ──► selection is the statement
  ├─ Ctrl/Cmd+Enter or Run
  │
  ├─ split statements
  │     ├─ 0 ──► "enter a SQL statement"
  │     └─ >1 ─► "run one statement at a time"
  │
  ├─ read-only connection + write ──► rejected
  │
  ├─ read/RETURNING ──► grid (capped rows, truncation badge)
  └─ write ──────────► affected rows / last insert id
```

- Every attempt is recorded in history when enabled.
- Errors show inline with the SQLite message; the previous result is cleared.
- **Clear** empties the editor and result.
- **Save** opens a name dialog and stores the query.
- **Export** writes the last result to CSV.

## 8. Saved queries

- Saved chips appear above the result panel.
- Click a chip to replace the editor content.
- Click the × to delete immediately (toast confirms).
- Save upserts by name.

## 9. History

```
open History
  │
  ├─ History(500) ──► list (newest first)
  ├─ filter text ───► client-side contains match
  ├─ expand ────────► SQL, error, timestamp, open-in-editor
  ├─ Clear ─────────► confirm ──► ClearHistory ──► empty state
  └─ Refresh ───────► reload
```

## 10. Dashboard

1. Loads when the view becomes active and a connection exists.
2. Reload key: view + connection id + `dataVersion` (when auto-refresh is on).
3. `DashboardStats(force)` returns storage, objects, tables, column types, activity,
   latency, totals, recent and `partial`.
4. Table chips open the data view; recent queries open in the editor.
5. If counting some tables failed, a warning lists them and `partial` is shown.

## 11. Settings

1. Edits are local until **Save settings**.
2. Save normalises ranges, persists, updates state, emits `settings:changed`.
3. Theme changes apply immediately and do not require saving.
4. **Reset demo** confirms, deletes and reseeds, then opens the demo.
5. **Forget** removes a recent profile; the open connection is unaffected.

## 12. Error and empty states

| Situation | UI |
| --- | --- |
| No database | Welcome with open/demo/recent |
| No tables | Dashboard empty state linking to the editor |
| Empty table page | “No rows to show.” inside the grid |
| Empty history | Explanatory empty state |
| Query error | Inline panel in the result area |
| Table load error | Error panel with retry |
| Startup error | Fatal screen with retry |
| Partial dashboard | Warning badge + affected table list |
| Save dialog cancelled | No toast, no file |

## 13. Shutdown

Closing the window closes the user database, app store and log file. Uncommitted
edits are never buffered by DBDesk; every successful mutation is written immediately.
