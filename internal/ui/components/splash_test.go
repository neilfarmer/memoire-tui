package components

import (
	"strings"
	"testing"
)

// TestSplashIncludesBannerAndSubtitle: the wide rendering should contain
// every line of the ANSI Shadow art plus the subtitle text.
func TestSplashIncludesBannerAndSubtitle(t *testing.T) {
	out := Splash(80, 24, "loading…")
	for i, line := range strings.Split(memoireArt, "\n") {
		if !strings.Contains(out, line) {
			t.Errorf("splash output missing art row %d: %q", i, line)
		}
	}
	if !strings.Contains(out, "loading…") {
		t.Error("splash output missing subtitle")
	}
}

// TestSplashNarrowFallback: when the terminal is narrower than the art
// can fit, Splash falls back to a plain "memoire" title rather than
// dumping a wrapped/garbled banner.
func TestSplashNarrowFallback(t *testing.T) {
	out := Splash(40, 12, "loading…")
	if strings.Contains(out, "███╗") {
		t.Error("narrow splash should not include ASCII art blocks")
	}
	if !strings.Contains(out, "memoire") {
		t.Error("narrow splash should include plain 'memoire' title")
	}
	if !strings.Contains(out, "loading…") {
		t.Error("narrow splash should still include subtitle")
	}
}

// TestSplashZeroSizeUsesDefaults: pre-WindowSizeMsg the app passes
// (0, 0); the function picks 80×24 defaults instead of returning empty.
func TestSplashZeroSizeUsesDefaults(t *testing.T) {
	out := Splash(0, 0, "")
	if !strings.Contains(out, "███╗") {
		t.Error("zero size should render the full ASCII art via defaults")
	}
}
