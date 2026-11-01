<script lang="ts">
  import { InsertRow, UpdateRow } from '../../wailsjs/go/main/App.js'
  import { model } from '../../wailsjs/go/models'
  import { errorText } from '../lib/format'
  import { bumpData, notify } from '../lib/stores'
  import Modal from './Modal.svelte'

  export let table: model.Table
  export let columns: model.ColumnMeta[] = []
  export let row: any[] | null = null
  export let keyIndexes: number[] = []
  export let keyNames: string[] = []
  export let onClose: () => void = () => {}

  interface Field {
    name: string
    type: string
    notNull: boolean
    dflt: string
    pk: boolean
    generated: boolean
    blob: boolean
  }

  const isUpdate = row !== null
  const keys = new Set(keyNames)

  const fields: Field[] = table.columns
    .filter((c) => c.hidden === 0)
    .map((c) => ({
      name: c.name,
      type: c.declType,
      notNull: c.notNull,
      dflt: c.default,
      pk: c.pk > 0 || keys.has(c.name),
      generated: c.hidden !== 0,
      blob: /BLOB/i.test(c.declType),
    }))

  const values: Record<string, string> = {}
  const nulls: Record<string, boolean> = {}
  const original: Record<string, any> = {}

  for (const f of fields) {
    const idx = columns.findIndex((c) => c.name === f.name)
    const v = isUpdate && idx >= 0 ? row![idx] : null
    original[f.name] = v
    values[f.name] = v === null || v === undefined ? '' : String(v)
    nulls[f.name] = isUpdate && (v === null || v === undefined)
  }

  let busy = false
  let error = ''

  function label(f: Field) {
    return f.type ? `${f.name} · ${f.type.toLowerCase()}` : f.name
  }

  function hint(f: Field) {
    if (f.pk) return 'row key'
    if (f.generated) return 'generated'
    if (f.notNull) return f.dflt ? `required, default ${f.dflt}` : 'required'
    return f.dflt ? `default ${f.dflt}` : 'nullable'
  }

  function rowKey(): model.KeyValue[] {
    return keyNames.map((name, i) => new model.KeyValue({ name, value: row![keyIndexes[i]] }))
  }

  function changed(name: string): boolean {
    const wasNull = original[name] === null || original[name] === undefined
    if (nulls[name]) return !wasNull
    if (wasNull) return values[name] !== ''
    return values[name] !== String(original[name])
  }

  async function save() {
    error = ''
    const list: model.KeyValue[] = []
    for (const f of fields) {
      if (f.generated || f.blob) continue
      if (isUpdate) {
        if (f.pk) continue
        if (!changed(f.name)) continue
        list.push(new model.KeyValue({ name: f.name, value: nulls[f.name] ? null : values[f.name] }))
      } else {
        if (!nulls[f.name] && values[f.name] === '') continue
        list.push(new model.KeyValue({ name: f.name, value: nulls[f.name] ? null : values[f.name] }))
      }
    }
    if (isUpdate && list.length === 0) {
      error = 'Nothing changed yet.'
      return
    }
    busy = true
    try {
      if (isUpdate) {
        await UpdateRow(new model.RowUpdateRequest({ table: table.name, keys: rowKey(), values: list }))
      } else {
        await InsertRow(new model.RowInsertRequest({ table: table.name, values: list }))
      }
      bumpData()
      notify(isUpdate ? 'Row updated' : 'Row inserted', 'success')
      onClose()
    } catch (e) {
      error = errorText(e)
    } finally {
      busy = false
    }
  }
</script>

<Modal open={true} title={isUpdate ? `Edit row · ${table.name}` : `Insert row · ${table.name}`} width={720} {onClose}>
  <form id="row-form" on:submit|preventDefault={save}>
    {#if error}
      <div class="error">{error}</div>
    {/if}
    <div class="fields">
      {#each fields as f (f.name)}
        <label class="field" class:keyfield={f.pk}>
          <span class="label-row">
            <span class="name" title={label(f)}>{f.name}</span>
            {#if f.type}<em>{f.type.toLowerCase()}</em>{/if}
            <span class="spacer"></span>
            {#if !f.blob && !f.generated}
              <span class="null-toggle" title="Store NULL">
                <input type="checkbox" bind:checked={nulls[f.name]} />
                NULL
              </span>
            {/if}
          </span>
          <input
            type="text"
            bind:value={values[f.name]}
            disabled={busy || nulls[f.name] || f.pk || f.generated || f.blob}
            placeholder={f.blob ? 'binary data (read-only)' : f.generated ? 'generated column' : f.pk ? 'row key' : hint(f)}
            spellcheck="false"
            autocomplete="off"
          />
          <small>{hint(f)}</small>
        </label>
      {/each}
    </div>
  </form>
  <div slot="footer" class="footer">
    <span class="muted footer-hint">
      {#if isUpdate}Editing {keyNames.join(' + ')}{:else}Blank values are left to the column default.{/if}
    </span>
    <button type="button" class="btn" disabled={busy} on:click={onClose}>Cancel</button>
    <button type="submit" form="row-form" class="btn primary" disabled={busy}>
      {busy ? 'Saving…' : isUpdate ? 'Save changes' : 'Insert row'}
    </button>
  </div>
</Modal>

<style>
  .error {
    margin-bottom: 12px;
    padding: 8px 11px;
    border: 1px solid var(--danger);
    border-radius: var(--radius-sm);
    background: var(--danger-soft);
    color: var(--danger);
    font-size: 12.5px;
    white-space: pre-wrap;
  }
  .fields {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px 16px;
    max-height: 58vh;
    overflow-y: auto;
    padding: 2px;
  }
  .label-row {
    display: flex;
    align-items: baseline;
    gap: 5px;
    min-width: 0;
  }
  .label-row .name {
    font-family: var(--font-mono);
    font-size: 11.5px;
    font-weight: 600;
    text-transform: none;
    letter-spacing: 0;
    color: var(--text);
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .label-row em {
    font-style: normal;
    font-size: 9.5px;
    text-transform: uppercase;
    color: var(--faint);
  }
  .keyfield .name {
    color: var(--warn);
  }
  .null-toggle {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 9.5px;
    letter-spacing: 0.06em;
    color: var(--faint);
    text-transform: uppercase;
  }
  .footer {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
  }
  .footer-hint {
    flex: 1;
    font-size: 11.5px;
  }
</style>
