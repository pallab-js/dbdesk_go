<script lang="ts">
  import {
    Activity,
    CircleAlert,
    HardDrive,
    Rows3,
    TableProperties,
    Timer,
  } from '@lucide/svelte'
  import type { model } from '../../wailsjs/go/models'
  import Chart from '../components/Chart.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import { baseOptions, palette } from '../lib/charts'
  import {
    formatBytes,
    formatDuration,
    formatNumber,
    sqlPreview,
    timeAgo,
  } from '../lib/format'
  import { openInEditor, openTable, view } from '../lib/stores'
  import { isDark } from '../lib/theme'

  export let data: model.DashboardData | null = null
  export let loading = false

  $: p = palette($isDark)
  $: totalRows = (data?.tables ?? []).reduce((n, t) => n + (t.rows || 0), 0)
  $: counted = (data?.tables ?? []).filter((t) => !t.error).length

  function axis(pal: ReturnType<typeof palette>) {
    return {
      x: {
        grid: { display: false },
        ticks: { color: pal.muted, font: { size: 10 }, maxRotation: 0 },
      },
      y: {
        beginAtZero: true,
        grid: { color: pal.grid, drawTicks: false },
        border: { display: false },
        ticks: { color: pal.muted, font: { size: 10 }, precision: 0 },
      },
    }
  }

  function options(pal: ReturnType<typeof palette>, horizontal = false) {
    return {
      ...baseOptions,
      plugins: {
        ...baseOptions.plugins,
        legend: {
          ...baseOptions.plugins.legend,
          labels: {
            ...baseOptions.plugins.legend.labels,
            color: pal.text,
          },
        },
        tooltip: {
          ...baseOptions.plugins.tooltip,
          backgroundColor: $isDark ? '#0b0f14' : '#ffffff',
          titleColor: pal.text,
          bodyColor: pal.text,
          borderColor: pal.grid,
          borderWidth: 1,
        },
      },
      scales: horizontal
        ? {
            x: {
              beginAtZero: true,
              grid: { color: pal.grid, drawTicks: false },
              border: { display: false },
              ticks: { color: pal.muted, font: { size: 10 }, precision: 0 },
            },
            y: {
              grid: { display: false },
              border: { display: false },
              ticks: { color: pal.text, font: { size: 11 } },
            },
          }
        : axis(pal),
    }
  }

  $: activityChart = data
    ? {
        type: 'bar' as const,
        data: {
          labels: data.activity.map((b) => b.label),
          datasets: [
            {
              label: 'Succeeded',
              data: data.activity.map((b) => b.success),
              backgroundColor: p.accentFill,
              hoverBackgroundColor: p.accent,
              borderRadius: 3,
              stack: 'queries',
            },
            {
              label: 'Failed',
              data: data.activity.map((b) => b.error),
              backgroundColor: p.danger,
              borderRadius: 3,
              stack: 'queries',
            },
          ],
        },
        options: {
          ...options(p),
          scales: {
            ...axis(p),
            x: { ...axis(p).x, ticks: { ...axis(p).x.ticks, maxTicksLimit: 8 } },
          },
        },
      }
    : null

  $: latencyChart = data
    ? {
        type: 'line' as const,
        data: {
          labels: data.latency.map((b) => b.label),
          datasets: [
            {
              label: 'p50',
              data: data.latency.map((b) => b.p50),
              borderColor: p.accent,
              backgroundColor: p.accentFill,
              pointRadius: 0,
              borderWidth: 2,
              tension: 0.32,
              fill: true,
            },
            {
              label: 'p95',
              data: data.latency.map((b) => b.p95),
              borderColor: p.warn,
              pointRadius: 0,
              borderWidth: 2,
              borderDash: [4, 3],
              tension: 0.32,
            },
          ],
        },
        options: {
          ...options(p),
          scales: {
            ...axis(p),
            x: { ...axis(p).x, ticks: { ...axis(p).x.ticks, maxTicksLimit: 8 } },
            y: {
              ...axis(p).y,
              ticks: {
                ...axis(p).y.ticks,
                precision: undefined,
                callback: (v: number) => `${v} ms`,
              },
            },
          },
        },
      }
    : null

  $: sizeChart = data
    ? {
        type: 'bar' as const,
        data: {
          labels: data.tables.slice(0, 8).map((t) => t.name),
          datasets: [
            {
              label: 'Rows',
              data: data.tables.slice(0, 8).map((t) => t.rows),
              backgroundColor: p.accentFill,
              hoverBackgroundColor: p.accent,
              borderRadius: 4,
              barThickness: 16,
            },
          ],
        },
        options: {
          ...options(p, true),
          plugins: { ...options(p, true).plugins, legend: { display: false } },
        },
      }
    : null

  $: typeChart = data
    ? {
        type: 'doughnut' as const,
        data: {
          labels: data.columnTypes.map((t) => t.type),
          datasets: [
            {
              data: data.columnTypes.map((t) => t.count),
              backgroundColor: data.columnTypes.map((_, i) => p.series[i % p.series.length]),
              borderWidth: 0,
              hoverOffset: 4,
            },
          ],
        },
        options: {
          ...options(p),
          cutout: '62%',
          scales: {},
          plugins: {
            ...options(p).plugins,
            legend: {
              position: 'right' as const,
              labels: { color: p.text, boxWidth: 8, boxHeight: 8, usePointStyle: true, pointStyle: 'circle' as const, font: { size: 11 } },
            },
          },
        },
      }
    : null

  const kpis = [
    { key: 'size', label: 'Database size', icon: HardDrive },
    { key: 'tables', label: 'Tables', icon: TableProperties },
    { key: 'rows', label: 'Rows', icon: Rows3 },
    { key: 'queries', label: 'Queries (24h)', icon: Activity },
    { key: 'errors', label: 'Error rate', icon: CircleAlert },
    { key: 'p95', label: 'p95 latency', icon: Timer },
  ] as const

  function kpiValue(key: string): string {
    if (!data) return '—'
    switch (key) {
      case 'size':
        return formatBytes(data.database.sizeBytes)
      case 'tables':
        return formatNumber(data.objects.tables)
      case 'rows':
        return formatNumber(totalRows)
      case 'queries':
        return formatNumber(data.totals.queries)
      case 'errors':
        return `${data.totals.errorRate.toFixed(1)}%`
      case 'p95':
        return formatDuration(data.totals.p95)
      default:
        return '—'
    }
  }

  function kpiSub(key: string): string {
    if (!data) return ''
    switch (key) {
      case 'size':
        return `${formatNumber(data.database.pageCount)} pages · ${formatBytes(data.database.pageSize)} each`
      case 'tables':
        return `${data.objects.views} views · ${data.objects.indexes} indexes`
      case 'rows':
        return `${counted} of ${data.tables.length} tables counted`
      case 'queries':
        return `${formatNumber(data.totals.today)} today`
      case 'errors':
        return `${formatNumber(data.totals.errors)} failed queries`
      case 'p95':
        return `p50 ${formatDuration(data.totals.p50)}`
      default:
        return ''
    }
  }
