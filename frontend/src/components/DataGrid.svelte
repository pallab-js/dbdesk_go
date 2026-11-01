<script lang="ts">
  import { ArrowDown, ArrowUp, KeyRound, Pencil, Trash2 } from '@lucide/svelte'
  import type { model } from '../../wailsjs/go/models'
  import { cellText } from '../lib/format'

  export let columns: model.ColumnMeta[] = []
  export let rows: any[][] = []
  export let keyIndexes: number[] = []
  export let editable = false
  export let sortable = false
  export let sortColumn = ''
  export let sortDir = ''
  export let onSort: (name: string) => void = () => {}
  export let onEdit: (row: any[]) => void = () => {}
  export let onDelete: (row: any[]) => void = () => {}

  // The backend appends a hidden rowid column; keep it out of the grid.
  $: visible = columns.map((c, i) => ({ c, i })).filter((x) => !x.c.hidden)
  $: keys = new Set(keyIndexes)

  function headerClick(name: string) {
    if (sortable) onSort(name)
  }
</script>

<div class="grid-wrap">
  <table>
    <thead>
      <tr>
        {#each visible as { c, i } (c.name + ':' + i)}
          <th
            class:sortable
            class:sorted={sortColumn === c.name}
            style="text-align:{(c.type || '').match(/INT|REAL|NUM|FLOA|DOUB/) ? 'right' : 'left'}"
            aria-sort={sortColumn === c.name ? (sortDir === 'desc' ? 'descending' : 'ascending') : 'none'}
            on:click={() => headerClick(c.name)}
            on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && (e.preventDefault(), headerClick(c.name))}
            tabindex={sortable ? 0 : -1}
          >
            <span class="th-inner" style="justify-content:{sortColumn === c.name && sortDir === 'desc' ? 'flex-end' : 'flex-start'}">
              {#if keys.has(i)}<KeyRound size={10} class="key" />{/if}
              <span class="name">{c.name}</span>
              {#if c.type}<span class="type">{c.type.toLowerCase()}</span>{/if}
              {#if sortColumn === c.name}
                {#if sortDir === 'desc'}<ArrowDown size={11} />{:else}<ArrowUp size={11} />{/if}
              {/if}
            </span>
          </th>
        {/each}
        {#if editable}
          <th class="actions-col"><span class="sr">Actions</span></th>
        {/if}
      </tr>
    </thead>
    <tbody>
      {#each rows as row, r (r)}
        <tr>
          {#each visible as { c, i } (c.name + ':' + i)}
            <td
              class:num={typeof row[i] === 'number'}
              class:null={row[i] === null || row[i] === undefined}
              title={cellText(row[i])}
            >
              {#if row[i] === null || row[i] === undefined}
                <span class="null-chip">NULL</span>
              {:else}
                {cellText(row[i])}
              {/if}
            </td>
          {/each}
          {#if editable}
            <td class="actions-col">
              <div class="row-actions">
                <button class="btn ghost icon sm" title="Edit row" aria-label="Edit row" on:click={() => onEdit(row)}>
                  <Pencil size={13} />
                </button>
                <button class="btn ghost icon sm danger-icon" title="Delete row" aria-label="Delete row" on:click={() => onDelete(row)}>
                  <Trash2 size={13} />
                </button>
              </div>
            </td>
          {/if}
        </tr>
      {:else}
        <tr>
          <td class="empty" colspan={visible.length + (editable ? 1 : 0)}>No rows to show.</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  .grid-wrap {
    flex: 1;
    min-height: 0;
    overflow: auto;
    background: var(--surface);
  }
  table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    font-family: var(--font-mono);
    font-size: 12px;
    table-layout: auto;
  }
  th {
    position: sticky;
    top: 0;
    z-index: 2;
    padding: 7px 11px;
    background: var(--bg-elevated);
    border-bottom: 1px solid var(--border-strong);
    border-right: 1px solid var(--border);
    color: var(--muted);
    font-family: var(--font-ui);
    font-size: 11px;
    font-weight: 600;
    white-space: nowrap;
    text-align: left;
    user-select: none;
  }
  th.sortable {
    cursor: pointer;
  }
  th.sortable:hover {
    color: var(--text);
  }
  th.sorted {
    color: var(--accent);
  }
  .th-inner {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .type {
    color: var(--faint);
    font-weight: 400;
    font-size: 9.5px;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }
  td {
    max-width: 360px;
    padding: 5px 11px;
    border-bottom: 1px solid var(--border);
    border-right: 1px solid var(--border);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-soft);
  }
  tbody tr:hover td {
    background: var(--surface-2);
    color: var(--text);
  }
  td.num {
    text-align: right;
  }
  td.null {
    color: var(--faint);
  }
  .null-chip {
    font-size: 10px;
    letter-spacing: 0.04em;
    color: var(--faint);
    font-style: italic;
  }
  .actions-col {
    position: sticky;
    right: 0;
    width: 64px;
    min-width: 64px;
    padding: 0 4px;
    border-right: 0;
    background: var(--surface);
  }
  thead .actions-col {
    background: var(--bg-elevated);
  }
  tbody tr:hover .actions-col {
    background: var(--surface-2);
  }
  .row-actions {
    display: flex;
    gap: 1px;
    justify-content: flex-end;
    opacity: 0;
    transition: opacity 0.1s ease;
  }
  tr:hover .row-actions,
  .row-actions:focus-within {
    opacity: 1;
  }
  .danger-icon:hover {
    color: var(--danger);
  }
  .empty {
    padding: 40px;
    text-align: center;
    color: var(--faint);
    font-family: var(--font-ui);
  }
  .sr {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
  }
</style>
