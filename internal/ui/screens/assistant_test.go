package screens

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/neilfarmer/memoire-tui/internal/api"
)

// TestAssistantModelToggle: pressing `m` while focus is outside the
// input textarea cycles the model. Regression for the ctrl+m binding
// which never fired because ctrl+m and Enter are the same byte (0x0D)
// in every terminal.
func TestAssistantModelToggle(t *testing.T) {
	a := newAssistant(api.New("https://example.invalid", "pat_test"))
	a.SetSize(120, 40)
	if a.model != "nova-lite" {
		t.Fatalf("default model: got %q, want nova-lite", a.model)
	}

	// In the input pane (default), 'm' is a literal character — must
	// not toggle the model.
	if a.pane != assistantPaneInput {
		t.Fatalf("expected boot pane to be input; got %d", a.pane)
	}
	a.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if a.model != "nova-lite" {
		t.Errorf("'m' in input pane should not toggle model; got %q", a.model)
	}

	// Tab once to move focus to convos pane.
	a.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	if a.pane != assistantPaneConvos {
		t.Fatalf("expected convos pane after tab; got %d", a.pane)
	}
	// Now 'm' should toggle.
	a.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if a.model != "nova-pro" {
		t.Errorf("after 'm' in convos pane, model should be nova-pro; got %q", a.model)
	}
	a.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if a.model != "nova-lite" {
		t.Errorf("second 'm' should cycle back to nova-lite; got %q", a.model)
	}
}

// TestRelativeTime covers the formatter used in the conversation list.
func TestRelativeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"garbage", "not-a-date", ""},
		{"just now", now.Add(-10 * time.Second).Format(time.RFC3339), "now"},
		{"minutes", now.Add(-5 * time.Minute).Format(time.RFC3339), "5m"},
		{"hours", now.Add(-3 * time.Hour).Format(time.RFC3339), "3h"},
		{"days", now.Add(-2 * 24 * time.Hour).Format(time.RFC3339), "2d"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := relativeTime(tc.in); got != tc.want {
				t.Errorf("relativeTime(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
