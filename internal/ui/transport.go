package ui

import (
	"fmt"
	"strings"

	"github.com/brandongregoryscott/tko/internal/engine"
)

// RenderTransport returns the transport bar string.
func RenderTransport(seq *engine.Sequencer, bank string) string {
	playIndicator := "▶"
	if seq.PlayState == engine.Stopped {
		playIndicator = "⏸"
	}

	bar := (seq.Position + 1) / 4
	beat := (seq.Position)%4 + 1
	barNum := bar + 1
	if seq.Position == 0 && seq.PlayState == engine.Stopped {
		barNum = 1
		beat = 1
	}

	swingLabel := ""
	if seq.Project.Swing > 0 {
		swingLabel = fmt.Sprintf(" Swing:%.0f%%", seq.Project.Swing*100)
	}

	parts := []string{
		fmt.Sprintf(" %s", playIndicator),
		fmt.Sprintf("Bank:%s", bank),
		fmt.Sprintf("BPM:%3.0f", seq.Project.BPM),
		fmt.Sprintf("Step:%02d/%d", seq.Position+1, seq.Project.NumSteps),
		fmt.Sprintf("Bar:%d Beat:%d", barNum, beat),
	}
	if swingLabel != "" {
		parts = append(parts, strings.TrimSpace(swingLabel))
	}

	line := strings.Join(parts, " │ ")
	return TransportStyle.Render(line)
}
