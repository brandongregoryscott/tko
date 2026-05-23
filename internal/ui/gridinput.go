package ui

import (
	"fmt"
	"time"

	"github.com/brandongregoryscott/tko/internal/engine"
	"github.com/brandongregoryscott/tko/internal/grid"

	tea "github.com/charmbracelet/bubbletea"
)

// handleGridKey dispatches a grid key press by row.
// Modifier keys (B/b/F/S) track state on both key-down and key-up.
func (m Model) handleGridKey(msg grid.GridKeyMsg) (tea.Model, tea.Cmd) {
	// Modifier tracking — handle both key-down and key-up.
	isMod := false
	switch {
	case msg.Y == 7 && msg.X == 12: // B — BPM modifier
		isMod = true
		if msg.State == 1 {
			m.gridMod |= grid.ModBPM
		} else {
			m.gridMod &^= grid.ModBPM
		}
	case msg.Y == 6 && msg.X == 12: // b — bank modifier
		isMod = true
		if msg.State == 1 {
			m.gridMod |= grid.ModBank
		} else {
			m.gridMod &^= grid.ModBank
		}
	case msg.Y == 6 && msg.X == 13: // F — folder modifier
		isMod = true
		if msg.State == 1 {
			m.gridMod |= grid.ModFolder
		} else {
			m.gridMod &^= grid.ModFolder
		}
	case msg.Y == 6 && msg.X == 15: // S — sample modifier
		isMod = true
		if msg.State == 1 {
			m.gridMod |= grid.ModSample
		} else {
			m.gridMod &^= grid.ModSample
		}
	}
	if isMod {
		return m.refreshGrid()
	}

	if msg.State != 1 {
		return m, nil // ignore key-up for non-modifier keys
	}

	switch {
	case msg.Y == 7:
		return m.handleGridTransport(msg.X)
	case msg.Y == 6:
		return m.handleGridControl(msg.X)
	default:
		return m.handleGridStep(msg.X, msg.Y)
	}
}

// handleGridTransport handles key presses on the transport row (row 7).
// Layout: ▶ ■ M ··· R B ◄ ▼ ►
func (m Model) handleGridTransport(col int) (tea.Model, tea.Cmd) {
	switch col {
	case 0: // ▶ — play/pause
		m.sequencer.TogglePlayPause()
		m.statusMsg = ""
		if m.sequencer.PlayState == engine.Playing {
			m.statusMsg = "Playing"
			m.refreshGridOnly()
			return m, m.scheduleTick()
		}
		m.statusMsg = "Stopped"
	case 1: // ■ — stop/reset
		m.sequencer.ResetPosition()
		m.statusMsg = "Stopped"
	case 2: // M — mute toggle for cursor track
		if m.cursorTrack < len(m.sequencer.Project.Tracks) {
			m.sequencer.ToggleMute(engine.TrackID(m.cursorTrack))
			tr := &m.sequencer.Project.Tracks[m.cursorTrack]
			if tr.Muted {
				m.statusMsg = tr.Name + " muted"
			} else {
				m.statusMsg = tr.Name + " unmuted"
			}
		}
	case 11: // R — randomize all
		m.randomizeAllSamples()
	case 13: // ◄ — step page left
		if m.stepPage > 0 {
			m.stepPage--
			m.alignStepPage()
		}
	case 14: // ▼ — cursor track down (modifier-aware)
		return m.handleGridArrow(false)
	case 15: // ► — step page right
		maxPage := (m.sequencer.Project.NumSteps - 1) / 16
		if m.stepPage < maxPage {
			m.stepPage++
			m.alignStepPage()
		}
	}
	return m.refreshGrid()
}

// handleGridControl handles key presses on the control row (row 6).
// Layout: 1 2 · 4 ··· r b F ▲ S
func (m Model) handleGridControl(col int) (tea.Model, tea.Cmd) {
	switch col {
	case 0: // page 0 (steps 0-15)
		return m.handleStepPageButton(0)
	case 1: // page 1 (steps 16-31)
		return m.handleStepPageButton(1)
	case 2: // page 2 (steps 32-47)
		return m.handleStepPageButton(2)
	case 3: // page 3 (steps 48-63)
		return m.handleStepPageButton(3)
	case 11: // r — randomize cursor track
		m.randomizeSample()
	case 14: // ▲ — cursor track up (modifier-aware)
		return m.handleGridArrow(true)
	}
	return m.refreshGrid()
}

