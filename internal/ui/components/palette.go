package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/styles"
)

// Command is a single palette entry. Name is what the user types; Display
// shows the label in the suggestion list; Group buckets entries (e.g.
// "Navigate", "Action"). Run is invoked when the user selects the command.
type Command struct {
	Name    string
	Display string
	Group   string
	Hint    string
	Run     func() tea.Cmd
}

// CommandPalette is a k9s-style overlay: text input + filtered suggestion
// list + Enter to run.
type CommandPalette struct {
	input    textinput.Model
	commands []Command
	cursor   int
	width    int
	height   int
}

// NewCommandPalette builds a palette with the given commands.
func NewCommandPalette(commands []Command) *CommandPalette {
	in := textinput.New()
	in.Placeholder = "type a command (e.g. tasks, export, quit)"
	in.Prompt = ":"
	in.Focus()
	return &CommandPalette{input: in, commands: commands}
}

// SetSize updates the rendering width/height.
func (p *CommandPalette) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.input.Width = width - 8
}

// HandleKey advances the palette state. Returns (selected_command, done).
// done=true means the palette is closing — either the user selected a
// command (cmd != nil) or pressed esc (cmd == nil).
func (p *CommandPalette) HandleKey(m tea.KeyMsg) (*Command, bool) {
	switch m.String() {
	case "esc":
		return nil, true
	case "enter":
		if c := p.selected(); c != nil {
			return c, true
		}
		return nil, true
	case "up", "ctrl+k":
		filtered := p.filtered()
		if len(filtered) == 0 {
			return nil, false
		}
		if p.cursor > 0 {
			p.cursor--
		}
		return nil, false
	case "down", "ctrl+j":
		filtered := p.filtered()
		if len(filtered) == 0 {
			return nil, false
		}
		if p.cursor < len(filtered)-1 {
			p.cursor++
		}
		return nil, false
	case "tab":
		// Tab autocompletes to the first match.
		if c := p.selected(); c != nil {
			p.input.SetValue(c.Name)
			p.input.CursorEnd()
		}
		return nil, false
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(m)
	_ = cmd
	// Keep cursor in range when filter shrinks.
	if filtered := p.filtered(); p.cursor >= len(filtered) {
		p.cursor = max(0, len(filtered)-1)
	}
	return nil, false
}

func (p *CommandPalette) filtered() []Command {
	q := strings.ToLower(strings.TrimSpace(p.input.Value()))
	if q == "" {
		return p.commands
	}
	out := make([]Command, 0, len(p.commands))
	for _, c := range p.commands {
		if strings.Contains(strings.ToLower(c.Name), q) ||
			strings.Contains(strings.ToLower(c.Display), q) {
			out = append(out, c)
		}
	}
	return out
}

func (p *CommandPalette) selected() *Command {
	filtered := p.filtered()
	if p.cursor >= len(filtered) || len(filtered) == 0 {
		return nil
	}
	c := filtered[p.cursor]
	return &c
}

// View renders the palette as a centred overlay.
func (p *CommandPalette) View() string {
	const maxRows = 10
	width := p.width - 8
	if width > 80 {
		width = 80
	}
	if width < 30 {
		width = 30
	}
	header := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("Command palette")
	rows := []string{header, "", p.input.View(), ""}
	filtered := p.filtered()
	if len(filtered) == 0 {
		rows = append(rows, styles.MutedText.Render("(no matches)"))
	} else {
		shown := filtered
		if len(shown) > maxRows {
			shown = shown[:maxRows]
		}
		var lastGroup string
		for i, c := range shown {
			if c.Group != "" && c.Group != lastGroup {
				rows = append(rows, lipgloss.NewStyle().Foreground(styles.Muted).Render(c.Group))
				lastGroup = c.Group
			}
			row := "  " + c.Display
			if c.Hint != "" {
				row += "  " + styles.MutedText.Render(c.Hint)
			}
			if i == p.cursor {
				row = styles.Selected.Render(row)
			}
			rows = append(rows, row)
		}
	}
	rows = append(rows, "", styles.MutedText.Render("↑/↓ select   ↵ run   esc cancel   tab autocomplete"))
	box := styles.BoxFocused.Width(width).Render(strings.Join(rows, "\n"))
	return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, box)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
