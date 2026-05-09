package components

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// OpenedURLMsg confirms a URL was handed off to the OS opener.
type OpenedURLMsg struct {
	URL string
	Err error
}

// OpenURL hands a URL to $BROWSER, falling back to the OS default opener.
//
// gosec flags subprocess + tainted input: that's intentional. The whole
// point of this helper is to launch a user-chosen browser on a user-chosen
// URL. Memoire's API responses are the source of "target" and the user's
// own shell config is the source of $BROWSER.
func OpenURL(target string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if browser := os.Getenv("BROWSER"); browser != "" {
			parts := strings.Fields(browser)
			args := append(parts[1:], target)
			cmd = exec.Command(parts[0], args...) //#nosec G204,G702 -- $BROWSER is user config; URL flow is by design
		} else {
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", target) //#nosec G204 -- intentional URL hand-off
			case "windows":
				cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target) //#nosec G204
			default:
				cmd = exec.Command("xdg-open", target) //#nosec G204
			}
		}
		err := cmd.Start()
		return OpenedURLMsg{URL: target, Err: err}
	}
}