// handleGridStep toggles a step in the track grid (rows 0-5).
func (m Model) handleGridStep(x, y int) (tea.Model, tea.Cmd) {
	trackIdx := m.trackOffset + y
	if trackIdx >= len(m.sequencer.Project.Tracks) {
		return m, nil
	}
	stepIdx := m.stepPage*16 + x
	if stepIdx >= m.sequencer.Project.NumSteps {
		return m, nil
	}
	m.sequencer.ToggleStep(engine.TrackID(trackIdx), stepIdx)
	return m.refreshGrid()
}

// handleGridArrow processes ▲/▼ with modifier-aware routing.
// isUp: true for ▲, false for ▼.
func (m Model) handleGridArrow(isUp bool) (tea.Model, tea.Cmd) {
	switch {
	case m.gridMod&grid.ModBPM != 0:
		delta := float64(5)
		if !isUp {
			delta = -5
		}
		m.sequencer.SetBPM(m.sequencer.Project.BPM + delta)
	case m.gridMod&grid.ModBank != 0:
		if isUp {
			m.cycleBank(1)
		} else {
			m.cycleBank(-1)
		}
	case m.gridMod&grid.ModFolder != 0:
		delta := 1
		if !isUp {
			delta = -1
		}
		m.cycleTrackFolderDelta(delta)
	case m.gridMod&grid.ModSample != 0:
		delta := 1
		if !isUp {
			delta = -1
		}
		m.cycleSample(delta)
	default:
		// No modifier: move cursor track.
		moved := false
		if isUp {
			if m.cursorTrack > 0 {
				m.cursorTrack--
				moved = true
			}
		} else {
			if m.cursorTrack < len(m.sequencer.Project.Tracks)-1 {
				m.cursorTrack++
				moved = true
			}
		}
		if moved {
			m.flashTrack = m.cursorTrack
			m.flashUntil = time.Now().Add(1 * time.Second)
			m.flashGen++
		}
		m.alignTrackOffset()
	}
	m.refreshGridOnly()
	if m.flashTrack >= 0 {
		gen := m.flashGen
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return flashClearMsg{gen: gen}
		})
	}
	return m, nil
}

// handleStepPageButton handles the page-navigation buttons.
// col 0 = page 0 (steps 0-15), col 1 = page 1 (steps 16-31), col 3 = page 3 (steps 48-63).
// Single tap navigates to the page; double-tap shrinks to 16 steps.
func (m Model) handleStepPageButton(col int) (tea.Model, tea.Cmd) {
	// Map column to page and minimum steps.
	type pageInfo struct{ page, minSteps int }
	pages := map[int]pageInfo{
		0: {0, 16},
		1: {1, 32},
		2: {2, 48},
		3: {3, 64},
	}
	info := pages[col]
	now := time.Now()

	if m.stepPage == info.page && now.Before(m.pageTapDeadline) {
		// Double-tap on active page: shrink to 16 steps, page 0.
		m.sequencer.SetNumSteps(16)
		m.statusMsg = "16 steps"
		m.stepPage = 0
		m.lastPageTap = -1
		m.pageTapDeadline = time.Time{}
		return m.refreshGrid()
	}

	m.lastPageTap = col
	m.pageTapDeadline = now.Add(400 * time.Millisecond)

	if m.stepPage == info.page {
		// Already on this page — wait for double-tap.
		return m, nil
	}

	// Navigate to this page. Ensure enough steps.
	if m.sequencer.Project.NumSteps < info.minSteps {
		m.sequencer.SetNumSteps(info.minSteps)
	}
	m.stepPage = info.page
	m.alignStepPage()
	m.statusMsg = fmt.Sprintf("%d steps", m.sequencer.Project.NumSteps)
	return m.refreshGrid()
}
