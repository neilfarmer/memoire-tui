package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/styles"
	"github.com/neilfarmer/memoire-tui/internal/ui/components/striped"
)

// Column re-exports the forked striped table column.
type Column = striped.Column

// Row re-exports the forked striped table row.
type Row = striped.Row

// Model re-exports the forked striped table model so screens can keep their
// existing field declarations (`tbl table.Model` becomes `tbl components.Model`).
type Model = striped.Model

// stripeBg is the background painted on every other row. Even rows are
// left at terminal default so the alternation is unambiguous regardless of
// the user's terminal palette.
var stripeBg = lipgloss.AdaptiveColor{Light: "#e2e8f0", Dark: "#1e2a3a"}

// NewTable returns a styled table.Model with alternating row backgrounds.
// The forked striped package adds a per-row Styles.RowStyler hook that the
// upstream bubbles/table v1.0.0 does not expose; we use it to paint
// alternating backgrounds without smuggling ANSI into cell values (which
// would confuse runewidth.Truncate inside renderRow).
//
// Callers pass logical column widths that should sum to contentWidth.
// bubbles/table's default Cell style has Padding(0, 1) which adds 2
// columns of horizontal padding per cell at render time; left unhandled
// the rendered row would be wider than contentWidth and wrap, producing
// visible blank lines between every-other row. We zero the cell padding
// here so the sum stays predictable.
func NewTable(cols []Column, rows []Row, height int) striped.Model {
	t := striped.New(
		striped.WithColumns(cols),
		striped.WithRows(rows),
		striped.WithFocused(true),
		striped.WithHeight(height),
	)
	s := striped.DefaultStyles()
	// Drop default Cell + Header padding — see godoc above. Header inherits
	// from DefaultStyles which sets Padding(0, 1); zero it explicitly.
	s.Cell = lipgloss.NewStyle()
	s.Header = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		BorderBottom(true).
		Foreground(styles.Primary).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0b0f14"}).
		Background(styles.Primary).
		Bold(true)
	s.Cell = s.Cell.Foreground(styles.Text)
	cell := s.Cell
	s.RowStyler = func(rowIndex int) lipgloss.Style {
		// Only stripe odd rows. Even rows render with terminal default so
		// the alternation is visible no matter the user's background color.
		if rowIndex%2 == 1 {
			return cell.Background(stripeBg)
		}
		return cell
	}
	t.SetStyles(s)
	return t
}

// WithStripeColumn is now a no-op — the striped fork handles row backgrounds
// at renderRow time, so callers do not need a synthetic stripe column.
func WithStripeColumn(cols []Column) []Column { return cols }

// Stripe is a no-op pass-through. Kept so existing call sites compile.
func Stripe(rows []Row, _ ...[]Column) []Row { return rows }

// FrameTable renders a table k9s-style: bold title with count badge, the
// table itself (no surrounding box — the table column header acts as the
// frame), and a key-hint footer.
func FrameTable(title string, count int, t striped.Model, hints []string, focused bool) string {
	heading := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(title)
	badge := CountBadge(count, "items")
	dot := lipgloss.NewStyle().Foreground(styles.Muted).Render("·")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, heading, "  ", dot, "  ", badge)
	if focused {
		headerRow = headerRow + "  " + lipgloss.NewStyle().Foreground(styles.Accent).Render("●")
	}
	hintRow := strings.Join(hints, "  ")
	if hintRow != "" {
		hintRow = lipgloss.NewStyle().Foreground(styles.Muted).Render(hintRow)
	}
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, "", t.View(), "", hintRow)
}
