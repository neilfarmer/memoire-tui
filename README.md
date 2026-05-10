# memoire TUI

Terminal client for memoire, distributed as a single static binary.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), Lipgloss, Bubbles, and Huh. No CGO; `go build` produces a static binary on every supported platform.

## Install

Pre-built binaries land on every [GitHub Release](https://github.com/neilfarmer/memoire-tui/releases). The snippets below grab the latest version, drop the binary at `~/.local/bin/memoire`, and never need `sudo`.

Make sure `~/.local/bin` is on your `PATH`. If not, add this once:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc   # or ~/.bashrc
```

### macOS (Apple Silicon)

```bash
mkdir -p ~/.local/bin
URL=$(curl -fsSL https://api.github.com/repos/neilfarmer/memoire-tui/releases/latest \
  | grep -oE '"browser_download_url": *"[^"]+darwin_arm64\.tar\.gz"' \
  | head -1 | cut -d'"' -f4)
curl -fsSL "$URL" | tar -xzf - -C ~/.local/bin memoire
chmod +x ~/.local/bin/memoire
xattr -d com.apple.quarantine ~/.local/bin/memoire 2>/dev/null || true
memoire --version
```

(Drop `xattr` if you don't see the macOS quarantine prompt.)

### macOS (Intel)

```bash
mkdir -p ~/.local/bin
URL=$(curl -fsSL https://api.github.com/repos/neilfarmer/memoire-tui/releases/latest \
  | grep -oE '"browser_download_url": *"[^"]+darwin_x86_64\.tar\.gz"' \
  | head -1 | cut -d'"' -f4)
curl -fsSL "$URL" | tar -xzf - -C ~/.local/bin memoire
chmod +x ~/.local/bin/memoire
memoire --version
```

### Linux x86_64

```bash
mkdir -p ~/.local/bin
URL=$(curl -fsSL https://api.github.com/repos/neilfarmer/memoire-tui/releases/latest \
  | grep -oE '"browser_download_url": *"[^"]+linux_x86_64\.tar\.gz"' \
  | head -1 | cut -d'"' -f4)
curl -fsSL "$URL" | tar -xzf - -C ~/.local/bin memoire
chmod +x ~/.local/bin/memoire
memoire --version
```

### Linux arm64

```bash
mkdir -p ~/.local/bin
URL=$(curl -fsSL https://api.github.com/repos/neilfarmer/memoire-tui/releases/latest \
  | grep -oE '"browser_download_url": *"[^"]+linux_arm64\.tar\.gz"' \
  | head -1 | cut -d'"' -f4)
curl -fsSL "$URL" | tar -xzf - -C ~/.local/bin memoire
chmod +x ~/.local/bin/memoire
memoire --version
```

### Verify checksums (optional)

```bash
URL=$(curl -fsSL https://api.github.com/repos/neilfarmer/memoire-tui/releases/latest \
  | grep -oE '"browser_download_url": *"[^"]+checksums\.txt"' \
  | head -1 | cut -d'"' -f4)
curl -fsSL "$URL"
```

Match the line for your archive against the `sha256sum` output of the file you downloaded.

### Build from source

```bash
git clone https://github.com/neilfarmer/memoire-tui
cd memoire-tui
make build            # produces bin/memoire
./bin/memoire
```

Or directly with Go:

```bash
go run ./cmd/memoire
```

## First-run setup

The first launch prompts for the API URL and Personal Access Token, then writes them to `~/.config/memoire-tui/config.toml` with `0600` permissions (or `$XDG_CONFIG_HOME/memoire-tui/config.toml` if set).

You can also pre-supply them via environment variables, which override the config file:

```bash
export MEMOIRE_API_URL="https://api.memoire.example.com"
export MEMOIRE_PAT="pat_..."
memoire
```

Generate a Personal Access Token in the web UI under **Settings → API Tokens**. PATs cannot create or revoke other PATs (the Tokens screen displays a banner and disables mutating actions when the session is PAT-authenticated).

## Flags

| Flag | Purpose |
|------|---------|
| `-h, --help` | show help and exit |
| `-v, --version` | print version and commit |
| `--config <path>` | use an alternate config file |
| `--no-color` | disable color output |

## Environment

| Variable | Effect |
|----------|--------|
| `MEMOIRE_API_URL` | API base URL (overrides config) |
| `MEMOIRE_PAT` | Personal Access Token (overrides config) |
| `EDITOR` | editor used for note / journal bodies (default `vi`) |
| `BROWSER` | command to open URLs (defaults to `open` on darwin, `xdg-open` on linux, `rundll32` on windows) |
| `XDG_CONFIG_HOME` | overrides config directory |
| `NO_COLOR` | set to disable color (mirrors `--no-color`) |

## Keys

The TUI uses an implicit focus model: at boot, focus is on the sidebar; `↵` drills into content; `esc` walks back up.

**Global**

| Key | Action |
|-----|--------|
| `↑/↓` | move cursor (sidebar when focused, table rows when content is focused) |
| `↵ enter` | drill in (sidebar → screen → detail/form) |
| `esc` | back one level (drill up → sidebar focus → quit confirm) |
| `:` | open command palette |
| `?` | toggle help overlay |
| `ctrl+r` | refresh current screen |
| `ctrl+q` | force quit |

**Universal per-screen actions** (any list/detail screen)

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

**Form mode** (any huh-form screen)

| Key | Action |
|-----|--------|
| `tab / shift+tab` | next / previous field |
| `enter` | new line inside body textareas |
| `ctrl+e` | open body in `$EDITOR` (notes, journal) |
| `ctrl+s` | save form |
| `esc` | cancel |

**Command palette** (`:`)

Filterable, k9s-style. Always includes screen jumps + heavy actions exposed by the current screen:

| Command | Screen | Action |
|---------|--------|--------|
| `auto-schedule`, `agenda` | Tasks | reschedule unscheduled tasks; agenda for next 7 days |
| `all-notes` | Notes | skip folder filter |
| `trends` | Health | 7-day summary |
| `force-refresh` | Feeds | force re-fetch articles |
| `export`, `test-notify` | Settings | export ZIP / send test ntfy |
| `new-conversation`, `clear-history`, `model-nova-lite`, `model-nova-pro` | Assistant | conversation + model controls |
| `help`, `quit`, `refresh` | Anywhere | global actions |
| `dashboard`, `tasks`, `notes`, … | Anywhere | jump to any screen by name |

Per-screen keys also appear in the help overlay (`?`) and the bottom status bar.

## Screens

| Screen | Notes |
|--------|-------|
| Dashboard | Today's tasks / habits / latest note summary |
| Tasks | Filter pills (all / todo / in-progress / done), sort cycle, table view. `:auto-schedule` and `:agenda` via palette. |
| Notes | Folder browser is the entry point: each folder shows note count; enter drills into the filtered notes list. Markdown rendering in detail; `ctrl+e` in the form opens `$EDITOR`. |
| Journal | Month calendar with day markers; one entry per day; mood + tags |
| Habits | 30-day ASCII history per habit; `space` toggles today |
| Goals | Status filter + form |
| Health | Date picker, totals, foods, exercises. `:trends` palette command for the 7-day summary. |
| Finances | Tabs: debts / income / expenses + summary header |
| Feeds | Two-pane (feeds / articles); inline article reader; favorite + mark read; `:force-refresh` palette command |
| Bookmarks | Search + tag filter |
| Favorites | Tag filter + remove |
| Settings | Account / Appearance / Notifications / Editor; `:export` and `:test-notify` palette commands |
| Tokens | List / create. Disabled with banner when session is PAT-authenticated. |
| Assistant | Multi-turn chat with conversations + model picker. Palette: `:new-conversation`, `:clear-history`, `:model-nova-lite`, `:model-nova-pro`. Spinner-then-render (no streaming). |
| Admin | Costs + DynamoDB / S3 stats. Renders "(unavailable)" for non-admin users. |

Not in the TUI:

- **Diagrams** (canvas-based) — terminal can't render the SPA's editing surface.
- **Nutrition** — backend has no separate `/nutrition` endpoint; nutrition data lives under `/health/{date}` foods. Use the Health screen.

## Running tests

```bash
make test              # go test ./... -count=1 -race
make test-cover        # coverage report
make lint              # go vet + gofmt -l
make security          # govulncheck
make build             # bin/memoire
make run               # go run ./cmd/memoire
```

## Logs

The binary writes a structured debug log on every run. Default destination is `/tmp/memoire-tui.log`. Override with `MEMOIRE_LOG=path` or disable with `MEMOIRE_LOG=off`.

Each entry is a single key=value text line: `key event`, `activate screen`, `api response`, `api error`. To capture panics + stack traces, run the binary with stderr redirected:

```bash
./bin/memoire 2>/tmp/memoire-stderr.log
```

## Layout

```
cmd/
  memoire/main.go               # entry point
  smoke/main.go                 # dev-only screen-frame dumper
internal/
  api/                          # HTTP client per feature (one file per feature)
  config/                       # TOML + env loader, first-run prompt
  logx/                         # file-based debug log
  styles/                       # adaptive color palette + shared lipgloss styles
  ui/
    app.go                      # root model + screen routing + palette + esc drill-up
    factories.go                # screen factory map
    keys.go                     # screen ordering + sidebar icons
    messages.go                 # tea.Msg types
    components/
      sidebar.go statusbar.go header.go    # chrome
      confirm.go help.go palette.go        # overlays
      markdown.go editor.go opener.go      # external integrations
      datepicker.go asciichart.go          # widgets
      table.go                              # NewTable wrapping the striped fork
      formkeys.go                           # huh keymap (enter = newline)
      striped/                              # vendored fork of bubbles/table v1.0.0
                                            # with a per-row RowStyler hook
    screens/                                # one Model per feature
```

The `striped` package is a forked copy of `charmbracelet/bubbles@v1.0.0/table` with a single addition: `Styles.RowStyler func(rowIndex int) lipgloss.Style`. The forked `renderRow` runs the hook to layer per-row backgrounds without smuggling ANSI through `runewidth.Truncate`.
