# Product Requirements Document (PRD)

## 1. Summary

DBDesk is a local desktop SQLite client for developers who want to inspect a
database, make a quick data change, run ad-hoc SQL, and understand what they have
been doing — without a browser-based tool, an account, or a server.

## 2. Problem

Working with a local SQLite file today usually means choosing between:

- a heavyweight database IDE that assumes a server,
- a web tool that requires uploading or exposing the file,
- the `sqlite3` shell, which is fast for queries but poor for browsing, editing and
  seeing history.

There is also no lightweight way to see how a database is shaped and how the user
has been querying it in one place. DBDesk closes that gap with a small native app.

## 3. Target users

- **Application developer** — inspects a dev database, fixes a row, checks a schema.
- **Data-curious engineer** — explores a downloaded `.db` file, runs grouping
  queries, exports a CSV.
- **Support/debug engineer** — opens a customer-provided SQLite file read-only and
  looks around safely.

## 4. Goals

1. Open a local SQLite file and be productive in under a minute.
2. Browse, sort, paginate, insert, edit and delete rows safely.
3. Run one SQL statement at a time with clear results and errors.
4. Export the current result to CSV.
5. Show a useful dashboard from database stats and the user's query history.
6. Keep everything local, with an explicit read-only mode.

## 5. Non-goals (v1)

- PostgreSQL, MySQL, or other engines.
- Multiple databases open at once.
- Schema designer / migrations.
- Data import, BLOB editing, multi-statement scripts.
- Cloud sync, accounts, collaboration, sharing.
- Query plan visualisation or profiling beyond duration.

## 6. User stories

### Opening

- As a developer, I can choose a `.db` file with a native dialog, so I do not have to
  type paths.
- As a developer, I can reopen a recent database in one click.
- As a support engineer, I can open a database read-only so I cannot modify it.
- As a first-time user, I can create and open a demo database to explore the app.

### Browsing

- As a developer, I can see tables and views in a sidebar and filter them.
- As a developer, I can page through rows and change page size.
- As a developer, I can sort by clicking a column header.
- As a developer, I can see column names, types and primary keys.
- As a developer, I can distinguish NULL from an empty string.

### Editing

- As a developer, I can insert a row and leave blank fields to their defaults.
- As a developer, I can edit a row and see which fields changed.
- As a developer, I can set a field to NULL explicitly.
- As a developer, I can delete a row after confirming the key values.
- As a developer, I cannot edit views, generated columns, or BLOBs.

### Querying

- As a developer, I can write SQL with syntax highlighting and completion.
- As a developer, I can run the whole editor or only the selected text.
- As a developer, I can see rows, affected count, duration and errors.
- As a developer, I can save named queries and reload them.
- As a developer, I can export the last result to CSV.

### Understanding usage

- As a developer, I can see database size, page count, journal mode and encoding.
- As a developer, I can see table row counts and column type distribution.
- As a developer, I can see query volume, failures and latency over the last day.
- As a developer, I can review, filter and clear query history.

### Preferences

- As a user, I can choose system, light or dark theme.
- As a user, I can cap returned rows, set a timeout, and control history retention.
- As a user, I can turn off history recording or dashboard auto-refresh.
- As a user, I can forget recent databases and reset the demo data.

## 7. Functional requirements

### FR-1 Connection management

- Native file picker filtered to SQLite files.
- Open read/write by default; opt-in read-only.
- Validate the file opens and is readable; report failures.
- Record path, name, mode and last-opened time.
- Allow closing the active database.
- Provide a recent list and a forget action.

### FR-2 Schema browsing

- List tables and views with column metadata.
- Show primary keys, foreign keys, indexes and DDL where available.
- Detect objects that cannot be edited.

### FR-3 Table data

- Server-side pagination with page size 50–1,000.
- Sorting by one column, ascending or descending.
- Total row count, with a clear fallback when counting fails.
- Hidden `rowid` support for tables without a usable primary key.

### FR-4 Row mutations

- Insert: blank fields defer to defaults; NULL is explicit.
- Update: only changed non-key fields are written.
- Delete: requires key values and confirmation.
- All mutations report affected rows and refresh the view.

### FR-5 SQL execution

- Exactly one statement per run.
- Selection, when present, is the statement.
- Row cap and timeout from settings.
- Read results in a grid; writes report affected rows.
- Errors shown inline and stored in history.

### FR-6 History

- Record SQL, kind, duration, row count, success and error.
- Timestamp every entry; prune by retention.
- Filter by text, expand details, copy back to editor, clear all.

### FR-7 Saved queries

- Save with a name; load into editor; delete.
- Persist across sessions.

### FR-8 Export

- Export the last result or current table page to CSV.
- Suggest a database-and-timestamp filename.
- Neutralise spreadsheet formulas.
- No file is created when the dialog is cancelled.

### FR-9 Dashboard

- Storage: size, page size, page count, free pages, journal mode, encoding, mode.
- Objects: tables, views, indexes, triggers.
- Tables: exact counts where possible; mark timeouts.
- Column types distribution.
- Query activity, failures, p50/p95 latency for 24 hours.
- Recent statements.
- Auto-refresh when data changes, if enabled.

### FR-10 Settings

- Theme mode.
- Row limit 100–10,000.
- Timeout 5–120 seconds.
- History on/off and retention 1–365 days.
- Dashboard auto-refresh on/off.
- Persist and apply immediately.

## 8. Non-functional requirements

- **Local-first:** no network calls, no telemetry.
- **Responsiveness:** table pages load within the timeout; UI never blocks on a
  failed query.
- **Safety:** identifiers quoted/validated; values parameterised; formula injection
  guarded; destructive actions confirmed.
- **Recoverability:** errors are visible and actionable; app state survives restarts.
- **Accessibility:** keyboard focus, labels on icon buttons, contrast in both themes,
  reduced-motion support.
- **Portability:** Linux, macOS and Windows through Wails.

## 9. Success criteria

- A new user can open the demo and reach a table, a query result and the dashboard
  without documentation.
- All row operations work on tables with integer primary keys and on tables that
  only expose `rowid`.
- Read-only connections reject writes in the UI and in the backend.
- `go test ./...`, `npm run check` and a production build pass.

## 10. Release scope

Version 0.1.0 is the first complete slice: SQLite, one connection, browse/edit/query/
export/history/dashboard/settings, with packaging for Linux and build instructions
for macOS and Windows.
