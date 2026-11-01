import { writable } from 'svelte/store'

export type View = 'dashboard' | 'data' | 'editor' | 'history' | 'import' | 'maintenance' | 'settings'

export type ToastKind = 'info' | 'success' | 'error'

export interface Toast {
  id: number
  kind: ToastKind
  message: string
}

export const view = writable<View>('dashboard')
export const activeTable = writable<string>('')
export const editorDraft = writable<string>('')
// Bumped whenever rows or schema change so views can reload cheaply.
export const dataVersion = writable(0)
export const toasts = writable<Toast[]>([])

let nextToast = 1

export function notify(message: string, kind: ToastKind = 'info') {
  const id = nextToast++
  toasts.update((list) => [...list, { id, kind, message }])
  setTimeout(() => dismissToast(id), 4500)
}

export function dismissToast(id: number) {
  toasts.update((list) => list.filter((t) => t.id !== id))
}

export function bumpData() {
  dataVersion.update((n) => n + 1)
}

export function openTable(name: string) {
  activeTable.set(name)
  view.set('data')
}

export function openInEditor(sql: string) {
  editorDraft.set(sql)
  view.set('editor')
}
