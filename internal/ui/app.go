package ui

import (
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/api"
	"github.com/neilfarmer/memoire-tui/internal/logx"
	"github.com/neilfarmer/memoire-tui/internal/styles"
	"github.com/neilfarmer/memoire-tui/internal/ui/components"
	"github.com/neilfarmer/memoire-tui/internal/ui/screens"
)

// ScreenFactory builds a Screen lazily on first navigation.
type ScreenFactory func(client *api.Client) screens.Screen

// App is the root Bubble Tea Model.
type App struct {
	client      *api.Client
	apiHost     string
	keys        GlobalKeys
	current     Screen
	sideCursor  int  // index into SidebarOrder; tracks current screen for sidebar render
	sideFocus   bool // true = arrows nav sidebar; false = arrows nav content
	registry    map[Screen]screens.Screen
	factories   map[Screen]ScreenFactory
	width       int
	height      int
	flash       string
	flashLevel  FlashLevel
	flashID     int64
	helpOpen    bool
	paletteOpen bool
	palette     *components.CommandPalette
	quitConfirm bool
	connected   bool
	lastConnAt  time.Time
}

// New builds the root App with the given API client and screen factories.
func New(client *api.Client, factories map[Screen]ScreenFactory) *App {
	host := ""
	if u, err := url.Parse(client.BaseURL); err == nil {
		host = u.Host
	}
	return &App{
		client:     client,
		apiHost:    host,
		keys:       DefaultKeys(),
		current:    ScreenDashboard,
		sideCursor: 0,
		sideFocus:  true,
		registry:   map[Screen]screens.Screen{},
		factories:  factories,
		connected:  true,
	}
}

func (a *App) Init() tea.Cmd {
	return a.activate(a.current)
}

func (a *App) activate(s Screen) tea.Cmd {
	first := false
	if _, ok := a.registry[s]; !ok {
		first = true
		factory, ok := a.factories[s]
		if !ok {
			a.registry[s] = screens.NewPlaceholder(SidebarLabels[s])
		} else {
			a.registry[s] = factory(a.client)
		}
	}
	a.current = s
	a.registry[s].SetSize(a.contentWidth(), a.contentHeight())
	logx.Debug("activate screen", "screen", string(s), "first_init", first)
	return a.registry[s].Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = m.Width, m.Height
		for _, s := range a.registry {
			s.SetSize(a.contentWidth(), a.contentHeight())
		}
	case tea.KeyMsg:
		// Filter terminal OSC responses that some terminals leak as keys
		// (e.g. background-color reply "]11;rgb:..."). Forwarding these to
		// the screen causes feedback loops + freezes.
		if isTerminalNoiseKey(m.String()) {
			return a, nil
		}
		logx.Debug("key event", "key", m.String(), "screen", string(a.current))
		if cmd, handled := a.handleKey(m); handled {
			return a, cmd
		}
	case NavigateMsg:
		return a, a.activate(m.To)
	case FlashMsg:
		return a, a.setFlash(m)
	case FlashClearMsg:
		if m.ID == a.flashID {
			a.flash = ""
		}
	case ErrorMsg:
		logx.Error("error msg", "err", m.Err)
		return a, a.setFlash(FlashMsg{Text: "error: " + m.Err.Error(), Level: FlashError, TTL: 6 * time.Second})
	case HelpToggleMsg:
		a.helpOpen = !a.helpOpen
		return a, nil
	}
	if cur, ok := a.registry[a.current]; ok {
		updated, cmd := cur.Update(msg)
		if s, ok := updated.(screens.Screen); ok {
			a.registry[a.current] = s
		}
		return a, cmd
	}
	return a, nil
}

