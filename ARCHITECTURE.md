# Memoire TUI — Architecture

## Status

Shipped as a separate Go module + binary at `github.com/neilfarmer/memoire-tui`.
This document captures the still-current design decisions; concrete file layout
is in [`README.md`](README.md), per-feature endpoint mapping in
[`FEATURES.md`](FEATURES.md), interaction details in [`UX.md`](UX.md).

---

## Framework Decision: Bubble Tea + Lipgloss + Bubbles + Huh

**Chosen stack:**
- `github.com/charmbracelet/bubbletea` — Elm-architecture TUI framework
- `github.com/charmbracelet/lipgloss` — Style/layout primitives (borders, padding, color, flex-like arrangement)
- `github.com/charmbracelet/bubbles` — Pre-built components: `list.Model`, `viewport.Model`, `textinput.Model`, `textarea.Model`, `spinner.Model`, `table.Model`, `key.Binding`
- `github.com/charmbracelet/huh` — Form library for structured create/edit flows with validation

**Rationale:**

1. **Bubble Tea** is the dominant Go TUI framework. Its Elm `Model / Update(Msg) / View()` pattern is clean, testable, and handles async API calls via `tea.Cmd` without blocking the render loop. It has first-class multi-screen support via nested models.
2. **Lipgloss** mirrors the CSS variable system in `frontend/index.html` well — surfaces, borders, text colors, and adaptive light/dark themes map cleanly to `lipgloss.AdaptiveColor`.
3. **Bubbles** ships `list.Model` (exactly what task/note/habit lists need), `viewport.Model` (for long content like journal entries and note bodies), and `textinput` / `textarea` for in-TUI editing — these alone cover 80% of the UI primitives needed.
4. **Huh** handles multi-field create/edit forms (task due date + priority + folder, goal description + target date, etc.) in a terminal-native way with validation and keyboard navigation.
5. The combination is proven at production scale (lazygit, soft-serve, mods, pop). No other Go TUI stack (tview, tcell raw, termui) offers the same ergonomics for a multi-screen CRUD app.

---

## Project Layout

The full as-shipped layout lives in [`README.md`](README.md). The original
locked design predicted:

- one `internal/api/<feature>.go` per memoire feature (kept)
- one `internal/ui/screens/<feature>.go` per screen (kept; nutrition + diagrams
  removed — see "Feature Adaptation Notes" below)
- shared `components/` for chrome (kept; gained `palette.go`,
  `formkeys.go`, `striped/`)

Two changes from the original spec worth highlighting:

- **No Nutrition screen.** The memoire backend has no `/nutrition` endpoint;
  nutrition data lives under `/health/{date}` foods. Use Health.
- **`internal/ui/components/striped/`** is a vendored fork of
  `charmbracelet/bubbles@v1.0.0/table` that adds a per-row
  `Styles.RowStyler` hook so we can paint alternating row backgrounds
  without smuggling ANSI through `runewidth.Truncate`.

---

## Auth Model

The TUI authenticates exclusively with a Personal Access Token (PAT).

**Resolution order (highest to lowest priority):**
1. `MEMOIRE_PAT` environment variable
2. `~/.config/memoire-tui/config.toml` — `[auth]` section, `pat` key
3. Interactive prompt on first run (stores to config file)

**API URL resolution:**
1. `MEMOIRE_API_URL` environment variable
2. `~/.config/memoire-tui/config.toml` — `[api]` section, `url` key

**Wire format:** Every request sets `Authorization: Bearer <pat>`.

**Constraint:** The `/tokens` Lambda rejects PAT-authenticated requests (JWT only). The tokens screen must display a clear notice and disable create/delete when the TUI session is PAT-authenticated. This is the only endpoint with this restriction.

**Config file example:**
```toml
[api]
url = "https://api.memoire.example.com"

[auth]
pat = "pat_abc123"
```

---

## Screen Routing & Focus Model

The root `App` (`internal/ui/app.go`) holds:
- `current Screen` — which feature is active
- `sideCursor int` — index into `SidebarOrder`; tracks the sidebar selection
- `sideFocus bool` — true when the sidebar owns arrow keys; false when content does
- `registry map[Screen]screens.Screen` — lazy-initialized screen models
- `factories map[Screen]ScreenFactory` — registered screen constructors
- `helpOpen`, `paletteOpen`, `quitConfirm` — overlay states
- `flash`, `flashLevel`, `flashID` — transient status with auto-dismiss

