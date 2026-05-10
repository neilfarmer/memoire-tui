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
// coloured selection. Row striping is handled separately by Stripe() — call
// it on your rows before SetRows.
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
	// Drop default Padding(0,1) — Stripe() pads each cell itself so the
	// background color reaches the column edge.
	s.Cell = lipgloss.NewStyle().Foreground(styles.Text)
	t.SetStyles(s)
	return t
}

// WithStripeColumn is a no-op kept for backwards compatibility with screens
// that still wrap their column lists. Stripe() draws full-row backgrounds
// now, so a dedicated stripe column is unnecessary.
func WithStripeColumn(cols []Column) []Column { return cols }

var stripeBgs = []lipgloss.AdaptiveColor{
	{Light: "#f8fafc", Dark: "#0e1620"},
	{Light: "#e2e8f0", Dark: "#1a2332"},
}

// Stripe pads each cell to its column width and applies an alternating row
// background so individual rows are visually distinct.
func Stripe(rows []Row, cols []Column) []Row {
	out := make([]Row, 0, len(rows))
	for i, r := range rows {
		bg := stripeBgs[i%len(stripeBgs)]
		styled := make(Row, 0, len(r))
		for ci, val := range r {
			width := 0
			if ci < len(cols) {
				width = cols[ci].Width
			}
			cellStyle := lipgloss.NewStyle().Background(bg).Padding(0, 1)
			if width > 0 {
				cellStyle = cellStyle.Width(width).MaxWidth(width)
			}
			styled = append(styled, cellStyle.Render(val))
		}
		out = append(out, styled)
	}
	return out
}

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
