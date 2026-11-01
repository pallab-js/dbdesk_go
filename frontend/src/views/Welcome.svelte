<script lang="ts">
  import { ChevronRight, Database, FolderOpen, ShieldCheck, Sparkles, Trash2 } from '@lucide/svelte'
  import type { model } from '../../wailsjs/go/models'
  import { formatDateTime } from '../lib/format'

  export let recent: model.ConnectionInfo[] = []
  export let onOpen: (readOnly: boolean) => void = () => {}
  export let onDemo: () => void = () => {}
  export let onOpenPath: (path: string, readOnly?: boolean) => void = () => {}
  export let onForget: (id: number, name: string) => void = () => {}

  let readOnly = false
</script>

<div class="welcome">
  <div class="hero">
    <div class="mark"><Database size={26} strokeWidth={2} /></div>
    <h1>Open a SQLite database</h1>
    <p>
      Browse tables, edit rows, run SQL and watch your own query telemetry. Everything stays on this
      machine.
    </p>

    <div class="actions">
      <button class="btn primary lg" on:click={() => onOpen(readOnly)}>
        <FolderOpen size={15} />
        Choose database file
      </button>
      <button class="btn lg" on:click={onDemo}>
        <Sparkles size={15} />
        Try the demo
      </button>
    </div>

    <label class="check readonly">
      <input type="checkbox" bind:checked={readOnly} />
      Open read-only (no writes, no row edits)
    </label>
  </div>

  {#if recent.length > 0}
    <section class="recent-panel panel">
      <div class="panel-head">
        <h2>Recent databases</h2>
        <span class="muted">{recent.length}</span>
      </div>
      <ul>
        {#each recent as c (c.id)}
          <li>
            <button class="recent-main" on:click={() => onOpenPath(c.path, c.readOnly)} title={c.path}>
              <span class="db-icon"><Database size={14} /></span>
              <span class="meta">
                <span class="name">{c.name}</span>
                <span class="path mono">{c.path}</span>
              </span>
              {#if c.readOnly}<span class="badge warn">read-only</span>{/if}
              <span class="when faint">{formatDateTime(c.lastOpenedAt)}</span>
              <ChevronRight size={14} />
            </button>
            <button
              class="btn ghost icon sm remove"
              title="Forget this database"
              aria-label="Forget database"
              on:click={() => onForget(c.id, c.name)}
            >
              <Trash2 size={13} />
            </button>
          </li>
        {/each}
      </ul>
    </section>
  {/if}

  <footer class="foot">
    <ShieldCheck size={13} />
    <span>Local-first · no telemetry leaves the app · SQLite only in this version</span>
  </footer>
</div>

<style>
  .welcome {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 22px;
    padding: 40px 28px;
  }
  .hero {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    max-width: 560px;
    padding-top: 4vh;
  }
  .mark {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 58px;
    height: 58px;
    margin-bottom: 18px;
    border-radius: 15px;
    background: var(--accent);
    color: var(--accent-contrast);
    box-shadow: 0 0 0 1px var(--accent-strong) inset, 0 18px 40px rgba(45, 212, 191, 0.18);
  }
  h1 {
    font-size: 25px;
    font-weight: 680;
    letter-spacing: -0.03em;
  }
  .hero p {
    margin-top: 8px;
    color: var(--muted);
    font-size: 13px;
    line-height: 1.6;
    max-width: 470px;
  }
  .actions {
    display: flex;
    gap: 9px;
    margin-top: 22px;
    flex-wrap: wrap;
    justify-content: center;
  }
  .btn.lg {
    height: 38px;
    padding: 0 16px;
    font-size: 13px;
    border-radius: 8px;
  }
  .readonly {
    margin-top: 13px;
    font-size: 11.5px;
    color: var(--faint);
  }
  .recent-panel {
    width: 100%;
    max-width: 640px;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 5px;
  }
  li {
    display: flex;
    align-items: center;
    gap: 2px;
    border-radius: var(--radius-sm);
  }
  li:hover {
    background: var(--surface-2);
  }
  .recent-main {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
    min-width: 0;
    padding: 7px 8px;
    border: 0;
    background: transparent;
    color: var(--text-soft);
    text-align: left;
  }
  .db-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 7px;
    background: var(--surface-3);
    color: var(--muted);
    flex: none;
  }
  .meta {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  .name {
    font-size: 12.5px;
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
    font-size: 10.5px;
    text-align: right;
  }
  .remove {
    margin-right: 4px;
  }
  .foot {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--faint);
    font-size: 11px;
    padding-bottom: 8px;
  }
</style>