func (a *App) handleKey(m tea.KeyMsg) (tea.Cmd, bool) {
	// Quit-confirm overlay traps everything until dismissed.
	if a.quitConfirm {
		switch m.String() {
		case "y", "enter":
			return tea.Quit, true
		case "n", "esc":
			a.quitConfirm = false
			return nil, true
		}
		return nil, true
	}
	// Help overlay traps esc + ?.
	if a.helpOpen {
		if m.String() == "?" || m.String() == "esc" {
			a.helpOpen = false
			return nil, true
		}
	}
	// Command palette traps everything; let the palette consume keys until
	// it emits a command or is closed via esc.
	if a.paletteOpen && a.palette != nil {
		cmd, done := a.palette.HandleKey(m)
		if done {
			a.paletteOpen = false
			if cmd != nil {
				return a.executeCommand(cmd), true
			}
		}
		return nil, true
	}
	if key.Matches(m, a.keys.Quit) {
		return tea.Quit, true
	}
	// Sidebar focus: arrows nav screens, enter drills into content. The
	// sidebar always wins arrow + enter when focused, even if the active
	// screen has a textarea (e.g. Assistant) — the textarea isn't really
	// receiving input until the user drills in.
	if a.sideFocus {
		switch m.String() {
		case "up":
			if a.sideCursor > 0 {
				a.sideCursor--
			}
			return a.activate(SidebarOrder[a.sideCursor]), true
		case "down":
			if a.sideCursor < len(SidebarOrder)-1 {
				a.sideCursor++
			}
			return a.activate(SidebarOrder[a.sideCursor]), true
		case "enter", "right":
			a.sideFocus = false
			return nil, true
		}
	}
	// `?` toggles help unless content is focused with a real text input.
	if m.String() == "?" {
		if !a.sideFocus && a.currentScreenIsEditing() {
			return nil, false
		}
		a.helpOpen = !a.helpOpen
		return nil, true
	}
	// `:` opens the command palette unless content is focused with a real
	// text input.
	if m.String() == ":" {
		if !a.sideFocus && a.currentScreenIsEditing() {
			return nil, false
		}
		a.palette = components.NewCommandPalette(a.commandList())
		a.paletteOpen = true
		return nil, true
	}
	// Esc drill-up:
	//   sidebar focused → quit confirm (regardless of which screen is shown)
	//   content focused, screen sub-mode → screen.OnEscape pops one level
	//   content focused, top level → return focus to sidebar
	if m.String() == "esc" {
		if a.sideFocus {
			a.quitConfirm = true
			return nil, true
		}
		// Content focused. Forms / chat textareas handle their own esc.
		if a.currentScreenIsEditing() {
			return nil, false
		}
		if cur, ok := a.registry[a.current]; ok {
			if e, ok := cur.(escapableScreen); ok {
				if e.OnEscape() {
					return nil, true
				}
			}
		}
		// Content is at its top level — return focus to sidebar.
		a.sideFocus = true
		return nil, true
	}
	return nil, false
}

// escapableScreen is implemented by screens that have a sub-mode tree
// (detail / form / overlay). OnEscape pops one level and returns true; if
// already at the top level, returns false so the App can quit.
type escapableScreen interface {
	OnEscape() bool
}

// textEditing is implemented by screens that have a focused text input
// (textarea, huh form input). When true, app-level single-character
// shortcuts must defer to the screen so the user can type literal chars
// like "?" or "n" into the input.
type textEditing interface {
	IsTextEditing() bool
}

// currentScreenIsEditing reports whether the active screen claims a text
// input is focused.
func (a *App) currentScreenIsEditing() bool {
	cur, ok := a.registry[a.current]
	if !ok {
		return false
	}
	if te, ok := cur.(textEditing); ok {
		return te.IsTextEditing()
	}
	return false
}

// isTerminalNoiseKey reports whether a tea KeyMsg.String() looks like a
// terminal OSC response or other escape-leakage rather than user input.
// The leaked OSC color-query reply looks like "]11;rgb:1c1c/1c1c/1c1c" —
// an introducer (']' or '[') followed immediately by digits, often with
// ";rgb:" embedded. Bare bracket keys ('[' or ']') must pass through
// untouched so users can type them.
func isTerminalNoiseKey(k string) bool {
	if k == "" {
		return false
	}
	if strings.Contains(k, ";rgb:") {
		return true
	}
	if len(k) >= 2 && (k[0] == ']' || k[0] == '[') && k[1] >= '0' && k[1] <= '9' {
		return true
	}
	return false
}

func (a *App) setFlash(f FlashMsg) tea.Cmd {
	a.flashID++
	id := a.flashID
	a.flash = f.Text
	a.flashLevel = f.Level
	ttl := f.TTL
	if ttl == 0 {
		ttl = 4 * time.Second
	}
	return tea.Tick(ttl, func(time.Time) tea.Msg { return FlashClearMsg{ID: id} })
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "loading..."
	}
	header := components.Header{
		Section:   SidebarLabels[a.current],
		Connected: a.connected,
		APIHost:   a.apiHost,
		Auth:      "PAT",
		Width:     a.width,
	}.View()

	items := make([]components.SidebarItem, 0, len(SidebarOrder))
	for _, s := range SidebarOrder {
		items = append(items, components.SidebarItem{
			Key:   string(s),
			Label: SidebarLabels[s],
			Icon:  SidebarIcon(s),
		})
	}
	side := components.Sidebar{
		Items:   items,
		Active:  string(SidebarOrder[a.sideCursor]),
		Width:   24,
		Focused: a.sideFocus,
	}.View()

	var content string
	if cur, ok := a.registry[a.current]; ok {
		content = cur.View()
	}
	// Hard-clip content height so an oversized screen view cannot push the
	// sidebar off-screen.
	contentBox := lipgloss.NewStyle().Width(a.contentWidth()).Height(a.contentHeight()).MaxHeight(a.contentHeight()).Render(content)

	var hints []string
	if a.sideFocus {
		hints = []string{
			styles.KeyHint("↑↓", "screens"),
			styles.KeyHint("↵", "enter screen"),
			styles.KeyHint(":", "command"),
			styles.KeyHint("esc", "quit"),
			styles.KeyHint("?", "help"),
		}
	} else {
		hints = []string{
			styles.KeyHint(":", "command"),
			styles.KeyHint("esc", "back"),
			styles.KeyHint("?", "help"),
		}
		if cur, ok := a.registry[a.current]; ok {
			extra := cur.StatusHints()
			hints = append(extra, hints...)
		}
	}
	status := components.StatusBar{
		Screen: SidebarLabels[a.current],
		Flash:  a.flash,
		Hints:  hints,
		Width:  a.width,
	}.View()

	body := lipgloss.JoinHorizontal(lipgloss.Top, side, contentBox)
	out := lipgloss.JoinVertical(lipgloss.Left, header, body, status)

	if a.helpOpen {
		return components.HelpView(a.width, a.height, a.helpSections(), a.helpOrder())
	}
	if a.paletteOpen && a.palette != nil {
		a.palette.SetSize(a.width, a.height)
		return a.palette.View()
	}
	if a.quitConfirm {
		return components.ConfirmView("Quit memoire?", a.width, a.height)
	}
	return out
}

