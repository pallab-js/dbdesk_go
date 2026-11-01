<script lang="ts">
  import Modal from './Modal.svelte'

  export let open = false
  export let title = 'Confirm'
  export let body = ''
  export let confirmLabel = 'Confirm'
  export let danger = false
  export let busy = false
  export let onConfirm: () => void = () => {}
  export let onCancel: () => void = () => {}
</script>

<Modal {open} {title} width={440} onClose={() => !busy && onCancel()}>
  <p class="body">{body}</p>
  <div slot="footer">
    <button class="btn" disabled={busy} on:click={onCancel}>Cancel</button>
    <button class="btn" class:danger class:primary={!danger} disabled={busy} on:click={onConfirm}>
      {busy ? 'Working…' : confirmLabel}
    </button>
  </div>
</Modal>

<style>
  .body {
    color: var(--text-soft);
    font-size: 13px;
    line-height: 1.55;
  }
</style>
