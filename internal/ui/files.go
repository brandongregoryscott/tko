package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/brandongregoryscott/tko/internal/midiexport"
	"github.com/brandongregoryscott/tko/internal/persistence"
	"github.com/brandongregoryscott/tko/internal/wavexport"

	tea "github.com/charmbracelet/bubbletea"
)

// maxNameIncrements caps how far nextProjectName will count before giving up
// and falling back to a timestamp suffix.
const maxNameIncrements = 10000

// openSaveDialog shows the file input for saving.
// If the project name already exists, a unique name is suggested instead.
func (m *Model) openSaveDialog() {
	name := m.lastSaveName
	if name == "" {
		name = "default"
	}
	name = nextProjectName(name, projectFileExists)
	m.fileInput.SetValue(name)
	m.fileInput.Placeholder = "project name"
	m.fileInput.Focus()
	m.focus = FocusSaveFile
	m.statusMsg = "Save — Enter to confirm, Esc to cancel"
}

// projectFileExists reports whether a project with the given name is already saved.
func projectFileExists(name string) bool {
	_, err := os.Stat(persistence.DefaultDir() + "/" + name + ".json")
	return err == nil
}

// nextProjectName suggests an unused project name based on name. Names ending
// in a number are incremented (gen-boombap-11 -> gen-boombap-12), preserving any
// zero padding; anything else falls back to a timestamp suffix.
func nextProjectName(name string, exists func(string) bool) string {
	if !exists(name) {
		return name
	}

	prefix, digits := splitTrailingDigits(name)
	if digits != "" {
		if n, err := strconv.ParseUint(digits, 10, 64); err == nil {
			width := 0
			if strings.HasPrefix(digits, "0") {
				width = len(digits)
			}
			for i := n + 1; i < n+maxNameIncrements; i++ {
				candidate := prefix + fmt.Sprintf("%0*d", width, i)
				if !exists(candidate) {
					return candidate
				}
			}
		}
	}

	return name + "-" + time.Now().Format("2006-01-02-150405")
}

// splitTrailingDigits splits name into everything before its trailing run of
// digits and the digits themselves. The digits are empty if name does not end
// in one.
func splitTrailingDigits(name string) (string, string) {
	i := len(name)
	for i > 0 && name[i-1] >= '0' && name[i-1] <= '9' {
		i--
	}
	return name[:i], name[i:]
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

// doExport writes the project as a MIDI file along with a rendered WAV loop
// sharing the same timestamped base name.
func (m Model) doExport() tea.Cmd {
	return func() tea.Msg {
		midPath := midiexport.DefaultPath()
		if err := midiexport.Export(m.sequencer.Project, midPath); err != nil {
			return StatusMsg("Export error: " + err.Error())
		}
		if m.audioLib == nil {
			return StatusMsg("Exported " + midPath + " (WAV skipped: samples not loaded)")
		}
		wavPath := strings.TrimSuffix(midPath, filepath.Ext(midPath)) + ".wav"
		if err := wavexport.Export(m.sequencer.Project, m.audioLib, wavPath, m.sampleRate); err != nil {
			return StatusMsg("Exported " + midPath + " (WAV error: " + err.Error() + ")")
		}
		return StatusMsg("Exported " + midPath + " + .wav")
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
