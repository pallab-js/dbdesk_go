<script lang="ts">
  import { CircleAlert, CircleCheck, Info, X } from '@lucide/svelte'
  import { dismissToast, toasts } from '../lib/stores'
</script>

<div class="toasts" aria-live="polite" aria-atomic="false">
  {#each $toasts as toast (toast.id)}
    <div class="toast {toast.kind}" role="status">
      <span class="icon">
        {#if toast.kind === 'success'}
          <CircleCheck size={15} />
        {:else if toast.kind === 'error'}
          <CircleAlert size={15} />
        {:else}
          <Info size={15} />
        {/if}
      </span>
      <span class="msg">{toast.message}</span>
      <button class="btn ghost icon sm close" aria-label="Dismiss" on:click={() => dismissToast(toast.id)}>
        <X size={13} />
      </button>
    </div>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    right: 16px;
    bottom: 16px;
    z-index: 80;
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 420px;
  }
  .toast {
    display: flex;
    align-items: flex-start;
    gap: 9px;
    padding: 9px 9px 9px 12px;
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-left: 3px solid var(--muted);
    border-radius: var(--radius-sm);
    box-shadow: var(--shadow);
    animation: slide 0.15s ease;
  }
  .toast.success {
    border-left-color: var(--ok);
  }
  .toast.error {
    border-left-color: var(--danger);
  }
  .toast.info {
    border-left-color: var(--accent);
  }
  .icon {
    display: flex;
    margin-top: 2px;
    color: var(--muted);
  }
  .success .icon {
    color: var(--ok);
  }
  .error .icon {
    color: var(--danger);
  }
  .info .icon {
    color: var(--accent);
  }
  .msg {
    flex: 1;
    font-size: 12.5px;
    line-height: 1.45;
    padding-top: 1px;
    word-break: break-word;
  }
  .close {
    flex: none;
    color: var(--muted);
  }
  @keyframes slide {
    from {
      opacity: 0;
      transform: translateY(5px);
    }
  }
</style>
