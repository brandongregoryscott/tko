package ui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	ColorGreen  = lipgloss.Color("#00FF00")
	ColorRed    = lipgloss.Color("#FF4444")
	ColorBlue   = lipgloss.Color("#4466FF")
	ColorGrey   = lipgloss.Color("#666666")
	ColorDim    = lipgloss.Color("#333333")
	ColorWhite  = lipgloss.Color("#FFFFFF")
	ColorBlack  = lipgloss.Color("#000000")
	ColorYellow = lipgloss.Color("#FFAA00")
	ColorOrange = lipgloss.Color("#FF8800")
)

// Cell styles for the step grid.
var (
	cellActive         = lipgloss.NewStyle().Foreground(ColorGreen)
	cellInactive       = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	cellCursor         = lipgloss.NewStyle().Background(ColorBlue).Foreground(ColorWhite)
	cellPlayhead       = lipgloss.NewStyle().Background(lipgloss.Color("#2A2A2A"))
	cellActivePlayhead = lipgloss.NewStyle().Foreground(ColorGreen).Background(lipgloss.Color("#2A2A2A"))
	cellCursorActive   = lipgloss.NewStyle().Background(lipgloss.Color("#6688FF")).Foreground(ColorWhite)
	cellBeat           = lipgloss.NewStyle().Background(lipgloss.Color("#222222"))
	cellActiveBeat     = lipgloss.NewStyle().Foreground(ColorGreen).Background(lipgloss.Color("#222222"))
)

// BeatDividerStyle is used for the beat boundary character in the divider row.
var BeatDividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

// Header and layout styles.
var (
	AppStyle       = lipgloss.NewStyle().Padding(1, 0)
	TransportStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).BorderForeground(ColorBlue)
	HelpStyle      = lipgloss.NewStyle().Foreground(ColorGrey)
	StatusStyle    = lipgloss.NewStyle().Foreground(ColorYellow)
	TrackStyle     = lipgloss.NewStyle().Padding(0, 1).Width(18)
	TrackSelected  = lipgloss.NewStyle().Padding(0, 1).Width(18).Background(ColorBlue).Foreground(ColorWhite)
	TrackMuted     = lipgloss.NewStyle().Padding(0, 1).Width(18).Foreground(ColorDim)
	GridLabelStyle = lipgloss.NewStyle().Foreground(ColorGrey).Width(3).Align(lipgloss.Center)
)

// RenderCell returns a styled string for a step cell (3 chars).
func RenderCell(active bool, isCursor bool, isPlayhead bool, isActiveAndPlayhead bool, isCursorActive bool, isBeat bool) string {
	ch := "[ ]"
	if active {
		ch = "[█]"
	}

	if isCursorActive {
		return cellCursorActive.Render(ch)
	}
	if isCursor {
		return cellCursor.Render(ch)
	}
	if isActiveAndPlayhead {
		return cellActivePlayhead.Render(ch)
	}
	if active {
		if isBeat {
			return cellActiveBeat.Render(ch)
		}
		return cellActive.Render(ch)
	}
	if isPlayhead {
		return cellPlayhead.Render(ch)
	}
	if isBeat {
		return cellBeat.Render(ch)
	}
	return cellInactive.Render(ch)
}
