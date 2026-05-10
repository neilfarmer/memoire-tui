package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/styles"
)

// Column wraps table.Column for readability.
type Column = table.Column

// Row wraps table.Row.
type Row = table.Row

// NewTable returns a styled table.Model. Column header row in cyan, primary-
// coloured selection.
//
// bubbles/table v1.0.0 truncates each cell with runewidth.Truncate AFTER
// applying the cell style, which means embedding ANSI background codes in
// cell values gets miscounted and corrupts the render. We can't do per-row
// striping at the cell level without forking the upstream package, so we
// rely on a strong selection style + the column-header bottom border for
// row separation. WithStripeColumn / Stripe / Stripe-with-cols below are
// kept as no-ops so existing call sites keep compiling while we figure
// out a custom striped table.
func NewTable(cols []Column, rows []Row, height int) table.Model {
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	s := table.DefaultStyles()
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
	t.SetStyles(s)
	return t
}

// WithStripeColumn is now a no-op (see NewTable for why).
func WithStripeColumn(cols []Column) []Column { return cols }

// Stripe is a no-op pass-through. It accepts the variadic column slice for
// backward compat with screens that already pass it.
func Stripe(rows []Row, _ ...[]Column) []Row { return rows }

// FrameTable renders a table k9s-style: bold title with count badge, the
// table itself (no surrounding box — the table column header acts as the
// frame), and a key-hint footer.
func FrameTable(title string, count int, t table.Model, hints []string, focused bool) string {
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
