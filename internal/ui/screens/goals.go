package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/api"
	"github.com/neilfarmer/memoire-tui/internal/styles"
	"github.com/neilfarmer/memoire-tui/internal/ui/components"
)

type goalMode int

const (
	goalList goalMode = iota
	goalDetail
	goalForm
	goalConfirmDelete
)

type Goals struct {
	client *api.Client
	width  int
	height int

	mode    goalMode
	loading bool
	err     error
	goals   []api.Goal
	view    []api.Goal
	filter  string
	tbl     components.Model
	form    *huh.Form
	formIn  goalFormState

	pendingDeleteID string
}

type goalFormState struct {
	id          string
	title       string
	description string
	category    string
	status      string
	deadline    string
	progress    string
}

type goalsLoadedMsg struct {
	goals []api.Goal
	err   error
}
type goalsMutatedMsg struct{ err error }

func newGoals(c *api.Client) *Goals {
	g := &Goals{client: c, filter: "all"}
	g.tbl = components.NewTable(goalCols(80), nil, 18)
	return g
}

func goalCols(w int) []components.Column {
	titleW := w - 14 - 12 - 12 - 14
	if titleW < 16 {
		titleW = 16
	}
	return []components.Column{
		{Title: "STATUS", Width: 12},
		{Title: "PROGRESS", Width: 14},
		{Title: "DEADLINE", Width: 12},
		{Title: "CATEGORY", Width: 14},
		{Title: "TITLE", Width: titleW},
	}
}

func (g *Goals) Init() tea.Cmd { return g.refresh() }

func (g *Goals) refresh() tea.Cmd {
	c := g.client
	return func() tea.Msg {
		out, err := c.ListGoals()
		return goalsLoadedMsg{goals: out, err: err}
	}
}

func (g *Goals) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case goalsLoadedMsg:
		g.loading = false
		g.err = m.err
		g.goals = m.goals
		g.refilter()
		return g, nil
	case goalsMutatedMsg:
		g.err = m.err
		g.mode = goalList
		return g, g.refresh()
	case tea.KeyMsg:
		return g.handleKey(m)
	}
	if g.mode == goalForm && g.form != nil {
		return g, updateForm(&g.form, msg, func() { g.mode = goalList }, g.submit)
	}
	if g.mode == goalList {
		var cmd tea.Cmd
		g.tbl, cmd = g.tbl.Update(msg)
		return g, cmd
	}
	return g, nil
}

func (g *Goals) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch g.mode {
	case goalConfirmDelete:
		return g, handleConfirmDelete(m.String(), g.deleteSelected, func() {
			g.mode = goalList
			g.pendingDeleteID = ""
		})
	case goalDetail:
		switch m.String() {
		case "esc":
			g.mode = goalList
		case "e":
			return g, g.startEdit()
		case "d":
			if idx := g.tbl.Cursor(); idx >= 0 && idx < len(g.view) {
				g.pendingDeleteID = g.view[idx].GoalID
				g.mode = goalConfirmDelete
			}
		}
		return g, nil
	case goalForm:
		return g, updateForm(&g.form, m, func() { g.mode = goalList }, g.submit)
	}
	switch m.String() {
	case "enter":
		if g.tbl.Cursor() < len(g.view) {
			g.mode = goalDetail
		}
		return g, nil
	case "n":
		return g, g.startNew()
	case "e":
		return g, g.startEdit()
	case "d":
		if idx := g.tbl.Cursor(); idx >= 0 && idx < len(g.view) {
			g.pendingDeleteID = g.view[idx].GoalID
			g.mode = goalConfirmDelete
		}
		return g, nil
	case "f":
		g.filter = nextGoalFilter(g.filter)
		g.refilter()
	case "r", "ctrl+r":
		return g, g.refresh()
	}
	var cmd tea.Cmd
	g.tbl, cmd = g.tbl.Update(m)
	return g, cmd
}

func nextGoalFilter(cur string) string {
	o := []string{"all", "active", "completed", "abandoned"}
	for i, x := range o {
		if x == cur {
			return o[(i+1)%len(o)]
		}
	}
	return "all"
}

func (g *Goals) refilter() {
	view := make([]api.Goal, 0, len(g.goals))
	for _, x := range g.goals {
		if g.filter != "all" && x.Status != g.filter {
			continue
		}
		view = append(view, x)
	}
	g.view = view
	rows := make([]components.Row, 0, len(view))
	for _, x := range view {
		deadline := orDash(x.Deadline)
		if deadline == "—" {
			deadline = orDash(x.TargetDate)
		}
		progress := fmt.Sprintf("%d%%  %s", x.Progress, components.Bar(x.Progress, 6))
		rows = append(rows, components.Row{
			orDash(x.Status),
			progress,
			deadline,
			truncate(orDash(x.Category), 14),
			x.Title,
		})
	}
	g.tbl.SetRows(rows)
}

func (g *Goals) startNew() tea.Cmd {
	g.formIn = goalFormState{status: "active"}
	g.form = g.newForm()
	g.mode = goalForm
	return g.form.Init()
}

func (g *Goals) startEdit() tea.Cmd {
	idx := g.tbl.Cursor()
	if idx >= len(g.view) {
		return nil
	}
	x := g.view[idx]
	g.formIn = goalFormState{
		id: x.GoalID, title: x.Title, description: x.Description,
		category: x.Category, status: x.Status,
		deadline: x.Deadline, progress: fmt.Sprintf("%d", x.Progress),
	}
	if g.formIn.deadline == "" && x.TargetDate != "" {
		g.formIn.deadline = x.TargetDate
	}
	g.form = g.newForm()
	g.mode = goalForm
	return g.form.Init()
}

