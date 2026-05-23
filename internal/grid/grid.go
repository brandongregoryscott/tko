// Package grid provides bidirectional communication with a Norns Grid
// (monome) hardware controller via the serialosc OSC daemon.
//
// The Controller handles device discovery, receives grid key events as
// Bubble Tea messages, and sends LED state updates to the hardware.
// The grid is optional — if no serialosc daemon is running, the
// Controller stays in Disconnected state and all operations are no-ops.
package grid

import (
	"fmt"
	"sync"
	"time"

	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/charmbracelet/bubbletea"
	"github.com/hypebeast/go-osc/osc"
)

// ConnectionState tracks the grid lifecycle.
type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Discovering
	Connected
)

func (s ConnectionState) String() string {
	switch s {
	case Disconnected:
		return "disconnected"
	case Discovering:
		return "discovering"
	case Connected:
		return "connected"
	}
	return "unknown"
}

// GridInfo holds identity information from serialosc discovery.
type GridInfo struct {
	ID   string // device serial number, e.g. "m0000045"
	Type string // device type, e.g. "monome 128"
	Port int    // OSC port for direct communication with the grid
}

// ---- Bubble Tea message types ----

// GridKeyMsg is sent when a grid key is pressed or released.
// State: 1 = key down, 0 = key up.
type GridKeyMsg struct {
	X, Y, State int
}

// GridConnectedMsg is sent when a grid device is discovered and connected.
type GridConnectedMsg struct{ Info GridInfo }

// GridDisconnectedMsg is sent when the grid device is disconnected.
type GridDisconnectedMsg struct{}

// GridErrorMsg is sent when a non-fatal error occurs.
type GridErrorMsg struct{ Err error }

func (m GridErrorMsg) Error() string { return m.Err.Error() }

// Modifier key bits for the grid control rows.
const (
	ModBPM    uint8 = 1 << iota // B — BPM control
	ModBank                     // b — bank cycling
	ModFolder                   // F — folder cycling
	ModSample                   // S — sample cycling
)

// GridState is a snapshot of sequencer + UI state used to render grid LEDs.
type GridState struct {
	Project     *engine.Project
	PlayState   engine.PlayState
	Position    int
	CursorTrack int
	TrackOffset int
	StepPage    int
	Modifiers   uint8
	FlashTrack  int       // cursor track that should flash (-1 = none)
	FlashUntil  time.Time // deadline for the flash effect
}

// ---- Controller ----

// Controller manages the lifecycle of a Norns Grid connection.
// It handles serialosc discovery, receives grid key events, and sends
// LED updates. All public methods are safe for concurrent use.
type Controller struct {
	mu    sync.Mutex
	state ConnectionState
	info  *GridInfo

	// sendMsg injects messages into Bubble Tea's event loop.
	// Typically wraps tea.Program.Send.
	sendMsg func(tea.Msg)

	// OSC communication.
	gridClient *osc.Client // sends LED data to the grid's OSC port
	gridServer *osc.Server // receives grid key events
	localPort  int         // our UDP listen port

	// Lifecycle.
	done   chan struct{}
	closed bool
}

// New creates a Controller. The sendMsg callback is called from background
// goroutines when grid events arrive; it should be program.Send so messages
// enter the Bubble Tea message loop.
func New(sendMsg func(tea.Msg)) *Controller {
	return &Controller{
		state:   Disconnected,
		sendMsg: sendMsg,
		done:    make(chan struct{}),
	}
}

// Start begins serialosc discovery and sets up the UDP listener for grid
// key events. It returns the local UDP port number or an error.
// Start spawns background goroutines and returns immediately.
func (c *Controller) Start() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	port, err := c.resolveLocalPort()
	if err != nil {
		return 0, fmt.Errorf("grid: %w", err)
	}
	c.localPort = port

	c.state = Discovering
	go c.discover()
	go c.reconnectLoop()

	return port, nil
}

// Close shuts down all goroutines, closes UDP sockets, and marks the
// device as disconnected.
func (c *Controller) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	close(c.done)

	if c.gridServer != nil {
		c.gridServer.CloseConnection()
	}
	c.gridClient = nil
	c.gridServer = nil
	c.state = Disconnected
	return nil
}

// ConnectionState returns the current connection state.
func (c *Controller) ConnectionState() ConnectionState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// Info returns device info (nil if not connected).
func (c *Controller) Info() *GridInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.info == nil {
		return nil
	}
	cp := *c.info
	return &cp
}

// IsConnected returns true if the grid is in Connected state.
func (c *Controller) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state == Connected
}

// RefreshLEDs computes the full 16×8 LED grid from the given state and
// sends it to the grid hardware. Safe to call from any goroutine.
// Returns immediately if not connected.
func (c *Controller) RefreshLEDs(s GridState) error {
	c.mu.Lock()
	if c.state != Connected || c.gridClient == nil {
		c.mu.Unlock()
		return nil
	}
	client := c.gridClient
	c.mu.Unlock()

	leds := buildLEDGrid(s)
	return sendFullGrid(client, leds)
}
