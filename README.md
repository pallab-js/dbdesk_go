# DBDesk

Local-first SQLite desktop studio built with Go, Wails v2, and Svelte 5. Browse, edit, query, import, and maintain SQLite databases in a fast, offline-native app.

## Features

- **Dashboard** — Database size, object counts, row estimates, query activity, latency (p50/p95), and storage health
- **Data Browser** — Paginated tables/views with sorting, insert/update/delete (by PK or rowid), inline actions
- **SQL Editor** — CodeMirror 6 with syntax highlighting, run selection or full buffer, `EXPLAIN QUERY PLAN`, results grid, saved queries, SQL formatting, CSV export
- **History** — Executed statements with duration, rows, status, and errors
- **Import** — Import CSV or JSON data into tables with header/truncate options
- **Maintenance** — VACUUM, ANALYZE, REINDEX, integrity checks, and SQLite pragma inspection
- **Settings** — Theme (system/light/dark), limits/timeouts/retention, recent databases, demo database
- **Read-only Mode** — Enforced at UI and backend for safe inspection
- **Demo DB** — Built-in e-commerce sample dataset

## Requirements

- Go 1.26+
- Node.js 22+
- Wails CLI v2.17+
- Linux: `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`

## Quick Start

```bash
npm install --prefix frontend
make dev
```

## Build & Test

```bash
make build  # build/bin/dbdesk
make test   # go test ./...
make check  # svelte-check
```

## Data Location

| OS | Path |
|---|---|
| Linux | `~/.local/share/dbdesk/` |
| macOS | `~/Library/Application Support/dbdesk/` |
| Windows | `%APPDATA%\dbdesk\` |

## License

[MIT](LICENSE)
