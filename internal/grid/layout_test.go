package grid

import (
	"testing"
	"time"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func makeProject(numSteps int) *engine.Project {
	p := engine.DefaultProject()
	p.NumSteps = numSteps
	for i := range p.Tracks {
		p.Tracks[i].Steps = [64]engine.StepState{}
		p.Tracks[i].Muted = false
		p.Tracks[i].Volume = 1.0
	}
	return p
}

func makeGridState(p *engine.Project) GridState {
	return GridState{
		Project:     p,
		PlayState:   engine.Stopped,
		Position:    0,
		CursorTrack: 0,
		TrackOffset: 0,
		StepPage:    0,
		Modifiers:   0,
		FlashTrack:  -1,
	}
}

// ---- stepBrightness ----

func TestStepBrightness(t *testing.T) {
	tests := []struct {
		name       string
		active     bool
		isPlayhead bool
		isBeat     bool
		muted      bool
		want       int
	}{
		{"off", false, false, false, false, ledOff},
		{"on", true, false, false, false, ledOn},
		{"muted overrides active", true, false, false, true, ledMuted},
		{"playhead on empty step", false, true, false, false, ledPlayhead},
		{"active beats playhead", true, true, false, false, ledOn},
		{"beat marker", false, false, true, false, ledBeat},
		{"active beats beat", true, false, true, false, ledOn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stepBrightness(tt.active, tt.isPlayhead, tt.isBeat, tt.muted)
			if got != tt.want {
				t.Errorf("stepBrightness(active=%v, playhead=%v, beat=%v, muted=%v) = %d, want %d",
					tt.active, tt.isPlayhead, tt.isBeat, tt.muted, got, tt.want)
			}
		})
	}
}

// ---- buildLEDGrid ----

func TestBuildLEDGridEmptyProject(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)
	leds := buildLEDGrid(s)

	for row := 0; row < 6; row++ {
		for col := 0; col < 16; col++ {
			expected := ledOff
			if col == 0 {
				expected = ledPlayhead // position 0 always visible
			} else if col%4 == 0 {
				expected = ledBeat
			}
			if leds[col][row] != expected {
				t.Errorf("empty project: leds[%d][%d] = %d, want %d", col, row, leds[col][row], expected)
			}
		}
	}
}

func TestBuildLEDGridActiveStep(t *testing.T) {
	p := makeProject(16)
	p.Tracks[0].Steps[0] = true
	s := makeGridState(p)
	leds := buildLEDGrid(s)

	if leds[0][0] != ledOn {
		t.Errorf("step (0,0) should be on, got %d", leds[0][0])
	}
}

func TestBuildLEDGridPlayheadPosition(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)
	s.Position = 4
	s.PlayState = engine.Playing
	leds := buildLEDGrid(s)

	for row := 0; row < 6; row++ {
		if leds[4][row] != ledPlayhead {
			t.Errorf("playhead at col 4, row %d: got %d, want ledPlayhead %d", row, leds[4][row], ledPlayhead)
		}
	}
}

func TestBuildLEDGridActiveOverridesPlayhead(t *testing.T) {
	p := makeProject(16)
	p.Tracks[0].Steps[4] = true
	s := makeGridState(p)
	s.Position = 4
	s.PlayState = engine.Playing
	leds := buildLEDGrid(s)

	if leds[4][0] != ledOn {
		t.Errorf("active step at playhead: got %d, want ledOn %d", leds[4][0], ledOn)
	}
}

func TestBuildLEDGridMutedTrack(t *testing.T) {
	p := makeProject(16)
	p.Tracks[0].Steps[0] = true
	p.Tracks[0].Muted = true
	s := makeGridState(p)
	leds := buildLEDGrid(s)

	if leds[0][0] != ledMuted {
		t.Errorf("muted active step: got %d, want ledMuted %d", leds[0][0], ledMuted)
	}
}

func TestBuildLEDGridTrackOffset(t *testing.T) {
	p := makeProject(16)
	p.Tracks[2].Steps[0] = true
	s := makeGridState(p)
	s.TrackOffset = 2
	leds := buildLEDGrid(s)

	if leds[0][0] != ledOn {
		t.Errorf("trackOffset=2: leds[0][0] = %d, want ledOn %d", leds[0][0], ledOn)
	}
}

func TestBuildLEDGridStepPage(t *testing.T) {
	p := makeProject(64)
	p.Tracks[0].Steps[20] = true
	s := makeGridState(p)
	s.StepPage = 1
	leds := buildLEDGrid(s)

	if leds[4][0] != ledOn {
		t.Errorf("stepPage=1 step=20: leds[4][0] = %d, want ledOn %d", leds[4][0], ledOn)
	}
}

func TestBuildLEDGridFlashTrack(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)
	s.FlashTrack = 2
	s.FlashUntil = time.Now().Add(time.Second)
	leds := buildLEDGrid(s)

	for col := 0; col < 16; col++ {
		// col 0 is the playhead (position 0) — flash doesn't overwrite playhead.
		expected := ledFlash
		if col == 0 {
			expected = ledPlayhead
		}
		if leds[col][2] != expected {
			t.Errorf("flash row 2, col %d: got %d, want %d", col, leds[col][2], expected)
		}
	}
}

