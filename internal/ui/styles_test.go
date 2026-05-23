package ui

import (
	"strings"
	"testing"
)

func TestRenderCellAllStates(t *testing.T) {
	tests := []struct {
		name                string
		active, cursor, playhead, activePlayhead, cursorActive, beat bool
		contains            string
	}{
		{"inactive", false, false, false, false, false, false, "[ ]"},
		{"active", true, false, false, false, false, false, "[█]"},
		{"cursor inactive", false, true, false, false, false, false, "[ ]"},
		{"cursor active", true, false, false, false, true, false, "[█]"},
		{"playhead inactive", false, false, true, false, false, false, "[ ]"},
		{"active at playhead", true, false, true, true, false, false, "[█]"},
		{"beat inactive", false, false, false, false, false, true, "[ ]"},
		{"active at beat", true, false, false, false, false, true, "[█]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := RenderCell(tt.active, tt.cursor, tt.playhead, tt.activePlayhead, tt.cursorActive, tt.beat)
			if !strings.Contains(cell, tt.contains) {
				t.Errorf("expected cell to contain %q, got %q", tt.contains, cell)
			}
			if len(cell) == 0 {
				t.Error("cell should not be empty")
			}
		})
	}
}

func TestRenderCellCursorTakesPriority(t *testing.T) {
	// Cursor-active beats everything.
	cell := RenderCell(true, true, true, true, true, true)
	if !strings.Contains(cell, "[█]") {
		t.Errorf("cursor-active should show active char, got %q", cell)
	}

	// Cursor on inactive step.
	cell = RenderCell(false, true, false, false, false, false)
	if !strings.Contains(cell, "[ ]") {
		t.Errorf("cursor on inactive should show inactive char, got %q", cell)
	}
}
