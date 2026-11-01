import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

export interface Palette {
  text: string
  muted: string
  grid: string
  accent: string
  accentFill: string
  ok: string
  danger: string
  warn: string
  series: string[]
}

export function palette(dark: boolean): Palette {
  return dark
    ? {
        text: '#c3ceda',
        muted: '#8b98a9',
        grid: 'rgba(139, 152, 169, 0.14)',
        accent: '#2dd4bf',
        accentFill: 'rgba(45, 212, 191, 0.55)',
        ok: '#34d399',
        danger: '#f87171',
        warn: '#fbbf24',
        series: ['#2dd4bf', '#60a5fa', '#c084fc', '#fbbf24', '#f87171', '#34d399', '#f472b6', '#94a3b8'],
      }
    : {
        text: '#334155',
        muted: '#5b6779',
        grid: 'rgba(15, 23, 42, 0.09)',
        accent: '#0d9488',
        accentFill: 'rgba(13, 148, 136, 0.55)',
        ok: '#059669',
        danger: '#dc2626',
        warn: '#b45309',
        series: ['#0d9488', '#2563eb', '#7c3aed', '#b45309', '#dc2626', '#059669', '#db2777', '#64748b'],
      }
}

export const baseOptions = {
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 180 },
  plugins: {
    legend: {
      display: true,
      labels: { boxWidth: 8, boxHeight: 8, usePointStyle: true, pointStyle: 'circle' as const, font: { size: 11 } },
    },
    tooltip: {
      padding: 9,
      cornerRadius: 6,
      titleFont: { size: 12 },
      bodyFont: { size: 12 },
    },
  },
  scales: {
    x: { grid: { display: false }, ticks: { font: { size: 10 }, maxRotation: 0 } },
    y: { beginAtZero: true, grid: { drawTicks: false }, ticks: { font: { size: 10 }, precision: 0 } },
  },
}

/** Svelte action: rebuilds the chart whenever the config object changes. */
export function chart(node: HTMLCanvasElement, config: any) {
  let instance = new Chart(node, config)
  return {
    update(next: any) {
      instance.destroy()
      instance = new Chart(node, next)
    },
    destroy() {
      instance.destroy()
    },
  }
}
