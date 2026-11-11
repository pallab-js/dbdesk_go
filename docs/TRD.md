# Technical Requirements Document (TRD)

## 1. Stack

| Layer | Choice | Version |
| --- | --- | --- |
| Desktop shell | Wails | v2.17.x |
| Backend | Go | 1.26 (module `dbdesk`) |
| SQLite driver | `modernc.org/sqlite` | pure Go, latest |
| Frontend | Svelte | 5.x |
| Language | TypeScript | 5.x |
| Build | Vite | ^7.1.0 |
| Editor | CodeMirror 6 + `@codemirror/lang-sql` | latest |
| Charts | Chart.js | latest |
| Icons | `@lucide/svelte` | latest |

No ORM, router, state library, CSS framework, or credential store is used.

## 2. Repository layout

```
.
├── app.go                 # Wails-bound facade
├── main.go                # window and asset configuration
├── go.mod
├── internal/
│   ├── config/paths.go    # OS data directory
│   ├── model/model.go     # shared JSON DTOs
│   ├── store/store.go     # app database + migrations
│   ├── sqlite/            # open, schema, preview, query, mutate, csv, split
│   ├── dashboard/         # telemetry + schema aggregation
│   └── demo/seed.go       # deterministic demo database
├── frontend/
│   ├── index.html         # theme bootstrap before bundle
│   ├── src/
│   │   ├── App.svelte
│   │   ├── main.ts
│   │   ├── components/
│   │   ├── views/
│   │   ├── lib/
│   │   └── styles/
│   └── wailsjs/           # generated bindings and models
├── build/                 # icons, platform metadata, launcher, output
├── docs/
└── Makefile
```

## 3. Backend requirements

### 3.1 Process lifecycle

- `OnStartup` resolves paths, opens the app store, runs migrations, loads and
  normalises settings, prunes old history, opens the log file.
- `OnShutdown` closes the user database, app store and log file.
- Startup errors are stored in `initErr` and returned from every bound call so the
  UI can show a fatal state.
- The app is built with embedded `frontend/dist` assets.

### 3.2 Window

- 1280×820, minimum 1024×680, title `DBDesk`, dark background colour.
- Single window; no tray, menus, or background operation.

### 3.3 Wails bindings

All methods are on `main.App`:

| Method | Purpose |
| --- | --- |
| `GetState() AppState` | connection, recent profiles, settings, paths, version |
| `PickDatabaseFile() string` | native open dialog |
| `OpenDatabase(path, readOnly) ConnectionInfo` | open + remember |
| `CloseDatabase()` | close active connection |
| `CreateDemoDatabase() ConnectionInfo` | seed if absent, then open |
| `ResetDemoDatabase() ConnectionInfo` | delete, reseed, open |
| `RecentConnections() []ConnectionInfo` | recent list |
| `ForgetConnection(id)` | remove profile |
| `Schema() Schema` | cached schema |
| `RefreshSchema() Schema` | re-introspect |
| `PreviewTable(req) TablePage` | paginated rows + metadata |
| `RunQuery(sql) QueryResult` | one statement |
| `InsertRow(req) MutationResult` | parameterised insert |
| `UpdateRow(req) MutationResult` | parameterised update by key |
| `DeleteRow(req) MutationResult` | parameterised delete by key |
| `ExportLastResultCSV() string` | save dialog + write |
| `DashboardStats(refresh) DashboardData` | storage + telemetry |
| `History(limit) []HistoryEntry` | stored statements |
| `ClearHistory()` | delete all history |
| `SavedQueries() []SavedQuery` | saved list |
| `SaveQuery(name, sql) SavedQuery` | upsert by name |
| `DeleteSavedQuery(id)` | remove |
| `GetSettings() Settings` | current settings |
| `SaveSettings(s) Settings` | normalise + persist |

Events: `db:changed` after open/close/demo/schema-affecting operations;
`settings:changed` after settings save.

### 3.4 Opening a database

- Resolve to an absolute path.
- DSN: `file:<slash-path>?mode=rw` or `?mode=ro`.
- `SetMaxOpenConns(1)` to serialise access.
- 10 s ping timeout; failure returns a readable message.
- Pragmas: `busy_timeout=5000`, `foreign_keys=ON`, and `query_only=ON` for read-only.
- Load schema once at open; refresh on demand.

### 3.5 Schema introspection

- Read `sqlite_schema` for tables/views and DDL.
- Use `PRAGMA table_xinfo` for columns including hidden/generated columns.
- Use `PRAGMA foreign_key_list` and `PRAGMA index_list`/`index_info`.
- Determine editability: tables are editable; views and virtual tables are not.
- Determine row key: primary-key columns, else `rowid` when available.

### 3.6 Preview

- Reject unknown table or sort column.
- Quote identifiers; only `ASC`/`DESC` from a boolean check.
- Page size clamped to 1–1,000; page clamped to ≥ 1.
- Append hidden `rowid` selection when the table lacks a usable primary key.
- Count rows with a short context; on failure return `countError` and `hasNext`
  based on page fullness.
- Strip hidden key columns from the export copy.

### 3.7 Query execution

- Split on top-level semicolons (strings, quoted identifiers, comments respected).
- Reject zero or more than one statement.
- Clamp row limit to 1,000 default / 10,000 max and timeout to 30 s default / 120 s max.
- Reject non-read statements on read-only connections.
- Scan `RETURNING` statements as reads.
- Summarise long cell values (10,000 chars) and BLOBs.
- Store the last successful result for CSV export.
- Invalidate cached counts after writes.

### 3.8 Mutations

