package ui

import (
	"github.com/brandongregoryscott/tko/internal/engine"
	"github.com/brandongregoryscott/tko/internal/grid"
	"github.com/brandongregoryscott/tko/internal/persistence"

	tea "github.com/charmbracelet/bubbletea"
)

// flashClearMsg clears the track cursor flash effect after its timeout.
// gen is the flash generation — only the most recent flash is honored.
type flashClearMsg struct{ gen int }

// Update handles incoming messages for the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UpdateDimensions(msg.Width, msg.Height)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case TickMsg:
		return m.handleTick()

	case StatusMsg:
		m.statusMsg = string(msg)
		return m, nil

	case LibraryLoadProgressMsg:
		m.loadTotal = msg.Total
		m.loadDone = msg.Loaded
		m.loadCurrent = msg.Current
		return m, nil

	case LibraryLoadedMsg:
		m.audioLib = msg.Lib
		m.player = msg.Player
		m.audioReady = true
		m.loadCurrent = ""
		banks := m.audioLib.Banks()
		if len(banks) > 0 {
			m.bank = banks[0]
		}
		// Auto-assign first sample from each folder.
		m.autoAssignSamples()
		// Try loading the last saved project.
		if proj, err := persistence.Load(persistence.DefaultDir() + "/default.json"); err == nil {
			m.sequencer.LoadProject(proj)
			for i := range m.sequencer.Project.Tracks {
				t := &m.sequencer.Project.Tracks[i]
				if t.Sample.Folder != "" && m.audioLib != nil {
					if name := m.audioLib.SampleName(t.Sample.Bank, t.Sample.Folder, t.Sample.Index); name != "" {
						t.Sample.Name = name
					}
				}
			}
		}
		m.statusMsg = "Ready"
		return m.refreshGrid()

	case LibraryLoadErrorMsg:
		m.loadErr = msg.Err
		return m, tea.Quit

	case ProjectLoadedMsg:
		m.sequencer.LoadProject(msg.Project)
		for i := range m.sequencer.Project.Tracks {
			t := &m.sequencer.Project.Tracks[i]
			if t.Sample.Folder != "" && m.audioLib != nil {
				if name := m.audioLib.SampleName(t.Sample.Bank, t.Sample.Folder, t.Sample.Index); name != "" {
					t.Sample.Name = name
				}
			}
		}
		m.statusMsg = "Project loaded"
		return m.refreshGrid()

	case grid.GridKeyMsg:
		return m.handleGridKey(msg)

	case grid.GridConnectedMsg:
		m.gridStatus = "Grid: " + msg.Info.Type
		m.gridMod = 0
		m.flashTrack = -1
		return m.refreshGrid()

	case grid.GridDisconnectedMsg:
		m.gridStatus = "Grid: disconnected"
		return m, nil

	case grid.GridErrorMsg:
		m.gridStatus = "Grid error: " + msg.Error()
		return m, nil

	case flashClearMsg:
		if msg.gen == m.flashGen {
			m.flashTrack = -1
			m.refreshGridOnly()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleTick() (tea.Model, tea.Cmd) {
	if m.sequencer.PlayState != engine.Playing {
		return m, nil
	}
	triggers := m.sequencer.Tick()
	if m.audioReady && m.player != nil {
		m.player.Trigger(triggers, m.sequencer.Project)
	}
	// Update grid LEDs for new playhead position.
	m.refreshGridOnly()
	return m, m.scheduleTick()
}

// refreshGrid sends the current sequencer state to the grid hardware.
func (m Model) refreshGrid() (tea.Model, tea.Cmd) {
	m.refreshGridOnly()
	return m, nil
}

// refreshGridOnly updates the grid LEDs without returning a command.
func (m *Model) refreshGridOnly() {
	if m.grid != nil && m.grid.IsConnected() {
		s := grid.GridState{
			Project:     m.sequencer.Project,
			PlayState:   m.sequencer.PlayState,
			Position:    m.sequencer.Position,
			CursorTrack: m.cursorTrack,
			TrackOffset: m.trackOffset,
			StepPage:    m.stepPage,
			Modifiers:   m.gridMod,
			FlashTrack:  m.flashTrack,
			FlashUntil:  m.flashUntil,
		}
		if err := m.grid.RefreshLEDs(s); err != nil {
			m.gridStatus = "Grid error: " + err.Error()
		} else if info := m.grid.Info(); info != nil {
			m.gridStatus = "Grid: " + info.Type
		}
	}
}
