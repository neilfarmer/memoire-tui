package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/api"
	"github.com/neilfarmer/memoire-tui/internal/styles"
	"github.com/neilfarmer/memoire-tui/internal/ui/components"
)

type assistantPane int

const (
	assistantPaneInput assistantPane = iota
	assistantPaneConvos
	assistantPaneMessages
)

type Assistant struct {
	client *api.Client
	width  int
	height int

	err         error
	convos      []api.Conversation
	messages    []api.ChatMessage
	currentConv string
	model       string
	pane        assistantPane

	input   textarea.Model
	view    viewport.Model
	convCur int
	sending bool
}

type convosLoadedMsg struct {
	convos []api.Conversation
	err    error
}
type convDetailMsg struct {
	convID string
	detail api.ConversationDetail
	err    error
}
type chatReplyMsg struct {
	resp api.ChatResponse
	err  error
}

func newAssistant(c *api.Client) *Assistant {
	a := &Assistant{client: c, model: "nova-lite"}
	a.input = textarea.New()
	a.input.Placeholder = "Message memoire… (ctrl+j to send)"
	a.input.SetHeight(3)
	a.input.ShowLineNumbers = false
	a.input.CharLimit = 0
	a.view = viewport.New(40, 20)
	a.input.Focus()
	return a
}

func (a *Assistant) Init() tea.Cmd {
	// Re-entering the screen (e.g. after pressing esc → sidebar →
	// enter Assistant again) needs to restore textarea focus so the
	// user can type. Without this they'd have to press tab twice to
	// cycle focus back.
	if a.pane == assistantPaneInput {
		a.input.Focus()
	}
	return a.refreshConvos()
}

func (a *Assistant) refreshConvos() tea.Cmd {
	c := a.client
	return func() tea.Msg {
		out, err := c.ListConversations()
		return convosLoadedMsg{convos: out, err: err}
	}
}

func (a *Assistant) loadConv(id string) tea.Cmd {
	c := a.client
	return func() tea.Msg {
		d, err := c.GetConversation(id)
		return convDetailMsg{convID: id, detail: d, err: err}
	}
}

func (a *Assistant) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case convosLoadedMsg:
		a.err = m.err
		a.convos = m.convos
		return a, nil
	case convDetailMsg:
		a.err = m.err
		a.currentConv = m.convID
		a.messages = m.detail.Messages
		a.refreshView()
		return a, nil
	case chatReplyMsg:
		a.sending = false
		if m.err != nil {
			a.err = m.err
			return a, nil
		}
		a.currentConv = m.resp.ConversationID
		a.messages = append(a.messages, api.ChatMessage{Role: "assistant", Content: m.resp.Reply})
		a.refreshView()
		return a, a.refreshConvos()
	case tea.KeyMsg:
		return a.handleKey(m)
	}
	switch a.pane {
	case assistantPaneInput:
		var cmd tea.Cmd
		a.input, cmd = a.input.Update(msg)
		return a, cmd
	case assistantPaneMessages:
		var cmd tea.Cmd
		a.view, cmd = a.view.Update(msg)
		return a, cmd
	}
	return a, nil
}

func (a *Assistant) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.String() {
	case "ctrl+j":
		return a, a.send()
	case "ctrl+l":
		return a, a.clearHistory()
	case "ctrl+n":
		return a, a.newConversation()
	case "ctrl+y":
		// Globally-bound model toggle (works in every pane). Plain `m`
		// only fires when input isn't focused so the user can type the
		// literal letter.
		a.toggleModel()
		return a, nil
	case "tab":
		a.pane = (a.pane + 1) % 3
		switch a.pane {
		case assistantPaneInput:
			a.input.Focus()
		default:
			a.input.Blur()
		}
		return a, nil
	}
	// `m` toggles models when the textarea isn't focused. ctrl+m is the
	// same byte as Enter in every terminal so the old binding was dead.
	if a.pane != assistantPaneInput && m.String() == "m" {
		a.toggleModel()
		return a, nil
	}
	if a.pane == assistantPaneConvos {
		switch m.String() {
		case "up", "k":
			if a.convCur > 0 {
				a.convCur--
			}
		case "down", "j":
			if a.convCur < len(a.convos)-1 {
				a.convCur++
			}
		case "enter":
			if a.convCur < len(a.convos) {
				return a, a.loadConv(a.convos[a.convCur].ConversationID)
			}
		}
		return a, nil
	}
	if a.pane == assistantPaneInput {
		var cmd tea.Cmd
		a.input, cmd = a.input.Update(m)
		return a, cmd
	}
	var cmd tea.Cmd
	a.view, cmd = a.view.Update(m)
	return a, cmd
}

func (a *Assistant) toggleModel() {
	if a.model == "nova-lite" {
		a.model = "nova-pro"
	} else {
		a.model = "nova-lite"
	}
}

