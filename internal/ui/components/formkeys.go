package components

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

// FormKeyMap returns a huh KeyMap with Text fields rebound so:
//
//	enter      → insert newline
//	tab        → advance to next field
//	shift+tab  → previous field
//	ctrl+s     → submit form (handled separately by each screen, but kept
//	             here as the documented submit key)
//	ctrl+e     → open the body in $EDITOR
//
// huh's default mapping treats enter as "next field" inside Text, which
// makes multi-paragraph editing impossible. We swap that so the textarea
// behaves like every other multi-line editor.
func FormKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Text.Next = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next"))
	km.Text.NewLine = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "new line"))
	km.Text.Submit = key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save"))
	return km
}
