package screens

import "testing"

// TestHabitWindow: the windowing math should keep the cursor visible
// regardless of where it is in the list. Each habit renders as a 4-line
// box, and height=22 leaves available=20 → maxVisible=5 habits.
func TestHabitWindow(t *testing.T) {
	cases := []struct {
		name             string
		cursor, total, h int
		wantStart        int
		wantEnd          int
	}{
		{"all fit, cursor at top", 0, 3, 22, 0, 3},
		{"all fit, cursor at bottom", 2, 3, 22, 0, 3},
		{"cursor inside first window", 0, 20, 22, 0, 5},
		{"cursor at edge of first window", 4, 20, 22, 0, 5},
		{"cursor advances past window", 5, 20, 22, 1, 6},
		{"cursor near end clamps to end", 19, 20, 22, 15, 20},
		{"cursor at very end", 19, 20, 22, 15, 20},
		{"tiny height shows at least one", 7, 20, 1, 7, 8},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end := habitWindow(tc.cursor, tc.total, tc.h)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Errorf("habitWindow(cursor=%d, total=%d, h=%d) = (%d, %d); want (%d, %d)",
					tc.cursor, tc.total, tc.h, start, end, tc.wantStart, tc.wantEnd)
			}
			if tc.cursor < tc.total {
				if tc.cursor < start || tc.cursor >= end {
					t.Errorf("cursor %d not inside window [%d, %d) — would be invisible",
						tc.cursor, start, end)
				}
			}
		})
	}
}
