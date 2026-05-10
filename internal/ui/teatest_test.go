package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/neilfarmer/memoire-tui/internal/api"
)

// fakeServer mimics the memoire API for end-to-end UI tests.
func fakeServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/tasks" && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"task_id":"t1","title":"Real task","status":"todo","priority":"high"}]`))
		case r.URL.Path == "/tasks/t1" && r.Method == "PUT":
			_, _ = w.Write([]byte(`{"task_id":"t1"}`))
		case r.URL.Path == "/notes" && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"note_id":"n1","title":"Real note"}]`))
		case r.URL.Path == "/notes/folders":
			_, _ = w.Write([]byte(`[]`))
		case r.URL.Path == "/notes/n1" && r.Method == "GET":
			_, _ = w.Write([]byte(`{"note_id":"n1","title":"Real note","body":"hi"}`))
		case r.URL.Path == "/notes/n1" && r.Method == "PUT":
			_, _ = w.Write([]byte(`{"note_id":"n1"}`))
		case r.URL.Path == "/habits" && r.Method == "GET":
			_, _ = w.Write([]byte(`[]`))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestEndToEnd_TasksLoadsAndShowsRow boots the App via teatest, drives real
// tea program loop, asserts the tasks list renders the row from the API.
func TestEndToEnd_TasksLoadsAndShowsRow(t *testing.T) {
	srv := fakeServer(t)
	client := api.New(srv.URL, "pat_test")
	app := New(client, DefaultFactories(client))

	tm := teatest.NewTestModel(t, app, teatest.WithInitialTermSize(120, 40))

	// Wait for dashboard to render, then jump to Tasks via ctrl+n.
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Dashboard")
	}, teatest.WithCheckInterval(time.Millisecond*50), teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Real task")
	}, teatest.WithCheckInterval(time.Millisecond*50), teatest.WithDuration(2*time.Second))

	tm.Send(tea.Quit())
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// TestEndToEnd_NotesEditOpensForm verifies that pressing 'e' on the Notes
// list opens the edit form populated with the note body fetched from the
// API. Reproduces the bug we just fixed: noteEditPrepMsg unhandled.
func TestEndToEnd_NotesEditOpensForm(t *testing.T) {
	srv := fakeServer(t)
	client := api.New(srv.URL, "pat_test")
	app := New(client, DefaultFactories(client))

	tm := teatest.NewTestModel(t, app, teatest.WithInitialTermSize(120, 40))

	// Sidebar is focused at boot. Down twice puts cursor on Notes.
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Real note")
	}, teatest.WithCheckInterval(time.Millisecond*50), teatest.WithDuration(2*time.Second))

	// Enter drills into content; then press 'e' on the list.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Title") &&
			strings.Contains(string(out), "<ctrl+s> save")
	}, teatest.WithCheckInterval(time.Millisecond*50), teatest.WithDuration(3*time.Second))

	// Cancel + quit to clean up.
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})
	tm.Send(tea.Quit())
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// TestEndToEnd_OSCResponseDoesNotFreeze sends synthetic OSC-like key events
// and verifies the program filters them rather than entering a feedback loop.
func TestEndToEnd_OSCResponseDoesNotFreeze(t *testing.T) {
	srv := fakeServer(t)
	client := api.New(srv.URL, "pat_test")
	app := New(client, DefaultFactories(client))
	tm := teatest.NewTestModel(t, app, teatest.WithInitialTermSize(120, 40))

	// Spam OSC-shaped key strings. The filter should drop them.
	for i := 0; i < 50; i++ {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]11;rgb:2828/2c2c/3434")})
	}
	tm.Send(tea.Quit())
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// TestEndToEnd_QuestionMarkInAssistantDoesNotOpenHelp verifies the fix for
// issue #4: pressing '?' inside the assistant input must reach the textarea
// rather than toggle the global help overlay.
func TestEndToEnd_QuestionMarkInAssistantDoesNotOpenHelp(t *testing.T) {
	srv := fakeServer(t)
	client := api.New(srv.URL, "pat_test")
	app := New(client, DefaultFactories(client))

	// Pre-activate the assistant screen so its IsTextEditing reports true.
	app.sideCursor = 0
	app.activate(ScreenAssistant)
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Send '?' directly to the App. With the fix, the global handler must
	// short-circuit and the help overlay must NOT open.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if app.helpOpen {
		t.Errorf("help overlay should not toggle while assistant input is focused")
	}
}

// TestEscDrillsUpThenQuits verifies the new esc semantics: from a sub-mode,
// esc returns to the list; from the list (top), esc opens the quit confirm.
func TestEscDrillsUpThenQuits(t *testing.T) {
	srv := fakeServer(t)
	client := api.New(srv.URL, "pat_test")
	app := New(client, DefaultFactories(client))
	app.activate(ScreenTasks)
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Drill into content (Tasks list).
	app.sideFocus = false
	// Press 'n' to open the new-task form.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// First esc closes the form. Should NOT open quit confirm.
	app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if app.quitConfirm {
		t.Fatalf("first esc should drill up, not open quit confirm")
	}

	// Second esc returns focus to sidebar. Still no quit confirm.
	app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if app.quitConfirm {
		t.Fatalf("second esc should return to sidebar, not quit confirm")
	}
	if !app.sideFocus {
		t.Fatalf("second esc should restore sidebar focus")
	}

	// Third esc on sidebar opens the quit confirm.
	app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if !app.quitConfirm {
		t.Errorf("esc at sidebar should open quit confirm")
	}

	// Pressing 'n' inside the quit confirm cancels it.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if app.quitConfirm {
		t.Errorf("quit confirm should dismiss on 'n'")
	}
}

// TestColonOpensPalette verifies ':' opens the command palette overlay.
func TestColonOpensPalette(t *testing.T) {
	srv := fakeServer(t)
	client := api.New(srv.URL, "pat_test")
	app := New(client, DefaultFactories(client))
	app.activate(ScreenDashboard)
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	if !app.paletteOpen {
		t.Fatalf("':' should open the command palette")
	}
	// Esc closes it.
	app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if app.paletteOpen {
		t.Errorf("esc inside palette should close it")
	}
}
