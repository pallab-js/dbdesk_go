<script lang="ts">
  import { Database, FolderOpen, Monitor, Moon, RotateCw, Sparkles, Sun, X } from '@lucide/svelte'
  import type { model } from '../../wailsjs/go/models'
  import { themeMode, cycleTheme } from '../lib/theme'

  export let connection: model.ConnectionInfo | null = null
  export let refreshing = false
  export let canRefresh = false
  export let onOpen: () => void = () => {}
  export let onDemo: () => void = () => {}
  export let onRefresh: () => void = () => {}
  export let onClose: () => void = () => {}

  $: themeIcon = $themeMode === 'light' ? Sun : $themeMode === 'dark' ? Moon : Monitor
  $: themeLabel = $themeMode === 'system' ? 'Theme: system' : $themeMode === 'light' ? 'Theme: light' : 'Theme: dark'
</script>

<header class="topbar">
  <div class="brand">
    <span class="mark"><Database size={15} strokeWidth={2.2} /></span>
    <span class="wordmark">DBDesk</span>
  </div>

  <span class="sep"></span>

  {#if connection}
    <div class="conn" title={connection.path}>
      <span class="dot" class:readonly={connection.readOnly}></span>
      <span class="name">{connection.name}</span>
      <span class="path">{connection.path}</span>
      {#if connection.readOnly}<span class="badge warn">read-only</span>{/if}
    </div>
  {:else}
    <span class="muted">No database open</span>
  {/if}

  <span class="spacer"></span>

  <div class="actions">
    <button class="btn ghost icon sm" title={themeLabel} aria-label={themeLabel} on:click={cycleTheme}>
      <svelte:component this={themeIcon} size={15} />
    </button>
    <button
      class="btn ghost icon sm"
      title="Refresh schema and stats"
      aria-label="Refresh schema and stats"
      disabled={!canRefresh || refreshing}
      on:click={onRefresh}
    >
      <RotateCw size={15} class={refreshing ? 'spin' : ''} />
    </button>
    {#if connection}
      <button class="btn ghost icon sm" title="Close database" aria-label="Close database" on:click={onClose}>
        <X size={15} />
      </button>
    {/if}
    <button class="btn sm" on:click={onDemo} title="Open the demo database (creates it on first use)">
      <Sparkles size={13} />
      Demo
    </button>
    <button class="btn sm primary" on:click={onOpen}>
      <FolderOpen size={13} />
      Open database
    </button>
  </div>
</header>

<style>
  .topbar {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 46px;
    padding: 0 14px;
    background: var(--bg-elevated);
    border-bottom: 1px solid var(--border);
    flex: none;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-right: 2px;
  }
  .mark {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 7px;
    background: var(--accent);
    color: var(--accent-contrast);
    box-shadow: 0 0 0 1px var(--accent-strong) inset;
  }
  .wordmark {
    font-size: 13.5px;
    font-weight: 700;
    letter-spacing: -0.015em;
  }
  .sep {
    width: 1px;
    height: 18px;
    background: var(--border);
    margin: 0 2px;
  }
  .conn {
    display: flex;
    align-items: center;
    gap: 7px;
    min-width: 0;
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--ok);
    box-shadow: 0 0 6px var(--ok);
    flex: none;
  }
  .dot.readonly {
    background: var(--warn);
    box-shadow: 0 0 6px var(--warn);
  }
  .name {
    font-weight: 600;
    font-size: 12.5px;
    white-space: nowrap;
  }
  .path {
    max-width: 380px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    direction: rtl;
    text-align: left;
    color: var(--faint);
    font-family: var(--font-mono);
    font-size: 11px;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }
</style>
