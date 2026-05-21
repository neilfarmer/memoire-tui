package screens

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

// notEmpty is a huh.Validate that requires a non-blank value.
func notEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("required")
	}
	return nil
}

// validateOptionalDate accepts blank or a YYYY-MM-DD calendar date.
func validateOptionalDate(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return fmt.Errorf("expected YYYY-MM-DD")
	}
	return nil
}

// validateOptionalInt accepts blank or a base-10 integer.
func validateOptionalInt(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if _, err := strconv.Atoi(s); err != nil {
		return fmt.Errorf("expected a whole number")
	}
	return nil
}

// validateOptionalFloat accepts blank or a base-10 float.
func validateOptionalFloat(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return fmt.Errorf("expected a number")
	}
	return nil
}

// parseInt returns 0 on blank input; rejects garbage like "7abc" (unlike
// fmt.Sscanf). Trims surrounding whitespace.
func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

// parseFloat is the float64 counterpart to parseInt.
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func intStr(n int) string       { return strconv.Itoa(n) }
func floatStr(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }

// SimpleItem is a generic list item carrying a title, description, and any
// underlying payload value.
type SimpleItem struct {
	TitleText string
	Desc      string
	Value     any
}

func (i SimpleItem) FilterValue() string { return i.TitleText + " " + i.Desc }
func (i SimpleItem) Title() string       { return i.TitleText }
func (i SimpleItem) Description() string { return i.Desc }

// MakeList builds a list.Model with our standard styling.
func MakeList(title string, items []list.Item, width, height int) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	return l
}
