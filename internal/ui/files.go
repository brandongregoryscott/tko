package ui

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/brandongregoryscott/tko/internal/midiexport"
	"github.com/brandongregoryscott/tko/internal/persistence"

	tea "github.com/charmbracelet/bubbletea"
)

// openSaveDialog shows the file input for saving.
// If the project name already exists, a timestamp is automatically appended.
func (m *Model) openSaveDialog() {
	name := m.lastSaveName
	if name == "" {
		name = "default"
	}
	path := persistence.DefaultDir() + "/" + name + ".json"
	if _, err := os.Stat(path); err == nil {
		name = name + "-" + time.Now().Format("2006-01-02-150405")
	}
	m.fileInput.SetValue(name)
	m.fileInput.Placeholder = "project name"
	m.fileInput.Focus()
	m.focus = FocusSaveFile
	m.statusMsg = "Save — Enter to confirm, Esc to cancel"
}

// openLoadDialog shows a selector of available project files.
func (m *Model) openLoadDialog() {
	m.fileList = listProjectFiles()
	m.fileCursor = 0
	m.focus = FocusLoadFile
	m.statusMsg = "Load — ↑↓ to select, Enter to confirm, Esc to cancel"
}

func listProjectFiles() []string {
	entries, err := os.ReadDir(persistence.DefaultDir())
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			names = append(names, strings.TrimSuffix(name, ".json"))
		}
	}
	sort.Strings(names)
	return names
}

// confirmSave writes the project using the current file input value.
func (m *Model) confirmSave() tea.Cmd {
	name := sanitizeFilename(m.fileInput.Value())
	if name == "" {
		name = "default"
	}
	m.lastSaveName = name
	m.focus = FocusGrid
	m.fileInput.Blur()
	path := persistence.DefaultDir() + "/" + name + ".json"
	proj := m.sequencer.Project
	return func() tea.Msg {
		if err := persistence.Save(proj, path); err != nil {
			return StatusMsg("Save error: " + err.Error())
		}
		return StatusMsg("Saved " + path)
	}
}

// confirmLoad reads the selected project file.
func (m *Model) confirmLoad() tea.Cmd {
	if m.fileCursor < 0 || m.fileCursor >= len(m.fileList) {
		m.focus = FocusGrid
		return func() tea.Msg { return StatusMsg("No project selected") }
	}
	name := m.fileList[m.fileCursor]
	m.lastSaveName = name
	m.focus = FocusGrid
	return func() tea.Msg {
		path := persistence.DefaultDir() + "/" + name + ".json"
		proj, err := persistence.Load(path)
		if err != nil {
			return StatusMsg("Load error: " + err.Error())
		}
		return ProjectLoadedMsg{Project: proj}
	}
}

// cancelDialog closes the file dialog without action.
func (m *Model) cancelDialog() {
	m.fileInput.Blur()
	m.focus = FocusGrid
	m.fileList = nil
	m.fileCursor = 0
	m.statusMsg = "Cancelled"
}

// doExport writes the project as a MIDI file.
func (m Model) doExport() tea.Cmd {
	return func() tea.Msg {
		path := midiexport.DefaultPath()
		if err := midiexport.Export(m.sequencer.Project, path); err != nil {
			return StatusMsg("Export error: " + err.Error())
		}
		return StatusMsg("Exported " + path)
	}
}

func sanitizeFilename(s string) string {
	// Strip extensions and unsafe characters.
	for i, c := range s {
		if c == '.' || c == '/' || c == '\\' || c == ':' || c == '*' ||
			c == '?' || c == '"' || c == '<' || c == '>' || c == '|' {
			return s[:i]
		}
	}
	return s
}
