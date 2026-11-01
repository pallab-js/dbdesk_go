<script lang="ts">
  import {
    Eye,
    History,
    LayoutDashboard,
    Search,
    Settings,
    SquareTerminal,
    Table2,
  } from '@lucide/svelte'
  import type { model } from '../../wailsjs/go/models'
  import { activeTable, openTable, view } from '../lib/stores'

  export let schema: model.Schema | null = null
  export let version = ''

  let filter = ''

  $: tables = (schema?.tables ?? []).filter((t) => t.type === 'table')
  $: views = (schema?.tables ?? []).filter((t) => t.type === 'view')
  $: filteredTables = filter.trim()
    ? tables.filter((t) => t.name.toLowerCase().includes(filter.trim().toLowerCase()))
    : tables
  $: filteredViews = filter.trim()
    ? views.filter((t) => t.name.toLowerCase().includes(filter.trim().toLowerCase()))
    : views

  const nav = [
    { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { id: 'data', label: 'Data', icon: Table2 },
    { id: 'editor', label: 'SQL editor', icon: SquareTerminal },
    { id: 'history', label: 'History', icon: History },
    { id: 'settings', label: 'Settings', icon: Settings },
  ] as const
</script>

<aside class="sidebar">
  <nav>
    {#each nav as item (item.id)}
      <button
        class="nav-item"
        class:active={$view === item.id}
        on:click={() => view.set(item.id)}
        title={item.label}
      >
        <svelte:component this={item.icon} size={15} />
        <span>{item.label}</span>
      </button>
    {/each}
  </nav>

  <div class="tables">
    <div class="section-head">
      <span>Schema</span>
      <span class="count">{tables.length + views.length}</span>
    </div>
    {#if schema}
      {#if tables.length + views.length > 12}
        <div class="filter">
          <Search size={12} />
          <input type="search" placeholder="Filter tables" bind:value={filter} spellcheck="false" />
        </div>
      {/if}
      <div class="list">
        {#each filteredTables as t (t.name)}
          <button
            class="table-item"
            class:active={$view === 'data' && $activeTable === t.name}
            title={`${t.name} · ${t.columns.length} columns`}
            on:click={() => openTable(t.name)}
          >
            <Table2 size={13} />
            <span class="truncate">{t.name}</span>
          </button>
        {/each}
        {#if filteredViews.length}
          <div class="sub-head">Views</div>
          {#each filteredViews as t (t.name)}
            <button
              class="table-item"
              class:active={$view === 'data' && $activeTable === t.name}
              title={`${t.name} · ${t.columns.length} columns`}
              on:click={() => openTable(t.name)}
            >
              <Eye size={13} />
              <span class="truncate">{t.name}</span>
            </button>
          {/each}
        {/if}
        {#if filteredTables.length === 0 && filteredViews.length === 0}
          <p class="none">No matching objects</p>
        {/if}
      </div>
    {:else}
      <p class="none">Open a database to browse its tables.</p>
    {/if}
  </div>

  <footer>
    <span class="faint">DBDesk</span>
    <span class="faint">{version ? `v${version}` : ''}</span>
  </footer>
</aside>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    width: 216px;
    flex: none;
    min-height: 0;
    background: var(--bg-elevated);
    border-right: 1px solid var(--border);
  }
  nav {
    display: flex;
    flex-direction: column;
    gap: 1px;
    padding: 9px 8px;
  }
  .nav-item {
    display: flex;
    align-items: center;
    gap: 9px;
    height: 31px;
    padding: 0 9px;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--muted);
    font-size: 12.5px;
    text-align: left;
  }
  .nav-item:hover {
    background: var(--surface-2);
    color: var(--text);
  }
  .nav-item.active {
    background: var(--accent-soft);
    color: var(--accent);
    font-weight: 600;
  }
  .tables {
    display: flex;
    flex-direction: column;
    min-height: 0;
    flex: 1;
    border-top: 1px solid var(--border);
    padding-top: 8px;
  }
  .section-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 15px 7px;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--faint);
  }
  .count {
    font-weight: 500;
    letter-spacing: 0;
  }
  .filter {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 10px 7px;
    padding: 0 8px;
    height: 27px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface);
    color: var(--faint);
  }
  .filter input {
    flex: 1;
    height: 25px;
    padding: 0;
    border: 0;
    background: transparent;
    font-size: 12px;
    outline: none;
  }
  .list {
    flex: 1;
    overflow-y: auto;
    padding: 0 8px 8px;
  }
  .table-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    height: 27px;
    padding: 0 7px;
    border: 0;
    border-radius: 5px;
    background: transparent;
    color: var(--muted);
    font-size: 12px;
    text-align: left;
  }
  .table-item :global(svg) {
    flex: none;
    opacity: 0.7;
  }
  .table-item:hover {
    background: var(--surface-2);
    color: var(--text);
  }
  .table-item.active {
    background: var(--surface-3);
    color: var(--text);
  }
  .sub-head {
    padding: 10px 7px 4px;
    font-size: 9.5px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--faint);
  }
  .none {
    padding: 8px 8px;
    color: var(--faint);
    font-size: 11.5px;
    line-height: 1.5;
  }
  footer {
    display: flex;
    justify-content: space-between;
    padding: 8px 14px;
    border-top: 1px solid var(--border);
    font-size: 10.5px;
  }
</style>
