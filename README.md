# DBDesk

A local-first SQLite desktop client built with [Go](https://go.dev), [Wails v2](https://wails.io), and [Svelte 5](https://svelte.dev). Browse schemas, run SQL, edit rows, export CSV, and explore your database with a built-in dashboard.

Everything runs on your machine — no accounts, no cloud services, no telemetry.

## Features

- **Dashboard** — Database size, table/view counts, row estimates, column type mix, query activity over time, p50/p95 latency, recent statements, and storage health.
- **Data browser** — Paginated table and view preview with column sorting, insert/update/delete with primary-key or `rowid` targeting.
- **SQL editor** — CodeMirror 6 with SQLite highlighting/completion, run selection or whole buffer, `EXPLAIN QUERY PLAN`, result grid, saved queries, and CSV export.
- **History** — Every executed statement with duration, row count, status, and errors. Filterable and clearable.
- **Settings** — Theme (system/light/dark), default row limit, query timeout, history retention, auto-refresh, recent databases, demo database management.
- **Read-only mode** — Enforced at the UI, connection flags, and backend guards.
- **Demo database** — Deterministic 8-table e-commerce sample with a view for quick exploration.

## Requirements

- [Go](https://go.dev/dl/) 1.24+ (tested with 1.26)
- [Node.js](https://nodejs.org/) 22+ and npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.17+
- **Linux**: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` (Ubuntu 24.04/26.04). On distros shipping WebKitGTK 4.0, adjust tags in `Makefile`.
- **macOS**: Xcode Command Line Tools; **Windows**: WebView2 Runtime

## Getting Started

```bash
# Install frontend deps
npm install --prefix frontend

# Run in development mode (Linux/Ubuntu 24.04+)
make dev
```

## Build

```bash
make build   # produces build/bin/dbdesk
make test    # go test ./...
make check   # svelte-check
```

## Install (Linux)

```bash
make install   # installs to /usr/local by default
make uninstall
```

## Data Location

| Platform | Path | Contents |
| --- | --- | --- |
| Linux (`$XDG_DATA_HOME`, default `~/.local/share`) | `~/.local/share/dbdesk/` | `app.db` (settings/connections/history), `demo/`, `logs/` |
| macOS | `~/Library/Application Support/dbdesk/` | Same structure |
| Windows | `%APPDATA%\dbdesk\` | Same structure |

Opened databases are accessed in-place; DBDesk only writes to them when you insert/update/delete or run write statements.

## Safety

- Parameterized queries where applicable; identifiers are validated against schema.
- Read-only mode prevents writes.
- Destructive actions (row delete/update) require confirmation in the UI.
- CSV export escapes formula-leading characters to mitigate spreadsheet injection.
- History can be cleared; retention is configurable.

## License

[MIT](LICENSE)