- Validate table and column names against the live schema.
- Insert: skip untouched blank fields so defaults apply; bind explicit NULLs.
- Update: key columns are immutable; only changed fields are sent; empty update is a
  no-op error.
- Delete: require a complete key.
- Execute in a transaction where the statement count requires it; report affected
  rows and use `RETURNING`/`last_insert_rowid` where available.
- Reject BLOBs and generated columns as targets.

### 3.9 CSV

- RFC 4180 via `encoding/csv`.
- Header from visible column names.
- Nil → empty; numbers formatted without scientific notation drift.
- Prefix `=`, `+`, `-`, `@` (after leading whitespace) with `'` to prevent formula
  injection.
- Save dialog default: `<database>-<YYYYMMDD-HHMMSS>.csv`.
- Cancelled dialog creates no file.

### 3.10 App store

- SQLite file at `<data>/app.db`.
- `PRAGMA user_version` migrations, version 1.
- Tables: `connections`, `query_history`, `saved_queries`, `settings`,
  `table_stats` (see SDA §5.2).
- Settings serialised as key/value JSON; unknown keys ignored.
- Connection profiles upserted by unique path.
- History capped at 20,000 rows at startup, plus retention days.

### 3.11 Dashboard aggregation

- Storage: `PRAGMA page_count`, `page_size`, `freelist_count`, `journal_mode`,
  `encoding`.
- Objects: count tables, views, indexes, triggers.
- Tables: exact `COUNT(*)` with a per-table timeout (2 s in preview, bounded in
  dashboard); failures mark the table and set `partial`.
- Query activity: hourly buckets for 24 h from `query_history`.
- Latency: p50/p95 per bucket computed in Go.
- Totals: queries in window, errors, error rate, p50/p95, today's count.
- Recent: last 8 statements.

### 3.12 Demo database

- Deterministic seed (fixed data, no randomness) at `<data>/demo/stonecrusher.db`.
- Eight tables (`categories`, `suppliers`, `products`, `customers`, `orders`,
  `order_items`, `payments`, `reviews`) and one view (`order_summary`).
- Seed only when missing; reset deletes and recreates.

## 4. Frontend requirements

### 4.1 Boot

1. `index.html` applies the saved theme synchronously to avoid a flash.
2. `main.ts` mounts `App.svelte`.
3. `App` calls `GetState`; if connected, loads schema; shows splash or fatal screen.
4. Subscribes to `db:changed` and `settings:changed`; unsubscribes on destroy.

### 4.2 State

- Svelte stores: `view`, `activeTable`, `editorDraft`, `dataVersion`, `toasts`.
- Local component state for page data, forms, filters and busy flags.
- Request tokens guard against out-of-order responses in schema, dashboard and table
  loading.

### 4.3 Views

- **Welcome** — open/demo actions, read-only opt-in, recent list.
- **Dashboard** — KPI cards, four charts, recent queries, storage table, table chips.
- **Data** — table selector, pager, page size, sort, CSV, insert, row actions,
  editor and delete dialogs.
- **Editor** — CodeMirror, run/clear/save/export, saved chips, result grid or write
  summary, inline errors.
- **History** — filter, reload, clear, expandable entries with open-in-editor.
- **Settings** — theme, execution and history limits, storage actions, recent list,
  about/shortcuts.

### 4.4 Charts

- Chart.js registered once in `lib/charts.ts`.
- A Svelte action rebuilds charts when config changes and destroys on unmount.
- Palette derives from the active theme; labels, grid and tooltips use theme colours.

### 4.5 Formatting and errors

- `formatBytes`, `formatNumber`, `formatDuration`, `formatDateTime`, `timeAgo`,
  `cellText`, `errorText`.
- Errors from bindings are displayed as returned (Go messages are user-facing) and
  also surfaced as toasts where a background action failed.

### 4.6 Accessibility

- Every icon-only button has `aria-label` and `title`.
- Dialogs use `role="dialog"`, `aria-modal`, Escape to close, backdrop click.
- Tables expose `aria-sort` on sortable headers.
- Toasts use `aria-live="polite"`.
- Reduced-motion media query disables animations.
- Focus-visible outline uses the accent colour.

## 5. Build and tooling

- `wails build -tags webkit2_41` compiles Go, runs `npm run build`, embeds assets and
  produces `build/bin/dbdesk`.
- The tag is required on distributions with WebKitGTK 4.1.
- `wails.json` sets name and output filename to `dbdesk`.
- `npm run check` (svelte-check) must report zero errors/warnings.
- `go test ./...` covers statement splitting, read-only enforcement, schema,
  preview/mutations, CSV safety, store migrations and dashboard aggregation.
- Make targets: `build`, `dev`, `run`, `test`, `check`, `clean`, `install`,
  `uninstall`.

## 6. Error handling conventions

- Go errors returned to the UI are short, lowercase-free sentences suitable for
  display, with no stack traces.
- SQLite driver noise is cleaned (`cleanError`) before display.
- UI never leaves a spinner running after a failed call.
- Partial data is preferred over blank screens (dashboard `partial`, preview
  `countError`).

## 7. Performance targets

- App cold start under 2 seconds on a typical desktop.
- 100-row page load under 150 ms for indexed tables up to ~1M rows.
- Dashboard under 1 second for databases with ≤ 20 tables; never blocks the UI.
- Frontend bundle under 1 MB (currently ~785 KB, mostly CodeMirror + Chart.js).

## 8. Known limitations

- One connection; a long query blocks other database calls until timeout.
- No query cancellation button (timeout only).
- Row counting is exact and may skip slow tables.
- Single-statement execution only.
- BLOBs summarised, not editable.
- History is local to the machine and stores SQL text in plaintext.