func TestBuildLEDGridFlashExpired(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)
	s.FlashTrack = 2
	s.FlashUntil = time.Now().Add(-time.Second)
	leds := buildLEDGrid(s)

	// Expired flash at col 0: still shows playhead (position 0).
	if leds[0][2] != ledPlayhead {
		t.Errorf("expired flash: leds[0][2] = %d, want ledPlayhead %d", leds[0][2], ledPlayhead)
	}
}

// ---- control row brightness ----

func TestControlRow6StepCountIndicators(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)
	leds := buildLEDGrid(s)

	if leds[0][6] != ledBtnActive {
		t.Errorf("16 steps: col 0 should be active, got %d", leds[0][6])
	}
	if leds[1][6] != ledBtnInact {
		t.Errorf("16 steps: col 1 should be inactive, got %d", leds[1][6])
	}
	if leds[3][6] != ledBtnInact {
		t.Errorf("16 steps: col 3 should be inactive, got %d", leds[3][6])
	}

	p.NumSteps = 64
	leds = buildLEDGrid(s)
	if leds[0][6] != ledBtnInact {
		t.Errorf("64 steps: col 0 should be inactive, got %d", leds[0][6])
	}
	if leds[3][6] != ledBtnActive {
		t.Errorf("64 steps: col 3 should be active, got %d", leds[3][6])
	}

	p.NumSteps = 48
	leds = buildLEDGrid(s)
	if leds[2][6] != ledBtnActive {
		t.Errorf("48 steps: col 2 should be active, got %d", leds[2][6])
	}
}

func TestControlRow6Modifiers(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)

	s.Modifiers = ModBank
	leds := buildLEDGrid(s)
	if leds[12][6] != ledBtnActive {
		t.Errorf("ModBank: col 12 should be active, got %d", leds[12][6])
	}

	s.Modifiers = ModFolder
	leds = buildLEDGrid(s)
	if leds[13][6] != ledBtnActive {
		t.Errorf("ModFolder: col 13 should be active, got %d", leds[13][6])
	}

	s.Modifiers = ModSample
	leds = buildLEDGrid(s)
	if leds[15][6] != ledBtnActive {
		t.Errorf("ModSample: col 15 should be active, got %d", leds[15][6])
	}
}

func TestControlRow7TransportButtons(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)
	s.PlayState = engine.Stopped
	leds := buildLEDGrid(s)

	if leds[0][7] != ledBtnInact {
		t.Errorf("stopped: play button should be inactive, got %d", leds[0][7])
	}

	s.PlayState = engine.Playing
	leds = buildLEDGrid(s)
	if leds[0][7] != ledBtnActive {
		t.Errorf("playing: play button should be active, got %d", leds[0][7])
	}
}

func TestControlRow7MuteIndicator(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)

	leds := buildLEDGrid(s)
	if leds[2][7] != ledUnmuted {
		t.Errorf("unmuted: col 2 should be %d, got %d", ledUnmuted, leds[2][7])
	}

	p.Tracks[0].Muted = true
	leds = buildLEDGrid(s)
	if leds[2][7] != ledMuted {
		t.Errorf("muted: col 2 should be %d, got %d", ledMuted, leds[2][7])
	}
}

func TestControlRow7NavigationButtons(t *testing.T) {
	p := makeProject(64)
	s := makeGridState(p)
	s.CursorTrack = 0
	s.StepPage = 1

	leds := buildLEDGrid(s)

	if leds[14][7] != ledBtnActive {
		t.Errorf("cursorTrack=0: down button should be active, got %d", leds[14][7])
	}
	if leds[13][7] != ledBtnActive {
		t.Errorf("stepPage=1: left button should be active, got %d", leds[13][7])
	}
	if leds[15][7] != ledBtnActive {
		t.Errorf("stepPage=1: right button should be active, got %d", leds[15][7])
	}
}

func TestControlRow7BPMIndicator(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)

	s.Modifiers = ModBPM
	leds := buildLEDGrid(s)
	if leds[12][7] != ledBtnActive {
		t.Errorf("ModBPM: col 12 should be active, got %d", leds[12][7])
	}
}

// ---- boundary conditions ----

func TestBuildLEDGridLastTrack(t *testing.T) {
	p := makeProject(16)
	p.Tracks[7].Steps[0] = true
	s := makeGridState(p)
	s.TrackOffset = 2
	leds := buildLEDGrid(s)

	if leds[0][5] != ledOn {
		t.Errorf("track 7 step 0: got %d, want ledOn %d", leds[0][5], ledOn)
	}
}

func TestBuildLEDGridBeyondNumSteps(t *testing.T) {
	p := makeProject(8)
	p.Tracks[0].Steps[7] = true
	s := makeGridState(p)
	leds := buildLEDGrid(s)

	if leds[7][0] != ledOn {
		t.Errorf("step 7: got %d, want ledOn %d", leds[7][0], ledOn)
	}
	if leds[8][0] != ledOff {
		t.Errorf("col 8 beyond NumSteps: got %d, want ledOff %d", leds[8][0], ledOff)
	}
}

func TestBuildLEDGridModifiersCombined(t *testing.T) {
	p := makeProject(16)
	s := makeGridState(p)
	s.Modifiers = ModBPM | ModBank

	leds := buildLEDGrid(s)
	if leds[12][7] != ledBtnActive {
		t.Errorf("ModBPM: col 12 row 7 should be active, got %d", leds[12][7])
	}
	if leds[12][6] != ledBtnActive {
		t.Errorf("ModBank: col 12 row 6 should be active, got %d", leds[12][6])
	}
}
