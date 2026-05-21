package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/neilfarmer/memoire-tui/internal/api"
)

func TestAppRendersAfterResize(t *testing.T) {
	client := api.New("https://example.com", "pat_test")
	app := New(client, DefaultFactories(client))
	// Init runs before first render in real usage, so call it.
	_ = app.Init()
	model, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	out := model.View()
	if out == "" {
		t.Fatal("empty view")
	}
	if !strings.Contains(out, "memoire") {
		t.Errorf("expected app name in header, got: %q", out[:min(200, len(out))])
	}
	if !strings.Contains(out, "Dashboard") {
		t.Errorf("expected sidebar entry Dashboard, got: %q", out[:min(200, len(out))])
	}
}

func TestArrowDownCyclesScreens(t *testing.T) {
	client := api.New("https://example.com", "pat_test")
	app := New(client, DefaultFactories(client))
	_ = app.Init()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	// Sidebar focused at boot. Down arrow moves to Tasks (preview-activate).
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyDown})
	a := model.(*App)
	if a.current != ScreenTasks {
		t.Errorf("expected current=tasks after ↓, got %s", a.current)
	}
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyUp})
	a = model.(*App)
	if a.current != ScreenDashboard {
		t.Errorf("expected current=dashboard after ↑, got %s", a.current)
	}
}

func TestHelpToggle(t *testing.T) {
	client := api.New("https://example.com", "pat_test")
	app := New(client, DefaultFactories(client))
	_ = app.Init()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view := model.View()
	if !strings.Contains(view, "Help") {
		t.Errorf("expected Help overlay, got: %q", view[:min(200, len(view))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestEscapeFromAssistantReturnsToSidebar regression-tests the user-reported
// bug: once inside the Assistant screen, the chat textarea is focused and
// IsTextEditing returns true. The old esc handler short-circuited on
// IsTextEditing and forwarded esc to the screen, which then forwarded it
// to the textarea, which silently swallowed it — so the user was stuck
// inside the screen with no way out.
//
// Asserts: after pressing esc inside Assistant, sidebar focus is restored
// and a subsequent esc triggers the quit-confirm overlay (the documented
// next step).
func TestEscapeFromAssistantReturnsToSidebar(t *testing.T) {
	client := api.New("https://example.com", "pat_test")
	app := New(client, DefaultFactories(client))
	_ = app.Init()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Walk down the sidebar until we land on Assistant. Cap iterations at
	// the sidebar length so an unreachable target doesn't hang the test.
	a := app
	for i := 0; i < len(SidebarOrder) && a.current != ScreenAssistant; i++ {
		m, _ := a.Update(tea.KeyMsg{Type: tea.KeyDown})
		a = m.(*App)
	}
	if a.current != ScreenAssistant {
		t.Fatalf("Assistant not reachable from sidebar; landed on %s", a.current)
	}
	// Drill into the screen (sidebar focus → content focus).
	m, _ := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	a = m.(*App)
	if a.sideFocus {
		t.Fatal("expected sideFocus=false after enter from sidebar")
	}

	// Press esc — the bug was this doing nothing. Expectation: focus
	// returns to sidebar.
	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyEscape})
	a = m.(*App)
	if !a.sideFocus {
		t.Errorf("expected esc to return focus to sidebar, but sideFocus=%v", a.sideFocus)
	}

	// One more esc should now trigger quit-confirm, proving we really
	// did pop back to the sidebar and didn't just no-op.
	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyEscape})
	a = m.(*App)
	if !a.quitConfirm {
		t.Errorf("expected quit confirm after second esc from sidebar; got quitConfirm=%v", a.quitConfirm)
	}
}

func TestIsTerminalNoiseKey(t *testing.T) {
	cases := []struct {
		in   string
		want bool
		why  string
	}{
		{"", false, "empty"},
		{"[", false, "bare left bracket — user typed it"},
		{"]", false, "bare right bracket — user typed it"},
		{"a", false, "letter"},
		{"ctrl+s", false, "modifier combo"},
		{"alt+]", false, "real alt-bracket combo"},
		{"]11;rgb:1c1c/1c1c/1c1c", true, "OSC 11 background color reply"},
		{"]10;rgb:ffff/ffff/ffff", true, "OSC 10 foreground reply"},
		{"[2~", true, "CSI sequence"},
		{"foo;rgb:bar", true, "embedded rgb payload"},
	}
	for _, c := range cases {
		if got := isTerminalNoiseKey(c.in); got != c.want {
			t.Errorf("isTerminalNoiseKey(%q) = %v, want %v (%s)", c.in, got, c.want, c.why)
		}
	}
}
