<script lang="ts">
  import {
    Bookmark,
    Download,
    Eraser,
    FileSearch,
    LoaderCircle,
    Play,
    Save,
    Trash2,
    TriangleAlert,
  } from '@lucide/svelte'
  import { basicSetup, EditorView } from 'codemirror'
  import { SQLite, sql } from '@codemirror/lang-sql'
  import { keymap } from '@codemirror/view'
  import { onMount } from 'svelte'
  import { ExportLastResultCSV, RunQuery, SavedQueries, SaveQuery, DeleteSavedQuery, ExplainQuery } from '../../wailsjs/go/main/App.js'
  import { model } from '../../wailsjs/go/models'
  import type { model as M } from '../../wailsjs/go/models'
  import DataGrid from '../components/DataGrid.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import Modal from '../components/Modal.svelte'
  import { errorText, formatDuration, formatNumber } from '../lib/format'
  import { editorDraft, notify } from '../lib/stores'

  export let schema: M.Schema | null = null
  export let rowLimit = 1000

  let host: HTMLDivElement
  let view: EditorView | null = null
  let result: M.QueryResult | null = null
  let error = ''
  let running = false
  let explaining = false
  let exporting = false
  let saved: M.SavedQuery[] = []
  let showSave = false
  let saveName = ''
  let saving = false

  function quote(name: string) {
    return `"${name.replace(/"/g, '""')}"`
  }

  function defaultSQL() {
    const t = schema?.tables.find((x) => x.type === 'table')
    return t ? `SELECT *\nFROM ${quote(t.name)}\nLIMIT 100;` : 'SELECT 1 AS hello;'
  }

  const theme = EditorView.theme({
    '&': {
      height: '100%',
      backgroundColor: 'transparent',
      color: 'var(--text)',
      fontSize: '12.5px',
    },
    '.cm-scroller': {
      fontFamily: 'var(--font-mono)',
      lineHeight: '1.65',
      overflow: 'auto',
    },
    '.cm-content': { padding: '10px 0', caretColor: 'var(--accent)' },
    '.cm-line': { padding: '0 12px' },
    '.cm-gutters': {
      backgroundColor: 'var(--bg-elevated)',
      color: 'var(--faint)',
      border: 'none',
      borderRight: '1px solid var(--border)',
      fontSize: '11px',
    },
    '.cm-activeLine': { backgroundColor: 'var(--surface-2)' },
    '.cm-activeLineGutter': { backgroundColor: 'var(--surface-2)', color: 'var(--text)' },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
      backgroundColor: 'var(--accent-soft)',
    },
    '.cm-cursor, .cm-dropCursor': { borderLeftColor: 'var(--accent)' },
    '.cm-panels': { backgroundColor: 'var(--surface-2)', color: 'var(--text)' },
    '.cm-tooltip': {
      backgroundColor: 'var(--surface-2)',
      border: '1px solid var(--border-strong)',
      borderRadius: '6px',
    },
  })

  onMount(() => {
    view = new EditorView({
      parent: host,
      doc: $editorDraft || defaultSQL(),
      extensions: [
        basicSetup,
        sql({ dialect: SQLite }),
        EditorView.lineWrapping,
        theme,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) editorDraft.set(update.state.doc.toString())
        }),
        keymap.of([
          {
            key: 'Mod-Enter',
            preventDefault: true,
            run: () => {
              run()
              return true
            },
          },
        ]),
      ],
    })
    view.focus()
    loadSaved()
    return () => view?.destroy()
  })

  async function loadSaved() {
    try {
      saved = await SavedQueries()
    } catch {
      saved = []
    }
  }

  function currentSQL(): string {
    if (!view) return ''
    const { from, to } = view.state.selection.main
    if (to > from) return view.state.sliceDoc(from, to)
    return view.state.doc.toString()
  }

  async function run() {
    const text = currentSQL().trim()
    if (!text) {
      notify('Write a query first', 'error')
      return
    }
    running = true
    explaining = false
    error = ''
    try {
      result = await RunQuery(text)
      if (result.truncated) {
        notify(`Showing the first ${formatNumber(result.rowCount)} rows`, 'info')
      }
    } catch (e) {
      result = null
      error = errorText(e)
    } finally {
      running = false
    }
  }

  async function explain() {
    const text = currentSQL().trim()
    if (!text) {
      notify('Write a query first', 'error')
      return
    }
    explaining = true
    running = false
    error = ''
    try {
      result = await ExplainQuery(text)
    } catch (e) {
      result = null
      error = errorText(e)
    } finally {
      explaining = false
    }
  }

  function loadQuery(text: string) {
    if (!view) return
    view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: text } })
    editorDraft.set(text)
    view.focus()
  }

  function clearEditor() {
    loadQuery('')
    result = null
    error = ''
  }

  async function exportCSV() {
    exporting = true
    try {
      const path = await ExportLastResultCSV()
      if (path) notify(`Exported to ${path}`, 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      exporting = false
    }
  }

  async function saveQuery() {
    saving = true
    try {
      await SaveQuery(saveName, view?.state.doc.toString() ?? '')
      showSave = false
      saveName = ''
      await loadSaved()
      notify('Query saved', 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      saving = false
    }
  }

  async function removeSaved(q: M.SavedQuery) {
    try {
      await DeleteSavedQuery(q.id)
      await loadSaved()
      notify(`Deleted “${q.name}”`, 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    }
  }

  $: sampleTable = schema?.tables.find((t) => t.type === 'table')?.name ?? ''
</script>

<div class="page">
  <section class="panel editor-panel">
    <div class="panel-head">
      <div class="row">
        <h2>SQL editor</h2>
        <span class="faint">selection runs alone · <span class="kbd">Ctrl</span> <span class="kbd">Enter</span></span>
      </div>
      <div class="row">
        <button class="btn sm" on:click={clearEditor} title="Clear editor">
          <Eraser size={13} />
          Clear
        </button>
        <button class="btn sm" on:click={() => (showSave = true)} title="Save this query">
          <Save size={13} />
          Save
        </button>
        <button class="btn sm" disabled={!result || exporting} on:click={exportCSV} title="Export results as CSV">
          <Download size={13} />
          Export
        </button>
        <button class="btn sm" disabled={running || explaining} on:click={explain} title="Run EXPLAIN QUERY PLAN for current selection">
          {#if explaining}<LoaderCircle size={13} class="spin" />{:else}<FileSearch size={13} />{/if}
          Explain
        </button>
        <button class="btn sm primary" disabled={running || explaining} on:click={run} title="Run query (Ctrl+Enter)">
          {#if running}<LoaderCircle size={13} class="spin" />{:else}<Play size={13} />{/if}
          Run
        </button>
      </div>
    </div>
    <div class="editor-host" bind:this={host}></div>
  </section>

  {#if saved.length > 0}
    <div class="saved">
      <Bookmark size={12} class="muted" />
      <span class="faint saved-label">Saved</span>
      {#each saved as q (q.id)}
        <span class="saved-chip">
          <button class="chip-name" title={q.sql} on:click={() => loadQuery(q.sql)}>{q.name}</button>
          <button class="chip-x" title="Delete saved query" aria-label="Delete saved query" on:click={() => removeSaved(q)}>
            <Trash2 size={10} />
          </button>
        </span>
      {/each}
    </div>
  {/if}

  <section class="panel result-panel">
    <div class="panel-head result-head">
      <div class="row">
        <h2>Result</h2>
        {#if result}
          <span class="badge accent">{result.kind || 'query'}</span>
          {#if result.isRead}
            <span class="badge">{formatNumber(result.rowCount)} rows</span>
          {:else if result.affected >= 0}
            <span class="badge ok">{formatNumber(result.affected)} affected</span>
          {/if}
          <span class="muted">{formatDuration(result.durationMs)}</span>
          {#if result.truncated}<span class="badge warn">row limit reached</span>{/if}
        {:else if running}
          <span class="muted">running…</span>
        {/if}
      </div>
      {#if result && !result.isRead && result.lastInsertId}
        <span class="faint">last insert id {result.lastInsertId}</span>
      {/if}
    </div>

    {#if error}
      <div class="query-error">
        <TriangleAlert size={16} />
        <pre>{error}</pre>
      </div>
    {:else if result && result.isRead}
      <div class="grid-host">
        <DataGrid columns={result.columns} rows={result.rows} />
      </div>
    {:else if result}
      <div class="write-ok">
        <p>Statement executed successfully.</p>
        <p class="muted">{formatNumber(result.affected)} row(s) affected in {formatDuration(result.durationMs)}.</p>
      </div>
    {:else}
      <EmptyState
        title="Run a statement"
        message="Write SQL above and press Run. Select part of the editor to execute only that part. Results export to CSV."
      >
        <button class="btn sm" on:click={() => loadQuery('SELECT name, type FROM sqlite_schema WHERE name NOT LIKE \'sqlite_%\' ORDER BY type, name;')}>
          List schema objects
        </button>
        {#if sampleTable}
          <button class="btn sm primary" on:click={() => loadQuery(`SELECT *\nFROM ${quote(sampleTable)}\nLIMIT ${Math.min(100, rowLimit)};`)}>
            Preview {sampleTable}
          </button>
        {/if}
      </EmptyState>
    {/if}
  </section>
</div>

<Modal open={showSave} title="Save query" width={440} onClose={() => (showSave = false)}>
  <label class="stack">
    <span class="field-label">Name</span>
    <input
      type="text"
      bind:value={saveName}
      placeholder="e.g. Orders by status"
      autocomplete="off"
      on:keydown={(e) => e.key === 'Enter' && saveName.trim() && saveQuery()}
    />
  </label>
  <div slot="footer">
    <button class="btn" on:click={() => (showSave = false)}>Cancel</button>
    <button class="btn primary" disabled={saving || !saveName.trim()} on:click={saveQuery}>
      {saving ? 'Saving…' : 'Save query'}
    </button>
  </div>
</Modal>

<style>
  .page {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px 16px 14px;
  }
  .editor-panel {
    display: flex;
    flex-direction: column;
    flex: 0 0 42%;
    min-height: 150px;
    overflow: hidden;
  }
  .panel-head {
    flex: none;
  }
  .panel-head h2 {
    font-size: 12.5px;
  }
  .editor-host {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }
  .saved {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
    padding: 1px 2px;
    flex: none;
  }
  .saved-label {
    font-size: 10.5px;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }
  .saved-chip {
    display: inline-flex;
    align-items: center;
    height: 23px;
    border: 1px solid var(--border);
    border-radius: 20px;
    background: var(--surface-2);
    overflow: hidden;
  }
  .chip-name {
    border: 0;
    background: transparent;
    color: var(--text-soft);
    font-size: 11.5px;
    padding: 0 8px;
    height: 100%;
    max-width: 220px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chip-name:hover {
    color: var(--accent);
  }
  .chip-x {
    display: flex;
    align-items: center;
    border: 0;
    border-left: 1px solid var(--border);
    background: transparent;
    color: var(--faint);
    height: 100%;
    padding: 0 6px;
  }
  .chip-x:hover {
    color: var(--danger);
    background: var(--danger-soft);
  }
  .result-panel {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .result-head {
    padding-top: 9px;
    padding-bottom: 9px;
  }
  .grid-host {
    flex: 1;
    min-height: 0;
    display: flex;
    overflow: hidden;
  }
  .query-error {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin: 12px;
    padding: 11px 13px;
    border: 1px solid var(--danger);
    border-radius: var(--radius-sm);
    background: var(--danger-soft);
    color: var(--danger);
  }
  .query-error pre {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 12px;
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--text);
  }
  .write-ok {
    padding: 26px;
    text-align: center;
    color: var(--ok);
  }
  .write-ok p:first-child {
    font-weight: 600;
  }
  .write-ok p + p {
    margin-top: 4px;
    font-size: 12px;
  }
  .field-label {
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }
</style>
