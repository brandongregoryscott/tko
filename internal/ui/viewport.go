package ui

// alignStepPage syncs the TUI viewport (cursorStep + scrollOffset) to match
// the grid's current stepPage. Call after any grid operation that changes stepPage.
func (m *Model) alignStepPage() {
	m.cursorStep = m.stepPage * 16
	if m.cursorStep >= m.sequencer.Project.NumSteps {
		m.cursorStep = m.sequencer.Project.NumSteps - 1
	}
	// Pin scrollOffset to the page boundary so the full 16-step page is visible.
	m.scrollOffset = m.stepPage * 16
	vis := m.visibleSteps()
	maxScroll := m.sequencer.Project.NumSteps - vis
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// alignTrackOffset adjusts trackOffset so cursorTrack stays visible.
func (m *Model) alignTrackOffset() {
	if m.cursorTrack < m.trackOffset {
		m.trackOffset = m.cursorTrack
	}
	if m.cursorTrack >= m.trackOffset+6 {
		m.trackOffset = m.cursorTrack - 5
	}
	maxOff := len(m.sequencer.Project.Tracks) - 6
	if maxOff < 0 {
		maxOff = 0
	}
	if m.trackOffset > maxOff {
		m.trackOffset = maxOff
	}
	if m.trackOffset < 0 {
		m.trackOffset = 0
	}
}

// visibleSteps returns how many step columns fit in the current terminal width.
// Accounts for the track panel (trackPanelWidth), separator ("  "), and grid-internal label area.
const leftOverhead = trackPanelWidth + 2 + trackLabelWidth + 1

func (m Model) visibleSteps() int {
	w := m.width
	if w < 40 {
		w = 40
	}
	n := (w - leftOverhead) / cellWidth
	if n < 1 {
		n = 1
	}
	if n > m.sequencer.Project.NumSteps {
		n = m.sequencer.Project.NumSteps
	}
	return n
}

// scrollToCursor adjusts scrollOffset so the cursor step is visible.
func (m *Model) scrollToCursor() {
	vis := m.visibleSteps()
	if vis >= m.sequencer.Project.NumSteps {
		m.scrollOffset = 0
		return
	}
	if m.cursorStep < m.scrollOffset {
		m.scrollOffset = m.cursorStep
	}
	if m.cursorStep >= m.scrollOffset+vis {
		m.scrollOffset = m.cursorStep - vis + 1
	}
	maxScroll := m.sequencer.Project.NumSteps - vis
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}
