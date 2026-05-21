package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/neilfarmer/memoire-tui/internal/styles"
)

// memoireArt is the ANSI Shadow rendering of "memoire" — 58 columns × 6 rows.
// Used as the splash banner during initial app boot and per-screen loading.
const memoireArt = `███╗   ███╗███████╗███╗   ███╗ ██████╗ ██╗██████╗ ███████╗
████╗ ████║██╔════╝████╗ ████║██╔═══██╗██║██╔══██╗██╔════╝
██╔████╔██║█████╗  ██╔████╔██║██║   ██║██║██████╔╝█████╗
██║╚██╔╝██║██╔══╝  ██║╚██╔╝██║██║   ██║██║██╔══██╗██╔══╝
██║ ╚═╝ ██║███████╗██║ ╚═╝ ██║╚██████╔╝██║██║  ██║███████╗
╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚═╝╚═╝  ╚═╝╚══════╝`

const (
	artWidth  = 58
	artHeight = 6
)

// Splash renders the memoire banner with an optional subtitle, centered
// inside (width × height). Falls back to a plain text title when the
// terminal is too narrow to fit the ASCII art.
func Splash(width, height int, subtitle string) string {
	if width < 1 || height < 1 {
		// Boot fallback: we haven't received a WindowSizeMsg yet, so
		// pick conservative defaults that look sane on 80×24.
		width, height = 80, 24
	}

	var body string
	if width < artWidth+2 {
		// Narrow terminal — degrade to a plain centered title.
		title := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("memoire")
		body = title
	} else {
		art := lipgloss.NewStyle().Foreground(styles.Primary).Render(memoireArt)
		body = art
	}

	if subtitle != "" {
		sub := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true).Render(subtitle)
		body = body + "\n\n" + lipgloss.PlaceHorizontal(lipgloss.Width(body), lipgloss.Center, sub)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, body)
}