func (g *Goals) newForm() *huh.Form {
	d := &g.formIn
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Title").Value(&d.title).Validate(notEmpty),
		huh.NewText().Title("Description").Value(&d.description).Lines(3),
		huh.NewInput().Title("Category").Value(&d.category),
		huh.NewSelect[string]().Title("Status").Options(
			huh.NewOption("Active", "active"),
			huh.NewOption("Completed", "completed"),
			huh.NewOption("Abandoned", "abandoned"),
		).Value(&d.status),
		huh.NewInput().Title("Deadline (YYYY-MM-DD)").Value(&d.deadline).Validate(validateOptionalDate),
		huh.NewInput().Title("Progress (0-100)").Value(&d.progress).Validate(validateOptionalInt),
	)).WithKeyMap(components.FormKeyMap())
}

func (g *Goals) submit() tea.Cmd {
	d := g.formIn
	progress, _ := parseInt(d.progress)
	in := api.GoalInput{
		Title:       d.title,
		Description: d.description,
		Category:    d.category,
		Status:      d.status,
		Deadline:    d.deadline,
		TargetDate:  d.deadline,
		Progress:    progress,
	}
	id := d.id
	c := g.client
	g.form = nil
	g.mode = goalList
	return func() tea.Msg {
		var err error
		if id == "" {
			_, err = c.CreateGoal(in)
		} else {
			_, err = c.UpdateGoal(id, in)
		}
		return goalsMutatedMsg{err: err}
	}
}

func (g *Goals) deleteSelected() tea.Cmd {
	id := g.pendingDeleteID
	g.pendingDeleteID = ""
	if id == "" {
		return nil
	}
	c := g.client
	return func() tea.Msg { return goalsMutatedMsg{err: c.DeleteGoal(id)} }
}

func (g *Goals) View() string {
	switch g.mode {
	case goalForm:
		if g.form != nil {
			return lipgloss.JoinVertical(lipgloss.Left, g.form.View(), "", components.FormHint())
		}
	case goalConfirmDelete:
		return components.ConfirmView("Delete this goal?", g.width, g.height)
	case goalDetail:
		return g.detailView()
	}
	if g.loading && len(g.goals) == 0 {
		return styles.MutedText.Render("Loading goals...")
	}
	header := renderPills(g.filter, []string{"all", "active", "completed", "abandoned"})
	g.tbl.SetColumns(goalCols(g.width - 6))
	if g.height-8 > 0 {
		g.tbl.SetHeight(g.height - 8)
	}
	hints := []string{
		styles.KeyHint("↵", "details"),
		styles.KeyHint("n", "new"),
		styles.KeyHint("e", "edit"),
		styles.KeyHint("d", "delete"),
		styles.KeyHint("f", "filter"),
	}
	body := components.FrameTable("Goals", len(g.view), g.tbl, hints, true)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func (g *Goals) detailView() string {
	idx := g.tbl.Cursor()
	if idx < 0 || idx >= len(g.view) {
		return ""
	}
	x := g.view[idx]
	rows := []string{
		components.Crumbs("Goals", x.Title),
		"",
		styles.Title.Render(x.Title),
		"",
		"Status:    " + styles.StatusColor(x.Status).Render(orDash(x.Status)),
		"Category:  " + orDash(x.Category),
		fmt.Sprintf("Progress:  %d%%  %s", x.Progress, components.Bar(x.Progress, 20)),
	}
	if x.Deadline != "" {
		rows = append(rows, "Deadline:  "+x.Deadline)
	}
	if x.TargetDate != "" {
		rows = append(rows, "Target:    "+x.TargetDate)
	}
	if x.Description != "" {
		rows = append(rows, "", components.RenderMarkdown(x.Description, g.width-6))
	}
	rows = append(rows, "", styles.MutedText.Render("<e> edit  <d> delete  <esc> back"))
	return styles.Box.Render(strings.Join(rows, "\n"))
}

func (g *Goals) Title() string { return "Goals" }
func (g *Goals) StatusHints() []string {
	return []string{
		styles.KeyHint("n", "new"),
		styles.KeyHint("↵", "open"),
		styles.KeyHint("e", "edit"),
		styles.KeyHint("d", "delete"),
		styles.KeyHint("f", "filter"),
	}
}
func (g *Goals) Help() []components.HelpEntry {
	return []components.HelpEntry{
		{Keys: "↑/↓", Desc: "select row"},
		{Keys: "↵ enter", Desc: "open detail"},
		{Keys: "esc", Desc: "back"},
		{Keys: "n", Desc: "new goal"},
		{Keys: "e", Desc: "edit"},
		{Keys: "d", Desc: "delete"},
		{Keys: "f", Desc: "cycle filter"},
		{Keys: "r", Desc: "refresh"},
	}
}
func (g *Goals) SetSize(w, h int) {
	g.width, g.height = w, h
	g.tbl.SetColumns(goalCols(w - 6))
	if h-8 > 0 {
		g.tbl.SetHeight(h - 8)
	}
}

// IsTextEditing reports that a form is active.
func (g *Goals) IsTextEditing() bool { return g.mode == goalForm }

func (g *Goals) OnEscape() bool {
	switch g.mode {
	case goalDetail, goalConfirmDelete:
		g.mode = goalList
		g.pendingDeleteID = ""
		return true
	case goalForm:
		g.mode = goalList
		g.form = nil
		return true
	}
	return false
}
