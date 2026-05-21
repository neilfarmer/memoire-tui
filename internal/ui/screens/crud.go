package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// updateForm runs the standard huh.Form lifecycle for a screen's edit
// form: "esc" cancels (via onCancel and nilling the pointer), "ctrl+s"
// submits, anything else forwards to the form. When the form reports
// huh.StateCompleted, submit is invoked. The form pointer is updated
// in place so callers don't have to do the type-assert dance themselves.
func updateForm(form **huh.Form, msg tea.Msg, onCancel func(), submit func() tea.Cmd) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc":
			*form = nil
			onCancel()
			return nil
		case "ctrl+s":
			return submit()
		}
	}
	f, cmd := (*form).Update(msg)
	if ff, ok := f.(*huh.Form); ok {
		*form = ff
	}
	if (*form).State == huh.StateCompleted {
		return submit()
	}
	return cmd
}

// handleConfirmDelete dispatches the y/n/esc keys of a confirm-delete
// prompt. "y" runs doDelete; "n" or "esc" runs onCancel. Other keys are
// swallowed. The caller already matched on the confirm-delete mode.
func handleConfirmDelete(key string, doDelete func() tea.Cmd, onCancel func()) tea.Cmd {
	switch key {
	case "y":
		return doDelete()
	case "n", "esc":
		onCancel()
	}
	return nil
}
