# Implementation Plan

## 1. Approach

Build the product in vertical slices: foundation and data engine first, then the
browsing/editing loop, then querying and telemetry, then polish and packaging. Each
slice ends with a runnable app and passing checks. The UI is built against real Go
bindings, not mocks, so type and behaviour drift surfaces early.

## 2. Phases

### Phase 0 — Scaffold

- Generate the Wails Svelte-TS template.
- Pin Vite to `^7.1.0` for `@sveltejs/vite-plugin-svelte@6` compatibility.
- Set module `dbdesk`, binary `dbdesk`, window size and dark background.
- Add frontend dependencies: Chart.js, CodeMirror 6, SQL language, Lucide.

**Done when:** `wails dev` opens an empty themed window and `go test ./...` passes.

### Phase 1 — Backend foundation

- `internal/config`: XDG/OS data paths, directory creation.
- `internal/model`: all DTOs with stable JSON names.
- `internal/store`: migrations, settings, connections, history, saved queries,
  cached stats, pruning.
- `internal/sqlite`: open/close, pragmas, schema introspection.
- `internal/demo`: deterministic seeder.
- Unit tests for store migrations and schema loading.

**Done when:** a Go test can create the app store, open a seeded database and read its
schema.

### Phase 2 — Data engine

- Paginated preview with PK/`rowid` key detection.
- Sorting whitelist and exact row counts with fallbacks.
- Single-statement splitter and executor with limits and read-only guard.
- Parameterised insert/update/delete.
- CSV writing with formula-injection guards.
- Tests for splitting, read-only rejection, preview, mutations and CSV.

**Done when:** every database operation works headlessly through Go tests.

### Phase 3 — Wails facade

- Implement `main.App` methods, mutex, timeouts, history recording, event emission.
- Regenerate bindings with `wails generate module` / build.
- Normalise settings and validate inputs at the boundary.

**Done when:** bindings exist and each method is callable from the frontend.

### Phase 4 — Design foundation

- Token file for dark/light themes; base CSS for controls, panels, badges, tables.
- Theme store with system/media sync and pre-bundle bootstrap in `index.html`.
- Formatting helpers, stores, Chart.js action, click-outside action.

**Done when:** both themes render correctly with no flash on launch.

### Phase 5 — Shell and shared components

- `App.svelte` boot/fatal/splash, event subscriptions, view switching.
- `TopBar`, `Sidebar`, `Modal`, `ConfirmDialog`, `Toasts`, `EmptyState`, `DataGrid`,
  `RowEditor`, `Chart`.
- Keyboard handling and accessibility attributes.

**Done when:** the shell can switch views and mount a grid with sample data.

### Phase 6 — Features

1. **Welcome** — open, demo, recent, read-only opt-in.
2. **Data** — preview, paging, sorting, page size, insert/edit/delete, CSV.
3. **Editor** — CodeMirror, run/selection, results, errors, saved queries, export.
4. **History** — list, filter, expand, open in editor, clear.
5. **Dashboard** — KPIs, four charts, recent, storage, table chips, partial state.
6. **Settings** — theme, limits, retention, auto-refresh, demo reset, recent list.

**Done when:** every PRD user story is reachable in the running app.

### Phase 7 — Verification

- `npm run check` — zero errors and warnings.
- `npm run build` — production bundle.
- `go test ./...` — all packages pass.
- `wails build -tags webkit2_41` — binary produced.
- Smoke launch — window opens, log records startup, no crash.
- Manual pass: demo → dashboard → table → edit/insert/delete → query → export →
  history → settings → theme toggle → close/reopen.

### Phase 8 — Packaging and docs

- `Makefile` with build/dev/run/test/check/clean/install/uninstall.
- `build/dbdesk.desktop` launcher and icon install rules.
- README with requirements, commands, data locations, safety model and limits.
- `docs/`: SDA, PRD, TRD, APP_FLOW, UI_UX, IMPLEMENTATION.
- `.gitignore` for build output, `node_modules`, `dist` and local artefacts.

## 3. Testing strategy

| Area | Method |
| --- | --- |
| Statement splitting | Table-driven Go tests for strings, comments, identifiers. |
| Read-only | Open a file read-only and assert writes are rejected. |
| Preview/mutations | Seed a temp DB, page it, insert/update/delete, verify rows. |
| CSV | Golden-style assertions for quoting, NULLs and formula prefixes. |
| Store | Migration idempotency, settings round-trip, history pruning. |
| Dashboard | Aggregation over seeded history and schema. |
| Frontend | `svelte-check` type and a11y checks; production build. |
| End-to-end | Manual smoke pass on the built binary. |

No browser automation is part of the repo; the native WebView is verified by launch
plus manual use.

## 4. Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| WebKitGTK version mismatch on Linux | Build with `-tags webkit2_41`; document the alternative. |
| Long table counts slow the dashboard | Per-table timeout, approximate fallback, `partial` flag. |
| Unintended writes | Read-only triple guard, confirmations, parameterised values. |
| CSV formula injection | Prefix dangerous leading characters. |
| Large results freeze the UI | Row cap, truncation flag, 10k char cell summary. |
| Binding drift after Go changes | Regenerate bindings during every Wails build. |
| Bundle growth | CodeMirror + Chart.js dominate; acceptable for a desktop app. |

## 5. Delivery checklist

- [x] Backend packages and tests
- [x] Wails bindings and facade
- [x] Theme tokens and base styles
- [x] Shared components
- [x] Welcome, Dashboard, Data, Editor, History, Settings
- [x] README, Makefile, desktop launcher
- [x] Planning docs
- [ ] Signed installers for macOS/Windows (future)

## 6. Future work

- Multiple connections and tabs.
- Schema editor (create/alter tables, indexes).
- Import CSV/JSON and BLOB editing.
- Multi-statement scripts with per-statement results.
- Query cancellation, explain plans, and slow-query hints.
- Export to JSON/Markdown and clipboard copy.
- Optional encrypted history and per-connection history filters.
- macOS `.app`/`.dmg` and Windows MSI packaging.
