package ui

import (
	"fmt"
	"time"

	"github.com/brandongregoryscott/tko/internal/audio"

	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model and returns the initial command.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnterAltScreen, tea.HideCursor}
	if m.grid != nil {
		cmds = append(cmds, m.startGridCmd())
	}
	if !m.audioReady && m.samplesRoot != "" {
		cmds = append(cmds, m.startLibraryLoad())
	}
	return tea.Batch(cmds...)
}

// startGridCmd returns a command that starts grid discovery.
// It runs after the Bubble Tea event loop has started, so p.Send()
// from background goroutines won't block.
func (m Model) startGridCmd() tea.Cmd {
	return func() tea.Msg {
		port, err := m.grid.Start()
		if err != nil {
			return StatusMsg("Grid error: " + err.Error())
		}
		return StatusMsg(fmt.Sprintf("Grid listening on port %d", port))
	}
}

// startLibraryLoad returns a command that loads samples in the background,
// sending progress messages via libSender and a final LibraryLoadedMsg or
// LibraryLoadErrorMsg when done.
func (m *Model) startLibraryLoad() tea.Cmd {
	return func() tea.Msg {
		go func() {
			lib, err := audio.NewLibraryWithProgress(m.samplesRoot, m.sampleRate, func(p audio.LoadProgress) {
				m.libSender(LibraryLoadProgressMsg(p))
			})
			if err != nil {
				m.libSender(LibraryLoadErrorMsg{Err: err})
				return
			}
			player := audio.NewPlayer(lib)
			m.libSender(LibraryLoadedMsg{Lib: lib, Player: player})
		}()
		return nil
	}
}

// scheduleTick schedules the next tick based on the current BPM and swing.
func (m Model) scheduleTick() tea.Cmd {
	dur := m.sequencer.TickDuration(m.sequencer.Position)
	return tea.Tick(dur, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}
