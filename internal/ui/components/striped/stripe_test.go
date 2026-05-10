package striped

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	// Force color output so the renderer emits ANSI escapes in tests.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// TestRowStylerEmitsBackground confirms the forked renderRow runs the
// RowStyler hook and the resulting ANSI background reaches the output.
func TestRowStylerEmitsBackground(t *testing.T) {
	cols := []Column{{Title: "A", Width: 6}, {Title: "B", Width: 6}}
	rows := []Row{{"x1", "y1"}, {"x2", "y2"}, {"x3", "y3"}}
	m := New(WithColumns(cols), WithRows(rows), WithFocused(true), WithHeight(5))
	s := DefaultStyles()
	s.RowStyler = func(rowIndex int) lipgloss.Style {
		if rowIndex%2 == 0 {
			return lipgloss.NewStyle().Background(lipgloss.Color("236"))
		}
		return lipgloss.NewStyle().Background(lipgloss.Color("238"))
	}
	m.SetStyles(s)
	out := m.View()
	// At least one background ANSI escape (\x1b[48;5;...) must appear.
	if !strings.Contains(out, "\x1b[48;5;236") && !strings.Contains(out, "\x1b[48;5;238") {
		t.Errorf("expected RowStyler background ANSI in output; got:\n%q", out)
	}
}