func (a *Assistant) refreshView() {
	rows := []string{}
	w := a.view.Width - 4
	if w < 20 {
		w = 20
	}
	for _, m := range a.messages {
		rows = append(rows, renderMessageBubble(m.Role, m.Content, w, a.model))
		rows = append(rows, "")
	}
	a.view.SetContent(strings.Join(rows, "\n"))
	a.view.GotoBottom()
}

// renderMessageBubble draws a single chat turn: a coloured role label
// followed by a bordered box containing the markdown-rendered content.
// User turns use accent (amber); assistant turns use primary (cyan).
func renderMessageBubble(role, content string, width int, model string) string {
	isUser := role == "user"
	var labelStyle lipgloss.Style
	var borderColor lipgloss.AdaptiveColor
	var label string
	if isUser {
		labelStyle = lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
		borderColor = styles.Accent
		label = "you"
	} else {
		labelStyle = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
		borderColor = styles.Primary
		label = model
		if label == "" {
			label = "memoire"
		}
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width)
	body := components.RenderMarkdown(content, width-4)
	return labelStyle.Render(label) + "\n" + box.Render(body)
}

func (a *Assistant) send() tea.Cmd {
	body := strings.TrimSpace(a.input.Value())
	if body == "" {
		return nil
	}
	a.input.Reset()
	a.messages = append(a.messages, api.ChatMessage{Role: "user", Content: body})
	a.refreshView()
	a.sending = true
	c := a.client
	model := a.model
	convID := a.currentConv
	return func() tea.Msg {
		resp, err := c.Chat(api.ChatRequest{Message: body, Model: model, ConversationID: convID})
		return chatReplyMsg{resp: resp, err: err}
	}
}

func (a *Assistant) clearHistory() tea.Cmd {
	c := a.client
	a.messages = nil
	a.currentConv = ""
	a.refreshView()
	return func() tea.Msg {
		_ = c.ClearAssistantHistory()
		return convosLoadedMsg{}
	}
}

func (a *Assistant) newConversation() tea.Cmd {
	a.messages = nil
	a.currentConv = ""
	a.refreshView()
	return nil
}

// currentConvTitle returns the title of the currently-loaded conversation
// or empty if none.
func (a *Assistant) currentConvTitle() string {
	for _, c := range a.convos {
		if c.ConversationID == a.currentConv {
			return c.Title
		}
	}
	return ""
}

