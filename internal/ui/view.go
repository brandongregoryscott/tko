package ui

import (
	"fmt"

	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/charmbracelet/lipgloss"
)

// View renders the entire TUI.
func (m Model) View() string {
	if !m.audioReady {
		return m.loadingView()
	}
	if m.showHelp {
		return m.helpView()
	}

	// Transport bar.
	transport := RenderTransport(m.sequencer, m.bank)

	// Track list (left sidebar).
	sampleCounts := map[string]int{}
	if m.audioLib != nil {
		sampleCounts = m.audioLib.Counts(m.bank)
	}
	trackList := RenderTrackList(m.sequencer.Project, m.cursorTrack, sampleCounts)

	// Step grid (main area).
	vis := m.visibleSteps()
	grid, _ := RenderGrid(
		m.sequencer.Project,
		m.cursorTrack, m.cursorStep,
		m.scrollOffset,
		vis,
		m.sequencer.PlayState, m.sequencer.Position,
	)

	// Layout: track list + grid side by side.
	body := lipgloss.JoinHorizontal(lipgloss.Top, trackList, "  ", grid)

	// Status bar.
	status := ""
	if m.statusMsg != "" {
		status = StatusStyle.Render(" " + m.statusMsg)
	}
	playStateLabel := "⏹"
	if m.sequencer.PlayState == engine.Playing {
		playStateLabel = "▶"
	}
	rightLabel := playStateLabel + " ?:help"
	if m.gridStatus != "" {
		rightLabel = m.gridStatus + "  " + rightLabel
	}
	rightStatus := HelpStyle.Render(rightLabel)

	statusBar := lipgloss.JoinHorizontal(lipgloss.Center, status, "  ", rightStatus)

	main := AppStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		transport,
		"",
		body,
		"",
		statusBar,
	))

	// File dialog below main content.
	if m.focus == FocusSaveFile || m.focus == FocusLoadFile {
		return main + "\n" + renderDialog(m)
	}

	return main
}

var dialogBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorBlue).
	Padding(0, 1)

func renderDialog(m Model) string {
	if m.focus == FocusLoadFile {
		return renderLoadDialog(m)
	}
	return renderSaveDialog(m)
}

func renderSaveDialog(m Model) string {
	label := lipgloss.NewStyle().Foreground(ColorYellow).Bold(true).Render("Save project:")
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		label+" "+m.fileInput.View(),
		HelpStyle.Render("Enter to confirm  Esc to cancel"),
	)
	return dialogBoxStyle.Render(content)
}

func renderLoadDialog(m Model) string {
	label := lipgloss.NewStyle().Foreground(ColorYellow).Bold(true).Render("Load project:")

	if len(m.fileList) == 0 {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			label,
			HelpStyle.Render("  (no projects found)"),
			"",
			HelpStyle.Render("Esc to cancel"),
		)
		return dialogBoxStyle.Render(content)
	}

	selectedStyle := lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(ColorWhite)

	var lines []string
	lines = append(lines, label)
	lines = append(lines, "")

	// Show a scrollable window into the file list.
	start := m.fileCursor - 4
	if start < 0 {
		start = 0
	}
	end := start + 10
	if end > len(m.fileList) {
		end = len(m.fileList)
		start = end - 10
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		prefix := "  "
		if i == m.fileCursor {
			prefix = "▶ "
			lines = append(lines, selectedStyle.Render(prefix+m.fileList[i]))
		} else {
			lines = append(lines, normalStyle.Render(prefix+m.fileList[i]))
		}
	}

	lines = append(lines, "")
	lines = append(lines, HelpStyle.Render("↑↓ to select  Enter to load  Esc to cancel"))

	return dialogBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// helpView renders the full help overlay.
func (m Model) helpView() string {
	title := lipgloss.NewStyle().Foreground(ColorYellow).Bold(true).Render("Key Bindings")

	// Render key bindings manually from FullHelp() groups.
	groups := m.keys.FullHelp()
	var sections []string
	keyStyle := lipgloss.NewStyle().Foreground(ColorGreen).Width(10)
	descStyle := lipgloss.NewStyle().Foreground(ColorWhite)
	for _, group := range groups {
		var lines []string
		for _, kb := range group {
			line := keyStyle.Render(kb.Keys()[0]) + " " + descStyle.Render(kb.Help().Desc)
			lines = append(lines, line)
		}
		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	return AppStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.JoinVertical(lipgloss.Left, sections...),
		"",
		HelpStyle.Render("Press ? to close help, q to quit"),
	))
}

// loadingView renders a progress bar while samples are loaded in the background.
func (m Model) loadingView() string {
	if m.loadErr != nil {
		return AppStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render("Error loading samples"),
			"",
			HelpStyle.Render(m.loadErr.Error()),
		))
	}

	total := m.loadTotal
	done := m.loadDone
	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}

	barWidth := 40
	filled := 0
	if total > 0 {
		filled = barWidth * done / total
	}

	var bar string
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	progressStyle := lipgloss.NewStyle().Foreground(ColorGreen)
	titleStyle := lipgloss.NewStyle().Foreground(ColorYellow).Bold(true)
	fileStyle := lipgloss.NewStyle().Foreground(ColorWhite).Italic(true)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		titleStyle.Render("Loading samples..."),
		"",
		progressStyle.Render(fmt.Sprintf("[%s] %d%% (%d/%d)", bar, pct, done, total)),
	)

	if m.loadCurrent != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			"",
			fileStyle.Render(m.loadCurrent),
		)
	}

	return AppStyle.Render(content)
}
