# DBDesk frontend

Svelte 5 + TypeScript UI served by the Wails WebView.

```sh
npm install
npm run dev     # Vite only; Go bindings are unavailable outside `wails dev`
npm run check   # svelte-check
npm run build   # dist/ bundled into the Go binary by Wails
```

See the repository [README](../README.md) and [docs](../docs/) for architecture,
flows and design decisions.
