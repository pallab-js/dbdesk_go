# DBDesk

Local-first SQLite desktop client built with Go, Wails v2, and Svelte 5.

## Features

- **Dashboard** — DB size, tables/views, row estimates, query activity, latency (p50/p95), and storage health
- **Data Browser** — Paginated tables/views, sorting, insert/update/delete (PK or rowid)
- **SQL Editor** — CodeMirror 6, run selection or buffer, `EXPLAIN QUERY PLAN`, results grid, saved queries, CSV export
- **History** — Executed statements with duration, rows, status, errors
- **Settings** — Theme (system/light/dark), limits/timeouts/retention, recent DBs, demo DB
- **Read-only Mode** — Enforced at UI and backend
- **Demo DB** — Built-in e-commerce sample

## Requirements

- Go 1.24+
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
