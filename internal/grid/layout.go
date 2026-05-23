package grid

import (
	"time"

	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/hypebeast/go-osc/osc"
)

// LED brightness constants (0-15).
const (
	ledOff       = 0
	ledBeat      = 1
	ledCursor    = 4
	ledPlayhead  = 8
	ledOn        = 15
	ledBtnInact  = 4
	ledBtnActive = 15
	ledMuted     = 2
	ledUnmuted   = 12
	ledFlash     = 4 // brightness used for the track-select flash fill
)

// buildLEDGrid computes a [16][8]int brightness array from grid state.
// Columns are X, rows are Y (matching the grid coordinate system).
func buildLEDGrid(s GridState) [16][8]int {
	var leds [16][8]int

	// Rows 0-5: track step grid.
	for row := 0; row < 6; row++ {
		trackIdx := s.TrackOffset + row
		if trackIdx >= len(s.Project.Tracks) {
			break
		}
		t := &s.Project.Tracks[trackIdx]
		for col := 0; col < 16; col++ {
			step := s.StepPage*16 + col
			if step >= s.Project.NumSteps {
				break
			}
			active := bool(t.Steps[step])
			isPlayhead := step == s.Position
			isBeat := step%4 == 0
			leds[col][row] = stepBrightness(active, isPlayhead, isBeat, t.Muted)
		}
	}

	// Track cursor flash: fill the selected track's row with a dim
	// glow so the user can see which track is active.
	if s.FlashTrack >= 0 && timeNow().Before(s.FlashUntil) {
		trackRow := s.FlashTrack - s.TrackOffset
		if trackRow >= 0 && trackRow < 6 {
			for col := 0; col < 16; col++ {
				if leds[col][trackRow] == ledOff || leds[col][trackRow] == ledBeat {
					leds[col][trackRow] = ledFlash
				}
			}
		}
	}

	// Row 6: controls (1 2 3 4 ··· r b F ▲ S).
	for col := 0; col < 16; col++ {
		leds[col][6] = controlRow6Brightness(s, col)
	}

	// Row 7: controls (▶ ■ M ··· R B ◄ ▼ ►).
	for col := 0; col < 16; col++ {
		leds[col][7] = controlRow7Brightness(s, col)
	}

	return leds
}

// stepBrightness returns the LED brightness for a single step cell.
func stepBrightness(active, isPlayhead, isBeat, muted bool) int {
	if active {
		if muted {
			return ledMuted
		}
		return ledOn
	}
	if isPlayhead {
		return ledPlayhead
	}
	if isBeat {
		return ledBeat
	}
	return ledOff
}

// controlRow6Brightness returns brightness for row 6.
// Layout: 1 2 3 4 ··· r b F ▲ S
//
//	col 0-3:  step page buttons
//	col 4-10: reserved
//	col 11:   r — randomize cursor track
//	col 12:   b — bank modifier
//	col 13:   F — folder modifier
//	col 14:   ▲ — cursor track up
//	col 15:   S — sample modifier
func controlRow6Brightness(s GridState, col int) int {
	switch col {
	case 0: // 16 steps
		if s.Project.NumSteps == 16 {
			return ledBtnActive
		}
		return ledBtnInact
	case 1: // 32 steps
		if s.Project.NumSteps == 32 {
			return ledBtnActive
		}
		return ledBtnInact
	case 2: // 48 steps
		if s.Project.NumSteps == 48 {
			return ledBtnActive
		}
		return ledBtnInact
	case 3: // 64 steps
		if s.Project.NumSteps == 64 {
			return ledBtnActive
		}
		return ledBtnInact
	case 11: // r — randomize cursor track
		return ledBtnInact
	case 12: // b — bank modifier
		if s.Modifiers&ModBank != 0 {
			return ledBtnActive
		}
		return ledBtnInact
	case 13: // F — folder modifier
		if s.Modifiers&ModFolder != 0 {
			return ledBtnActive
		}
		return ledBtnInact
	case 14: // ▲ — cursor track up
		if s.CursorTrack > 0 {
			return ledBtnActive
		}
		return ledBtnInact
	case 15: // S — sample modifier
		if s.Modifiers&ModSample != 0 {
			return ledBtnActive
		}
		return ledBtnInact
	default: // cols 4-10 — reserved
		return ledOff
	}
}

// controlRow7Brightness returns brightness for row 7.
// Layout: ▶ ■ M ··· R B ◄ ▼ ►
//
//	col 0:    ▶ — play/pause
//	col 1:    ■ — stop/reset
//	col 2:    M — mute toggle
//	col 3-10: reserved
//	col 11:   R — randomize all tracks
//	col 12:   B — BPM modifier
//	col 13:   ◄ — step page left
//	col 14:   ▼ — cursor track down
//	col 15:   ► — step page right
func controlRow7Brightness(s GridState, col int) int {
	switch col {
	case 0: // ▶ — play/pause
		if s.PlayState == engine.Playing {
			return ledBtnActive
		}
		return ledBtnInact
	case 1: // ■ — stop/reset
		return ledBtnInact
	case 2: // M — mute toggle
		if s.CursorTrack >= 0 && s.CursorTrack < len(s.Project.Tracks) {
			if s.Project.Tracks[s.CursorTrack].Muted {
				return ledMuted
			}
		}
		return ledUnmuted
	case 11: // R — randomize all
		return ledBtnInact
	case 12: // B — BPM modifier
		if s.Modifiers&ModBPM != 0 {
			return ledBtnActive
		}
		return ledBtnInact
	case 13: // ◄ — step page left
		if s.StepPage > 0 {
			return ledBtnActive
		}
		return ledBtnInact
	case 14: // ▼ — cursor track down
		if s.CursorTrack < len(s.Project.Tracks)-1 {
			return ledBtnActive
		}
		return ledBtnInact
	case 15: // ► — step page right
		maxPage := (s.Project.NumSteps - 1) / 16
		if s.StepPage < maxPage {
			return ledBtnActive
		}
		return ledBtnInact
	default: // cols 3-10 — reserved
		return ledOff
	}
}

// sendFullGrid sends the full 16x8 LED state to the grid as two
// <prefix>/grid/led/level/map messages covering the left and right 8x8 quadrants.
func sendFullGrid(client *osc.Client, leds [16][8]int) error {
	ledPath := oscPrefix + "/grid/led/level/map"

	// Left quadrant: cols 0-7.
	left := make([]int32, 64)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			left[y*8+x] = int32(leds[x][y])
		}
	}

	// Right quadrant: cols 8-15.
	right := make([]int32, 64)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			right[y*8+x] = int32(leds[x+8][y])
		}
	}

	args := []interface{}{int32(0), int32(0)}
	for _, v := range left {
		args = append(args, v)
	}
	msg := osc.NewMessage(ledPath, args...)
	if err := client.Send(msg); err != nil {
		return err
	}

	args = []interface{}{int32(8), int32(0)}
	for _, v := range right {
		args = append(args, v)
	}
	msg = osc.NewMessage(ledPath, args...)
	return client.Send(msg)
}

// timeNow is a variable so tests can override it.
var timeNow = func() time.Time { return time.Now() }
