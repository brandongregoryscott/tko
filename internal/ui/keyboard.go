package ui

import (
	"fmt"

	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// File dialog mode — only Enter, Esc, and text input keys work.
	if m.focus == FocusSaveFile {
		return m.handleDialogKey(msg, true)
	}
	if m.focus == FocusLoadFile {
		return m.handleDialogKey(msg, false)
	}

	// Global keys.
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.HelpToggle):
		m.showHelp = !m.showHelp
		if !m.showHelp {
			m.statusMsg = ""
		}
		return m, nil

	case key.Matches(msg, m.keys.PlayPause):
		m.sequencer.TogglePlayPause()
		m.statusMsg = ""
		m.showHelp = false
		if m.sequencer.PlayState == engine.Playing {
			m.statusMsg = "Playing"
			m.refreshGridOnly()
			return m, m.scheduleTick()
		}
		m.statusMsg = "Stopped"
		return m.refreshGrid()

	case key.Matches(msg, m.keys.ResetPos):
		m.sequencer.ResetPosition()
		m.statusMsg = "Stopped"
		m.showHelp = false
		return m.refreshGrid()

	case key.Matches(msg, m.keys.Save):
		m.openSaveDialog()
		return m, nil

	case key.Matches(msg, m.keys.Load):
		m.openLoadDialog()
		return m, nil

	case key.Matches(msg, m.keys.ExportMIDI):
		return m, m.doExport()

	case key.Matches(msg, m.keys.BPMUp):
		m.sequencer.SetBPM(m.sequencer.Project.BPM + 5)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.BPMDown):
		m.sequencer.SetBPM(m.sequencer.Project.BPM - 5)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.BPMUpFine):
		m.sequencer.SetBPM(m.sequencer.Project.BPM + 1)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.BPMDownFine):
		m.sequencer.SetBPM(m.sequencer.Project.BPM - 1)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.CycleBankNext):
		m.cycleBank(1)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.CycleBankPrev):
		m.cycleBank(-1)
		return m.refreshGrid()
	}

	// Grid movement and editing.
	switch {
	case key.Matches(msg, m.keys.CursorUp):
		m.handleUpDownKey(true)
		return m, nil

	case key.Matches(msg, m.keys.CursorDown):
		m.handleUpDownKey(false)
		return m, nil

	case key.Matches(msg, m.keys.CursorLeft):
		if m.cursorStep > 0 {
			m.cursorStep--
			m.scrollToCursor()
		}
		m.showHelp = false
		return m, nil

	case key.Matches(msg, m.keys.CursorRight):
		if m.cursorStep < m.sequencer.Project.NumSteps-1 {
			m.cursorStep++
			m.scrollToCursor()
		}
		m.showHelp = false
		return m, nil

	case key.Matches(msg, m.keys.FirstStep):
		m.cursorStep = 0
		m.showHelp = false
		return m, nil

	case key.Matches(msg, m.keys.LastStep):
		m.cursorStep = m.sequencer.Project.NumSteps - 1
		m.showHelp = false
		return m, nil

	case key.Matches(msg, m.keys.ToggleStep):
		m.sequencer.ToggleStep(engine.TrackID(m.cursorTrack), m.cursorStep)
		m.showHelp = false
		return m.refreshGrid()

	case key.Matches(msg, m.keys.CycleNext):
		m.cycleSample(1)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.CyclePrev):
		m.cycleSample(-1)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.RandomSample):
		m.randomizeSample()
		return m.refreshGrid()

	case key.Matches(msg, m.keys.RandomizeAll):
		m.randomizeAllSamples()
		return m.refreshGrid()

	case key.Matches(msg, m.keys.MuteTrack):
		m.sequencer.ToggleMute(engine.TrackID(m.cursorTrack))
		tr := &m.sequencer.Project.Tracks[m.cursorTrack]
		if tr.Muted {
			m.statusMsg = tr.Name + " muted"
		} else {
			m.statusMsg = tr.Name + " unmuted"
		}
		return m.refreshGrid()

	case key.Matches(msg, m.keys.VolUp):
		t := &m.sequencer.Project.Tracks[m.cursorTrack]
		m.sequencer.SetVolume(t.ID, t.Volume+0.1)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.VolDown):
		t := &m.sequencer.Project.Tracks[m.cursorTrack]
		m.sequencer.SetVolume(t.ID, t.Volume-0.1)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.Steps16):
		m.sequencer.SetNumSteps(16)
		m.statusMsg = "16 steps"
		return m.refreshGrid()

	case key.Matches(msg, m.keys.Steps32):
		m.sequencer.SetNumSteps(32)
		m.statusMsg = "32 steps"
		return m.refreshGrid()

	case key.Matches(msg, m.keys.Steps48):
		m.sequencer.SetNumSteps(48)
		m.statusMsg = "48 steps"
		return m.refreshGrid()

	case key.Matches(msg, m.keys.Steps64):
		m.sequencer.SetNumSteps(64)
		m.statusMsg = "64 steps"
		return m.refreshGrid()

	case key.Matches(msg, m.keys.DupTrack):
		src := engine.TrackID(m.cursorTrack)
		dst := m.sequencer.DuplicateTrack(src)
		if dst < 0 {
			m.statusMsg = "No empty track to duplicate into"
		} else {
			m.statusMsg = m.sequencer.Project.Tracks[dst].Name + " created"
		}
		return m.refreshGrid()

	case key.Matches(msg, m.keys.CycleFolder):
		m.cycleTrackFolder()
		return m.refreshGrid()

	case key.Matches(msg, m.keys.ClearTrack):
		t := &m.sequencer.Project.Tracks[m.cursorTrack]
		if t.Sample.Folder == "" {
			m.statusMsg = "Track already empty"
		} else {
			name := t.Name
			m.sequencer.ClearTrack(engine.TrackID(m.cursorTrack))
			m.statusMsg = name + " cleared"
		}
		return m.refreshGrid()

	case key.Matches(msg, m.keys.SwingUp):
		m.sequencer.SetSwing(m.sequencer.Project.Swing + 0.1)
		m.statusMsg = fmt.Sprintf("Swing: %.1f", m.sequencer.Project.Swing)
		return m.refreshGrid()

	case key.Matches(msg, m.keys.SwingDown):
		m.sequencer.SetSwing(m.sequencer.Project.Swing - 0.1)
		m.statusMsg = fmt.Sprintf("Swing: %.1f", m.sequencer.Project.Swing)
		return m.refreshGrid()
	}

	return m, nil
}

func (m *Model) handleDialogKey(msg tea.KeyMsg, isSave bool) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if isSave {
			return m, m.confirmSave()
		}
		return m, m.confirmLoad()

	case tea.KeyEsc:
		m.cancelDialog()
		return m, nil

	case tea.KeyCtrlC:
		m.cancelDialog()
		return m, nil
	}

	if isSave {
		var cmd tea.Cmd
		m.fileInput, cmd = m.fileInput.Update(msg)
		return m, cmd
	}

	// Load mode: navigate file list.
	switch msg.String() {
	case "up", "k":
		if m.fileCursor > 0 {
			m.fileCursor--
		}
	case "down", "j":
		if m.fileCursor < len(m.fileList)-1 {
			m.fileCursor++
		}
	}
	return m, nil
}

func (m *Model) handleUpDownKey(isUp bool) {
	if isUp && m.cursorTrack > 0 {
		m.cursorTrack--
		m.showHelp = false
		return
	}

	if !isUp && m.cursorTrack < 7 {
		m.cursorTrack++
		m.showHelp = false
		return
	}
}