**Implicit focus model.** No explicit focus toggle key. The user moves
through depth via `enter` (drill in) and `esc` (drill up):

```
sidebar → screen list → detail / form → form-field text-input
   ↑       ↑              ↑                  ↑
   └── esc — esc ────────┘                  esc cancels form
                                            (or ctrl+s saves)
```

`OnEscape() bool` is implemented by every screen with sub-modes; App calls
it when content is focused and falls back to `sideFocus = true` if the
screen reports it's already at the top. Pressing esc while the sidebar is
focused opens the quit confirm.

A k9s-style command palette (`:`) overlays a filterable command list.
Screens contribute heavy actions via `PaletteCommands() []components.Command`;
the App always adds screen-jump and global commands.

---

## Async API Calls

All HTTP calls use `tea.Cmd` (non-blocking). Pattern:

```go
func fetchTasks(client *api.Client, userID string) tea.Cmd {
    return func() tea.Msg {
        tasks, err := client.ListTasks()
        if err != nil {
            return errMsg{err}
        }
        return tasksLoadedMsg{tasks}
    }
}
```

Loading states use `spinner.Model`. Errors surface in the status bar with a 4-second auto-dismiss.

---

## Feature Adaptation Notes

| Feature | Adaptation |
|---------|-----------|
| Diagrams | Excluded — terminal can't render the SPA's canvas editor. |
| Nutrition | Excluded — backend has no `/nutrition` endpoint; nutrition data lives under `/health/{date}` foods. |
| Note images/attachments | List attachment filenames. "Open" key launches `xdg-open`/`open` with presigned URL via `$BROWSER`. |
| Note rich text | Render markdown via `glamour` (charmbracelet/glamour, fixed dark style). Edit in `$EDITOR` (tmp file, write back) via `ctrl+e`. |
| Assistant chat | Spinner-then-render (no streaming — backend returns full reply). Multiline input via `enter` (rebound to newline by `formkeys.go`) or `ctrl+e`. Output in scrollable viewport. |
| Export | Trigger via `:export` palette command, display presigned URL. No ZIP extraction in TUI. |
| Admin | Read-only stats table. Renders "(unavailable)" for non-admin users. |
| Feeds articles | Read articles inline via viewport. External open via `$BROWSER` for full page. |
| Notes folders | Folder browser is the entry point; `enter` drills into a filtered notes list. |

---

## Stability notes

A few patterns that bit the early implementation and are now load-bearing
constraints:

- **Pin a fixed colour profile up front** (`lipgloss.SetColorProfile(termenv.TrueColor)`).
  Auto-detection issues an OSC 11 background-colour query against the terminal,
  which on some terminals (Ghostty, certain iTerm/tmux combos) replies via
  stdin. bubbletea reads the reply as fake key events and the resulting
  feedback loop freezes the program.
- **Pin glamour's style** (`glamour.WithStandardStyle("dark")`). The default
  `WithAutoStyle` does the same OSC query per render and reproduces the
  freeze inside any markdown panel.
- **Use `tea.ExecProcess` for `$EDITOR`.** Calling `exec.Command.Run()` while
  bubbletea owns the terminal corrupts the alt-screen.
- **Don't smuggle ANSI into bubbles/table cell values.** Its renderRow calls
  `runewidth.Truncate` BEFORE applying the cell style, miscounts byte length
  vs display width, and chops the ANSI mid-sequence. The `striped` fork
  exposes a per-row style hook so backgrounds layer at renderRow time.
- **Drop `Padding(0, 1)` from bubbles/table Cell + Header.** The default
  padding adds 2 chars per column at render time, exceeding the column
  widths the screen calculated, wrapping rows onto a second line. Keep the
  total predictable by zeroing the padding.

## Key design constraints (from CLAUDE.md)

- No emojis anywhere — UI strings, help text, status bar, comments.
- Single static binary. No CGO.
- Conventional Commits + release-please for version automation.
