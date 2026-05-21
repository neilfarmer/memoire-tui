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
