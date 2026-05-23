package persistence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-project.json")

	proj := engine.DefaultProject()
	proj.BPM = 128
	proj.NumSteps = 32
	proj.Swing = 0.3
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "beatbox", Folder: "kick", Index: 2, Name: "kick3"}
	proj.Tracks[0].Steps[0] = true
	proj.Tracks[0].Steps[3] = true
	proj.Tracks[0].Volume = 0.7
	proj.Tracks[0].Muted = true

	if err := Save(proj, path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.BPM != 128 {
		t.Errorf("BPM: got %f, want 128", loaded.BPM)
	}
	if loaded.NumSteps != 32 {
		t.Errorf("NumSteps: got %d, want 32", loaded.NumSteps)
	}
	if loaded.Swing != 0.3 {
		t.Errorf("Swing: got %f, want 0.3", loaded.Swing)
	}
	tr := loaded.Tracks[0]
	if tr.Sample.Folder != "kick" || tr.Sample.Index != 2 || tr.Sample.Name != "kick3" {
		t.Errorf("Sample: got %+v", tr.Sample)
	}
	if !bool(tr.Steps[0]) || !bool(tr.Steps[3]) || bool(tr.Steps[1]) {
		t.Error("Steps not preserved")
	}
	if tr.Volume != 0.7 {
		t.Errorf("Volume: got %f", tr.Volume)
	}
	if !tr.Muted {
		t.Error("Muted not preserved")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/project.json")
	if err == nil {
		t.Error("should return error for nonexistent file")
	}
}

func TestLoadClampsOutOfRangeValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	// Write a project with out-of-range values manually.
	raw := `{
  "version": 1,
  "bpm": 999,
  "swing": 5.0,
  "num_steps": 0,
  "tracks": [
    {"id": 0, "name": "Track 1", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false},
    {"id": 1, "name": "Track 2", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false},
    {"id": 2, "name": "Track 3", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false},
    {"id": 3, "name": "Track 4", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false},
    {"id": 4, "name": "Track 5", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false},
    {"id": 5, "name": "Track 6", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false},
    {"id": 6, "name": "Track 7", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false},
    {"id": 7, "name": "Track 8", "sample": {"bank": "", "folder": "", "index": 0, "name": ""}, "steps": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0], "volume": 1.0, "muted": false}
  ]
}`
	if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.BPM != 140 {
		t.Errorf("BPM should be clamped to 140, got %f", loaded.BPM)
	}
	if loaded.Swing != 0 {
		t.Errorf("Swing should be clamped to 0, got %f", loaded.Swing)
	}
	if loaded.NumSteps != 16 {
		t.Errorf("NumSteps should be clamped to 16, got %d", loaded.NumSteps)
	}
}

func TestSaveCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "project.json")

	proj := engine.DefaultProject()
	if err := Save(proj, path); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Error("file should exist after save")
	}
}
