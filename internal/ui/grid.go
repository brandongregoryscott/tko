package ui

import (
	"fmt"
	"strings"

	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/charmbracelet/lipgloss"
)

// Colors for each track type, cycling through a palette.
var trackColors = []lipgloss.Color{
	ColorGreen,
	ColorRed,
	ColorYellow,
	ColorBlue,
	ColorOrange,
	lipgloss.Color("#FF69B4"),
	lipgloss.Color("#00FFFF"),
	lipgloss.Color("#BB88FF"),
}

// Grid dimensions.
const (
	cellWidth       = 3 // 3 chars per step cell ([ ] or [█])
	trackLabelWidth = 14
)

type gridRender struct {
	Grid   string
	Width  int
	Height int
}

// RenderGrid renders the step grid as a styled string.
func RenderGrid(
	proj *engine.Project,
	cursorTrack, cursorStep int,
	scrollOffset int,
	visibleSteps int,
	playState engine.PlayState,
	position int,
) (string, int) {
	maxScroll := proj.NumSteps - visibleSteps
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollOffset > maxScroll {
		scrollOffset = maxScroll
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	var lines []string

	// Column header (step numbers).
	beatLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")).Bold(true).Width(3).Align(lipgloss.Center)
	header := strings.Repeat(" ", trackLabelWidth+1)
	for s := scrollOffset; s < scrollOffset+visibleSteps && s < proj.NumSteps; s++ {
		isPlayhead := playState == engine.Playing && s == position
		isBeat := s%4 == 0
		label := fmt.Sprintf("%3d", s+1)
		var styled string
		if isPlayhead {
			styled = lipgloss.NewStyle().Foreground(ColorYellow).Width(3).Align(lipgloss.Center).Render(label)
		} else if isBeat {
			styled = beatLabelStyle.Render(label)
		} else {
			styled = GridLabelStyle.Render(label)
		}
		header += styled
	}
	lines = append(lines, header)

	// Divider with thicker marks at beat boundaries.
	var divBuilder strings.Builder
	divBuilder.WriteString(strings.Repeat("─", trackLabelWidth+1))
	for s := scrollOffset; s < scrollOffset+visibleSteps && s < proj.NumSteps; s++ {
		if s%4 == 0 {
			divBuilder.WriteString("━┻━")
		} else {
			divBuilder.WriteString("───")
		}
	}
	lines = append(lines, BeatDividerStyle.Render(divBuilder.String()))

	// Track rows.
	for i := range proj.Tracks {
		t := &proj.Tracks[i]
		color := trackColors[i%len(trackColors)]
		label := fmt.Sprintf(" %-*s", trackLabelWidth-1, truncate(t.Name, trackLabelWidth-2))

		trackStyle := lipgloss.NewStyle().Foreground(color)
		if t.Muted {
			trackStyle = lipgloss.NewStyle().Foreground(ColorDim)
		}
		lines = append(lines, trackStyle.Render(label)+" "+renderStepRow(t, i, cursorTrack, cursorStep, scrollOffset, visibleSteps, proj.NumSteps, playState, position))
	}

	// Scroller hint.
	if proj.NumSteps > visibleSteps {
		pct := float64(scrollOffset) / float64(maxScroll)
		bar := scrollBar(pct, visibleSteps*cellWidth)
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorDim).Render(bar))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...), scrollOffset
}

func renderStepRow(
	t *engine.Track,
	trackIdx, cursorTrack, cursorStep int,
	scrollOffset, visibleSteps, numSteps int,
	playState engine.PlayState,
	position int,
) string {
	var b strings.Builder
	for s := scrollOffset; s < scrollOffset+visibleSteps && s < numSteps; s++ {
		active := bool(t.Steps[s])
		isCursor := trackIdx == cursorTrack && s == cursorStep
		isPlayhead := playState == engine.Playing && s == position
		isActiveAndPlayhead := active && isPlayhead
		isCursorActive := isCursor && active
		isBeat := s%4 == 0

		cell := RenderCell(active, isCursor, isPlayhead, isActiveAndPlayhead, isCursorActive, isBeat)
		b.WriteString(cell)
	}
	return b.String()
}

func scrollBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	pos := int(pct * float64(width-1))
	if pos < 0 {
		pos = 0
	}

	bar := strings.Repeat("─", width)
	runes := []rune(bar)
	if pos < len(runes) {
		runes[pos] = '█'
	}
	return strings.Repeat(" ", trackLabelWidth+1) + string(runes)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
