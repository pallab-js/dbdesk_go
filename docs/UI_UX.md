# UI and UX

## 1. Design direction

**Graphite + teal.** A dense, calm developer tool: dark graphite surfaces, a single
teal accent for action and focus, colour reserved for status (green success, amber
warning, red danger). Light mode is the same system on paper-white surfaces. The
interface stays out of the way: no hero illustrations inside the workspace, no
decorative gradients on content, no animation longer than 200 ms.

Principles:

1. **Data first.** Chrome is quiet; tables, SQL and numbers carry the page.
2. **One accent.** Teal means “interactive or selected”, nothing else.
3. **Dense but breathable.** 12–13 px UI text, 30 px controls, 10–14 px gaps.
4. **Obvious state.** Busy, empty, failed, read-only and partial are all visible.

## 2. Theme

Mode is `system | light | dark`, persisted in `localStorage` under `dbdesk.theme`.
`index.html` applies it before the bundle loads; the top bar cycles it and Settings
offers a segmented control.

### Core tokens (dark / light)

| Token | Dark | Light |
| --- | --- | --- |
| `--bg` | `#0b0f14` | `#f6f8fa` |
| `--bg-elevated` | `#0e141b` | `#eef2f5` |
| `--surface` | `#11171f` | `#ffffff` |
| `--surface-2` | `#161e28` | `#f0f3f6` |
| `--surface-3` | `#1c2632` | `#e7ecf1` |
| `--border` | `#243040` | `#d8dee6` |
| `--border-strong` | `#34465a` | `#bcc7d3` |
| `--text` | `#e6edf3` | `#0f172a` |
| `--text-soft` | `#c3ceda` | `#334155` |
| `--muted` | `#8b98a9` | `#5b6779` |
| `--faint` | `#5d6b7c` | `#7c8899` |
| `--accent` | `#2dd4bf` | `#0d9488` |
| `--ok` | `#34d399` | `#059669` |
| `--warn` | `#fbbf24` | `#b45309` |
| `--danger` | `#f87171` | `#dc2626` |

Radii: 10 px panels, 7 px controls, 20 px pills. Fonts: system UI for chrome,
platform monospace for SQL, identifiers and data.

## 3. Layout

```
┌──────────────────────────────────────────────────────────────┐
│ Top bar  brand · connection · theme/refresh/close · demo/open│ 46 px
├───────────┬──────────────────────────────────────────────────┤
│ Sidebar   │ Main view                                        │
│ 216 px    │                                                  │
│ nav       │  Dashboard / Data / Editor / History / Settings  │
│ schema    │                                                  │
│ footer    │                                                  │
└───────────┴──────────────────────────────────────────────────┘
```

- Window 1280×820, minimum 1024×680.
- Top bar: brand mark, active database name and path, read-only badge, theme cycle,
  refresh, close, Demo, Open.
- Sidebar: five nav items, then schema with a filter (when > 12 objects), tables,
  a Views subheading and object counts. Footer shows the version.
- Main views own their scrolling; only the content area scrolls, not the window.

## 4. Component language

### Controls

- `.btn` 30 px, `.sm` 26 px; primary teal fill, default surface, ghost for icon-only,
  danger keeps a neutral border until hover.
- Icon buttons always pair an icon with `title` and `aria-label`.
- Inputs match control height, focus to teal border, placeholders in faint text.
- Segmented control for mutually exclusive choices (theme).
- Badges: neutral, accent, ok, warn, danger; pill shaped, 19 px.
- Keyboard hints use bordered `.kbd` chips.

### Panels

- Header with a 12.5 px title and an action on the right.
- 1 px border, surface fill, 10 px radius; no drop shadows except dialogs and toasts.

### Tables

- Sticky header on the elevated background; sticky right action column.
- Monospace data, right-aligned numbers, `NULL` rendered as a faint italic chip.
- Row hover lifts the surface; action icons reveal on hover and stay visible on
  keyboard focus.
- Sortable headers show an arrow and `aria-sort`.

### Dialogs

- Backdrop `rgba(2,6,12,.62)` with 2 px blur.
- Dialog rises 6 px in 140 ms; Escape and backdrop click close.
- `RowEditor` is a two-column form with per-field type, NULL toggle and hint, and a
  scrollable field area capped at 58 vh.
- Delete and reset actions use a dedicated confirm dialog with the key values shown.

### Toasts

- Bottom-right stack, 420 px max, coloured left edge by kind, auto-dismiss 4.5 s,
  manual close, `aria-live="polite"`.

### Empty and error states

- Centered icon in a soft accent square, short title, one explanatory sentence, up to
  two actions.
- Table load and history errors use a bordered danger panel with retry.
- Startup failure gets a full-window panel with the raw message and retry.

## 5. View notes

### Welcome

- 58 px brand tile, 25 px headline, one-sentence promise.
- Primary “Choose database file”, secondary “Try the demo”.
- Read-only checkbox below the actions.
- Recent list with path, read-only badge, timestamp, delete.
- Footnote: local-first, SQLite only.

### Dashboard

- Six KPI cards: database size, tables, rows, 24 h queries, error rate, p95.
- Charts: activity bars (success/failed stacked), latency line (p50/p95), column-type
  doughnut, largest-tables horizontal bars.
- Lower row: recent queries (open in editor) and storage/health definition list.
- Table chips at the bottom open the data view.
- Header badges: journal mode, encoding, partial counts, updated time.

### Data

- Toolbar carries table selector, column/row counts, read-only badge, pager, page
  size, refresh, export and insert.
- Grid fills the remaining height; footer states total rows or count unavailability.
- Sort, edit and delete are the only row interactions; no inline cell editing in v1.

### Editor

- Editor panel takes 42% of height (min 150 px); result panel fills the rest.
- Run is the only primary button. Clear, Save and Export are secondary.
- Saved query chips sit between the panels.
- Result header shows kind, row count or affected rows, duration and truncation.
- Empty state offers “List schema objects” and “Preview <table>”.

### History

- Search field, reload, destructive clear.
- Rows show status dot, kind, truncated SQL, duration, row count, relative time.
- Expanded detail shows full SQL, error box, exact timestamp and open-in-editor.

### Settings

- Two columns: preferences on the left, recent databases and about on the right.
- Unsaved changes badge appears in the header; Save is disabled until dirty.
- Theme applies instantly; other preferences save on demand.
- About lists version, engine, shortcuts and the parameterisation guarantee.

## 6. Interaction and motion

- Hover and focus transitions: 100–150 ms ease.
- Charts animate for 180 ms, or not at all with reduced motion.
- Spinners only inside the control or panel that is waiting.
- No optimistic row edits: wait for the backend, then reload.
- Destructive actions always confirm; non-destructive ones act immediately.

## 7. Accessibility

- Visible `:focus-visible` outline in accent, 2 px offset 1 px.
- Dialogs labelled by title, modal semantics, Escape closes.
- Sortable table headers are focusable and respond to Enter.
- Status never relies on colour alone: labels, icons, and text accompany it.
- Contrast targets: body text ≥ 4.5:1 on its surface in both themes.
- `prefers-reduced-motion` disables animation and transition duration.
- Scrollbars are thin and theme-aware.

## 8. Copy

- Short, direct, sentence case. No exclamation marks.
- Errors state what failed and what to do: “Open a database first”, “run one
  statement at a time (found 3)”.
- Empty states explain the next action rather than apologising.
- Numbers are localised; durations switch from ms to seconds; bytes use binary units.