func (a *Assistant) View() string {
	// Layout budget per row:
	//   header(2) + body(messages_box + input_box) = a.height
	// where messages_box = viewHeight + 2 border and input_box = 5
	// (textarea Height(3) + 2 border). The status hints live in the
	// app's global status bar so the screen doesn't need its own.
	const (
		convoWidth         = 30
		headerLines        = 2
		inputBoxLines      = 5
		messagesBorderLine = 2
	)
	// body width budget: convos(convoWidth+2 border) + spacer(2) +
	// right(rightWidth+2 border) must equal a.width, so
	// rightWidth = a.width - convoWidth - 6. The old constant 3 was
	// off and pushed the body 3 cols past the app's contentBox,
	// forcing lipgloss to wrap every row and double the rendered
	// height — which is what hid the input box on real terminals.
	rightWidth := a.width - convoWidth - 6
	if rightWidth < 30 {
		rightWidth = 30
	}
	viewHeight := a.height - headerLines - inputBoxLines - messagesBorderLine
	if viewHeight < 5 {
		viewHeight = 5
	}
	a.view.Width = rightWidth - 2
	a.view.Height = viewHeight
	a.refreshView()

	header := a.renderHeader()
	convos := a.renderConvos(convoWidth, a.height-headerLines)
	messages := a.renderMessages(rightWidth, viewHeight)
	input := a.renderInput(rightWidth)

	right := lipgloss.JoinVertical(lipgloss.Left, messages, input)
	body := lipgloss.JoinHorizontal(lipgloss.Top, convos, "  ", right)
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (a *Assistant) renderHeader() string {
	title := a.currentConvTitle()
	if title == "" {
		title = "New conversation"
	}
	titleStyled := styles.Title.Render(title)

	chips := []string{}
	if a.sending {
		chips = append(chips, styles.ChipAccent.Render("thinking…"))
	}
	modelChip := styles.ChipPrimary.Render("model: " + a.model)
	chips = append(chips, modelChip)
	right := lipgloss.JoinHorizontal(lipgloss.Top, chips...)

	leftW := lipgloss.Width(titleStyled)
	rightW := lipgloss.Width(right)
	gap := a.width - leftW - rightW - 2
	if gap < 1 {
		gap = 1
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		titleStyled,
		lipgloss.NewStyle().Width(gap).Render(""),
		right,
	)
	rule := lipgloss.NewStyle().Foreground(styles.Border).Render(strings.Repeat("─", maxInt(a.width-2, 10)))
	return lipgloss.JoinVertical(lipgloss.Left, bar, rule)
}

func (a *Assistant) renderConvos(width, height int) string {
	rows := []string{styles.Heading.Render("Conversations")}
	if len(a.convos) == 0 {
		rows = append(rows, styles.MutedText.Render("(no conversations yet)"))
	}
	for i, c := range a.convos {
		title := c.Title
		if title == "" {
			title = "untitled"
		}
		active := c.ConversationID == a.currentConv
		titleWidth := width - 4
		ts := relativeTime(c.UpdatedAt)
		if ts == "" {
			ts = relativeTime(c.CreatedAt)
		}
		titleStr := truncate(title, titleWidth-len(ts)-2)
		line := titleStr + "  " + styles.MutedText.Render(ts)
		if active {
			line = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("● ") + line
		} else {
			line = "  " + line
		}
		if i == a.convCur && a.pane == assistantPaneConvos {
			line = styles.Selected.Render(line)
		}
		rows = append(rows, line)
	}
	box := styles.Box
	if a.pane == assistantPaneConvos {
		box = styles.BoxFocused
	}
	if height-2 > 0 {
		box = box.Height(height - 2)
	}
	return box.Width(width).Render(strings.Join(rows, "\n"))
}

func (a *Assistant) renderMessages(width, height int) string {
	box := styles.Box
	if a.pane == assistantPaneMessages {
		box = styles.BoxFocused
	}
	if len(a.messages) == 0 {
		// Empty state must fit inside the box's inner width
		// (width - 4: -2 border, -2 padding) on a single line each.
		// Anything wider forces lipgloss to wrap and the rendered
		// height grows past the budget, clipping the input box.
		empty := lipgloss.JoinVertical(lipgloss.Center,
			styles.Heading.Render("Start a conversation"),
			"",
			styles.MutedText.Render("type below • ctrl+j sends"),
		)
		inner := lipgloss.Place(width-4, height, lipgloss.Center, lipgloss.Center, empty)
		return box.Width(width).Height(height).Render(inner)
	}
	return box.Width(width).Height(height).Render(a.view.View())
}

func (a *Assistant) renderInput(width int) string {
	a.input.SetWidth(width - 4)
	box := styles.Box
	if a.pane == assistantPaneInput {
		box = styles.BoxFocused
	}
	prompt := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("> ")
	body := lipgloss.JoinHorizontal(lipgloss.Top, prompt, a.input.View())
	return box.Width(width).Render(body)
}

func relativeTime(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try a few looser forms before giving up.
		for _, layout := range []string{"2006-01-02T15:04:05Z", "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err = time.Parse(layout, s); err == nil {
				break
			}
		}
		if err != nil {
			return ""
		}
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *Assistant) Title() string { return "Assistant" }
func (a *Assistant) StatusHints() []string {
	return []string{
		styles.KeyHint("ctrl+j", "send"),
		styles.KeyHint("ctrl+y", "model"),
		styles.KeyHint("ctrl+n", "new"),
		styles.KeyHint("ctrl+l", "clear"),
		styles.KeyHint("tab", "pane"),
	}
}
func (a *Assistant) Help() []components.HelpEntry {
	return []components.HelpEntry{
		{Keys: "ctrl+j", Desc: "send message"},
		{Keys: "ctrl+y", Desc: "toggle model (nova-lite / nova-pro) — works in every pane"},
		{Keys: "m", Desc: "alternate model toggle — only fires outside the input pane"},
		{Keys: "ctrl+l", Desc: "clear current conversation"},
		{Keys: "ctrl+n", Desc: "start a new conversation"},
		{Keys: "tab", Desc: "cycle pane (input / conversations / messages)"},
		{Keys: "enter (in conversations)", Desc: "open conversation"},
		{Keys: ":model-nova-lite / :model-nova-pro", Desc: "switch model via the command palette"},
	}
}
func (a *Assistant) SetSize(w, h int) { a.width, a.height = w, h }

// IsTextEditing reports that the input textarea owns focus, so the App
// should not steal single-character shortcuts (e.g. "?") that the user is
// trying to type into the chat input.
func (a *Assistant) IsTextEditing() bool { return a.pane == assistantPaneInput }

func (a *Assistant) OnEscape() bool {
	// Assistant has no sub-mode tree to pop. Return false so the app
	// pops focus back to the sidebar. We deliberately do NOT blur the
	// input here: blurring it would mean the textarea is still blurred
	// when the user later re-enters the screen, breaking typing.
	// Init() refocuses on re-entry as a belt-and-braces guarantee.
	return false
}

func (a *Assistant) PaletteCommands() []components.Command {
	return []components.Command{
		{Name: "new-conversation", Display: "Start a new conversation", Group: "Assistant", Run: func() tea.Cmd { return a.newConversation() }},
		{Name: "clear-history", Display: "Clear current conversation", Group: "Assistant", Run: func() tea.Cmd { return a.clearHistory() }},
		{Name: "model-nova-lite", Display: "Switch model: nova-lite", Group: "Assistant", Run: func() tea.Cmd { a.model = "nova-lite"; return nil }},
		{Name: "model-nova-pro", Display: "Switch model: nova-pro", Group: "Assistant", Run: func() tea.Cmd { a.model = "nova-pro"; return nil }},
	}
}
