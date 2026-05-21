package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Regression tests for the "delete the row, not the row above/below" bug
// (commit 246bf15). The bug: pressing 'd' on a list set mode = ConfirmDelete
// but the same key fell through to bubbles/table.Update, where 'd' is bound to
// HalfPageDown. The cursor moved before the confirm flow read it, so pressing
// 'y' deleted the wrong row.
//
// The fix captures the selected row's ID at the moment 'd' is pressed and
// returns early so the table never sees the key. These tests assert each
// fixed screen captures the correct ID on row 0 and that pressing 'y' issues
// DELETE for that ID — independent of any subsequent cursor movement.

func pressRune(t *testing.T, m tea.Model, r rune) tea.Model {
	t.Helper()
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return drainCmd(t, out, cmd)
}

func TestTasksDeleteCapturesRowZero(t *testing.T) {
	fb := newFakeBackend(t)
	m := runScreen(t, newTasks(fb.client()))
	tasks := m.(*Tasks)
	if len(tasks.filtered) < 2 {
		t.Fatalf("need >=2 tasks to test cursor regression; got %d", len(tasks.filtered))
	}
	want := tasks.filtered[0].TaskID

	m = pressRune(t, m, 'd')
	tasks = m.(*Tasks)
	if tasks.mode != taskConfirmDelete {
		t.Fatalf("after 'd', expected taskConfirmDelete; got %d", tasks.mode)
	}
	if tasks.pendingDeleteID != want {
		t.Errorf("pendingDeleteID = %q, want %q", tasks.pendingDeleteID, want)
	}

	pressRune(t, m, 'y')
	if !fb.saw("DELETE", "/tasks/"+want) {
		t.Errorf("expected DELETE /tasks/%s; saw %v", want, fb.calls)
	}
}

func TestGoalsDeleteCapturesRowZero(t *testing.T) {
	fb := newFakeBackend(t)
	m := runScreen(t, newGoals(fb.client()))
	g := m.(*Goals)
	if len(g.view) < 2 {
		t.Fatalf("need >=2 goals to test cursor regression; got %d", len(g.view))
	}
	want := g.view[0].GoalID

	m = pressRune(t, m, 'd')
	g = m.(*Goals)
	if g.mode != goalConfirmDelete {
		t.Fatalf("after 'd', expected goalConfirmDelete; got %d", g.mode)
	}
	if g.pendingDeleteID != want {
		t.Errorf("pendingDeleteID = %q, want %q", g.pendingDeleteID, want)
	}

	pressRune(t, m, 'y')
	if !fb.saw("DELETE", "/goals/"+want) {
		t.Errorf("expected DELETE /goals/%s; saw %v", want, fb.calls)
	}
}

func TestBookmarksDeleteCapturesRowZero(t *testing.T) {
	fb := newFakeBackend(t)
	m := runScreen(t, newBookmarks(fb.client()))
	b := m.(*Bookmarks)
	if len(b.items) < 2 {
		t.Fatalf("need >=2 bookmarks; got %d", len(b.items))
	}
	want := b.items[0].BookmarkID

	m = pressRune(t, m, 'd')
	b = m.(*Bookmarks)
	if b.mode != bookmarkConfirmDelete {
		t.Fatalf("after 'd', expected bookmarkConfirmDelete; got %d", b.mode)
	}
	if b.pendingDeleteID != want {
		t.Errorf("pendingDeleteID = %q, want %q", b.pendingDeleteID, want)
	}

	pressRune(t, m, 'y')
	if !fb.saw("DELETE", "/bookmarks/"+want) {
		t.Errorf("expected DELETE /bookmarks/%s; saw %v", want, fb.calls)
	}
}

func TestFavoritesDeleteCapturesRowZero(t *testing.T) {
	fb := newFakeBackend(t)
	m := runScreen(t, newFavorites(fb.client()))
	f := m.(*Favorites)
	if len(f.items) < 2 {
		t.Fatalf("need >=2 favorites; got %d", len(f.items))
	}
	want := f.items[0].FavoriteID

	m = pressRune(t, m, 'd')
	f = m.(*Favorites)
	if !f.confirm {
		t.Fatalf("after 'd', expected confirm=true")
	}
	if f.pendingDeleteID != want {
		t.Errorf("pendingDeleteID = %q, want %q", f.pendingDeleteID, want)
	}

	pressRune(t, m, 'y')
	if !fb.saw("DELETE", "/favorites/"+want) {
		t.Errorf("expected DELETE /favorites/%s; saw %v", want, fb.calls)
	}
}

func TestFeedsDeleteCapturesRowZero(t *testing.T) {
	fb := newFakeBackend(t)
	m := runScreen(t, newFeeds(fb.client()))
	f := m.(*Feeds)
	if len(f.feeds) < 2 {
		t.Fatalf("need >=2 feeds; got %d", len(f.feeds))
	}
	want := f.feeds[0].FeedID

	m = pressRune(t, m, 'd')
	f = m.(*Feeds)
	if f.mode != feedsConfirmDeleteFeed {
		t.Fatalf("after 'd', expected feedsConfirmDeleteFeed; got %d", f.mode)
	}
	if f.pendingDeleteID != want {
		t.Errorf("pendingDeleteID = %q, want %q", f.pendingDeleteID, want)
	}

	pressRune(t, m, 'y')
	if !fb.saw("DELETE", "/feeds/"+want) {
		t.Errorf("expected DELETE /feeds/%s; saw %v", want, fb.calls)
	}
}

func TestFinancesDeleteCapturesRowZero(t *testing.T) {
	fb := newFakeBackend(t)
	m := runScreen(t, newFinances(fb.client()))
	f := m.(*Finances)
	if len(f.debts) < 2 {
		t.Fatalf("need >=2 debts; got %d", len(f.debts))
	}
	want := f.debts[0].DebtID

	m = pressRune(t, m, 'd')
	f = m.(*Finances)
	if f.mode != financeConfirmDelete {
		t.Fatalf("after 'd', expected financeConfirmDelete; got %d", f.mode)
	}
	if f.pendingDeleteID != want {
		t.Errorf("pendingDeleteID = %q, want %q", f.pendingDeleteID, want)
	}
	if f.pendingDeleteTab != tabDebts {
		t.Errorf("pendingDeleteTab = %d, want tabDebts(0)", f.pendingDeleteTab)
	}

	pressRune(t, m, 'y')
	if !fb.saw("DELETE", "/debts/"+want) {
		t.Errorf("expected DELETE /debts/%s; saw %v", want, fb.calls)
	}
}

// TestTasksDeleteCancelClearsPending: pressing 'd' then 'n' should reset
// state cleanly so a subsequent 'd' captures the (possibly different) row.
func TestTasksDeleteCancelClearsPending(t *testing.T) {
	fb := newFakeBackend(t)
	m := runScreen(t, newTasks(fb.client()))
	m = pressRune(t, m, 'd')
	if m.(*Tasks).pendingDeleteID == "" {
		t.Fatal("pendingDeleteID should be set after 'd'")
	}
	m = pressRune(t, m, 'n')
	tasks := m.(*Tasks)
	if tasks.mode != taskList {
		t.Errorf("after 'n', expected taskList; got %d", tasks.mode)
	}
	if tasks.pendingDeleteID != "" {
		t.Errorf("pendingDeleteID should be cleared after cancel; got %q", tasks.pendingDeleteID)
	}
}
