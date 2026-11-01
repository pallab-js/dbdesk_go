# DBDesk

A local-first SQLite desktop client: browse tables, edit rows, run SQL, export CSV,
and watch a dashboard built from your own query history and database stats.

Everything runs on your machine. There is no account, no cloud service, and no
telemetry that leaves the app.

## Features

- **Dashboard** — database size, table/row counts, column type mix, per-hour query
  activity, p50/p95 latency, recent statements, and storage health.
- **Data browser** — paginated table and view preview, column sorting, insert/update/
  delete with primary-key or `rowid` targeting.
- **SQL editor** — CodeMirror 6 with SQLite highlighting and completion, run the whole
  buffer or just the selection, result grid, saved queries and CSV export.
- **History** — every statement with duration, row count and error, filterable and
  clearable.
- **Settings** — theme (system/light/dark), row limit, timeout, history retention,
  auto-refresh, recent databases, demo reset.
- **Read-only mode** — enforced in the connection, pragmas and backend guards.
- **Demo database** — deterministic 8-table e-commerce sample with a view.

## Requirements

- Go 1.24+ (developed with Go 1.26)
- Node.js 22+ and npm
- Wails v2.17+
- Linux: GTK 3 and WebKitGTK 4.1 development packages
  (`libgtk-3-dev`, `libwebkit2gtk-4.1-dev` on Debian/Ubuntu)
- macOS: Xcode command line tools; Windows: WebView2 runtime

## Development

```sh
npm install --prefix frontend
make dev          # wails dev -tags webkit2_41
```

On distributions still shipping WebKitGTK 4.0, remove `-tags webkit2_41` from the
Makefile (or run `wails dev` directly).

## Build

```sh
make build        # produces build/bin/dbdesk
make test         # go test ./...
make check        # svelte-check
```

`make install` copies the binary to `$PREFIX/bin` (default `/usr/local`), the
launcher to `share/applications`, and the icon to `share/icons`. `make uninstall`
removes them.

## Where data lives

Linux (`$XDG_DATA_HOME` defaults to `~/.local/share`):

| Path | Contents |
| --- | --- |
| `~/.local/share/dbdesk/app.db` | app store: connections, history, saved queries, settings |
| `~/.local/share/dbdesk/demo/ecommerce.db` | generated demo database |
| `~/.local/share/dbdesk/logs/app.log` | app log |

macOS uses `~/Library/Application Support/dbdesk`, Windows uses
`%APPDATA%\dbdesk`. Your databases are opened in place; DBDesk only writes to them
when you edit rows or run a writing statement.

## Safety model

- One database connection at a time.
- Read-only mode uses a read-only DSN plus `PRAGMA query_only=ON` plus a backend
  statement guard.
- Table and column identifiers are validated against the live schema and quoted;
  row values are always bound as parameters.
- The SQL editor runs exactly one statement per execution (selected text counts as
  the whole statement).
- Query results default to 1,000 rows (configurable up to 10,000) and 30 seconds
  (configurable up to 120).
- CSV cells beginning with `=`, `+`, `-` or `@` are prefixed to prevent spreadsheet
  formula injection.

## Limits in this version

- SQLite only, one connection at a time.
- No schema designer, import, or multi-statement scripts.
- BLOB columns are displayed summarised and are not editable.
- Generated columns and views are read-only.

## Documentation

- [Solution design](docs/SDA.md)
- [Product requirements](docs/PRD.md)
- [Technical requirements](docs/TRD.md)
- [Application flow](docs/APP_FLOW.md)
- [UI and UX](docs/UI_UX.md)
- [Implementation plan](docs/IMPLEMENTATION.md)
