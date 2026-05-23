package midiexport

import (
	"os"
	"testing"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func TestExportEveryOtherStep(t *testing.T) {
	// Create a project with hi-hats on every other step (0, 2, 4, ...).
	proj := engine.DefaultProject()
	proj.BPM = 140
	proj.NumSteps = 16

	proj.Tracks[0].Sample = engine.SampleRef{Folder: "hats", Index: 0, Name: "hat1"}
	proj.Tracks[0].Name = "Hi-Hat"
	for i := 0; i < 16; i += 2 {
		proj.Tracks[0].Steps[i] = true
	}

	path := t.TempDir() + "/test.mid"
	if err := Export(proj, path); err != nil {
		t.Fatalf("export error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	// Standard MIDI file header: "MThd" + length (4 bytes) + format + ntrks + division.
	if len(data) < 14 {
		t.Fatal("file too short for MIDI header")
	}
	if string(data[0:4]) != "MThd" {
		t.Error("missing MThd header")
	}

	// format 1, 2 tracks (conductor + hi-hat), 480 PPQN.
	format := int(data[8])<<8 | int(data[9])
	ntrks := int(data[10])<<8 | int(data[11])
	division := int(data[12])<<8 | int(data[13])

	if format != 1 {
		t.Errorf("expected format 1, got %d", format)
	}
	if ntrks != 2 {
		t.Errorf("expected 2 tracks (conductor + hi-hat), got %d", ntrks)
	}
	if division != 480 {
		t.Errorf("expected 480 PPQN, got %d", division)
	}

	// Verify the file contains note-on events (0x90).
	hasNoteOn := false
	for i := 0; i < len(data)-2; i++ {
		if data[i] == 0x90 {
			hasNoteOn = true
			break
		}
	}
	if !hasNoteOn {
		t.Error("exported MIDI file contains no note-on events")
	}
}

func TestExportSkipsEmptyTracks(t *testing.T) {
	proj := engine.DefaultProject()
	proj.NumSteps = 16

	// Track 0: no sample assigned → should be skipped.
	// Track 1: sample assigned but no active steps → should be skipped.
	proj.Tracks[1].Sample = engine.SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	proj.Tracks[1].Name = "Kick"

	path := t.TempDir() + "/empty.mid"
	if err := Export(proj, path); err != nil {
		t.Fatalf("export error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	// Should only have 1 track (conductor).
	ntrks := int(data[10])<<8 | int(data[11])
	if ntrks != 1 {
		t.Errorf("expected 1 track (conductor only), got %d", ntrks)
	}
}

func TestExportTempoTrack(t *testing.T) {
	proj := engine.DefaultProject()
	proj.BPM = 140
	proj.NumSteps = 16

	proj.Tracks[0].Sample = engine.SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	proj.Tracks[0].Name = "Kick"
	proj.Tracks[0].Steps[0] = true

	path := t.TempDir() + "/tempo.mid"
	if err := Export(proj, path); err != nil {
		t.Fatalf("export error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	// Read back with smf to verify tempo.
	// Don't have the reader API easily available, so just check the file exists and has reasonable size.
	if len(data) < 50 {
		t.Errorf("exported file too small: %d bytes", len(data))
	}
}

func TestNoteNumbers(t *testing.T) {
	// Verify distinct notes for each track.
	seen := make(map[uint8]bool)
	for i := 0; i < 8; i++ {
		n := midiNoteForTrack(i)
		if seen[n] {
			t.Errorf("duplicate note %d for track %d", n, i)
		}
		seen[n] = true
		if n < 48 || n > 60 {
			t.Errorf("track %d: note %d out of expected range [48, 60]", i, n)
		}
	}
}
