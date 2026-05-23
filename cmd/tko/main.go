package main

import (
	"fmt"
	"os"
	"time"

	"github.com/brandongregoryscott/tko/internal/engine"
	"github.com/brandongregoryscott/tko/internal/grid"
	"github.com/brandongregoryscott/tko/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

var version = "dev"

func main() {
	// Initialize audio.
	sampleRate := beep.SampleRate(44100)
	if err := speaker.Init(sampleRate, sampleRate.N(time.Second/10)); err != nil {
		fmt.Fprintf(os.Stderr, "audio init error: %v\n", err)
		os.Exit(1)
	}
	defer speaker.Close()

	seq := engine.NewSequencer()

	// Create model without library — samples load asynchronously with a progress bar.
	// Auto-assignment and last-project loading happen when the library is ready.
	m := ui.New(seq, nil, nil, "samples", sampleRate)

	// Grid and sender must be attached before NewProgram because NewProgram
	// copies the model by value — modifications after NewProgram are invisible.
	var p *tea.Program
	g := grid.New(func(msg tea.Msg) { p.Send(msg) })
	m.AttachGrid(g)
	defer g.Close()
	m.SetSender(func(msg tea.Msg) { p.Send(msg) })

	p = tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
