package ui

import (
	"testing"
)

func TestDefaultKeyMapHasAllBindings(t *testing.T) {
	km := DefaultKeyMap()

	// Verify all key bindings are non-nil.
	bindings := map[string]bool{
		"Quit":         len(km.Quit.Keys()) > 0,
		"PlayPause":    len(km.PlayPause.Keys()) > 0,
		"ResetPos":     len(km.ResetPos.Keys()) > 0,
		"Save":         len(km.Save.Keys()) > 0,
		"Load":         len(km.Load.Keys()) > 0,
		"ExportMIDI":   len(km.ExportMIDI.Keys()) > 0,
		"HelpToggle":   len(km.HelpToggle.Keys()) > 0,
		"ToggleStep":   len(km.ToggleStep.Keys()) > 0,
		"CursorUp":     len(km.CursorUp.Keys()) > 0,
		"CursorDown":   len(km.CursorDown.Keys()) > 0,
		"CursorLeft":   len(km.CursorLeft.Keys()) > 0,
		"CursorRight":  len(km.CursorRight.Keys()) > 0,
		"CycleNext":    len(km.CycleNext.Keys()) > 0,
		"CyclePrev":    len(km.CyclePrev.Keys()) > 0,
		"RandomSample": len(km.RandomSample.Keys()) > 0,
		"MuteTrack":    len(km.MuteTrack.Keys()) > 0,
		"VolUp":        len(km.VolUp.Keys()) > 0,
		"VolDown":      len(km.VolDown.Keys()) > 0,
		"DupTrack":     len(km.DupTrack.Keys()) > 0,
		"ClearTrack":   len(km.ClearTrack.Keys()) > 0,
		"CycleFolder":  len(km.CycleFolder.Keys()) > 0,
	}
	for name, ok := range bindings {
		if !ok {
			t.Errorf("%s binding is empty", name)
		}
	}
}

func TestShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.ShortHelp()
	if len(help) == 0 {
		t.Error("ShortHelp should return bindings")
	}
}

func TestFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.FullHelp()
	if len(help) == 0 {
		t.Error("FullHelp should return bindings")
	}
}
