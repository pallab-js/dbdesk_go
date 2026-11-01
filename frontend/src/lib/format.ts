export function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let i = 0
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024
    i++
  }
  return `${value >= 100 || i === 0 ? Math.round(value) : value.toFixed(1)} ${units[i]}`
}

export function formatNumber(n: number | null | undefined): string {
  if (n === null || n === undefined) return '—'
  return new Intl.NumberFormat().format(n)
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${Math.round(ms)} ms`
  return `${(ms / 1000).toFixed(2)} s`
}

export function formatDateTime(ms: number): string {
  if (!ms) return '—'
  return new Date(ms).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

export function timeAgo(ms: number): string {
  const diff = Date.now() - ms
  if (diff < 45_000) return 'just now'
  if (diff < 3_600_000) return `${Math.round(diff / 60_000)}m ago`
  if (diff < 86_400_000) return `${Math.round(diff / 3_600_000)}h ago`
  if (diff < 7 * 86_400_000) return `${Math.round(diff / 86_400_000)}d ago`
  return new Date(ms).toLocaleDateString()
}

/** Text shown for one result cell. Blobs arrive already summarized by Go. */
export function cellText(v: unknown): string {
  if (v === null || v === undefined) return ''
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

export function errorText(e: unknown): string {
  if (!e) return ''
  if (typeof e === 'string') return e
  if (e instanceof Error) return e.message
  return String(e)
}

/** Width of a scrollbar-safe single-line SQL preview. */
export function sqlPreview(sql: string, max = 120): string {
  const line = sql.replace(/\s+/g, ' ').trim()
  return line.length > max ? line.slice(0, max) + '…' : line
}

export function formatSQL(sql: string): string {
  // Basic SQL formatter
  const keywords = ['SELECT', 'FROM', 'WHERE', 'JOIN', 'LEFT JOIN', 'RIGHT JOIN', 'INNER JOIN', 'OUTER JOIN', 'ON', 'GROUP BY', 'ORDER BY', 'HAVING', 'LIMIT', 'OFFSET', 'INSERT INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE FROM', 'CREATE TABLE', 'ALTER TABLE', 'DROP TABLE', 'UNION', 'UNION ALL', 'AND', 'OR'];
  let formatted = sql.replace(/\s+/g, ' ').trim();
  // Add newlines before major keywords
  formatted = formatted.replace(/\b(SELECT|FROM|WHERE|JOIN|LEFT JOIN|RIGHT JOIN|INNER JOIN|GROUP BY|ORDER BY|HAVING|LIMIT|INSERT INTO|UPDATE|DELETE FROM|UNION|UNION ALL)\b/gi, '\n$1');
  // Add newlines for AND/OR in WHERE
  formatted = formatted.replace(/\b(AND|OR)\b/gi, '\n  $1');
  // Indent after SELECT
  formatted = formatted.replace(/\nSELECT\s+/gi, '\nSELECT\n  ');
  formatted = formatted.replace(/,\s*/g, ',\n  ');
  return formatted.trim();
}
