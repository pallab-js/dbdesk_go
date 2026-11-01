<script lang="ts">
  export let open = false
  export let title = ''
  export let width = 520
  export let onClose: () => void = () => {}

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape' && open) {
      e.stopPropagation()
      onClose()
    }
  }
</script>

<svelte:window on:keydown={onKey} />

{#if open}
  <div class="backdrop" role="presentation" on:mousedown|self={onClose}>
    <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="modal-title" style="max-width:{width}px">
      <header>
        <h2 id="modal-title">{title}</h2>
        <button class="btn ghost icon sm" aria-label="Close dialog" on:click={onClose}>✕</button>
      </header>
      <div class="content">
        <slot />
      </div>
      {#if $$slots.footer}
        <footer>
          <slot name="footer" />
        </footer>
      {/if}
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 8vh 24px 24px;
    background: rgba(2, 6, 12, 0.62);
    backdrop-filter: blur(2px);
    overflow-y: auto;
    animation: fade 0.12s ease;
  }
  .dialog {
    width: 100%;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    animation: rise 0.14s ease;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
  }
  h2 {
    font-size: 13.5px;
    font-weight: 600;
  }
  .content {
    padding: 16px;
  }
  footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid var(--border);
    background: var(--bg-elevated);
    border-radius: 0 0 var(--radius) var(--radius);
  }
  @keyframes fade {
    from {
      opacity: 0;
    }
  }
  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
  }
</style>
