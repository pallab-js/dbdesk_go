<script lang="ts">
  import { LoaderCircle, TriangleAlert } from '@lucide/svelte'
  import { onMount } from 'svelte'
  import {
    CloseDatabase,
    CreateDemoDatabase,
    DashboardStats,
    ForgetConnection,
    GetState,
    IsFullscreen,
    OpenDatabase,
    PickDatabaseFile,
    RefreshSchema,
    Schema,
    ToggleFullscreen,
  } from '../wailsjs/go/main/App.js'
  import { EventsOn } from '../wailsjs/runtime/runtime.js'
  import { model } from '../wailsjs/go/models'
  import Sidebar from './components/Sidebar.svelte'
  import Toasts from './components/Toasts.svelte'
  import TopBar from './components/TopBar.svelte'
  import { errorText } from './lib/format'
  import { activeTable, bumpData, dataVersion, notify, view } from './lib/stores'
  import Dashboard from './views/Dashboard.svelte'
  import Data from './views/Data.svelte'
  import Editor from './views/Editor.svelte'
  import History from './views/History.svelte'
  import Import from './views/Import.svelte'
  import Maintenance from './views/Maintenance.svelte'
  import Settings from './views/Settings.svelte'
  import Welcome from './views/Welcome.svelte'

  let state: model.AppState | null = null
  let schema: model.Schema | null = null
  let dashboard: model.DashboardData | null = null
  let booting = true
  let fatal = ''
  let dashLoading = false
  let refreshing = false
  let schemaToken = 0
  let dashToken = 0

  onMount(() => {
    void boot()
    const offDb = EventsOn('db:changed', () => void onDbChanged())
    const offSettings = EventsOn('settings:changed', () => void refreshState())
    
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'F11') {
        e.preventDefault()
        void ToggleFullscreen()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    
    return () => {
      offDb()
      offSettings()
      window.removeEventListener('keydown', handleKeyDown)
    }
  })

  async function boot() {
    booting = true
    fatal = ''
    try {
      state = await GetState()
      if (state.connection) await loadSchema()
    } catch (e) {
      fatal = errorText(e)
    } finally {
      booting = false
    }
  }

  async function refreshState() {
    try {
      const next = await GetState()
      state = next
      if (!next.connection) {
        schema = null
        dashboard = null
        activeTable.set('')
      }
    } catch (e) {
      notify(errorText(e), 'error')
    }
  }

  async function loadSchema(refresh = false) {
    const token = ++schemaToken
    try {
      const next = refresh ? await RefreshSchema() : await Schema()
      if (token !== schemaToken) return
      schema = next
      if ($activeTable && !next.tables.some((t) => t.name === $activeTable)) {
        activeTable.set('')
      }
    } catch (e) {
      if (token !== schemaToken) return
      notify(errorText(e), 'error')
    }
  }

  async function loadDashboard(force = false) {
    if (!state?.connection) {
      dashboard = null
      return
    }
    const token = ++dashToken
    dashLoading = true
    try {
      const next = await DashboardStats(force)
      if (token === dashToken) dashboard = next
    } catch (e) {
      if (token === dashToken) notify(errorText(e), 'error')
    } finally {
      if (token === dashToken) dashLoading = false
    }
  }

  async function onDbChanged() {
    await refreshState()
    bumpData()
    if (state?.connection) await loadSchema(true)
  }

  // Dashboard reload key: view + connection + content version (unless auto-refresh is off).
  $: autoRefresh = state?.settings?.autoRefresh ?? true
  $: dashKey = state?.connection
    ? [$view, state.connection.id, autoRefresh ? $dataVersion : 0].join('|')
    : ''
  $: if (dashKey && $view === 'dashboard') loadDashboard(false)

  async function openDialog(readOnly: boolean) {
    try {
      const path = await PickDatabaseFile()
      if (path) await openPath(path, readOnly)
    } catch (e) {
      notify(errorText(e), 'error')
    }
  }

  async function openPath(path: string, readOnly = false) {
    refreshing = true
    try {
      await OpenDatabase(path, readOnly)
      activeTable.set('')
      await loadSchema()
      await refreshState()
      view.set('dashboard')
      notify(`Opened ${path.split(/[\\/]/).pop()}`, 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      refreshing = false
    }
  }

  async function openDemo() {
    refreshing = true
    try {
      const conn = await CreateDemoDatabase()
      activeTable.set('')
      await loadSchema()
      await refreshState()
      view.set('dashboard')
      notify(`Demo database ready · ${conn.name}`, 'success')
    } catch (e) {
      notify(errorText(e), 'error')
    } finally {
      refreshing = false
    }
  }

  async function closeDb() {
    try {
      await CloseDatabase()
      schema = null
      dashboard = null
      activeTable.set('')
      view.set('dashboard')
      notify('Database closed', 'info')
    } catch (e) {
      notify(errorText(e), 'error')
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

  async function refreshAll() {
    if (!state?.connection) return
    refreshing = true
    try {
      await Promise.allSettled([loadSchema(true), loadDashboard(true)])
      bumpData()
    } finally {
      refreshing = false
    }
  }
</script>

{#if booting}
  <div class="splash">
    <div class="splash-mark">D</div>
    <LoaderCircle size={18} class="spin" />
    <p class="muted">Starting DBDesk…</p>
  </div>
{:else if fatal}
  <div class="fatal">
    <div class="panel fatal-panel">
      <TriangleAlert size={22} />
      <h1>DBDesk could not start</h1>
      <pre>{fatal}</pre>
      <p class="muted">App data lives in your local data folder. Close any other copy of DBDesk and try again.</p>
      <button class="btn primary" on:click={boot}>Try again</button>
    </div>
  </div>
{:else if state}
  <div class="app">
    <TopBar
      connection={state.connection ?? null}
      {refreshing}
      canRefresh={!!state.connection}
      onOpen={() => openDialog(false)}
      onDemo={openDemo}
      onRefresh={refreshAll}
      onClose={closeDb}
    />
    <div class="workspace">
      <Sidebar {schema} version={state.version} />
      <main>
        {#if !state.connection}
          <Welcome
            recent={state.recent}
            onOpen={openDialog}
            onDemo={openDemo}
            onOpenPath={(p, ro) => openPath(p, ro)}
            onForget={forget}
          />
        {:else if $view === 'dashboard'}
          <Dashboard data={dashboard} loading={dashLoading} />
        {:else if $view === 'data'}
          <Data {schema} />
        {:else if $view === 'editor'}
          <Editor {schema} rowLimit={state.settings.rowLimit} />
        {:else if $view === 'history'}
          <History />
        {:else if $view === 'import'}
          <Import {schema} />
        {:else if $view === 'maintenance'}
          <Maintenance connection={state.connection} />
        {:else}
          <Settings
            settings={state.settings}
            recent={state.recent}
            appDataDir={state.appDataDir}
            version={state.version}
            onOpenPath={(p, ro) => openPath(p, ro)}
            onDemo={openDemo}
          />
        {/if}
      </main>
    </div>
    <Toasts />
  </div>
{/if}

<style>
  :global(html),
  :global(body) {
    height: 100%;
  }
  .app {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }
  .workspace {
    display: flex;
    flex: 1;
    min-height: 0;
  }
  main {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    min-height: 0;
    background: var(--bg);
  }
  .splash,
  .fatal {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 14px;
    height: 100%;
    padding: 24px;
  }
  .splash-mark {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 48px;
    border-radius: 13px;
    background: var(--accent);
    color: var(--accent-contrast);
    font-size: 23px;
    font-weight: 700;
  }
  .splash p {
    font-size: 12.5px;
  }
  .fatal-panel {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    max-width: 520px;
    padding: 28px;
    text-align: center;
    color: var(--danger);
  }
  .fatal-panel h1 {
    font-size: 17px;
    color: var(--text);
  }
  .fatal-panel pre {
    width: 100%;
    margin: 0;
    padding: 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--text-soft);
    font-family: var(--font-mono);
    font-size: 11.5px;
    white-space: pre-wrap;
    text-align: left;
  }
  .fatal-panel p {
    font-size: 12px;
    line-height: 1.5;
  }
</style>
