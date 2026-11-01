# Solution Design Architecture (SDA)

## 1. Purpose

DBDesk is a desktop application for working with local SQLite databases. It combines
the three things a developer usually opens separate tools for:

1. a schema/data browser with row editing,
2. a SQL editor with history and saved queries,
3. a dashboard over the database's storage stats and the user's own query telemetry.

The application is single-user, offline, and local-first. The only network access is
none; all persistence lives in the user's data directory or in the database files the
user opens.

## 2. System context

```
┌─────────────────────────────────────────────────────────────┐
│ Desktop machine                                             │
│                                                             │
│  ┌──────────────┐   Wails bindings    ┌─────────────────┐   │
│  │ Svelte 5 UI  │ ──────────────────► │ Go application  │   │
│  │ (WebView)    │ ◄────────────────── │ (main.App)      │   │
│  └──────────────┘   events/results    └────────┬────────┘   │
│                                                │            │
│                       ┌────────────────────────┼──────────┐ │
│                       │                        │          │ │
│                ┌──────▼──────┐        ┌────────▼───────┐  │ │
│                │ User SQLite │        │ App store      │  │ │
│                │ database(s) │        │ app.db         │  │ │
│                └─────────────┘        └────────────────┘  │ │
│                                                             │
│                       ┌────────────────┐                    │
│                       │ Demo database  │                    │
│                       │ ecommerce.db   │                    │
│                       └────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

No server process, no remote API, no background daemon. Wails embeds the compiled
frontend into the Go binary and serves it to a platform WebView.

## 3. Layers

### 3.1 Presentation (Svelte 5 + TypeScript)

- `frontend/src/App.svelte` is the shell: boot, state, schema/dashboard loading, event
  wiring, routing between views.
- Views: `Welcome`, `Dashboard`, `Data`, `Editor`, `History`, `Settings`.
- Components: `TopBar`, `Sidebar`, `DataGrid`, `RowEditor`, `Modal`, `ConfirmDialog`,
  `Toasts`, `EmptyState`, `Chart`.
- Cross-cutting libraries: `theme.ts` (system/light/dark), `stores.ts` (view, active
  table, toasts, data version), `format.ts`, `charts.ts` (Chart.js), `clickOutside.ts`.
- Styling is plain CSS with design tokens; no CSS framework.
- The UI talks to Go only through generated Wails bindings
  (`frontend/wailsjs/go/main/App.js`) and generated model constructors.

### 3.2 Application facade (Go, `main` package)

`App` in `app.go` is the only exported surface. It owns:

- process lifecycle (`startup`, `shutdown`),
- the active database connection (`*sqlite.DB`),
- the app store (`*store.Store`),
- current settings and recent connections,
- event emission (`db:changed`, `settings:changed`),
- input validation at the trust boundary,
- conversion between internal errors and user-facing messages.

The facade serialises database access with a read/write mutex, applies per-call
timeouts and row limits, records history, and delegates all SQL work to `internal`.

### 3.3 Domain packages (`internal/`)

| Package | Responsibility |
| --- | --- |
| `config` | Resolve the OS data directory and its files. |
| `model` | JSON DTOs shared by Go and the generated TS models. |
| `store` | App database: migrations, connections, history, saved queries, settings, cached stats. |
| `sqlite` | Open/close, schema introspection, paginated preview, query execution, row mutations, CSV export, statement splitting. |
| `dashboard` | Aggregation of storage stats, schema counts, table row counts and query telemetry. |
| `demo` | Deterministic demo database seeder. |

`internal` packages do not import Wails; they are testable with plain Go tests.

## 4. Data flow

### 4.1 Open a database

1. UI calls `PickDatabaseFile` (native dialog) or a recent path.
2. `OpenDatabase(path, readOnly)` closes the previous connection, opens the new one
   with the appropriate DSN and pragmas, verifies it is readable, upserts a
   connection profile in `app.db`, and emits `db:changed`.
3. UI reloads `GetState`, `Schema`, and the dashboard.

### 4.2 Browse and edit rows

1. `PreviewTable` builds a `SELECT` with quoted identifiers, optional whitelisted
   sort column/direction, `LIMIT/OFFSET`, and a hidden `rowid` where needed.
2. The response carries `ColumnMeta`, rows, key indexes/names and editability.
3. `InsertRow` / `UpdateRow` / `DeleteRow` validate identifiers against the live
   schema, require key values for update/delete, bind all values, and run inside a
   transaction where needed.
4. On success the UI bumps `dataVersion`, which reloads the page and dashboard.

### 4.3 Run SQL

1. `RunQuery(sqlText)` asks `sqlite.SplitStatements` to split top-level semicolons.
   Exactly one statement is accepted.
2. Read statements and `RETURNING` queries are scanned with a row cap; other
   statements are executed and report affected rows/last insert id.
3. Every execution is appended to `query_history` when history is enabled.
4. The last result is retained in memory for CSV export.

### 4.4 Dashboard

`DashboardStats(refresh)` reads storage pragmas and schema objects from the open
database, counts table rows with a per-table timeout (falling back to approximate
counts), and aggregates query history from the app store into hourly buckets,
latency percentiles, totals and recent entries. Results may be flagged `partial`
when some tables could not be counted.

## 5. Persistence

### 5.1 User database

Opened in place. DBDesk never copies or rewrites the file except through normal
SQLite writes the user requests. WAL files created by SQLite stay next to the
database.

### 5.2 App store (`app.db`)

SQLite database with `PRAGMA user_version` migrations. Tables:

- `connections` — path (unique), display name, read-only flag, timestamps.
- `query_history` — SQL, kind, duration, row count, success, error, timestamp.
- `saved_queries` — name, SQL, timestamps.
- `settings` — key/value JSON.
- `table_stats` — cached row counts per connection/table.

History is pruned at startup using the retention setting.

### 5.3 Settings

Defaults: row limit 1,000, timeout 30 s, history on, retention 30 days, dashboard
auto-refresh on. Values are normalised on read and write to the documented ranges.

## 6. Key decisions

| Decision | Rationale |
| --- | --- |
| Wails v2 + Svelte 5 | Small native window, Go backend, no Node runtime in production. |
| `modernc.org/sqlite` | Pure-Go driver; no cgo toolchain, predictable cross-compilation. |
| No ORM | Introspection and user SQL are the domain; an ORM adds no value. |
| No router/state library | Five views and a handful of stores; built-in Svelte stores suffice. |
| Chart.js | Mature canvas charts with small integration surface. |
| CodeMirror 6 | Lightweight, themeable SQL editor with SQLite dialect. |
| CSS variables | One token file drives both themes; no framework or runtime CSS. |
| Generated bindings | Types stay in sync with Go DTOs; no hand-written API client. |
| Read-only triple guard | DSN + pragma + statement check prevents accidental writes. |
| Single connection | Keeps state, transactions and the UI predictable in v1. |

## 7. Security and trust boundaries

- File paths arrive from native dialogs or saved profiles; opening a path is the
  user's explicit action.
- SQL text is user-authored by design. Structural identifiers used by DBDesk's own
  generated statements are validated against live schema metadata and quoted.
- Values are always bound parameters, never concatenated.
- Header names are quoted with doubled quotes; sort column/direction are
  whitelisted, not interpolated from free text.
- CSV exports neutralise spreadsheet formula prefixes.
- The app emits no telemetry and performs no network requests.

## 8. Failure handling

- Startup failures surface as a fatal screen with the error and a retry action.
- Database open failures keep the previous connection closed and report the cause.
- Query errors preserve the message, are written to history, and show inline.
- Dashboard counting failures mark the table and set `partial`; the rest still renders.
- CSV export cancels cleanly when the save dialog is dismissed.

## 9. Out of scope for v1

PostgreSQL/MySQL support, multiple simultaneous connections, schema editing,
imports, BLOB editing, multi-statement scripts, plugins, cloud sync, and user
accounts.
