package striped

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestNoExtraNewlinesPerRow confirms each rendered row is a single line.
// Previously the user reported visible blank lines between every other
// row when a RowStyler was set; this catches the regression at unit-test
// granularity.
func TestNoExtraNewlinesPerRow(t *testing.T) {
	cols := []Column{{Title: "STATUS", Width: 12}, {Title: "TITLE", Width: 30}}
	rows := []Row{
		{"todo", "Flash the shed"},
		{"todo", "Finish observability"},
		{"todo", "eat food"},
		{"todo", "Look into"},
		{"todo", "Setup"},
	}
	m := New(WithColumns(cols), WithRows(rows), WithFocused(true), WithHeight(10))
	s := DefaultStyles()
	cell := s.Cell.Foreground(lipgloss.Color("15"))
	s.Cell = cell
	s.RowStyler = func(i int) lipgloss.Style {
		if i%2 == 1 {
			return cell.Background(lipgloss.Color("236"))
		}
		return cell
	}
	m.SetStyles(s)
	for i := range rows {
		out := m.renderRow(i)
		nls := strings.Count(out, "\n")
		if nls > 0 {
			t.Errorf("row %d: %d newlines (expected 0)\nrow=%q", i, nls, out)
		}
	}
	fmt.Println("rendered viewport content:\n" + m.View())
}
