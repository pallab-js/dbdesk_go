# DBDesk Repository Structure

```text
dbdesk/
├── app.go                # Wails App facade (bindings)
├── main.go               # App entrypoint and window config
├── go.mod / go.sum
├── wails.json
├── Makefile
├── README.md
├── LICENSE
├── .gitignore
├── docs/
│   └── STRUCTURE.md
├── build/                # Build artifacts (generated)
├── frontend/
│   ├── package.json
│   ├── index.html
│   ├── src/
│   │   ├── App.svelte
│   │   ├── main.ts
│   │   ├── app.css
│   │   ├── components/   # UI components (DataGrid, Modal, etc.)
│   │   ├── views/        # Dashboard, Editor, Data, History, Settings, Welcome
│   │   ├── lib/          # stores, utils, formatting
│   │   └── wailsjs/      # Generated Wails bindings
│   └── vite.config.ts
└── internal/
    ├── config/           # Paths and app config
    ├── dashboard/        # Dashboard stats aggregation
    ├── demo/             # Demo DB generator
    ├── model/            # Shared types/models
    ├── sqlite/           # SQLite driver wrapper, queries, schema
    └── store/            # App DB (settings/history/saved queries)
```

## Key Modules

- **app.go**: Exposes methods to frontend (OpenDatabase, RunQuery, ExplainQuery, CRUD operations, dashboard stats).
- **internal/sqlite**: Connection management, schema introspection, query execution, pagination, mutations.
- **internal/store**: Persistent app state (recent connections, query history, saved queries, settings).
- **internal/dashboard**: Computes KPIs and chart data from the open DB + history.
- **frontend/src/views/Editor.svelte**: SQL editor with CodeMirror 6, run/explain, results grid, saved queries.
- **frontend/src/components/DataGrid.svelte**: Paginated results with sorting and inline editing.
