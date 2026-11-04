package ui

import (
	"path/filepath"
	"testing"

	"github.com/brandongregoryscott/tko/internal/audio"
	"github.com/brandongregoryscott/tko/internal/engine"
	"github.com/brandongregoryscott/tko/internal/persistence"
	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadedProjectPreservesSampleNamesAfterReordering(t *testing.T) {
	for _, startup := range []bool{false, true} {
		name := "file dialog"
		if startup {
			name = "startup"
		}
		t.Run(name, func(t *testing.T) {
			lib := audio.NewTestLibrary(map[string][]string{"kit": {"snare"}}, map[string]map[string][]string{
				"kit": {"snare": {"Snare (2)", "Snare (7)", "Snare (69)", "Snare (70)"}},
			})
			proj := engine.DefaultProject()
			// Snare (69) occupied index 1 in alphabetical order.
			proj.Tracks[0].Sample = engine.SampleRef{Bank: "kit", Folder: "snare", Index: 1, Name: "Snare (69)"}
			// Legacy references without names still resolve by index.
			proj.Tracks[1].Sample = engine.SampleRef{Bank: "kit", Folder: "snare", Index: 0}
			m := New(engine.NewSequencer(), lib, nil, "", 44100)
			var updated tea.Model
			if startup {
				t.Chdir(t.TempDir())
				if err := persistence.Save(proj, filepath.Join("projects", "default.json")); err != nil {
					t.Fatal(err)
				}
				updated, _ = m.Update(LibraryLoadedMsg{Lib: lib})
			} else {
				updated, _ = m.Update(ProjectLoadedMsg{Project: proj})
			}
			m = updated.(Model)
			if got := m.sequencer.Project.Tracks[0].Sample; got.Name != "Snare (69)" || got.Index != 2 {
				t.Fatalf("saved sample changed: %+v", got)
			}
			if got := m.sequencer.Project.Tracks[1].Sample; got.Name != "Snare (2)" || got.Index != 0 {
				t.Fatalf("index fallback failed: %+v", got)
			}
			m.cycleSample(1)
			if got := m.sequencer.Project.Tracks[0].Sample.Name; got != "Snare (70)" {
				t.Fatalf("next sample = %q, want Snare (70)", got)
			}
			m.cycleSample(-1)
			if got := m.sequencer.Project.Tracks[0].Sample.Name; got != "Snare (69)" {
				t.Fatalf("previous sample = %q, want Snare (69)", got)
			}
		})
	}
}
