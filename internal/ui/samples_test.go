package ui

import (
	"testing"

	"github.com/brandongregoryscott/tko/internal/audio"
	"github.com/brandongregoryscott/tko/internal/engine"
)

// testLibrary returns a Library with known test data:
// Bank "beatbox": kick (3 samples), snare (2 samples), hat (4 samples)
func testLibrary() *audio.Library {
	return audio.NewTestLibrary(
		map[string][]string{
			"beatbox": {"hat", "kick", "snare"},
			"loops":   {"bass", "pad"},
		},
		map[string]map[string][]string{
			"beatbox": {
				"kick":  {"kick1", "kick2", "kick3"},
				"snare": {"snare1", "snare2"},
				"hat":   {"hat1", "hat2", "hat3", "hat4"},
			},
			"loops": {
				"bass": {"bass1", "bass2"},
				"pad":  {"pad1"},
			},
		},
	)
}

// testModel creates a Model with a fresh sequencer and the test library.
func testModel() Model {
	lib := testLibrary()
	m := New(engine.NewSequencer(), lib, nil, "", 0)
	m.bank = "beatbox"
	return m
}

func TestCycleSampleForward(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "beatbox", Folder: "kick", Index: 0, Name: "kick1",
	}

	m.cycleSample(1)
	tr := m.sequencer.Project.Tracks[0]
	if tr.Sample.Index != 1 {
		t.Errorf("expected index 1, got %d", tr.Sample.Index)
	}
	if tr.Sample.Name != "kick2" {
		t.Errorf("expected name kick2, got %s", tr.Sample.Name)
	}
}

func TestCycleSampleWraps(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "beatbox", Folder: "kick", Index: 2, Name: "kick3",
	}

	m.cycleSample(1)
	tr := m.sequencer.Project.Tracks[0]
	if tr.Sample.Index != 0 {
		t.Errorf("expected wrap to index 0, got %d", tr.Sample.Index)
	}
	if tr.Sample.Name != "kick1" {
		t.Errorf("expected name kick1, got %s", tr.Sample.Name)
	}
}

func TestCycleSampleBackward(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "beatbox", Folder: "kick", Index: 0, Name: "kick1",
	}

	m.cycleSample(-1)
	tr := m.sequencer.Project.Tracks[0]
	if tr.Sample.Index != 2 {
		t.Errorf("expected wrap to index 2, got %d", tr.Sample.Index)
	}
	if tr.Sample.Name != "kick3" {
		t.Errorf("expected name kick3, got %s", tr.Sample.Name)
	}
}

func TestCycleSampleEmptyFolderCyclesFolder(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	// Track has no folder → should call cycleTrackFolder instead.
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{}

	m.cycleSample(1)
	tr := m.sequencer.Project.Tracks[0]
	if tr.Sample.Folder == "" {
		t.Error("folder should be assigned")
	}
}

func TestCycleSampleNoLibrary(t *testing.T) {
	m := New(engine.NewSequencer(), nil, nil, "", 0)
	m.cursorTrack = 0
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "beatbox", Folder: "kick", Index: 0, Name: "kick1",
	}

	// Should not panic when audioLib is nil.
	m.cycleSample(1)
}

func TestRandomizeSamplePicksRandom(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	track := &m.sequencer.Project.Tracks[0]
	track.Sample = engine.SampleRef{
		Bank: "beatbox", Folder: "kick", Index: 0, Name: "kick1",
	}

	m.randomizeSample()
	if track.Sample.Folder != "kick" {
		t.Errorf("folder should stay kick, got %s", track.Sample.Folder)
	}
	if track.Sample.Name == "" {
		t.Error("should have a sample name")
	}
}

func TestRandomizeSampleNoFolder(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	track := &m.sequencer.Project.Tracks[0]
	track.Sample = engine.SampleRef{}

	m.randomizeSample()
	if track.Sample.Folder == "" {
		t.Error("should assign a random folder")
	}
}

func TestRandomizeSampleNoLibrary(t *testing.T) {
	m := New(engine.NewSequencer(), nil, nil, "", 0)
	m.cursorTrack = 0

	m.randomizeSample()
	if m.statusMsg == "" {
		t.Error("should set status message about no library")
	}
}

func TestRandomizeAllSamples(t *testing.T) {
	m := testModel()
	proj := m.sequencer.Project
	proj.NumSteps = 16

	// Set up tracks with steps and folder assignments.
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "beatbox", Folder: "kick", Index: 0, Name: "kick1"}
	proj.Tracks[0].Steps[0] = true
	proj.Tracks[1].Sample = engine.SampleRef{Bank: "beatbox", Folder: "snare", Index: 0, Name: "snare1"}
	proj.Tracks[1].Steps[0] = true
	// Track 2 has a folder but no steps — should be skipped.
	proj.Tracks[2].Sample = engine.SampleRef{Bank: "beatbox", Folder: "hat", Index: 0, Name: "hat1"}

	m.randomizeAllSamples()

	// Tracks with steps should have been randomized.
	if proj.Tracks[0].Sample.Name == "" {
		t.Error("track 0 should have a sample name")
	}
	if proj.Tracks[1].Sample.Name == "" {
		t.Error("track 1 should have a sample name")
	}
	// Track 2 had no steps — should be unchanged.
	if proj.Tracks[2].Sample.Name == "" {
		t.Error("track 2 sample name should still be set (not randomized)")
	}
}

func TestAutoAssignSamples(t *testing.T) {
	m := testModel()
	m.bank = "beatbox"

	m.autoAssignSamples()

	// Should assign first sample from each folder to a track.
	names := make(map[string]bool)
	for i := 0; i < 3; i++ {
		names[m.sequencer.Project.Tracks[i].Sample.Folder] = true
	}
	if !names["hat"] || !names["kick"] || !names["snare"] {
		t.Errorf("first 3 tracks should have hat, kick, snare; got %v", names)
	}
}

