package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/brandongregoryscott/tko/internal/engine"
)

// DefaultDir returns the directory for saved projects (relative to CWD).
func DefaultDir() string {
	return "projects"
}

// Save writes a project to a JSON file. Creates directories as needed.
func Save(proj *engine.Project, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads a project from a JSON file.
func Load(path string) (*engine.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var proj engine.Project
	if err := json.Unmarshal(data, &proj); err != nil {
		return nil, err
	}
	// Clamp fields to safe ranges on load.
	if proj.BPM < 20 || proj.BPM > 300 {
		proj.BPM = 140
	}
	if proj.NumSteps < 1 || proj.NumSteps > 64 {
		proj.NumSteps = 16
	}
	if proj.Swing < 0 || proj.Swing > 1 {
		proj.Swing = 0
	}
	return &proj, nil
}
