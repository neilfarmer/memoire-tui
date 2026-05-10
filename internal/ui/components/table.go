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

// stripeBgs are the alternating row background colors. Even rows take the
// first entry, odd rows the second.
var stripeBgs = []lipgloss.AdaptiveColor{
	{Light: "#f8fafc", Dark: "#0e1620"},
	{Light: "#e2e8f0", Dark: "#1a2332"},
}

// NewTable returns a styled table.Model with alternating row backgrounds.
// The forked striped package adds a per-row Styles.RowStyler hook that the
// upstream bubbles/table v1.0.0 does not expose; we use it to paint
// alternating backgrounds without smuggling ANSI into cell values (which
// would confuse runewidth.Truncate inside renderRow).
func NewTable(cols []Column, rows []Row, height int) striped.Model {
	t := striped.New(
		striped.WithColumns(cols),
		striped.WithRows(rows),
		striped.WithFocused(true),
		striped.WithHeight(height),
	)
	s := striped.DefaultStyles()
	s.Header = s.Header.
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
		bg := stripeBgs[rowIndex%len(stripeBgs)]
		return cell.Background(bg)
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
