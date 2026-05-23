package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func TestTruncate(t *testing.T) {
	// Short string, no truncation.
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("short: got %q", got)
	}

	// Exact length.
	if got := truncate("hello", 5); got != "hello" {
		t.Errorf("exact: got %q", got)
	}

	// Overflow — should end with ellipsis and be at most max bytes (not counting
	// multi-byte ellipsis, since max is a byte count and ellipsis is 3 bytes).
	got := truncate("hello world", 8)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("overflow should end with ellipsis: got %q", got)
	}
	// "hello w…" = 7 ASCII bytes + 3-byte ellipsis = 10 bytes.
	if utf8.RuneCountInString(got) != 8 {
		t.Errorf("overflow rune count: got %d, want 8 (%q)", utf8.RuneCountInString(got), got)
	}
}

func TestScrollBar(t *testing.T) {
	// scrollBar returns: padding (trackLabelWidth+1 spaces) + bar of '─' with one '█'.
	bar := scrollBar(0, 20)
	// Should contain a position indicator.
	if !strings.Contains(bar, "█") {
		t.Error("scrollbar should have a position indicator")
	}

	// 0% — indicator at the first bar position (index 0 of bar portion).
	prefixLen := trackLabelWidth + 1
	runes := []rune(bar)
	if runes[prefixLen] != '█' {
		t.Errorf("0%%: first bar rune should be indicator, got %c", runes[prefixLen])
	}

	// 100% — indicator at the last bar position.
	full := scrollBar(1.0, 20)
	fullRunes := []rune(full)
	lastBarIdx := prefixLen + 19
	if fullRunes[lastBarIdx] != '█' {
		t.Errorf("100%%: last bar rune should be indicator, got %c", fullRunes[lastBarIdx])
	}

	// 50% — indicator near the middle.
	mid := scrollBar(0.5, 20)
	midRunes := []rune(mid)
	found := false
	for i := prefixLen + 8; i <= prefixLen+11; i++ {
		if midRunes[i] == '█' {
			found = true
			break
		}
	}
	if !found {
		t.Error("50%%: indicator should be near the middle")
	}
}

func TestScrollBarClamp(t *testing.T) {
	prefixLen := trackLabelWidth + 1

	// Negative should clamp to 0.
	bar := scrollBar(-0.5, 10)
	runes := []rune(bar)
	if runes[prefixLen] != '█' {
		t.Error("negative pct: indicator should be at first position")
	}

	// > 1 should clamp to 1.
	bar = scrollBar(2.0, 10)
	runes = []rune(bar)
	if runes[prefixLen+9] != '█' {
		t.Error(">1 pct: indicator should be at last position")
	}
}

func TestRenderGridEmptyProject(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16

	grid, scrollOff := RenderGrid(proj, 0, 0, 0, 8, engine.Stopped, 0)
	if scrollOff != 0 {
		t.Errorf("scrollOffset=%d, want 0", scrollOff)
	}
	if grid == "" {
		t.Fatal("grid should not be empty")
	}

	for i := 0; i < 8; i++ {
		name := proj.Tracks[i].Name
		if !strings.Contains(grid, truncate(name, trackLabelWidth-2)) {
			t.Errorf("grid should contain track name %q", name)
		}
	}

	if !strings.Contains(grid, "  1") {
		t.Error("grid should contain step number 1")
	}
}

func TestRenderGridActiveStep(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16
	proj.Tracks[0].Steps[0] = true

	grid, _ := RenderGrid(proj, 0, 0, 0, 8, engine.Stopped, 0)
	if !strings.Contains(grid, "[█]") {
		t.Error("grid should contain active step indicator")
	}
}

func TestRenderGridPlaying(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16
	proj.Tracks[0].Steps[0] = true

	grid, _ := RenderGrid(proj, 0, 0, 0, 8, engine.Playing, 0)
	if grid == "" {
		t.Fatal("grid should not be empty when playing")
	}
}

func TestRenderGridScrolled(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 64

	grid, scrollOff := RenderGrid(proj, 0, 16, 16, 8, engine.Stopped, 0)
	if scrollOff != 16 {
		t.Errorf("scrollOffset=%d, want 16", scrollOff)
	}
	if !strings.Contains(grid, " 17") {
		t.Error("scrolled grid should show step 17")
	}
}

func TestRenderGridMutedTrack(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16
	proj.Tracks[0].Muted = true
	proj.Tracks[0].Steps[0] = true

	grid, _ := RenderGrid(proj, 0, 0, 0, 8, engine.Stopped, 0)
	if grid == "" {
		t.Fatal("grid should not be empty with muted track")
	}
}

func TestRenderGridScrollBar(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 64

	grid, _ := RenderGrid(proj, 0, 0, 0, 8, engine.Stopped, 0)
	if !strings.Contains(grid, "█") {
		t.Error("scrolled grid should have a scrollbar indicator")
	}
}

func TestRenderStepRow(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16
	proj.Tracks[0].Steps[0] = true
	proj.Tracks[0].Steps[3] = true

	row := renderStepRow(&proj.Tracks[0], 0, 0, 0, 0, 8, 16, engine.Stopped, 0)
	if row == "" {
		t.Fatal("step row should not be empty")
	}
	if !strings.Contains(row, "[█]") {
		t.Error("active step should be visible in row")
	}
}

func TestRenderStepRowCursor(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16

	row := renderStepRow(&proj.Tracks[0], 0, 0, 2, 0, 8, 16, engine.Stopped, 0)
	if row == "" {
		t.Fatal("step row with cursor should not be empty")
	}
}

func TestRenderStepRowPlayhead(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16

	row := renderStepRow(&proj.Tracks[0], 0, 0, 0, 0, 8, 16, engine.Playing, 3)
	if row == "" {
		t.Fatal("step row with playhead should not be empty")
	}
}
