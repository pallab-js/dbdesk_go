<script lang="ts">
  export let title = ''
  export let message = ''
</script>

<div class="empty">
  <div class="mark"><slot name="icon" /></div>
  <h3>{title}</h3>
  {#if message}<p>{message}</p>{/if}
  <div class="actions"><slot /></div>
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 40px 24px;
    text-align: center;
  }
  .mark {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    margin-bottom: 2px;
    border-radius: 10px;
    background: var(--accent-soft);
    color: var(--accent);
  }
  .mark:empty {
    display: none;
  }
  h3 {
    font-size: 14px;
    font-weight: 600;
  }
  p {
    max-width: 420px;
    color: var(--muted);
    font-size: 12.5px;
    line-height: 1.55;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 8px;
    margin-top: 6px;
  }
  .actions:empty {
    display: none;
  }
</style>