</script>

<div class="page">
  <header class="page-head">
    <div>
      <h1>Dashboard</h1>
      <p class="muted">Storage, schema shape and your own query telemetry for the last 24 hours.</p>
    </div>
    <div class="head-meta">
      {#if data}
        {#if data.partial}<span class="badge warn">partial counts</span>{/if}
        <span class="badge">journal {data.database.journalMode}</span>
        <span class="badge">{data.database.encoding}</span>
        <span class="faint">updated {timeAgo(data.generatedAt)}</span>
      {/if}
    </div>
  </header>

  {#if !data}
    <div class="panel">
      <EmptyState title={loading ? 'Reading database…' : 'No statistics yet'} message="Open a database to see its size, tables and query activity." />
    </div>
  {:else}
    <section class="kpis">
      {#each kpis as k (k.key)}
        <div class="panel kpi">
          <div class="kpi-top">
            <span class="kpi-label">{k.label}</span>
            <svelte:component this={k.icon} size={14} />
          </div>
          <div class="kpi-value">{kpiValue(k.key)}</div>
          <div class="kpi-sub">{kpiSub(k.key)}</div>
        </div>
      {/each}
    </section>

    {#if data.tables.length === 0}
      <div class="panel">
        <EmptyState title="This database has no tables yet" message="Use the SQL editor to create a table, then come back." >
          <button class="btn primary" on:click={() => view.set('editor')}>Open SQL editor</button>
        </EmptyState>
      </div>
    {:else}
      <section class="charts">
        <div class="panel chart-card wide">
          <div class="panel-head">
            <h2>Query activity</h2>
            <span class="muted">{formatNumber(data.totals.queries)} queries · {data.totals.errors} failed</span>
          </div>
          <div class="panel-body"><Chart config={activityChart} height={190} /></div>
        </div>

        <div class="panel chart-card">
          <div class="panel-head">
            <h2>Latency by hour</h2>
            <span class="muted">p50 / p95</span>
          </div>
          <div class="panel-body"><Chart config={latencyChart} height={190} /></div>
        </div>

        <div class="panel chart-card">
          <div class="panel-head">
            <h2>Column types</h2>
            <span class="muted">{data.columnTypes.reduce((n, t) => n + t.count, 0)} columns</span>
          </div>
          <div class="panel-body"><Chart config={typeChart} height={190} /></div>
        </div>

        <div class="panel chart-card wide">
          <div class="panel-head">
            <h2>Largest tables</h2>
            <span class="muted">exact row counts</span>
          </div>
          <div class="panel-body"><Chart config={sizeChart} height={Math.max(100, Math.min(8, data.tables.length) * 26)} /></div>
        </div>
      </section>

      <section class="lower">
        <div class="panel">
          <div class="panel-head">
            <h2>Recent queries</h2>
            <button class="btn ghost sm" on:click={() => view.set('history')}>All history</button>
          </div>
          {#if data.recent.length === 0}
            <p class="empty-hint">Queries you run will appear here.</p>
          {:else}
            <ul class="recent">
              {#each data.recent as h (h.id)}
                <li>
                  <button class="recent-row" on:click={() => openInEditor(h.sql)} title="Open in SQL editor">
                    <span class="status" class:bad={!h.success}></span>
                    <span class="sql mono truncate">{h.sql}</span>
                    <span class="muted dur">{formatDuration(h.durationMs)}</span>
                    <span class="faint when">{timeAgo(h.createdAt)}</span>
                  </button>
                </li>
              {/each}
            </ul>
          {/if}
        </div>

        <div class="panel">
          <div class="panel-head"><h2>Storage &amp; health</h2></div>
          <dl class="meta">
            <div><dt>File</dt><dd class="mono" title={data.database.path}>{data.database.path}</dd></div>
            <div><dt>Size</dt><dd>{formatBytes(data.database.sizeBytes)}</dd></div>
            <div><dt>Page size</dt><dd>{formatBytes(data.database.pageSize)}</dd></div>
            <div><dt>Pages</dt><dd>{formatNumber(data.database.pageCount)}</dd></div>
            <div><dt>Free pages</dt><dd>{formatNumber(data.database.freePages)}</dd></div>
            <div><dt>Journal</dt><dd>{data.database.journalMode}</dd></div>
            <div><dt>Encoding</dt><dd>{data.database.encoding}</dd></div>
            <div><dt>Mode</dt><dd>{data.database.readOnly ? 'read-only' : 'read/write'}</dd></div>
            <div><dt>Triggers</dt><dd>{formatNumber(data.objects.triggers)}</dd></div>
            <div><dt>Views</dt><dd>{formatNumber(data.objects.views)}</dd></div>
          </dl>
          {#if data.tables.some((t) => t.error)}
            <p class="warn-note">
              Some tables were too slow to count: {data.tables.filter((t) => t.error).map((t) => t.name).join(', ')}
            </p>
          {/if}
        </div>
      </section>

      <section class="panel table-stats">
        <div class="panel-head">
          <h2>Tables</h2>
          <span class="muted">Click a table to browse its rows</span>
        </div>
        <div class="table-list">
          {#each data.tables as t (t.name)}
            <button class="table-chip" on:click={() => openTable(t.name)} title={`Browse ${t.name}`}>
              <span class="mono">{t.name}</span>
              <span class="badge" class:warn={t.error}>{t.error ? 'timed out' : formatNumber(t.rows)}</span>
            </button>
          {/each}
        </div>
      </section>
    {/if}
  {/if}
</div>

<style>
  .page {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 18px 20px 32px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .page-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
  }
  .page-head h1 {
    font-size: 19px;
    font-weight: 650;
    letter-spacing: -0.02em;
  }
  .page-head p {
    margin-top: 2px;
    font-size: 12.5px;
  }
  .head-meta {
    display: flex;
    align-items: center;
    gap: 7px;
    font-size: 11.5px;
  }
  .kpis {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 10px;
  }
  .kpi {
    padding: 11px 13px;
  }
  .kpi-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    color: var(--faint);
  }
  .kpi-label {
    font-size: 10.5px;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }
  .kpi-value {
    margin-top: 5px;
    font-size: 21px;
    font-weight: 650;
    letter-spacing: -0.02em;
    font-variant-numeric: tabular-nums;
  }
  .kpi-sub {
    margin-top: 1px;
    color: var(--faint);
    font-size: 11px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .charts {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }
  .chart-card.wide {
    grid-column: span 2;
  }
  .lower {
    display: grid;
    grid-template-columns: minmax(0, 1.35fr) minmax(280px, 1fr);
    gap: 12px;
  }
  .recent {
    list-style: none;
    margin: 0;
    padding: 5px;
  }
  .recent-row {
    display: flex;
    align-items: center;
    gap: 9px;
    width: 100%;
    padding: 5px 8px;
    border: 0;
    border-radius: 5px;
    background: transparent;
    color: var(--text-soft);
    text-align: left;
  }
  .recent-row:hover {
    background: var(--surface-2);
  }
  .status {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--ok);
    flex: none;
  }
  .status.bad {
    background: var(--danger);
  }
  .sql {
    flex: 1;
    font-size: 11.5px;
    color: var(--text-soft);
  }
  .dur {
    font-size: 11px;
    font-variant-numeric: tabular-nums;
  }
  .when {
    width: 58px;
    text-align: right;
    font-size: 10.5px;
  }
  .meta {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0;
    margin: 0;
    padding: 6px 0;
  }
  .meta > div {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    padding: 4px 14px;
    border-bottom: 1px solid var(--border);
    min-width: 0;
  }
  .meta > div:nth-last-child(-n + 2) {
    border-bottom: 0;
  }
  dt {
    color: var(--faint);
    font-size: 11.5px;
    flex: none;
  }
  dd {
    margin: 0;
    font-size: 11.5px;
    text-align: right;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 62%;
  }
  .meta > div:first-child {
    grid-column: span 2;
  }
  .meta > div:first-child dd {
    max-width: 78%;
    direction: rtl;
  }
  .warn-note {
    margin: 0 14px 12px;
    padding: 7px 9px;
    border-radius: var(--radius-sm);
    background: var(--warn-soft);
    color: var(--warn);
    font-size: 11.5px;
    line-height: 1.45;
  }
  .table-stats {
    padding-bottom: 6px;
  }
  .table-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    padding: 12px 14px;
  }
  .table-chip {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    height: 27px;
    padding: 0 5px 0 10px;
    border: 1px solid var(--border);
    border-radius: 20px;
    background: var(--surface-2);
    color: var(--text-soft);
    font-size: 11.5px;
  }
  .table-chip:hover {
    border-color: var(--accent);
    color: var(--text);
  }
  @media (max-width: 1080px) {
    .charts {
      grid-template-columns: 1fr;
    }
    .chart-card.wide {
      grid-column: span 1;
    }
    .lower {
      grid-template-columns: 1fr;
    }
  }
</style>
