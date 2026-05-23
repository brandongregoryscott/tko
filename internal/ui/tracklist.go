package ui

import (
	"fmt"
	"strings"

	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/charmbracelet/lipgloss"
)

const trackPanelWidth = 24

// RenderTrackList returns the track sidebar panel.
func RenderTrackList(proj *engine.Project, cursorTrack int, libSampleCounts map[string]int) string {
	baseStyle := lipgloss.NewStyle().Width(trackPanelWidth).MaxWidth(trackPanelWidth)

	var lines []string
	header := baseStyle.Foreground(ColorGrey).Bold(true).Render(" TRK  NAME         SMP  VOL")
	lines = append(lines, header)
	lines = append(lines, baseStyle.Foreground(ColorDim).Render(strings.Repeat("─", trackPanelWidth)))

	for i := range proj.Tracks {
		t := &proj.Tracks[i]
		muteCh := " "
		if t.Muted {
			muteCh = "M"
		}
		sampleName := "—"
		if t.Sample.Folder != "" {
			count := libSampleCounts[t.Sample.Folder]
			if i == cursorTrack && t.Sample.Name != "" {
				sampleName = t.Sample.Name
			} else {
				sampleName = fmt.Sprintf("%s:%d/%d", truncate(t.Sample.Folder, 5), t.Sample.Index+1, count)
			}
		}
		line := fmt.Sprintf(" %2d  %-11s %s %3.0f%%",
			i+1, truncate(t.Name, 11), muteCh, t.Volume*100)

		style := baseStyle.Foreground(trackColors[i%len(trackColors)])
		if t.Muted {
			style = baseStyle.Foreground(ColorDim)
		}
		if i == cursorTrack {
			style = style.Background(ColorBlue)
			// Prevent the foreground and background colors overlapping, causing the text to be invisible
			if style.GetForeground() == ColorBlue {
				style = style.Foreground(ColorWhite)
			}
		}

		// Sample info line, truncated to fit.
		line2 := fmt.Sprintf("   ↳ %s", truncate(sampleName, trackPanelWidth-5))
		style2 := baseStyle.Foreground(ColorGrey)
		if i == cursorTrack {
			style2 = style2.Background(ColorBlue).Foreground(ColorWhite)
		}

		lines = append(lines, style.Render(line))
		lines = append(lines, style2.Render(line2))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
