<script lang="ts">
  import {
    ChevronLeft,
    ChevronRight,
    Download,
    Filter,
    LoaderCircle,
    Plus,
    RefreshCw,
    Search,
    TriangleAlert,
  } from '@lucide/svelte'
  import { DeleteRow, ExportLastResultCSV, PreviewTable } from '../../wailsjs/go/main/App.js'
  import { model } from '../../wailsjs/go/models'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import DataGrid from '../components/DataGrid.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import RowEditor from '../components/RowEditor.svelte'
  import type { model as M } from '../../wailsjs/go/models'
  import { errorText, formatNumber } from '../lib/format'
  import { activeTable, bumpData, dataVersion, notify, openTable } from '../lib/stores'

  export let schema: M.Schema | null = null

  let pageData: M.TablePage | null = null
  let loading = false
  let error = ''
  let page = 1
  let pageSize = 100
  let sortColumn = ''
  let sortDir = 'asc'
  let lastTable = ''

  let editorOpen = false
  let editorRow: any[] | null = null
  let deleteRow: any[] | null = null
  let deleting = false
  let exporting = false
  let requestToken = 0

  const pageSizes = [50, 100, 250, 500, 1000]

  $: tables = schema?.tables ?? []
  $: selected = tables.find((t) => t.name === $activeTable) ?? null
  $: if (schema && !$activeTable && tables.length > 0) openTable(tables[0].name)
  $: if ($activeTable && $activeTable !== lastTable) {
    lastTable = $activeTable
    page = 1
    sortColumn = ''
    sortDir = 'asc'
    pageData = null
  }
  // Reload when paging, sorting or data version changes.
  $: loadKey = [$activeTable, page, pageSize, sortColumn, sortDir, $dataVersion].join('|')
  $: if ($activeTable && $activeTable === lastTable) load(loadKey)

  async function load(_key: string) {
    const token = ++requestToken
    loading = true
    error = ''
    try {
      const res = await PreviewTable(
        new model.TablePageRequest({
          table: $activeTable,
          page,
          pageSize,
          sortColumn,
          sortDir,
        }),
      )
      if (token !== requestToken) return
      pageData = res
      if (res.page !== page) page = res.page
    } catch (e) {
      if (token !== requestToken) return
      error = errorText(e)
      pageData = null
    } finally {
      if (token === requestToken) loading = false
    }
  }

  function sortBy(name: string) {
    if (sortColumn === name) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc'
    } else {
      sortColumn = name
      sortDir = 'asc'
    }
    page = 1
  }

  function changeSize(e: Event) {
    pageSize = Number((e.target as HTMLSelectElement).value)
    page = 1
  }

  function beginInsert() {
    editorRow = null
    editorOpen = true
  }

  function beginEdit(row: any[]) {
    editorRow = row
    editorOpen = true
  }

  function keysFor(row: any[]): model.KeyValue[] {
    if (!pageData) return []
    return pageData.keyNames.map(
      (name, i) => new model.KeyValue({ name, value: row[pageData!.keyIndexes[i]] }),
    )
  }

  async function confirmDelete() {
    if (!pageData || !deleteRow) return
    deleting = true
    try {
      await DeleteRow(
        new model.RowDeleteRequest({ table: pageData.table, keys: keysFor(deleteRow) }),
      )
      notify('Row deleted', 'success')
      deleteRow = null
      bumpData()
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      deleting = false
    }
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

  $: pageCount =
    pageData && pageData.total >= 0 ? Math.max(1, Math.ceil(pageData.total / pageData.pageSize)) : null
  $: keyPreview = deleteRow && pageData
    ? pageData.keyNames
        .map((n, i) => `${n} = ${JSON.stringify(deleteRow![pageData!.keyIndexes[i]])}`)
        .join(', ')
    : ''
</script>

<div class="page">
  <header class="toolbar">
    <div class="left">
      <select class="table-select" aria-label="Select table" value={$activeTable} on:change={(e) => openTable((e.target as HTMLSelectElement).value)}>
        {#each tables as t (t.name)}
          <option value={t.name}>{t.name}{t.type === 'view' ? ' (view)' : ''}</option>
        {/each}
      </select>
      {#if selected}
        <span class="badge">{formatNumber(selected.columns.length)} columns</span>
        <span class="badge">{formatNumber(pageData?.total ?? 0)} rows</span>
        {#if pageData && !pageData.editable}
          <span class="badge warn">{selected.type === 'view' ? 'view · read-only' : 'read-only'}</span>
        {/if}
      {/if}
      {#if loading}<LoaderCircle size={14} class="spin muted" />{/if}
    </div>

    <div class="right">
      <div class="pager">
        <button class="btn ghost icon sm" disabled={page <= 1 || loading} title="Previous page" on:click={() => (page = page - 1)}>
          <ChevronLeft size={15} />
        </button>
        <span class="page-label">
          {#if pageData}
            page {pageData.page}{#if pageCount} / {formatNumber(pageCount)}{/if}
          {:else}—{/if}
        </span>
        <button class="btn ghost icon sm" disabled={!pageData?.hasNext || loading} title="Next page" on:click={() => (page = page + 1)}>
          <ChevronRight size={15} />
        </button>
      </div>
      <select value={String(pageSize)} aria-label="Rows per page" on:change={changeSize} title="Rows per page">
        {#each pageSizes as n (n)}
          <option value={String(n)}>{n} / page</option>
        {/each}
      </select>
      <button class="btn sm icon" title="Refresh rows" aria-label="Refresh rows" disabled={loading} on:click={() => bumpData()}>
        <RefreshCw size={13} />
      </button>
      <button class="btn sm" disabled={!pageData || exporting} on:click={exportCSV} title="Export the current page as CSV">
        <Download size={13} />
        Export
      </button>
      <button class="btn sm primary" disabled={!pageData?.editable} on:click={beginInsert} title={pageData?.editable ? 'Insert a row' : 'This object is read-only'}>
        <Plus size={13} />
        Insert row
      </button>
    </div>
  </header>

  {#if !schema || !selected}
    <div class="panel">
      <EmptyState title="No table selected" message="Open a database and pick a table in the sidebar." />
    </div>
  {:else if error}
    <div class="panel error-panel">
      <TriangleAlert size={18} />
      <div>
        <h2>Could not load rows</h2>
        <p class="mono">{error}</p>
      </div>
      <button class="btn" on:click={() => bumpData()}>Try again</button>
    </div>
  {:else}
    <div class="panel grid-panel">
      {#if pageData}
        <DataGrid
          columns={pageData.columns}
          rows={pageData.rows}
          keyIndexes={pageData.keyIndexes}
          editable={pageData.editable}
          sortable
          {sortColumn}
          {sortDir}
          onSort={sortBy}
          onEdit={beginEdit}
          onDelete={(row) => (deleteRow = row)}
        />
        <footer class="grid-foot">
          <span class="muted">
            {#if pageData.countError}
              row count unavailable · {pageData.rows.length} rows on this page
            {:else}
              {formatNumber(pageData.total)} rows total
            {/if}
          </span>
          <span class="spacer"></span>
          <span class="faint">
            {#if pageData.editable}Use the row icons to edit or delete · {/if}Click a column header to sort
          </span>
        </footer>
      {:else}
        <div class="loading-body"><LoaderCircle class="spin" size={18} /> Loading rows…</div>
      {/if}
    </div>
  {/if}
</div>

{#if editorOpen && selected && pageData}
  <RowEditor
    table={selected}
    columns={pageData.columns}
    row={editorRow}
    keyIndexes={pageData.keyIndexes}
    keyNames={pageData.keyNames}
    onClose={() => (editorOpen = false)}
  />
{/if}

<ConfirmDialog
  open={deleteRow !== null}
  title="Delete row"
  body={`Delete this row from ${selected?.name ?? 'the table'}? ${keyPreview ? '(' + keyPreview + ')' : ''} This cannot be undone.`}
  confirmLabel="Delete row"
  danger
  busy={deleting}
  onConfirm={confirmDelete}
  onCancel={() => (deleteRow = null)}
/>

<style>
  .page {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    padding: 12px 16px 14px;
    gap: 10px;
  }
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }
  .left,
  .right {
    display: flex;
    align-items: center;
    gap: 7px;
    flex-wrap: wrap;
  }
  .table-select {
    min-width: 190px;
    height: 30px;
    font-family: var(--font-mono);
    font-size: 12px;
    background-color: var(--surface-2);
  }
  .pager {
    display: flex;
    align-items: center;
    gap: 1px;
    padding: 1px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .page-label {
    min-width: 74px;
    padding: 0 4px;
    text-align: center;
    font-size: 11.5px;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }
  .grid-panel {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .grid-foot {
    display: flex;
    align-items: center;
    padding: 6px 12px;
    border-top: 1px solid var(--border);
    background: var(--bg-elevated);
    font-size: 11.5px;
  }
  .loading-body {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 60px;
    color: var(--muted);
  }
  .error-panel {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 16px;
    border-color: var(--danger);
    color: var(--danger);
  }
  .error-panel h2 {
    font-size: 13px;
  }
  .error-panel p {
    margin-top: 4px;
    color: var(--text-soft);
    white-space: pre-wrap;
  }
  .error-panel .btn {
    margin-left: auto;
  }
</style>
