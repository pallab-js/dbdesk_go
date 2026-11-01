<script lang="ts">
  import {
    Check,
    Database,
    Monitor,
    Moon,
    RotateCcw,
    Save,
    Sun,
    Trash2,
  } from '@lucide/svelte'
  import { ForgetConnection, ResetDemoDatabase, SaveSettings } from '../../wailsjs/go/main/App.js'
  import { model } from '../../wailsjs/go/models'
  import type { model as M } from '../../wailsjs/go/models'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import { errorText, formatDateTime } from '../lib/format'
  import { notify } from '../lib/stores'
  import { setTheme, themeMode } from '../lib/theme'
  import type { ThemeMode } from '../lib/theme'

  export let settings: M.Settings
  export let recent: M.ConnectionInfo[] = []
  export let appDataDir = ''
  export let version = ''
  export let onOpenPath: (path: string, readOnly?: boolean) => void = () => {}
  export let onDemo: () => void = () => {}

  let draft = { ...settings }
  let saving = false
  let resetting = false
  let confirmReset = false

  $: dirty =
    draft.rowLimit !== settings.rowLimit ||
    draft.timeoutSeconds !== settings.timeoutSeconds ||
    draft.historyEnabled !== settings.historyEnabled ||
    draft.retentionDays !== settings.retentionDays ||
    draft.autoRefresh !== settings.autoRefresh

  const themes: { id: ThemeMode; label: string; icon: typeof Sun }[] = [
    { id: 'system', label: 'System', icon: Monitor },
    { id: 'light', label: 'Light', icon: Sun },
    { id: 'dark', label: 'Dark', icon: Moon },
  ]

  async function save() {
    saving = true
    try {
      settings = await SaveSettings(new model.Settings({ ...draft }))
      draft = { ...settings }
      notify('Settings saved', 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      saving = false
    }
  }

  async function forget(id: number, name: string) {
    try {
      await ForgetConnection(id)
      notify(`Removed ${name} from recent databases`, 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    }
  }

  async function resetDemo() {
    resetting = true
    try {
      await ResetDemoDatabase()
      confirmReset = false
      notify('Demo database recreated and opened', 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      resetting = false
    }
  }
</script>

<div class="page">
  <header class="page-head">
    <div>
      <h1>Settings</h1>
      <p class="muted">Preferences are stored locally in the app data folder.</p>
    </div>
    <div class="row">
      {#if dirty}<span class="badge warn">unsaved changes</span>{/if}
      <button class="btn primary" disabled={!dirty || saving} on:click={save}>
        {#if saving}<RotateCcw size={13} class="spin" />{:else}<Save size={13} />{/if}
        {saving ? 'Saving…' : 'Save settings'}
      </button>
    </div>
  </header>

  <div class="columns">
    <div class="col">
      <section class="panel">
        <div class="panel-head"><h2>Appearance</h2></div>
        <div class="panel-body stack">
          <div class="setting">
            <div>
              <div class="title">Theme</div>
              <div class="desc">Applies immediately. “System” follows your desktop preference.</div>
            </div>
            <div class="segmented">
              {#each themes as t (t.id)}
                <button class:active={$themeMode === t.id} on:click={() => setTheme(t.id)}>
                  <svelte:component this={t.icon} size={12} />
                  {t.label}
                </button>
              {/each}
            </div>
          </div>
          <div class="setting">
            <div>
              <div class="title">Auto-refresh dashboard</div>
              <div class="desc">Reload statistics when database content changes.</div>
            </div>
            <label class="check"><input type="checkbox" bind:checked={draft.autoRefresh} /> Enabled</label>
          </div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-head"><h2>Query execution</h2></div>
        <div class="panel-body form-grid">
          <label class="field">
            <span>Row limit</span>
            <input type="number" min="100" max="10000" step="100" bind:value={draft.rowLimit} />
            <small>Maximum rows returned per query (100–10,000).</small>
          </label>
          <label class="field">
            <span>Timeout (seconds)</span>
            <input type="number" min="5" max="120" step="5" bind:value={draft.timeoutSeconds} />
            <small>Queries and table loads stop after this long.</small>
          </label>
          <label class="field">
            <span>History retention (days)</span>
            <input type="number" min="1" max="365" bind:value={draft.retentionDays} />
            <small>Older statements are pruned at startup.</small>
          </label>
          <div class="field">
            <span>Query history</span>
            <label class="check history-check">
              <input type="checkbox" bind:checked={draft.historyEnabled} />
              Record statements, timings and errors
            </label>
            <small>Turning this off keeps existing history but stops new entries.</small>
          </div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-head"><h2>Storage</h2></div>
        <div class="panel-body stack">
          <div class="path-row">
            <span class="muted">App data</span>
            <code class="mono">{appDataDir || '—'}</code>
          </div>
          <div class="path-row">
            <span class="muted">Demo database</span>
            <code class="mono">ecommerce.db (8 tables, 1 view)</code>
          </div>
          <div class="row buttons">
            <button class="btn sm" on:click={onDemo}><Database size={13} /> Open demo</button>
            <button class="btn sm danger" disabled={resetting} on:click={() => (confirmReset = true)}>
              <RotateCcw size={13} />
              Reset demo data
            </button>
          </div>
          <p class="desc">
            Resetting deletes the demo file and seeds it again with the same deterministic data. Your other databases are untouched.
          </p>
        </div>
      </section>
    </div>

    <div class="col">
      <section class="panel">
        <div class="panel-head">
          <h2>Recent databases</h2>
          <span class="muted">{recent.length}</span>
        </div>
        {#if recent.length === 0}
          <p class="empty-hint">Databases you open appear here.</p>
        {:else}
          <ul class="recent">
            {#each recent as c (c.id)}
              <li>
                <button class="recent-main" title={c.path} on:click={() => onOpenPath(c.path, c.readOnly)}>
                  <span class="db-icon"><Database size={13} /></span>
                  <span class="meta">
                    <span class="name">{c.name}</span>
                    <span class="path mono">{c.path}</span>
                  </span>
                  <span class="when faint">{formatDateTime(c.lastOpenedAt)}</span>
                </button>
                <button class="btn ghost icon sm remove" title="Forget this database" aria-label="Forget database" on:click={() => forget(c.id, c.name)}>
                  <Trash2 size={13} />
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <section class="panel">
        <div class="panel-head"><h2>About</h2></div>
        <div class="panel-body stack about">
          <div class="about-row"><span>Version</span><span class="mono">{version || '—'}</span></div>
          <div class="about-row"><span>Engine</span><span class="mono">SQLite (modernc, pure Go)</span></div>
          <div class="about-row"><span>Shell</span><span class="mono">Go + Wails + Svelte</span></div>
          <div class="divider"></div>
          <div class="shortcuts">
            <div><span class="kbd">Ctrl</span> <span class="kbd">Enter</span><span class="muted">Run query / selection</span></div>
            <div><span class="kbd">Esc</span><span class="muted">Close dialogs</span></div>
            <div><span><Check size={12} /></span><span class="muted">Identifiers are quoted and values parameterized</span></div>
          </div>
        </div>
      </section>
    </div>
  </div>
</div>

<ConfirmDialog
  open={confirmReset}
  title="Reset demo database"
  body="Delete the demo database and seed it again? Any edits you made to the demo are lost."
  confirmLabel="Reset demo"
  danger
  busy={resetting}
  onConfirm={resetDemo}
  onCancel={() => (confirmReset = false)}
/>

<style>
  .page {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 18px 20px 30px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .page-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 14px;
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
  .columns {
    display: grid;
    grid-template-columns: minmax(0, 1.25fr) minmax(300px, 1fr);
    gap: 14px;
    align-items: start;
  }
  .col {
    display: flex;
    flex-direction: column;
    gap: 14px;
    min-width: 0;
  }
  .setting {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
  }
  .title {
    font-size: 12.5px;
    font-weight: 550;
  }
  .desc {
    color: var(--faint);
    font-size: 11.5px;
    line-height: 1.5;
    margin-top: 1px;
  }
  .history-check {
    height: 30px;
    font-size: 12.5px;
  }
  .path-row {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .path-row .muted {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .path-row code {
    font-size: 11.5px;
    color: var(--text-soft);
    word-break: break-all;
  }
  .buttons {
    margin-top: 2px;
  }
  .recent {
    list-style: none;
    margin: 0;
    padding: 5px;
  }
  .recent li {
    display: flex;
    align-items: center;
    gap: 2px;
    border-radius: var(--radius-sm);
  }
  .recent li:hover {
    background: var(--surface-2);
  }
  .recent-main {
    display: flex;
    align-items: center;
    gap: 9px;
    flex: 1;
    min-width: 0;
    padding: 6px 7px;
    border: 0;
    background: transparent;
    text-align: left;
    color: var(--text-soft);
  }
  .db-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border-radius: 6px;
    background: var(--surface-3);
    color: var(--muted);
    flex: none;
  }
  .meta {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
  }
  .name {
    font-size: 12px;
    font-weight: 600;
    color: var(--text);
  }
  .path {
    font-size: 10.5px;
    color: var(--faint);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    direction: rtl;
    text-align: left;
  }
  .when {
    flex: none;
    font-size: 10px;
    width: 96px;
    text-align: right;
  }
  .remove {
    margin-right: 3px;
  }
  .about {
    gap: 6px;
  }
  .about-row {
    display: flex;
    justify-content: space-between;
    gap: 14px;
    font-size: 12px;
    color: var(--muted);
  }
  .about-row span:last-child {
    color: var(--text-soft);
  }
  .shortcuts {
    display: flex;
    flex-direction: column;
    gap: 7px;
    font-size: 11.5px;
  }
  .shortcuts > div {
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .shortcuts .muted:last-child {
    margin-left: 4px;
  }
  @media (max-width: 1080px) {
    .columns {
      grid-template-columns: 1fr;
    }
  }
</style>
