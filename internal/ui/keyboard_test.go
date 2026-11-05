package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/brandongregoryscott/tko/internal/audio"
	"github.com/brandongregoryscott/tko/internal/engine"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpKeyDoesNotCycleSamples(t *testing.T) {
	m := testModel()
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{Bank: "beatbox", Folder: "kick", Index: 1, Name: "kick2"}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	if key.Matches(msg, m.keys.CyclePrev) {
		t.Error("? is bound to previous sample as well as help")
	}
	for _, wantHelp := range []bool{true, false} {
		updated, _ := m.handleKeyMsg(msg)
		m = updated.(Model)
		if m.showHelp != wantHelp {
			t.Errorf("showHelp = %v, want %v", m.showHelp, wantHelp)
		}
		if got := m.sequencer.Project.Tracks[0].Sample.Index; got != 1 {
			t.Errorf("help changed sample to %d", got)
		}
	}
}

func TestSampleCycleKeys(t *testing.T) {
	for _, tt := range []struct {
		key  string
		want int
	}{
		{"]", 5}, {"[", 3}, {"}", 14}, {"{", 19},
	} {
		t.Run(tt.key, func(t *testing.T) {
			m := testModel()
			names := make([]string, 25)
			for i := range names {
				names[i] = fmt.Sprintf("kick%d", i+1)
			}
			m.audioLib = audio.NewTestLibrary(map[string][]string{"beatbox": {"kick"}}, map[string]map[string][]string{"beatbox": {"kick": names}})
			m.showHelp = true
			m.cursorTrack = 1
			m.sequencer.Project.Tracks[1].Sample = engine.SampleRef{Bank: "beatbox", Folder: "kick", Index: 4, Name: "kick5"}
			updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			m = updated.(Model)
			if got := m.sequencer.Project.Tracks[1].Sample.Index; got != tt.want {
				t.Errorf("sample index = %d, want %d", got, tt.want)
			}
			if m.showHelp {
				t.Error("sample cycling left help open, hiding the updated sample")
			}
		})
	}
}

func TestHelpShowsSampleCycleKeys(t *testing.T) {
	m := testModel()
	help := m.helpView()
	for _, want := range []string{"[", "]", "{", "}", "samples +10", "samples -10"} {
		if !strings.Contains(help, want) {
			t.Errorf("help does not show sample binding %q", want)
		}
	}
}

func TestOldSampleAliasesDoNothing(t *testing.T) {
	for _, input := range []string{"/", "U"} {
		m := testModel()
		before := *m.sequencer.Project
		updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(input)})
		m = updated.(Model)
		if *m.sequencer.Project != before || m.showHelp {
			t.Errorf("removed alias %q still changes the model", input)
		}
	}
}
