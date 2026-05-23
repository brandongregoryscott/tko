package engine

import (
	"encoding/json"
)

// StepState is true when a step is active (triggers sample playback).
type StepState bool

// TrackID is a zero-based index (0-7).
type TrackID int

// SampleRef identifies a sample in the library.
type SampleRef struct {
	Bank   string `json:"bank"`   // bank name (top-level dir under samples/)
	Folder string `json:"folder"` // subdirectory name within the bank
	Index  int    `json:"index"`  // position within the folder
	Name   string `json:"name"`   // display name (filename without extension)
}

// Steps is a 64-step sequencer pattern serialized as an array of 0/1.
type Steps [64]StepState

// MarshalJSON serializes steps as a full 64-element array like [0,0,1,0,...].
func (s Steps) MarshalJSON() ([]byte, error) {
	a := make([]int8, 64)
	for i := 0; i < 64; i++ {
		if s[i] {
			a[i] = 1
		}
	}
	return json.Marshal(a)
}

// UnmarshalJSON parses an array of 0/1 into a Steps array.
func (s *Steps) UnmarshalJSON(data []byte) error {
	*s = Steps{}
	var a []int8
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	for i, v := range a {
		if i >= 64 {
			break
		}
		if v == 1 {
			s[i] = true
		}
	}
	return nil
}

// Track holds the state for one sequencer track.
type Track struct {
	ID     TrackID   `json:"id"`
	Name   string    `json:"name"`
	Sample SampleRef `json:"sample"`
	Steps  Steps     `json:"steps"`
	Volume float64   `json:"volume"` // 0.0 to 1.0
	Muted  bool      `json:"muted"`
}

// PlayState describes whether the sequencer is running.
type PlayState int

const (
	Stopped PlayState = iota
	Playing
)

// Project is the complete top-level state saved to disk.
type Project struct {
	Version  int      `json:"version"`
	BPM      float64  `json:"bpm"`
	Swing    float64  `json:"swing"`
	NumSteps int      `json:"num_steps"`
	Tracks   [8]Track `json:"tracks"`
}

// DefaultProject returns a fresh project with sensible defaults.
func DefaultProject() *Project {
	p := &Project{
		Version:  1,
		BPM:      140,
		Swing:    0,
		NumSteps: 16,
	}
	defaultNames := []string{"Kick", "Snare", "Hi-Hat", "Cymbal", "Perc 1", "Perc 2", "Perc 3", "Perc 4"}
	for i := range p.Tracks {
		p.Tracks[i] = Track{
			ID:     TrackID(i),
			Name:   defaultNames[i],
			Volume: 1.0,
		}
	}
	return p
}
