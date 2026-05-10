# Memoire TUI — UX

Layout, key bindings, interaction patterns. Per-screen quick reference also
appears in the `?` overlay at runtime.

## Window layout

```
+------------------------------------------------------------+
| ◆ memoire │ <Section>                  auth PAT  api.host  ● online
| ────────────────────────────────────────────────────────────
| ╭────────╮  <screen content>
| │ Sidebar│
| │  list  │
| │        │
| ╰────────╯
| ────────────────────────────────────────────────────────────
| Section   ↑↓ screens   ↵ enter screen   :  command   ?  help
+------------------------------------------------------------+
```

- **Header** — app name, active section, host + auth chips, online dot.
- **Sidebar** — all screens with their icons. Currently focused screen is
  highlighted; preview-activates as you arrow up/down.
- **Status bar** — flash messages (4s TTL) on the left, screen + global
  key hints on the right. The hints adapt to whether sidebar or content
  is focused.
- **Overlays** — help (`?`), command palette (`:`), confirm (delete /
  quit), one-shot dialogs (token secret).

## Focus model

Implicit. No explicit toggle key.

```
sidebar focused (boot)
   │
   │  ↵ / →     drill into content
   ▼
content focused: list / table
   │
   │  ↵         drill into detail
   ▼
content focused: detail
   │
   │  e         drill into edit form (or n for new)
   ▼
content focused: form (text-input)
```

`esc` walks back up the same path. From the sidebar, `esc` opens the quit
confirm. While the sidebar is focused, all global keys (arrows, esc, `?`,
`:`) belong to the App regardless of which screen is shown — the textarea
in (e.g.) Assistant only gets keys after the user has drilled into it.

## Global keys

| Key | Action |
|-----|--------|
| `↑/↓` | move cursor (sidebar when focused, table rows when content is focused) |
| `↵` | drill in (sidebar → content → detail → form) |
| `esc` | back one level (drill up → sidebar → quit confirm) |
| `:` | open command palette |
| `?` | toggle help overlay |
| `ctrl+r` | refresh current screen |
| `ctrl+q` | force quit (no confirm) |

## Universal per-screen keys

Every list/table screen accepts:

| Key | Action |
|-----|--------|
| `n` | new entry |
| `e` | edit selected |
| `d` | delete selected (with `y/n` confirm) |
| `o` | open URL externally (where applicable) |
| `/` | filter list (built-in) |
| `f` | cycle filter pill |
| `s` | cycle sort |
| `r` | refresh |

## Form mode

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | next / previous field |
| `enter` | new line inside body textareas |
| `ctrl+e` | open body in `$EDITOR` (notes, journal) |
| `ctrl+s` | save form |
| `esc` | cancel |

## Command palette (`:`)

K9s-style overlay. Type to filter, `tab` to autocomplete, `↵` to run, `esc`
to cancel. Always includes:

- Every screen by name (`tasks`, `notes`, `journal`, …)
- `help`, `quit`, `refresh`
- Per-screen heavy actions exposed via `PaletteCommands()`:

| Command | Screen |
|---------|--------|
| `auto-schedule`, `agenda` | Tasks |
| `all-notes` | Notes |
| `trends` | Health |
| `force-refresh` | Feeds |
| `export`, `test-notify` | Settings |
| `new-conversation`, `clear-history`, `model-nova-lite`, `model-nova-pro` | Assistant |

## Per-screen interactions

### Notes
- Folder browser is the entry point. Each folder shows note count; "All
  notes" is always present at the top.
- `↵` on a folder drills into the filtered notes list.
- `↵` on a note opens the detail (markdown via `glamour`).
- `e` on detail opens the edit form (title, body, tags). Folder is fixed
  to the current filter — no folder picker in the form.

### Journal
- Calendar dots on dates with entries (loaded from `GET /journal`).
- `↑↓ ←→` step week/day; `t` jumps to today.
- Mood is a select (`great/good/okay/bad/terrible`); body via `enter` =
  newline + `ctrl+e` = `$EDITOR`.

### Tasks
- Filter pills cycle on `f` (all → todo → in-progress → done).
- Sort cycles on `s` (smart → due → priority → title).
- `:auto-schedule` reschedules unscheduled tasks via the API.
- `:agenda` shows the next 7 days from `/tasks/calendar`.

### Habits
- 30-day history rendered as `■` (done) / `·` (empty).
- `space` toggles today.

### Health
- Date picker top bar: `←/→` step day, `t` jumps to today.
- `:trends` opens the 7-day summary.
- Foods + exercises tracked together; nutrition has no separate screen.

### Finances
- `tab` cycles tabs (debts / income / expenses).
- Summary header pulls from `/finances/summary`.

### Feeds
- Left pane feeds, right pane articles. `tab` switches panes.
- `↵` opens article inline via `/feeds/article-text`.
- `o` opens externally, `h` favorites, `r` marks read.
- `:force-refresh` re-fetches articles bypassing the 30-min cache.

### Bookmarks / Favorites
- Standard list table. `o` opens the URL externally.

### Assistant
- Three panes: conversations (left), messages (center, viewport), input
  (bottom textarea). `tab` cycles panes.
- `ctrl+s` sends. `ctrl+l` clears the current conversation.
- Model + new-conversation switching via the palette (`:model-nova-lite`,
  `:new-conversation`).
- Streaming is not used; a spinner indicates "sending" and the reply is
  rendered all at once via `glamour`.

### Settings
- Sectioned read-only view; `e` opens an edit form.
- `:export` triggers `/export` and prints the presigned URL.
- `:test-notify` sends a test ntfy notification.

### Tokens
- When the API returns 403 to `GET /tokens` (PAT auth), the screen shows
  a banner and disables create/delete.
- After create, the plaintext token is shown once in a centred dialog;
  `↵` or `esc` dismisses.

### Admin
- Two read-only tables (Costs and Stats). Renders "(unavailable)" for
  non-admin users.

## Theming

- Adaptive colour palette (`internal/styles/styles.go`) auto-switches with
  terminal background.
- Primary cyan, accent amber. Borders rounded by default.
- Tables stripe odd rows with a subtle tinted background; even rows render
  at terminal default for clear alternation.
- `--no-color` (or `NO_COLOR=1`) disables colour entirely.