func TestAutoAssignSamplesNoLibrary(t *testing.T) {
	m := New(engine.NewSequencer(), nil, nil, "", 0)
	// Should not panic.
	m.autoAssignSamples()
}

func TestCycleTrackFolder(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	m.bank = "loops"
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "loops", Folder: "bass", Index: 0, Name: "bass1",
	}

	m.cycleTrackFolder()
	tr := m.sequencer.Project.Tracks[0]
	if tr.Sample.Folder != "pad" {
		t.Errorf("expected folder pad, got %s", tr.Sample.Folder)
	}
	if tr.Sample.Name != "pad1" {
		t.Errorf("expected name pad1, got %s", tr.Sample.Name)
	}
}

func TestCycleTrackFolderWraps(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	m.bank = "loops"
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "loops", Folder: "pad", Index: 0, Name: "pad1",
	}

	m.cycleTrackFolder()
	tr := m.sequencer.Project.Tracks[0]
	if tr.Sample.Folder != "bass" {
		t.Errorf("expected wrap to bass, got %s", tr.Sample.Folder)
	}
}

func TestCycleTrackFolderDelta(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	m.bank = "beatbox"
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "beatbox", Folder: "kick", Index: 0, Name: "kick1",
	}

	m.cycleTrackFolderDelta(1)
	if m.sequencer.Project.Tracks[0].Sample.Folder != "snare" {
		t.Errorf("expected folder snare, got %s", m.sequencer.Project.Tracks[0].Sample.Folder)
	}

	m.cycleTrackFolderDelta(-1)
	if m.sequencer.Project.Tracks[0].Sample.Folder != "kick" {
		t.Errorf("expected back to kick, got %s", m.sequencer.Project.Tracks[0].Sample.Folder)
	}
}

func TestCycleTrackFolderDeltaWraps(t *testing.T) {
	m := testModel()
	m.cursorTrack = 0
	m.bank = "beatbox"
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{
		Bank: "beatbox", Folder: "hat", Index: 0, Name: "hat1",
	}

	// "hat" is last alphabetically (hat, kick, snare), so +1 wraps to kick.
	m.cycleTrackFolderDelta(1)
	tr := m.sequencer.Project.Tracks[0]
	if tr.Sample.Folder != "kick" {
		t.Errorf("expected wrap to kick, got %s", tr.Sample.Folder)
	}
}

func TestCycleBank(t *testing.T) {
	m := testModel()
	m.bank = "beatbox"
	proj := m.sequencer.Project
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "beatbox", Folder: "kick", Index: 0, Name: "kick1"}
	proj.Tracks[1].Sample = engine.SampleRef{Bank: "beatbox", Folder: "snare", Index: 0, Name: "snare1"}
	// Track 2: empty folder — should be skipped.

	m.cycleBank(1)

	if m.bank != "loops" {
		t.Errorf("bank should be loops, got %s", m.bank)
	}
	// Track 0: kick doesn't exist in loops, so it gets a fallback folder.
	tr0 := proj.Tracks[0]
	if tr0.Sample.Bank != "loops" {
		t.Errorf("track 0 bank should be loops, got %s", tr0.Sample.Bank)
	}
	// Track 1: snare doesn't exist in loops either, fallback assigned.
	tr1 := proj.Tracks[1]
	if tr1.Sample.Bank != "loops" {
		t.Errorf("track 1 bank should be loops, got %s", tr1.Sample.Bank)
	}
	// Track 2: empty folder, should stay empty.
	tr2 := proj.Tracks[2]
	if tr2.Sample.Folder != "" {
		t.Error("track 2 should remain empty")
	}
}

func TestCycleBankBackward(t *testing.T) {
	m := testModel()
	m.bank = "loops"
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{Bank: "loops", Folder: "bass", Index: 0, Name: "bass1"}

	m.cycleBank(-1)

	if m.bank != "beatbox" {
		t.Errorf("bank should be beatbox, got %s", m.bank)
	}
	tr0 := m.sequencer.Project.Tracks[0]
	if tr0.Sample.Bank != "beatbox" {
		t.Errorf("track 0 bank should be beatbox, got %s", tr0.Sample.Bank)
	}
}

func TestCycleBankSameFolderNamePreserved(t *testing.T) {
	// Create a library where both banks share a folder name.
	lib := audio.NewTestLibrary(
		map[string][]string{
			"bank-a": {"shared"},
			"bank-b": {"shared", "unique-b"},
		},
		map[string]map[string][]string{
			"bank-a": {"shared": {"a1", "a2"}},
			"bank-b": {"shared": {"b1", "b2"}, "unique-b": {"ub1"}},
		},
	)
	m := New(engine.NewSequencer(), lib, nil, "", 0)
	m.bank = "bank-a"
	m.sequencer.Project.Tracks[0].Sample = engine.SampleRef{Bank: "bank-a", Folder: "shared", Index: 1, Name: "a2"}

	m.cycleBank(1)

	tr0 := m.sequencer.Project.Tracks[0]
	if tr0.Sample.Bank != "bank-b" {
		t.Errorf("bank should be bank-b, got %s", tr0.Sample.Bank)
	}
	if tr0.Sample.Folder != "shared" {
		t.Errorf("folder should stay shared, got %s", tr0.Sample.Folder)
	}
	// Index should reset to 0 and name updated from bank-b.
	if tr0.Sample.Name != "b1" {
		t.Errorf("expected name b1, got %s", tr0.Sample.Name)
	}
}
