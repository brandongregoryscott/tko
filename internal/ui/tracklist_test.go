package ui

import (
	"strings"
	"testing"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func TestRenderTrackList(t *testing.T) {
	proj := engine.DefaultProject()
	counts := map[string]int{"kick": 5, "snare": 3}

	panel := RenderTrackList(proj, 0, counts)
	if panel == "" {
		t.Fatal("track list should not be empty")
	}

	// Header.
	if !strings.Contains(panel, "TRK") {
		t.Error("track list should have header")
	}

	// Track names.
	if !strings.Contains(panel, truncate(proj.Tracks[0].Name, 11)) {
		t.Error("track list should contain track 0 name")
	}

	// Track numbers.
	if !strings.Contains(panel, " 1  ") {
		t.Error("track list should contain track numbers")
	}
}

func TestRenderTrackListMuted(t *testing.T) {
	proj := engine.DefaultProject()
	proj.Tracks[0].Muted = true

	panel := RenderTrackList(proj, 0, nil)
	if !strings.Contains(panel, "M") {
		t.Error("muted track should show M indicator")
	}
}

func TestRenderTrackListWithSample(t *testing.T) {
	proj := engine.DefaultProject()
	proj.Tracks[0].Sample = engine.SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	counts := map[string]int{"kick": 5}

	panel := RenderTrackList(proj, 0, counts)
	if !strings.Contains(panel, "kick1") {
		t.Error("selected track should show sample name")
	}
}

func TestRenderTrackListUnselectedSample(t *testing.T) {
	proj := engine.DefaultProject()
	proj.Tracks[0].Sample = engine.SampleRef{Folder: "kick", Index: 2, Name: "kick3"}
	proj.Tracks[1].Sample = engine.SampleRef{Folder: "snare", Index: 0, Name: "snare1"}
	counts := map[string]int{"kick": 5, "snare": 3}

	// Cursor on track 1 — track 0 should show folder:idx/count, not name.
	panel := RenderTrackList(proj, 1, counts)
	if !strings.Contains(panel, "kick:3/5") {
		t.Error("unselected track should show folder:idx/count pattern")
	}
}
