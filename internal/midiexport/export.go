package midiexport

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/brandongregoryscott/tko/internal/engine"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Export writes the project as a Standard MIDI File format 1 to the given path.
func Export(proj *engine.Project, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	clock := smf.MetricTicks(480) // 480 ticks per quarter note
	ticksPerStep := clock.Ticks16th()

	// Track 0: tempo and time signature.
	conductor := smf.Track{}
	conductor.Add(0, smf.MetaTrackSequenceName("tko"))
	conductor.Add(0, smf.MetaMeter(4, 4))
	conductor.Add(0, smf.MetaTempo(proj.BPM))
	conductor.Close(0)

	s := smf.New()
	s.TimeFormat = clock
	if err := s.Add(conductor); err != nil {
		return fmt.Errorf("add conductor track: %w", err)
	}

	// One track per sequencer track that has an assigned sample and at least one active step.
	for i := range proj.Tracks {
		t := &proj.Tracks[i]
		if t.Sample.Folder == "" || !trackHasSteps(t) {
			continue
		}

		note := midiNoteForTrack(i)
		mt := smf.Track{}
		mt.Add(0, smf.MetaTrackSequenceName(t.Name))
		mt.Add(0, smf.MetaInstrument(t.Sample.Name))

		// Write note-on/note-off for each active step.
		// Track.Add uses delta ticks (relative to the previous event), so we track
		// the absolute position and compute deltas.
		var currentTick uint32
		for step := 0; step < proj.NumSteps; step++ {
			if !bool(t.Steps[step]) {
				continue
			}
			vel := uint8(t.Volume * 100)
			if vel < 1 {
				vel = 1
			}
			if vel > 127 {
				vel = 127
			}
			startTick := uint32(step) * ticksPerStep
			// Delta from current position to this note-on.
			if startTick > currentTick {
				mt.Add(startTick-currentTick, midi.NoteOn(0, note, vel))
			} else {
				mt.Add(0, midi.NoteOn(0, note, vel))
			}
			currentTick = startTick

			dur := ticksPerStep / 2
			mt.Add(dur, midi.NoteOff(0, note))
			currentTick += dur
		}
		mt.Close(0)

		if err := s.Add(mt); err != nil {
			return fmt.Errorf("add track %s: %w", t.Name, err)
		}
	}

	return s.WriteFile(path)
}

// DefaultPath returns a timestamped export path.
func DefaultPath() string {
	ts := time.Now().Format("20060102_150405")
	return filepath.Join("projects", fmt.Sprintf("tko_%s.mid", ts))
}

func trackHasSteps(t *engine.Track) bool {
	for _, s := range t.Steps {
		if bool(s) {
			return true
		}
	}
	return false
}

// midiNoteForTrack maps track index to a distinct MIDI note (C3 through G3).
func midiNoteForTrack(idx int) uint8 {
	notes := []uint8{48, 50, 52, 53, 55, 57, 59, 60} // C3, D3, E3, F3, G3, A3, B3, C4
	if idx < 0 || idx >= len(notes) {
		return 60
	}
	return notes[idx]
}
