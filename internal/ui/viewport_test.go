package ui

import (
	"testing"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func TestVisibleStepsBodyFitsTerminal(t *testing.T) {
	// The total rendered width must not exceed the terminal width.
	// Body = trackPanelWidth (24) + separator (2) + grid internal padding (trackLabelWidth+1) + visibleSteps*cellWidth.
	// Minimum body width with 1 visible step = 44; below that we accept overflow.
	minBody := trackPanelWidth + 2 + trackLabelWidth + 1 + 1*cellWidth
	widths := []int{60, 80, 100, 120, 140, 160, 186, 200, 240}

	for _, w := range widths {
		proj := engine.DefaultProject()
		proj.NumSteps = 64
		m := Model{
			sequencer: engine.NewSequencer(),
		}
		m.sequencer.Project = proj
		m.width = w

		vis := m.visibleSteps()

		bodyWidth := trackPanelWidth + 2 + trackLabelWidth + 1 + vis*cellWidth
		if bodyWidth > w && w >= minBody {
			t.Errorf("width=%d: body overflow: %d > %d (visibleSteps=%d)", w, bodyWidth, w, vis)
		}
		if vis < 1 {
			t.Errorf("width=%d: visibleSteps must be >= 1, got %d", w, vis)
		}
	}
}

func TestScrollToCursorNeverExceedsMaxScroll(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 64
	m := Model{
		sequencer: engine.NewSequencer(),
	}
	m.sequencer.Project = proj
	m.width = 120

	vis := m.visibleSteps()
	maxScroll := proj.NumSteps - vis

	// Scroll to last step from position 0.
	m.cursorStep = 63
	m.scrollToCursor()

	if m.scrollOffset > maxScroll {
		t.Errorf("scrollOffset %d exceeds maxScroll %d (vis=%d)", m.scrollOffset, maxScroll, vis)
	}
	if m.scrollOffset < 0 {
		t.Errorf("scrollOffset %d is negative", m.scrollOffset)
	}

	// Verify the cursor step is visible.
	if m.cursorStep < m.scrollOffset || m.cursorStep >= m.scrollOffset+vis {
		t.Errorf("cursorStep %d not in visible range [%d, %d) (vis=%d)",
			m.cursorStep, m.scrollOffset, m.scrollOffset+vis, vis)
	}
}

func TestScrollFollowsCursor(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 64
	m := Model{
		sequencer: engine.NewSequencer(),
	}
	m.sequencer.Project = proj
	m.width = 80 // narrow terminal

	// Move cursor right one step at a time; verify it stays visible.
	for step := 0; step < 64; step++ {
		m.cursorStep = step
		m.scrollToCursor()
		vis := m.visibleSteps()
		if m.cursorStep < m.scrollOffset || m.cursorStep >= m.scrollOffset+vis {
			t.Errorf("step %d: cursorStep %d not visible [%d, %d) (vis=%d)",
				step, m.cursorStep, m.scrollOffset, m.scrollOffset+vis, vis)
		}
	}

	// Move cursor left one step at a time.
	for step := 63; step >= 0; step-- {
		m.cursorStep = step
		m.scrollToCursor()
		vis := m.visibleSteps()
		if m.cursorStep < m.scrollOffset || m.cursorStep >= m.scrollOffset+vis {
			t.Errorf("step %d: cursorStep %d not visible [%d, %d) (vis=%d)",
				step, m.cursorStep, m.scrollOffset, m.scrollOffset+vis, vis)
		}
	}
}

func TestVisibleStepsNeverExceedsNumSteps(t *testing.T) {
	proj := engine.DefaultProject()
	m := Model{
		sequencer: engine.NewSequencer(),
	}
	m.sequencer.Project = proj
	m.width = 500 // huge terminal

	for _, n := range []int{16, 32, 48, 64} {
		proj.NumSteps = n
		vis := m.visibleSteps()
		if vis > n {
			t.Errorf("NumSteps=%d: visibleSteps=%d exceeds NumSteps", n, vis)
		}
	}
}

func TestAlignStepPage(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 64
	m := Model{
		sequencer: engine.NewSequencer(),
	}
	m.sequencer.Project = proj
	m.width = 120

	m.stepPage = 1
	m.alignStepPage()

	if m.cursorStep != 16 {
		t.Errorf("stepPage=1: cursorStep=%d, want 16", m.cursorStep)
	}
	if m.scrollOffset != 16 {
		t.Errorf("stepPage=1: scrollOffset=%d, want 16", m.scrollOffset)
	}

	// Page 3 (last page) with narrow terminal — scrollOffset should clamp.
	m.width = 60
	m.stepPage = 3
	m.alignStepPage()
	if m.cursorStep != 48 {
		t.Errorf("stepPage=3: cursorStep=%d, want 48", m.cursorStep)
	}
	vis := m.visibleSteps()
	maxScroll := proj.NumSteps - vis
	if m.scrollOffset > maxScroll {
		t.Errorf("scrollOffset=%d exceeds maxScroll=%d", m.scrollOffset, maxScroll)
	}
	if m.scrollOffset < 0 {
		t.Errorf("scrollOffset should not be negative, got %d", m.scrollOffset)
	}
}

func TestAlignTrackOffset(t *testing.T) {
	proj := engine.DefaultProject()
	m := Model{
		sequencer: engine.NewSequencer(),
	}
	m.sequencer.Project = proj

	// Cursor at track 0, offset 0 — no change.
	m.cursorTrack = 0
	m.trackOffset = 0
	m.alignTrackOffset()
	if m.trackOffset != 0 {
		t.Errorf("trackOffset should stay 0, got %d", m.trackOffset)
	}

	// Cursor below visible range — offset follows.
	m.cursorTrack = 7
	m.trackOffset = 0
	m.alignTrackOffset()
	if m.trackOffset != 2 {
		t.Errorf("cursorTrack=7: trackOffset=%d, want 2", m.trackOffset)
	}

	// Cursor above visible range — offset follows.
	m.cursorTrack = 0
	m.trackOffset = 2
	m.alignTrackOffset()
	if m.trackOffset != 0 {
		t.Errorf("cursorTrack=0: trackOffset=%d, want 0", m.trackOffset)
	}
}
