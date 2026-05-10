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

// NewTable returns a styled table.Model. The model is k9s-like: column header
// row in cyan, alternating row striping, primary-coloured selection.
func NewTable(cols []Column, rows []Row, height int) table.Model {
	// Inject a leading "stripe" column. Each row's stripe cell is filled in
	// by Stripe(rows) before being passed to SetRows.
	cols = withStripeColumn(cols)
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

// WithStripeColumn prepends an empty column used by Stripe() to render the
// alternating per-row colored bar. Width 2 (one glyph + one space). Call this
// inside any *Cols helper that feeds NewTable / SetColumns.
func WithStripeColumn(cols []Column) []Column {
	out := make([]Column, 0, len(cols)+1)
	out = append(out, Column{Title: " ", Width: 2})
	out = append(out, cols...)
	return out
}

// withStripeColumn is the unexported alias kept for the NewTable internal call.
func withStripeColumn(cols []Column) []Column { return WithStripeColumn(cols) }

// Stripe wraps each row with its alternating leading bar. Call this on the
// rows you build before SetRows.
func Stripe(rows []Row) []Row {
	colors := []lipgloss.AdaptiveColor{
		{Light: "#0891b2", Dark: "#22d3ee"}, // cyan
		{Light: "#d97706", Dark: "#fbbf24"}, // amber
	}
	out := make([]Row, 0, len(rows))
	for i, r := range rows {
		c := colors[i%len(colors)]
		bar := lipgloss.NewStyle().Foreground(c).Bold(true).Render("▎")
		out = append(out, append(Row{bar}, r...))
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
