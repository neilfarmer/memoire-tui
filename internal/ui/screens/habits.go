package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/api"
	"github.com/neilfarmer/memoire-tui/internal/styles"
	"github.com/neilfarmer/memoire-tui/internal/ui/components"
)

type habitMode int

const (
	habitList habitMode = iota
	habitForm
	habitConfirmDelete
)

type Habits struct {
	client *api.Client

	mode    habitMode
	width   int
	height  int
	loading bool
	err     error

	habits []api.Habit
	cursor int
	form   *huh.Form
	formIn habitFormState
}

type habitFormState struct {
	id         string
	name       string
	desc       string
	frequency  string
	notifyTime string
	timeOfDay  string
}

type habitsLoadedMsg struct {
	habits []api.Habit
	err    error
}

type habitsMutatedMsg struct{ err error }

func newHabits(c *api.Client) *Habits {
	return &Habits{client: c}
}

func (h *Habits) Init() tea.Cmd { return h.refresh() }

func (h *Habits) refresh() tea.Cmd {
	c := h.client
	return func() tea.Msg {
		habits, err := c.ListHabits()
		return habitsLoadedMsg{habits: habits, err: err}
	}
}

func (h *Habits) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case habitsLoadedMsg:
		h.loading = false
		h.err = m.err
		h.habits = m.habits
		if h.cursor >= len(h.habits) {
			h.cursor = len(h.habits) - 1
		}
		if h.cursor < 0 {
			h.cursor = 0
		}
		return h, nil
	case habitsMutatedMsg:
		h.err = m.err
		h.mode = habitList
		return h, h.refresh()
	case tea.KeyMsg:
		return h.handleKey(m)
	}
	if h.mode == habitForm && h.form != nil {
		return h, updateForm(&h.form, msg, func() { h.mode = habitList }, h.submit)
	}
	return h, nil
}

func (h *Habits) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch h.mode {
	case habitConfirmDelete:
		return h, handleConfirmDelete(m.String(), h.deleteSelected, func() { h.mode = habitList })
	case habitForm:
		return h, updateForm(&h.form, m, func() { h.mode = habitList }, h.submit)
	}
	switch m.String() {
	case "up", "k":
		if h.cursor > 0 {
			h.cursor--
		}
	case "down", "j":
		if h.cursor < len(h.habits)-1 {
			h.cursor++
		}
	case " ", "space":
		if len(h.habits) > 0 {
			return h, h.toggleSelected("")
		}
	case "n":
		return h, h.startNew()
	case "e":
		return h, h.startEdit()
	case "d":
		if len(h.habits) > 0 {
			h.mode = habitConfirmDelete
		}
	case "r", "ctrl+r":
		return h, h.refresh()
	}
	return h, nil
}

func (h *Habits) toggleSelected(date string) tea.Cmd {
	if h.cursor >= len(h.habits) {
		return nil
	}
	id := h.habits[h.cursor].HabitID
	c := h.client
	return func() tea.Msg {
		_, err := c.ToggleHabit(id, date)
		return habitsMutatedMsg{err: err}
	}
}

func (h *Habits) startNew() tea.Cmd {
	h.formIn = habitFormState{frequency: "daily", timeOfDay: "anytime"}
	h.form = h.newForm("New habit")
	h.mode = habitForm
	return h.form.Init()
}

func (h *Habits) startEdit() tea.Cmd {
	if h.cursor >= len(h.habits) {
		return nil
	}
	x := h.habits[h.cursor]
	h.formIn = habitFormState{
		id: x.HabitID, name: x.Name, desc: x.Description,
		frequency: x.Frequency, notifyTime: x.NotifyTime, timeOfDay: x.TimeOfDay,
	}
	h.form = h.newForm("Edit habit")
	h.mode = habitForm
	return h.form.Init()
}

func (h *Habits) newForm(title string) *huh.Form {
	d := &h.formIn
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Name").Value(&d.name).Validate(notEmpty),
		huh.NewText().Title("Description").Value(&d.desc).Lines(3),
		huh.NewSelect[string]().Title("Frequency").Options(
			huh.NewOption("Daily", "daily"),
			huh.NewOption("Weekly", "weekly"),
		).Value(&d.frequency),
		huh.NewSelect[string]().Title("Time of day").Options(
			huh.NewOption("Anytime", "anytime"),
			huh.NewOption("Morning", "morning"),
			huh.NewOption("Afternoon", "afternoon"),
			huh.NewOption("Evening", "evening"),
		).Value(&d.timeOfDay),
		huh.NewInput().Title("Notify time (HH:MM UTC, optional)").Value(&d.notifyTime),
	)).WithKeyMap(components.FormKeyMap())
}

