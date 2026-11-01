import { derived, get, writable } from 'svelte/store'

export type ThemeMode = 'system' | 'light' | 'dark'

const KEY = 'dbdesk.theme'
const media = window.matchMedia('(prefers-color-scheme: dark)')

function storedMode(): ThemeMode {
  const v = localStorage.getItem(KEY)
  return v === 'light' || v === 'dark' || v === 'system' ? v : 'system'
}

const systemDark = writable(media.matches)
media.addEventListener('change', (e) => systemDark.set(e.matches))

export const themeMode = writable<ThemeMode>(storedMode())
export const isDark = derived(
  [themeMode, systemDark],
  ([mode, system]) => mode === 'dark' || (mode === 'system' && system),
)

// Keep the DOM attribute and localStorage in sync with the chosen mode.
themeMode.subscribe((mode) => {
  localStorage.setItem(KEY, mode)
  systemDark.set(media.matches)
})
isDark.subscribe((dark) => {
  document.documentElement.dataset.theme = dark ? 'dark' : 'light'
})

export function setTheme(mode: ThemeMode) {
  themeMode.set(mode)
}

export function cycleTheme() {
  const order: ThemeMode[] = ['system', 'light', 'dark']
  const next = order[(order.indexOf(get(themeMode)) + 1) % order.length]
  themeMode.set(next)
}
