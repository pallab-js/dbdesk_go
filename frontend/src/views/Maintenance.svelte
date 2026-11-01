<script lang="ts">
  import { Activity, Database, LoaderCircle, ShieldCheck, Wrench } from '@lucide/svelte'
  import { onMount } from 'svelte'
  import { GetPragmas, RunMaintenance, SetPragma } from '../../wailsjs/go/main/App.js'
  import { model } from '../../wailsjs/go/models'
  import type { model as M } from '../../wailsjs/go/models'
  import { errorText, formatDuration } from '../lib/format'
  import { notify } from '../lib/stores'

  export let connection: M.ConnectionInfo | null = null

  let pragmas: M.PragmaInfo[] = []
  let loading = false
  let running = ''
  let results: M.MaintenanceResult[] = []

  onMount(() => {
    void load()
  })

  async function load() {
    if (!connection) return
    loading = true
    try {
      pragmas = await GetPragmas()
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      loading = false
    }
  }

  async function run(action: string) {
    running = action
    try {
      const res = await RunMaintenance(action)
      results = [res, ...results].slice(0, 10)
      if (res.success) {
        notify(`${action} completed`, 'success')
      } else {
        notify(res.message, 'error')
      }
      await load()
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      running = ''
    }
  }

  const actions = [
    { id: 'vacuum', label: 'VACUUM', desc: 'Reclaim unused space and defragment database' },
    { id: 'analyze', label: 'ANALYZE', desc: 'Update query planner statistics' },
    { id: 'reindex', label: 'REINDEX', desc: 'Rebuild indexes' },
    { id: 'integrity_check', label: 'Integrity Check', desc: 'Check database for corruption' },
  ]
</script>

<div class="page">
  <header class="head">
    <div class="title">
      <Wrench size={16} />
      <h1>Maintenance & Pragmas</h1>
    </div>
    <span class="faint">Database optimization and configuration</span>
  </header>

  <section class="panel">
    <div class="panel-head">
      <div class="row">
        <Database size={14} />
        <h2>Database Maintenance</h2>
      </div>
    </div>
    <div class="actions">
      {#each actions as a}
        <button class="action-card" disabled={!!running || !connection} on:click={() => run(a.id)}>
          <div class="action-head">
            <span class="action-label">{a.label}</span>
            {#if running === a.id}<LoaderCircle size={14} class="spin" />{/if}
          </div>
          <p class="action-desc">{a.desc}</p>
        </button>
      {/each}
    </div>
  </section>

  {#if results.length > 0}
    <section class="panel">
      <div class="panel-head">
        <div class="row">
          <Activity size={14} />
          <h2>Recent Results</h2>
        </div>
      </div>
      <div class="results">
        {#each results as r}
          <div class="result-row">
            <span class="badge {r.success ? 'ok' : 'warn'}">{r.action}</span>
            <span class="result-msg">{r.message}</span>
            <span class="spacer"></span>
            <span class="faint">{formatDuration(r.durationMs)}</span>
          </div>
        {/each}
      </div>
    </section>
  {/if}

  <section class="panel">
    <div class="panel-head">
      <div class="row">
        <ShieldCheck size={14} />
        <h2>Pragmas</h2>
      </div>
      <button class="btn sm" disabled={loading} on:click={load}>Refresh</button>
    </div>
    <div class="pragma-table">
      {#if loading}
        <div class="loading"><LoaderCircle class="spin" size={14} /> Loading…</div>
      {:else if pragmas.length === 0}
        <div class="empty">No pragma information available</div>
      {:else}
        {#each pragmas as p}
          <div class="pragma-row">
            <span class="pragma-name">PRAGMA {p.name}</span>
            <span class="pragma-value">{p.value}</span>
          </div>
        {/each}
      {/if}
    </div>
  </section>
</div>

<style>
  .page { flex: 1; min-height: 0; display: flex; flex-direction: column; gap: 12px; padding: 16px; overflow: auto; }
  .head { display: flex; flex-direction: column; gap: 4px; }
  .title { display: flex; align-items: center; gap: 8px; }
  .title h1 { font-size: 15px; font-weight: 600; }
  .actions { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 8px; padding: 12px; }
  .action-card { text-align: left; padding: 10px 12px; border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface-2); display: flex; flex-direction: column; gap: 4px; cursor: pointer; }
  .action-card:hover:not(:disabled) { border-color: var(--accent); }
  .action-card:disabled { opacity: 0.5; cursor: not-allowed; }
  .action-head { display: flex; align-items: center; justify-content: space-between; }
  .action-label { font-weight: 600; font-size: 12.5px; }
  .action-desc { font-size: 11.5px; color: var(--muted); }
  .results { display: flex; flex-direction: column; }
  .result-row { display: flex; align-items: center; gap: 10px; padding: 8px 12px; border-bottom: 1px solid var(--border); }
  .result-row:last-child { border-bottom: none; }
  .result-msg { font-size: 12px; color: var(--text-soft); }
  .pragma-table { display: flex; flex-direction: column; }
  .pragma-row { display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; border-bottom: 1px solid var(--border); }
  .pragma-row:last-child { border-bottom: none; }
  .pragma-name { font-family: var(--font-mono); font-size: 12px; color: var(--text-soft); }
  .pragma-value { font-family: var(--font-mono); font-size: 12px; }
  .loading, .empty { padding: 16px; text-align: center; color: var(--muted); display: flex; align-items: center; justify-content: center; gap: 8px; }
</style>
