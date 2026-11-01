<script lang="ts">
  import {
    ChevronDown,
    ChevronRight,
    LoaderCircle,
    RefreshCw,
    Search,
    SquarePen,
    Trash2,
    TriangleAlert,
  } from '@lucide/svelte'
  import { onMount } from 'svelte'
  import { ClearHistory, History } from '../../wailsjs/go/main/App.js'
  import type { model } from '../../wailsjs/go/models'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import { errorText, formatDateTime, formatDuration, formatNumber, timeAgo } from '../lib/format'
  import { bumpData, notify, openInEditor } from '../lib/stores'

  let entries: model.HistoryEntry[] = []
  let loading = false
  let error = ''
  let filter = ''
  let expanded: number | null = null
  let confirmClear = false
  let clearing = false

  $: filtered = filter.trim()
    ? entries.filter((e) => e.sql.toLowerCase().includes(filter.trim().toLowerCase()))
    : entries
  $: stats = {
    total: entries.length,
    failed: entries.filter((e) => !e.success).length,
  }

  onMount(load)

  async function load() {
    loading = true
    error = ''
    try {
      entries = await History(500)
    } catch (e) {
      error = errorText(e)
    } finally {
      loading = false
    }
  }

  async function clearAll() {
    clearing = true
    try {
      await ClearHistory()
      entries = []
      expanded = null
      confirmClear = false
      bumpData()
      notify('Query history cleared', 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      clearing = false
    }
  }
</script>

<div class="page">
  <header class="page-head">
    <div>
      <h1>History</h1>
      <p class="muted">
        {formatNumber(stats.total)} statements stored locally · {formatNumber(stats.failed)} failed
      </p>
    </div>
    <div class="row">
      <div class="search">
        <Search size={12} />
        <input type="search" aria-label="Filter statements" placeholder="Filter statements" bind:value={filter} spellcheck="false" />
      </div>
      <button class="btn sm icon" title="Reload history" aria-label="Reload history" disabled={loading} on:click={load}>
        <RefreshCw size={13} />
      </button>
      <button class="btn sm danger" disabled={entries.length === 0 || loading} on:click={() => (confirmClear = true)}>
        <Trash2 size={13} />
        Clear
      </button>
    </div>
  </header>

  {#if error}
    <div class="panel error-panel">
      <TriangleAlert size={16} />
      <span class="mono">{error}</span>
    </div>
  {:else if loading && entries.length === 0}
    <div class="panel loading"><LoaderCircle size={17} class="spin" /> Loading history…</div>
  {:else if filtered.length === 0}
    <div class="panel">
      <EmptyState
        title={entries.length === 0 ? 'No queries yet' : 'No matching statements'}
        message={entries.length === 0
          ? 'Statements run in the SQL editor are recorded here, with timing and errors.'
          : 'Try a different filter, or clear it to see every statement.'}
      />
    </div>
  {:else}
    <div class="panel list-panel">
      {#each filtered as h (h.id)}
        <div class="entry" class:open={expanded === h.id}>
          <button class="entry-row" aria-expanded={expanded === h.id} on:click={() => (expanded = expanded === h.id ? null : h.id)}>
            <span class="caret">
              {#if expanded === h.id}<ChevronDown size={13} />{:else}<ChevronRight size={13} />{/if}
            </span>
            <span class="status" class:bad={!h.success} role="img" aria-label={h.success ? "succeeded" : "failed"} title={h.success ? "succeeded" : "failed"}></span>
            <span class="kind">{h.kind || 'sql'}</span>
            <span class="sql mono truncate">{h.sql}</span>
            <span class="dur muted">{formatDuration(h.durationMs)}</span>
            <span class="rows muted">{formatNumber(h.rowCount)} rows</span>
            <span class="when faint" title={formatDateTime(h.createdAt)}>{timeAgo(h.createdAt)}</span>
          </button>
          {#if expanded === h.id}
            <div class="detail">
              <pre class="mono">{h.sql}</pre>
              {#if h.error}
                <div class="err mono">{h.error}</div>
              {/if}
              <div class="detail-foot">
                <span class="faint">{formatDateTime(h.createdAt)} · id {h.id}</span>
                <span class="spacer"></span>
                <button class="btn sm" on:click={() => openInEditor(h.sql)}>
                  <SquarePen size={12} />
                  Open in editor
                </button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<ConfirmDialog
  open={confirmClear}
  title="Clear query history"
  body="Delete every stored statement, including timings and errors? This cannot be undone."
  confirmLabel="Clear history"
  danger
  busy={clearing}
  onConfirm={clearAll}
  onCancel={() => (confirmClear = false)}
/>

<!-- Small inline search icon to avoid importing a second icon name. -->
<style>
  .page {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 18px 18px;
  }
  .page-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
  }
  .page-head h1 {
    font-size: 19px;
    font-weight: 650;
    letter-spacing: -0.02em;
  }
  .page-head p {
    margin-top: 2px;
    font-size: 12.5px;
  }
  .search {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 30px;
    padding: 0 9px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--bg-elevated);
    color: var(--faint);
    width: 240px;
  }
  .search input {
    flex: 1;
    height: 28px;
    padding: 0;
    border: 0;
    background: transparent;
    outline: none;
    font-size: 12px;
  }
  .list-panel {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 4px;
  }
  .entry {
    border-radius: var(--radius-sm);
  }
  .entry.open {
    background: var(--surface-2);
  }
  .entry-row {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    height: 31px;
    padding: 0 8px;
    border: 0;
    background: transparent;
    color: var(--text-soft);
    text-align: left;
  }
  .entry-row:hover {
    background: var(--surface-2);
    border-radius: var(--radius-sm);
  }
  .caret {
    display: flex;
    color: var(--faint);
  }
  .status {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--ok);
    flex: none;
  }
  .status.bad {
    background: var(--danger);
  }
  .kind {
    width: 58px;
    flex: none;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--faint);
  }
  .sql {
    flex: 1;
    font-size: 11.5px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .dur {
    width: 64px;
    text-align: right;
    font-size: 11px;
    font-variant-numeric: tabular-nums;
  }
  .rows {
    width: 72px;
    text-align: right;
    font-size: 11px;
  }
  .when {
    width: 62px;
    text-align: right;
    font-size: 10.5px;
  }
  .detail {
    padding: 4px 12px 12px 32px;
  }
  .detail pre {
    margin: 0;
    padding: 10px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    font-size: 11.5px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 260px;
    overflow: auto;
  }
  .err {
    margin-top: 8px;
    padding: 7px 10px;
    background: var(--danger-soft);
    border: 1px solid var(--danger);
    border-radius: var(--radius-sm);
    color: var(--danger);
    font-size: 11.5px;
    white-space: pre-wrap;
  }
  .detail-foot {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
    font-size: 11px;
  }
  .error-panel {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 13px;
    border-color: var(--danger);
    color: var(--danger);
  }
  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 50px;
    color: var(--muted);
  }
</style>