func (a *App) helpSections() map[string][]components.HelpEntry {
	out := map[string][]components.HelpEntry{
		"Global": {
			{Keys: "↑/↓", Desc: "navigate (sidebar when focused, rows in content)"},
			{Keys: "↵ enter", Desc: "drill in (sidebar → screen → detail/form)"},
			{Keys: "esc", Desc: "back one level (quit confirm at sidebar)"},
			{Keys: ":", Desc: "open command palette"},
			{Keys: "?", Desc: "toggle help"},
			{Keys: "ctrl+r", Desc: "refresh current screen"},
			{Keys: "ctrl+q", Desc: "force quit"},
		},
		"Universal screen actions": {
			{Keys: "n", Desc: "new entry"},
			{Keys: "e", Desc: "edit selected"},
			{Keys: "d", Desc: "delete selected"},
			{Keys: "o", Desc: "open URL in browser (where applicable)"},
			{Keys: "/", Desc: "filter list"},
			{Keys: "f", Desc: "cycle filter pill"},
			{Keys: "s", Desc: "cycle sort (where applicable)"},
			{Keys: "r", Desc: "refresh"},
		},
	}
	if cur, ok := a.registry[a.current]; ok {
		if entries := cur.Help(); len(entries) > 0 {
			out[cur.Title()] = entries
		}
	}
	return out
}

func (a *App) helpOrder() []string {
	order := []string{"Global", "Universal screen actions"}
	if cur, ok := a.registry[a.current]; ok {
		order = append(order, cur.Title())
	}
	return order
}

// commandList builds the palette commands available right now: every screen
// jump plus any per-screen heavy actions exposed via PaletteCommands().
func (a *App) commandList() []components.Command {
	cmds := []components.Command{}
	for _, s := range SidebarOrder {
		s := s
		cmds = append(cmds, components.Command{
			Name:    string(s),
			Display: SidebarLabels[s],
			Group:   "Navigate",
			Hint:    "screen",
			Run: func() tea.Cmd {
				a.sideCursor = sidebarIndex(s)
				return a.activate(s)
			},
		})
	}
	if cur, ok := a.registry[a.current]; ok {
		if pc, ok := cur.(paletteContributor); ok {
			cmds = append(cmds, pc.PaletteCommands()...)
		}
	}
	cmds = append(cmds,
		components.Command{Name: "help", Display: "Help", Group: "App", Run: func() tea.Cmd { a.helpOpen = true; return nil }},
		components.Command{Name: "quit", Display: "Quit", Group: "App", Run: func() tea.Cmd { return tea.Quit }},
		components.Command{Name: "refresh", Display: "Refresh current screen", Group: "App", Run: func() tea.Cmd {
			if cur, ok := a.registry[a.current]; ok {
				return cur.Init()
			}
			return nil
		}},
	)
	return cmds
}

// paletteContributor lets a screen add its own heavy actions to the palette.
type paletteContributor interface {
	PaletteCommands() []components.Command
}

// executeCommand runs a palette selection and returns the resulting tea.Cmd.
func (a *App) executeCommand(c *components.Command) tea.Cmd {
	if c == nil || c.Run == nil {
		return nil
	}
	return c.Run()
}

func sidebarIndex(s Screen) int {
	for i, x := range SidebarOrder {
		if x == s {
			return i
		}
	}
	return 0
}

func (a *App) contentWidth() int {
	w := a.width - 26 // sidebar(24) + 2 padding
	if w < 40 {
		w = 40
	}
	return w
}

func (a *App) contentHeight() int {
	h := a.height - 5 // header(2) + statusbar(2) + spacing(1)
	if h < 10 {
		h = 10
	}
	return h
}

// SetConnection toggles the header indicator. Currently only used by initial
// reachability check.
func (a *App) SetConnection(ok bool) {
	a.connected = ok
	if ok {
		a.lastConnAt = time.Now()
	}
}

// keep the styles import in use for KeyHint and adaptive colors.
var _ = styles.Primary
