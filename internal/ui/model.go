package ui

import (
	"time"

	"github.com/brandongregoryscott/tko/internal/audio"
	"github.com/brandongregoryscott/tko/internal/engine"
	"github.com/brandongregoryscott/tko/internal/grid"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep/v2"
)

// FocusArea indicates which UI region receives keyboard input.
type FocusArea int

const (
	FocusGrid FocusArea = iota
	FocusSaveFile
	FocusLoadFile
)

// TickMsg is sent each time the sequencer advances a step.
type TickMsg struct{}

// Model is the top-level Bubble Tea model.
type Model struct {
	sequencer *engine.Sequencer
	audioLib  *audio.Library
	player    *audio.Player
	keys      KeyMap

	// UI state
	width        int
	height       int
	focus        FocusArea
	cursorTrack  int
	cursorStep   int
	scrollOffset int
	showHelp     bool
	statusMsg    string
	bank         string

	// Norns Grid integration.
	grid        *grid.Controller
	gridStatus  string
	trackOffset int // grid vertical scroll (0, 6 — tracks visible on grid rows 0-5)
	stepPage    int // grid horizontal page (0..3, each page is 16 steps)

	// Grid modifier keys (bitmask of grid.Mod*).
	gridMod uint8

	// Step-page double-tap state (for shrinking step count).
	lastPageTap     int
	pageTapDeadline time.Time

	// Track cursor flash state.
	flashTrack int
	flashUntil time.Time
	flashGen   int // incremented each flash; stale clear messages are ignored

	// File dialog.
	fileInput    textinput.Model
	fileList     []string // available project files for load selector
	fileCursor   int      // selected index in fileList
	lastSaveName string   // remembers the last saved project name

	audioReady bool

	// Async library loading.
	samplesRoot string
	sampleRate  beep.SampleRate
	libSender   func(tea.Msg)
	loadTotal   int
	loadDone    int
	loadCurrent string
	loadErr     error
}

// New creates a new UI model.
// Pass nil for lib and player when samples should be loaded asynchronously
// (samplesRoot and sampleRate must be set).
func New(seq *engine.Sequencer, lib *audio.Library, player *audio.Player, samplesRoot string, sampleRate beep.SampleRate) Model {
	ti := textinput.New()
	ti.Placeholder = "filename"
	ti.CharLimit = 64
	ti.Width = 30
	ti.Prompt = "> "

	bank := ""
	if lib != nil {
		banks := lib.Banks()
		if len(banks) > 0 {
			bank = banks[0]
		}
	}
	m := Model{
		sequencer:   seq,
		audioLib:    lib,
		player:      player,
		keys:        DefaultKeyMap(),
		focus:       FocusGrid,
		fileInput:   ti,
		audioReady:  lib != nil && player != nil,
		bank:        bank,
		flashTrack:  -1,
		samplesRoot: samplesRoot,
		sampleRate:  sampleRate,
	}
	return m
}

// SetSender sets the callback used by async loading to pump messages into the TUI.
func (m *Model) SetSender(sender func(tea.Msg)) {
	m.libSender = sender
}

// AttachGrid sets the grid controller. Call before p.Run() or during Update.
func (m *Model) AttachGrid(g *grid.Controller) {
	m.grid = g
}

// SetStatus sets a transient status message.
func (m *Model) SetStatus(msg string) {
	m.statusMsg = msg
}

// AudioReady returns whether audio is initialized.
func (m Model) AudioReady() bool {
	return m.audioReady
}

// KeyMap returns the key bindings for the help bubble.
func (m Model) KeyMap() KeyMap {
	return m.keys
}

// UpdateDimensions updates the terminal dimensions.
func (m *Model) UpdateDimensions(w, h int) {
	m.width = w
	m.height = h
}

// StatusMsg carries a status bar message from an async command.
type StatusMsg string

// ProjectLoadedMsg signals that a project was loaded from disk.
type ProjectLoadedMsg struct {
	Project *engine.Project
}

// LibraryLoadProgressMsg carries a single sample-load completion event.
type LibraryLoadProgressMsg audio.LoadProgress

// LibraryLoadedMsg signals that the sample library has finished loading.
type LibraryLoadedMsg struct {
	Lib    *audio.Library
	Player *audio.Player
}

// LibraryLoadErrorMsg signals a fatal error during sample loading.
type LibraryLoadErrorMsg struct {
	Err error
}