func (h *Habits) submit() tea.Cmd {
	d := h.formIn
	in := api.HabitInput{
		Name:        strings.TrimSpace(d.name),
		Description: d.desc,
		Frequency:   d.frequency,
		TimeOfDay:   d.timeOfDay,
		NotifyTime:  strings.TrimSpace(d.notifyTime),
	}
	id := d.id
	c := h.client
	h.form = nil
	h.mode = habitList
	return func() tea.Msg {
		var err error
		if id == "" {
			_, err = c.CreateHabit(in)
		} else {
			_, err = c.UpdateHabit(id, in)
		}
		return habitsMutatedMsg{err: err}
	}
}

func (h *Habits) deleteSelected() tea.Cmd {
	if h.cursor >= len(h.habits) {
		return nil
	}
	id := h.habits[h.cursor].HabitID
	c := h.client
	return func() tea.Msg { return habitsMutatedMsg{err: c.DeleteHabit(id)} }
}

func (h *Habits) View() string {
	if h.mode == habitForm {
		if h.form != nil {
			return lipgloss.JoinVertical(lipgloss.Left, h.form.View(), "", components.FormHint())
		}
	}
	if h.mode == habitConfirmDelete {
		return components.ConfirmView("Delete this habit?", h.width, h.height)
	}
	if h.loading && len(h.habits) == 0 {
		return styles.MutedText.Render("Loading habits...")
	}
	if len(h.habits) == 0 {
		return styles.MutedText.Render("No habits yet. Press n to create one.")
	}
	// Window the habit list so the cursor stays visible. Each rendered
	// habit is 4 lines (top border + name + history + bottom border);
	// we slice into a window that follows the cursor and reserve two
	// lines for "N more above/below" hints.
	start, end := habitWindow(h.cursor, len(h.habits), h.height)
	rows := []string{}
	if start > 0 {
		rows = append(rows, styles.MutedText.Render(fmt.Sprintf("↑ %d more above", start)))
	}
	for i := start; i < end; i++ {
		x := h.habits[i]
		head := fmt.Sprintf("%-30s  streak %d  best %d", truncate(x.Name, 30), x.CurrentStreak, x.BestStreak)
		hist := renderHistory(x.History)
		row := head + "\n" + styles.MutedText.Render(hist)
		if i == h.cursor {
			row = styles.BoxFocused.Render(row)
		} else {
			row = styles.Box.Render(row)
		}
		rows = append(rows, row)
	}
	if end < len(h.habits) {
		rows = append(rows, styles.MutedText.Render(fmt.Sprintf("↓ %d more below", len(h.habits)-end)))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// habitWindow returns [start, end) into a habits slice such that cursor
// is visible given the available terminal height. Each habit renders as
// a 4-line box. Two lines are reserved for the "more above/below" hints
// so the window doesn't push them off-screen on edge resizes.
func habitWindow(cursor, total, height int) (int, int) {
	const linesPerHabit = 4
	available := height - 2
	if available < linesPerHabit {
		available = linesPerHabit
	}
	maxVisible := available / linesPerHabit
	if maxVisible < 1 {
		maxVisible = 1
	}
	if maxVisible >= total {
		return 0, total
	}
	start := 0
	if cursor >= maxVisible {
		start = cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return start, end
}

func renderHistory(history []api.HabitHistory) string {
	// Show the last 30 days. Map by date for lookup.
	byDate := map[string]bool{}
	for _, h := range history {
		byDate[h.Date] = h.Done
	}
	out := strings.Builder{}
	for i := 29; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if byDate[d] {
			out.WriteString("■ ")
		} else {
			out.WriteString("· ")
		}
	}
	return out.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func (h *Habits) Title() string { return "Habits" }
func (h *Habits) StatusHints() []string {
	return []string{
		styles.KeyHint("space", "toggle today"),
		styles.KeyHint("n", "new"),
		styles.KeyHint("e", "edit"),
		styles.KeyHint("d", "delete"),
	}
}
func (h *Habits) Help() []components.HelpEntry {
	return []components.HelpEntry{
		{Keys: "↑/↓ or j/k", Desc: "select habit"},
		{Keys: "space", Desc: "toggle today's completion"},
		{Keys: "n", Desc: "new habit"},
		{Keys: "e", Desc: "edit habit"},
		{Keys: "d", Desc: "delete habit"},
		{Keys: "r", Desc: "refresh"},
	}
}
func (h *Habits) SetSize(w, ht int) { h.width, h.height = w, ht }

// IsTextEditing reports that a form is active.
func (h *Habits) IsTextEditing() bool { return h.mode == habitForm }

// OnEscape: habits has no sub-modes beyond form/confirm — pop them.
func (h *Habits) OnEscape() bool {
	switch h.mode {
	case habitForm:
		h.mode = habitList
		h.form = nil
		return true
	case habitConfirmDelete:
		h.mode = habitList
		return true
	}
	return false
}
