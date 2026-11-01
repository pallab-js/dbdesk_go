<script lang="ts">
  import { FileUp, LoaderCircle, Table } from '@lucide/svelte'
  import { ImportData } from '../../wailsjs/go/main/App.js'
  import { model } from '../../wailsjs/go/models'
  import type { model as M } from '../../wailsjs/go/models'
  import { errorText } from '../lib/format'
  import { bumpData, notify } from '../lib/stores'

  export let schema: M.Schema | null = null

  let selectedTable = ''
  let format = 'csv'
  let data = ''
  let hasHeader = true
  let truncate = false
  let importing = false
  let result: M.ImportResult | null = null

  $: tables = schema?.tables.filter((t) => t.type === 'table') ?? []
  $: if (!selectedTable && tables.length > 0) selectedTable = tables[0].name

  async function importFile(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    const text = await file.text()
    data = text
    if (file.name.endsWith('.json')) format = 'json'
    else format = 'csv'
    input.value = ''
  }

  async function doImport() {
    if (!selectedTable || !data.trim()) {
      notify('Select table and provide data', 'error')
      return
    }
    importing = true
    result = null
    try {
      const res = await ImportData(
        new model.ImportRequest({
          table: selectedTable,
          format,
          data,
          hasHeader,
          truncate,
          columnMap: [],
        }),
      )
      result = res
      if (res.inserted > 0) {
        notify(`Imported ${res.inserted} rows`, 'success')
        bumpData()
      }
      if (res.errors > 0) {
        notify(`${res.errors} rows had errors`, 'info')
      }
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      importing = false
    }
  }
</script>

<div class="page">
  <header class="head">
    <div class="title">
      <FileUp size={16} />
      <h1>Import Data</h1>
    </div>
    <span class="faint">Import CSV or JSON into tables</span>
  </header>

  <section class="panel">
    <div class="panel-head">
      <h2>Import Settings</h2>
    </div>
    <div class="form">
      <label class="field">
        <span class="label">Target Table</span>
        <select bind:value={selectedTable}>
          {#each tables as t}
            <option value={t.name}>{t.name}</option>
          {/each}
        </select>
      </label>
      <div class="row">
        <label class="field">
          <span class="label">Format</span>
          <select bind:value={format}>
            <option value="csv">CSV</option>
            <option value="json">JSON</option>
          </select>
        </label>
        {#if format === 'csv'}
          <label class="check">
            <input type="checkbox" bind:checked={hasHeader} />
            <span>First row is header</span>
          </label>
        {/if}
        <label class="check">
          <input type="checkbox" bind:checked={truncate} />
          <span>Truncate table before import</span>
        </label>
      </div>
      <div class="upload">
        <input type="file" accept=".csv,.json,.txt" on:change={importFile} id="file" />
        <label for="file" class="btn"><FileUp size={13} /> Choose file</label>
        {#if data}<span class="faint">{data.split('\n').length} lines loaded</span>{/if}
      </div>
      <label class="field">
        <span class="label">Data</span>
        <textarea bind:value={data} placeholder={format === 'csv' ? 'name,email\nAlice,alice@example.com' : '[{"name": "Alice", "email": "..."}]'}></textarea>
      </label>
      <div class="actions">
        <button class="btn primary" disabled={importing || !selectedTable || !data.trim()} on:click={doImport}>
          {#if importing}<LoaderCircle size={13} class="spin" />{:else}<Table size={13} />{/if}
          Import
        </button>
      </div>
    </div>
  </section>

  {#if result}
    <section class="panel">
      <div class="panel-head"><h2>Import Result</h2></div>
      <div class="result">
        <div class="stat"><span class="num">{result.inserted}</span><span class="muted">inserted</span></div>
        <div class="stat"><span class="num warn">{result.errors}</span><span class="muted">errors</span></div>
        {#if result.details}<pre class="details">{result.details}</pre>{/if}
      </div>
    </section>
  {/if}
</div>

<style>
  .page { flex: 1; min-height: 0; display: flex; flex-direction: column; gap: 12px; padding: 16px; overflow: auto; }
  .head { display: flex; flex-direction: column; gap: 4px; }
  .title { display: flex; align-items: center; gap: 8px; }
  .title h1 { font-size: 15px; font-weight: 600; }
  .form { display: flex; flex-direction: column; gap: 12px; padding: 12px; }
  .field { display: flex; flex-direction: column; gap: 6px; }
  .label { font-size: 12px; color: var(--text-soft); }
  .row { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
  .check { display: flex; align-items: center; gap: 6px; font-size: 12px; }
  .upload { display: flex; align-items: center; gap: 10px; }
  input[type=file] { display: none; }
  textarea { min-height: 200px; font-family: var(--font-mono); font-size: 12px; resize: vertical; }
  .actions { display: flex; }
  .result { display: flex; gap: 20px; padding: 12px; align-items: center; flex-wrap: wrap; }
  .stat { display: flex; flex-direction: column; align-items: center; gap: 2px; }
  .num { font-size: 24px; font-weight: 600; }
  .num.warn { color: var(--warning); }
  .details { width: 100%; background: var(--surface-2); padding: 8px; border-radius: var(--radius-sm); font-size: 11px; overflow: auto; }
</style>
