package audio

import (
	"testing"
)

func makeTestLib() *Library {
	return NewTestLibrary(
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

func TestBanks(t *testing.T) {
	lib := makeTestLib()
	banks := lib.Banks()
	if len(banks) != 2 {
		t.Errorf("expected 2 banks, got %d", len(banks))
	}
	if banks[0] != "beatbox" || banks[1] != "loops" {
		t.Errorf("banks not sorted: got %v", banks)
	}
}

func TestFolders(t *testing.T) {
	lib := makeTestLib()

	folders := lib.Folders("beatbox")
	if len(folders) != 3 {
		t.Errorf("expected 3 folders, got %d", len(folders))
	}
	if folders[0] != "hat" || folders[1] != "kick" || folders[2] != "snare" {
		t.Errorf("folders not sorted: got %v", folders)
	}

	if lib.Folders("nonexistent") != nil {
		t.Error("nonexistent bank should return nil")
	}
}

func TestNumSamples(t *testing.T) {
	lib := makeTestLib()

	if n := lib.NumSamples("beatbox", "kick"); n != 3 {
		t.Errorf("kick: got %d, want 3", n)
	}
	if n := lib.NumSamples("beatbox", "snare"); n != 2 {
		t.Errorf("snare: got %d, want 2", n)
	}
	if n := lib.NumSamples("beatbox", "hat"); n != 4 {
		t.Errorf("hat: got %d, want 4", n)
	}
	if n := lib.NumSamples("nonexistent", "kick"); n != 0 {
		t.Errorf("nonexistent bank: got %d, want 0", n)
	}
	if n := lib.NumSamples("beatbox", "nonexistent"); n != 0 {
		t.Errorf("nonexistent folder: got %d, want 0", n)
	}
}

func TestSampleName(t *testing.T) {
	lib := makeTestLib()

	if name := lib.SampleName("beatbox", "kick", 0); name != "kick1" {
		t.Errorf("got %q, want kick1", name)
	}
	if name := lib.SampleName("beatbox", "kick", 2); name != "kick3" {
		t.Errorf("got %q, want kick3", name)
	}
	if name := lib.SampleName("beatbox", "kick", -1); name != "" {
		t.Errorf("negative index: got %q, want empty", name)
	}
	if name := lib.SampleName("beatbox", "kick", 99); name != "" {
		t.Errorf("out of bounds index: got %q, want empty", name)
	}
	if name := lib.SampleName("nonexistent", "kick", 0); name != "" {
		t.Errorf("nonexistent bank: got %q, want empty", name)
	}
}

func TestCounts(t *testing.T) {
	lib := makeTestLib()
	counts := lib.Counts("beatbox")
	if len(counts) != 3 {
		t.Errorf("expected 3 entries, got %d", len(counts))
	}
	if counts["kick"] != 3 {
		t.Errorf("kick: got %d, want 3", counts["kick"])
	}
	if counts["snare"] != 2 {
		t.Errorf("snare: got %d, want 2", counts["snare"])
	}
	if counts["hat"] != 4 {
		t.Errorf("hat: got %d, want 4", counts["hat"])
	}

	if counts := lib.Counts("nonexistent"); len(counts) != 0 {
		t.Error("nonexistent bank should return empty map")
	}
}

func TestBuffer(t *testing.T) {
	lib := makeTestLib()

	// Buffers are pre-allocated with correct lengths but nil entries.
	if buf := lib.Buffer("beatbox", "kick", 0); buf != nil {
		t.Error("test library buffers should be nil (pre-allocated slots)")
	}
	if buf := lib.Buffer("beatbox", "kick", -1); buf != nil {
		t.Error("negative index should return nil")
	}
	if buf := lib.Buffer("beatbox", "kick", 99); buf != nil {
		t.Error("out of bounds index should return nil")
	}
	if buf := lib.Buffer("nonexistent", "kick", 0); buf != nil {
		t.Error("nonexistent bank should return nil")
	}
}

func TestNewLibraryLoadsSamples(t *testing.T) {
	lib, err := NewLibrary("testdata/library", 44100)
	if err != nil {
		t.Fatal(err)
	}

	banks := lib.Banks()
	if len(banks) != 1 || banks[0] != "samplified-flodelity" {
		t.Errorf("banks: got %v, want [samplified-flodelity]", banks)
	}

	folders := lib.Folders("samplified-flodelity")
	if len(folders) != 2 {
		t.Errorf("folders: got %d, want 2 (%v)", len(folders), folders)
	}

	// Verify names are loaded (stripped of .wav extension).
	if name := lib.SampleName("samplified-flodelity", "kicks", 0); name != "FLO - Kick 01" {
		t.Errorf("kick name: got %q, want 'FLO - Kick 01'", name)
	}
	if name := lib.SampleName("samplified-flodelity", "open-hats", 0); name != "FLO - Open Hat 01" {
		t.Errorf("open-hat name: got %q, want 'FLO - Open Hat 01'", name)
	}

	// Verify sample counts.
	if n := lib.NumSamples("samplified-flodelity", "kicks"); n != 1 {
		t.Errorf("kicks count: got %d, want 1", n)
	}
	if n := lib.NumSamples("samplified-flodelity", "open-hats"); n != 1 {
		t.Errorf("open-hats count: got %d, want 1", n)
	}

	// Buffers should be real, non-nil beep.Buffer values.
	if buf := lib.Buffer("samplified-flodelity", "kicks", 0); buf == nil {
		t.Error("kick buffer should not be nil")
	}
	if buf := lib.Buffer("samplified-flodelity", "open-hats", 0); buf == nil {
		t.Error("open-hat buffer should not be nil")
	}

	// Counts.
	counts := lib.Counts("samplified-flodelity")
	if len(counts) != 2 {
		t.Errorf("counts: got %d entries, want 2", len(counts))
	}
}

func TestNewLibraryWithProgress(t *testing.T) {
	var progressCalls []LoadProgress
	lib, err := NewLibraryWithProgress("testdata/library", 44100, func(p LoadProgress) {
		progressCalls = append(progressCalls, p)
	})
	if err != nil {
		t.Fatal(err)
	}

	if lib == nil {
		t.Fatal("lib should not be nil")
	}
	if len(progressCalls) != 2 {
		t.Errorf("expected 2 progress calls, got %d", len(progressCalls))
	}
	last := progressCalls[len(progressCalls)-1]
	if last.Loaded != 2 || last.Total != 2 {
		t.Errorf("final progress: loaded=%d total=%d, want 2/2", last.Loaded, last.Total)
	}
}

func TestNewLibraryNonexistentDir(t *testing.T) {
	_, err := NewLibrary("testdata/nonexistent", 44100)
	if err == nil {
		t.Error("should return error for nonexistent directory")
	}
}
