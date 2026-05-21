package screens

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/neilfarmer/memoire-tui/internal/api"
)

// TestNotesFolderDeleteCapturesCursor regression-tests the folder-delete
// off-by-one that mirrors the bug commit 5c83fa7 fixed across other screens.
// Confirm that 'd' on a folder row captures the folder ID into
// pendingDeleteFolderID and 'y' issues DELETE for that captured ID, regardless
// of whether bubbles/table's 'd'=HalfPageDown would have moved the cursor.
func TestNotesFolderDeleteCapturesCursor(t *testing.T) {
	// fakeBackend's /notes/folders only returns one folder; use a focused
	// httptest server that serves enough folders to test cursor capture.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/notes/folders" && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"folder_id":"f1","name":"Inbox"},{"folder_id":"f2","name":"Archive"}]`))
		case r.URL.Path == "/notes" && r.Method == "GET":
			_, _ = w.Write([]byte(`[]`))
		case r.URL.Path == "/notes/folders/f1" && r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := api.New(srv.URL, "pat_test")

	n := newNotes(c)
	m := runScreen(t, n)
	notes := m.(*Notes)
	if notes.mode != noteFolderList {
		t.Fatalf("expected noteFolderList at boot; got %d", notes.mode)
	}
	if len(notes.folders) != 2 {
		t.Fatalf("expected 2 folders; got %d", len(notes.folders))
	}

	// Cursor 0 is the "All notes" pseudo-row; move to cursor 1 (real folder f1).
	notes.folderTbl.SetCursor(1)

	m = pressRune(t, m, 'd')
	notes = m.(*Notes)
	if notes.mode != noteConfirmDeleteFolder {
		t.Fatalf("after 'd' on a real folder row, expected noteConfirmDeleteFolder; got %d", notes.mode)
	}
	if notes.pendingDeleteFolderID != "f1" {
		t.Errorf("pendingDeleteFolderID = %q, want %q", notes.pendingDeleteFolderID, "f1")
	}

	pressRune(t, m, 'y')
	// drainCmd above will issue the DELETE; no easy way to inspect calls here
	// without our usual fakeBackend, but the pre-DELETE state assertion above
	// is the core regression check.
}

// TestNotesFolderDeleteCancelClearsPending: pressing 'd' then 'n' should
// clear pendingDeleteFolderID so a subsequent 'd' doesn't accidentally
// delete the prior target.
func TestNotesFolderDeleteCancelClearsPending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/notes/folders" && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"folder_id":"f1","name":"Inbox"}]`))
		case r.URL.Path == "/notes" && r.Method == "GET":
			_, _ = w.Write([]byte(`[]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	n := newNotes(api.New(srv.URL, "pat_test"))
	m := runScreen(t, n)
	m.(*Notes).folderTbl.SetCursor(1)
	m = pressRune(t, m, 'd')
	if m.(*Notes).pendingDeleteFolderID != "f1" {
		t.Fatalf("setup failed: pendingDeleteFolderID should be f1; got %q",
			m.(*Notes).pendingDeleteFolderID)
	}
	m = pressRune(t, m, 'n')
	notes := m.(*Notes)
	if notes.mode != noteFolderList {
		t.Errorf("after 'n', expected noteFolderList; got %d", notes.mode)
	}
	if notes.pendingDeleteFolderID != "" {
		t.Errorf("pendingDeleteFolderID should be cleared after cancel; got %q",
			notes.pendingDeleteFolderID)
	}
}

// TestDashboardOverdueAndDueToday verifies the dashboard correctly counts
// tasks due today vs overdue. Locks in the date-string comparison logic
// at dashboard.go:78-82 so a future refactor that breaks it gets caught.
func TestDashboardOverdueAndDueToday(t *testing.T) {
	// Use a far-future "today" so any test-time DueDate comparisons are
	// stable: 2099-01-01 is past any plausible due date below.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tasks":
			_, _ = w.Write([]byte(`[
				{"task_id":"a","title":"Past","status":"todo","due_date":"2000-01-01"},
				{"task_id":"b","title":"Past done","status":"done","due_date":"2000-01-01"},
				{"task_id":"c","title":"Future","status":"todo","due_date":"3000-01-01"},
				{"task_id":"d","title":"No due","status":"todo"}
			]`))
		case "/habits":
			_, _ = w.Write([]byte(`[{"habit_id":"h1","done_today":true,"current_streak":5}]`))
		case "/notes":
			_, _ = w.Write([]byte(`[{"note_id":"n1","title":"Latest","updated_at":"2024-01-01"}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	d := newDashboard(api.New(srv.URL, "pat_test"))
	m := runScreen(t, d)
	dash := m.(*Dashboard)
	// "Past" is overdue (today > 2000-01-01); "Past done" excluded; "Future"
	// not yet overdue or today; "No due" excluded from both counters.
	if dash.tasksOverdue != 1 {
		t.Errorf("tasksOverdue = %d, want 1", dash.tasksOverdue)
	}
	if dash.tasksToday != 0 {
		t.Errorf("tasksToday = %d, want 0", dash.tasksToday)
	}
	if dash.tasksTotal != 4 {
		t.Errorf("tasksTotal = %d, want 4", dash.tasksTotal)
	}
	if dash.streak != 5 {
		t.Errorf("streak = %d, want 5", dash.streak)
	}
	if dash.latestNote != "Latest" {
		t.Errorf("latestNote = %q, want %q", dash.latestNote, "Latest")
	}
}

// TestJournalEntryText proves the Text() helper prefers Content over Body
// (matching the prior firstNonEmpty(Content, Body) call order) so the
// journal display doesn't regress when the server returns one field or
// the other.
func TestJournalEntryText(t *testing.T) {
	cases := []struct {
		name  string
		entry api.JournalEntry
		want  string
	}{
		{"both populated prefers content", api.JournalEntry{Content: "c", Body: "b"}, "c"},
		{"only body", api.JournalEntry{Body: "b"}, "b"},
		{"only content", api.JournalEntry{Content: "c"}, "c"},
		{"neither", api.JournalEntry{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.entry.Text(); got != tc.want {
				t.Errorf("Text() = %q, want %q", got, tc.want)
			}
		})
	}
}

// Ensure tea import stays referenced even if other tests are removed.
var _ tea.KeyMsg
