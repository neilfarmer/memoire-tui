package screens

import (
	"testing"

	"github.com/neilfarmer/memoire-tui/internal/ui/components"
)

// TestNotesEditorClosedRebuildsForm: when $EDITOR returns with new body
// content, the form must be rebuilt so huh's text widget re-reads from
// the formData.body binding pointer. Without the rebuild, huh's internal
// textarea buffer (frozen at the pre-editor state) overwrites
// formData.body on the next keypress and the user's editor work is lost.
//
// This test asserts (a) formData.body now matches the editor content and
// (b) the *huh.Form pointer changed, proving newForm() was called and the
// stale textarea buffer is gone.
func TestNotesEditorClosedRebuildsForm(t *testing.T) {
	fb := newFakeBackend(t)
	n := newNotes(fb.client())
	m := runScreen(t, n)
	notes := m.(*Notes)

	// Put the screen into an edit form (mode=noteForm with a non-nil form).
	notes.formData = noteFormState{id: "n1", title: "T", body: "old body"}
	notes.form = notes.newForm("Edit note")
	notes.mode = noteForm
	formBefore := notes.form

	out, _ := notes.Update(components.EditorClosedMsg{Content: "from editor"})
	after := out.(*Notes)

	if after.formData.body != "from editor" {
		t.Errorf("formData.body = %q, want %q", after.formData.body, "from editor")
	}
	if after.form == nil {
		t.Fatal("form should be rebuilt, not nil")
	}
	if after.form == formBefore {
		t.Error("form pointer unchanged — huh's stale textarea buffer will overwrite the editor content on the next keypress")
	}
}

// TestJournalEditorClosedRebuildsForm: same regression coverage for the
// journal screen, which also exposes ctrl+e in its edit form.
func TestJournalEditorClosedRebuildsForm(t *testing.T) {
	fb := newFakeBackend(t)
	j := newJournal(fb.client())
	m := runScreen(t, j)
	journal := m.(*Journal)

	journal.formData = journalFormState{title: "T", body: "old body"}
	journal.form = journal.newForm()
	journal.mode = journalForm
	formBefore := journal.form

	out, _ := journal.Update(components.EditorClosedMsg{Content: "from editor"})
	after := out.(*Journal)

	if after.formData.body != "from editor" {
		t.Errorf("formData.body = %q, want %q", after.formData.body, "from editor")
	}
	if after.form == formBefore {
		t.Error("form pointer unchanged — editor content will be overwritten on next keypress")
	}
}
