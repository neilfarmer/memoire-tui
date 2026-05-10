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

Global:

| Key | Action |
|-----|--------|
| `?` | toggle help overlay |
| `ctrl+q` | quit |
| `1`–`9` | jump to the first 9 sidebar entries |
| `g <letter>` | leader nav (`g t` = tasks, `g n` = notes, …) |
| `ctrl+r` | refresh current screen |

Per-screen keys are listed in the help overlay and the bottom status bar.

## Screens

| Screen | Notes |
|--------|-------|
| Dashboard | Today's tasks / habits / latest note summary |
| Tasks | Filter / sort / group + create-edit-delete + auto-schedule + 7-day agenda |
| Notes | Folder tree + markdown rendering. `ctrl+e` in the body editor opens `$EDITOR`. |
| Journal | Month calendar with markers; one entry per day; mood + tags |
| Habits | 30-day ASCII history per habit; `space` toggles today |
| Goals | Status filter + form |
| Health | Date picker, totals, foods, exercises, 7-day summary |
| Nutrition | Date picker, meal log, totals row |
| Finances | Tabs: debts / income / expenses + summary header |
| Feeds | Two-pane (feeds / articles); inline article reader; favorite + mark read |
| Bookmarks | Search + tag filter |
| Favorites | Tag filter + remove |
| Settings | Account / Appearance / Notifications / Editor + export + test-notification |
| Tokens | List / create. Disabled and labeled when session is PAT-authenticated. |
| Assistant | Multi-turn chat with model picker + conversations. Spinner-then-render (no streaming). |
| Admin | Costs + DynamoDB / S3 stats. Sidebar entry stays available; the screen shows "(unavailable)" for non-admin users. |

Diagrams (canvas-based) is intentionally not in the TUI.

## Running tests

```bash
make test              # go test ./...
make lint              # go vet + gofmt
make security          # govulncheck
```

## Layout

```
cmd/memoire/main.go             # entry point
internal/
  api/                          # HTTP client per feature
  config/                       # TOML + env loader, first-run prompt
  styles/                       # adaptive color palette + shared lipgloss styles
  ui/
    app.go                      # root model + screen routing
    keys.go                     # global key bindings
    messages.go                 # tea.Msg types
    factories.go                # screen factory map
    components/                 # sidebar, statusbar, header, confirm, help, markdown, editor, opener, datepicker, asciichart
    screens/                    # one Model per feature
```
