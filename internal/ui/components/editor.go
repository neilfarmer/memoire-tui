package components

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// EditorClosedMsg returns from EditExternal once the editor exits.
type EditorClosedMsg struct {
	Content string
	Err     error
}

// EditExternal launches $EDITOR (default: vi) on a tmpfile seeded with the
// given content. Uses tea.ExecProcess so bubbletea suspends the alt-screen
// + raw mode while the editor runs and restores terminal state on exit.
// Calling exec.Command.Run() directly while tea owns the terminal corrupts
// the session.
func EditExternal(initial, ext string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	f, err := os.CreateTemp(os.TempDir(), "memoire-*"+normaliseExt(ext))
	if err != nil {
		return func() tea.Msg { return EditorClosedMsg{Err: err} }
	}
	path := f.Name()
	if _, err := f.WriteString(initial); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return func() tea.Msg { return EditorClosedMsg{Err: err} }
	}
	_ = f.Close()

	cmd := buildEditorCommand(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(execErr error) tea.Msg {
		defer os.Remove(path)
		if execErr != nil {
			return EditorClosedMsg{Err: fmt.Errorf("%s: %w", editor, execErr)}
		}
		buf, err := os.ReadFile(path) //#nosec G304 -- path is the tmpfile we just created above
		if err != nil {
			return EditorClosedMsg{Err: err}
		}
		return EditorClosedMsg{Content: string(buf)}
	})
}

// buildEditorCommand assembles the exec.Cmd that launches $EDITOR on path.
// On Unix it delegates to /bin/sh so EDITOR can contain quoted paths and
// flags ("EDITOR='code -w'", "EDITOR='/Applications/Sublime Text/subl'").
// strings.Fields-based splitting would mangle both. On Windows it falls
// back to whitespace-splitting since /bin/sh isn't available.
func buildEditorCommand(editor, path string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		parts := strings.Fields(editor)
		args := append(parts[1:], path)
		return exec.Command(parts[0], args...) //#nosec G204 -- $EDITOR is user config
	}
	// Single-quote the tmpfile path defensively, escaping any embedded
	// single quotes. The path is from os.CreateTemp so this is precaution
	// rather than a known attack vector.
	quoted := "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
	return exec.Command("/bin/sh", "-c", editor+" "+quoted) //#nosec G204 -- $EDITOR is user config
}

func normaliseExt(ext string) string {
	if ext == "" {
		return ".md"
	}
	if filepath.Ext(ext) == "" {
		return "." + strings.TrimPrefix(ext, ".")
	}
	return ext
}
