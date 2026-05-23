package ui

import (
	"strings"
	"testing"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func TestRenderTransportStopped(t *testing.T) {
	seq := engine.NewSequencer()
	bar := RenderTransport(seq, "beatbox")

	if !strings.Contains(bar, "BPM") {
		t.Error("transport should show BPM")
	}
	if !strings.Contains(bar, "beatbox") {
		t.Error("transport should show bank name")
	}
	if !strings.Contains(bar, "Step:01") {
		t.Error("transport should show step 1 when stopped at 0")
	}
	if !strings.Contains(bar, "⏸") {
		t.Error("transport should show pause icon when stopped")
	}
}

func TestRenderTransportPlaying(t *testing.T) {
	seq := engine.NewSequencer()
	seq.PlayState = engine.Playing
	seq.Position = 7

	bar := RenderTransport(seq, "loops")
	if !strings.Contains(bar, "▶") {
		t.Error("transport should show play icon when playing")
	}
	if !strings.Contains(bar, "Step:08") {
		t.Error("transport should show step 8 at position 7")
	}
	if !strings.Contains(bar, "Bar:3 Beat:4") {
		t.Error("transport should show Bar:3 Beat:4 at position 7")
	}
}

func TestRenderTransportWithSwing(t *testing.T) {
	seq := engine.NewSequencer()
	seq.Project.Swing = 0.5

	bar := RenderTransport(seq, "beatbox")
	if !strings.Contains(bar, "Swing:50%") {
		t.Error("transport should show swing when > 0")
	}
}

func TestRenderTransportNoSwing(t *testing.T) {
	seq := engine.NewSequencer()
	seq.Project.Swing = 0

	bar := RenderTransport(seq, "beatbox")
	if strings.Contains(bar, "Swing") {
		t.Error("transport should not show swing when 0")
	}
}
