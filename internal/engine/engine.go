package engine

import (
	"fmt"
	"math"
	"time"
)

// Sequencer owns the runtime state: the project, play state, and current position.
type Sequencer struct {
	Project   *Project
	PlayState PlayState
	Position  int // current step index (0-based), ranges 0..NumSteps-1
}

// NewSequencer creates a sequencer with a default project.
func NewSequencer() *Sequencer {
	return &Sequencer{
		Project:   DefaultProject(),
		PlayState: Stopped,
	}
}

// LoadProject replaces the current project and resets playback.
func (s *Sequencer) LoadProject(p *Project) {
	s.Project = p
	s.Position = 0
	s.PlayState = Stopped
}

// Tick advances the sequencer by one step and returns which tracks should trigger.
// Returns nil when the sequencer is stopped.
func (s *Sequencer) Tick() []TrackID {
	if s.PlayState != Playing {
		return nil
	}
	s.Position = (s.Position + 1) % s.Project.NumSteps

	var triggers []TrackID
	for i := range s.Project.Tracks {
		t := &s.Project.Tracks[i]
		if !t.Muted && bool(t.Steps[s.Position]) && t.Sample.Folder != "" {
			triggers = append(triggers, t.ID)
		}
	}
	return triggers
}

// ToggleStep flips the step state at the given track and step position.
func (s *Sequencer) ToggleStep(track TrackID, step int) {
	if step < 0 || step >= 64 || int(track) < 0 || int(track) >= len(s.Project.Tracks) {
		return
	}
	s.Project.Tracks[track].Steps[step] = !s.Project.Tracks[track].Steps[step]
}

// SetStep explicitly sets the step state.
func (s *Sequencer) SetStep(track TrackID, step int, active bool) {
	if step < 0 || step >= 64 || int(track) < 0 || int(track) >= len(s.Project.Tracks) {
		return
	}
	s.Project.Tracks[track].Steps[step] = StepState(active)
}

// CycleSample advances the sample for a track by delta (+1 forward, -1 backward).
// maxIdx is the number of samples available in the folder (provided by the audio library).
func (s *Sequencer) CycleSample(track TrackID, delta, maxIdx int) {
	if int(track) < 0 || int(track) >= len(s.Project.Tracks) || maxIdx <= 0 {
		return
	}
	t := &s.Project.Tracks[track]
	idx := t.Sample.Index + delta
	if idx < 0 {
		idx = maxIdx - 1
	} else if idx >= maxIdx {
		idx = 0
	}
	t.Sample.Index = idx
}

// SetBPM clamps and sets the tempo.
func (s *Sequencer) SetBPM(bpm float64) {
	s.Project.BPM = math.Max(20, math.Min(300, bpm))
}

// SetNumSteps clamps and sets the number of active steps.
func (s *Sequencer) SetNumSteps(n int) {
	if n < 1 {
		n = 1
	}
	if n > 64 {
		n = 64
	}
	s.Project.NumSteps = n
	if s.Position >= n {
		s.Position = 0
	}
}

// ToggleMute toggles the mute state of a track.
func (s *Sequencer) ToggleMute(track TrackID) {
	if int(track) < 0 || int(track) >= len(s.Project.Tracks) {
		return
	}
	s.Project.Tracks[track].Muted = !s.Project.Tracks[track].Muted
}

// SetVolume sets the volume for a track (clamped 0-1).
func (s *Sequencer) SetVolume(track TrackID, vol float64) {
	if int(track) < 0 || int(track) >= len(s.Project.Tracks) {
		return
	}
	s.Project.Tracks[track].Volume = math.Max(0, math.Min(1, vol))
}

// ResetPosition stops playback and resets to step 0.
func (s *Sequencer) ResetPosition() {
	s.PlayState = Stopped
	s.Position = 0
}

// TogglePlayPause toggles between playing and stopped.
func (s *Sequencer) TogglePlayPause() {
	if s.PlayState == Playing {
		s.PlayState = Stopped
	} else {
		s.PlayState = Playing
	}
}

// TickDuration returns the duration of a step at the current BPM, with swing applied.
// step is the 0-based step index. Swing delays odd-numbered steps and shortens even ones.
func (s *Sequencer) TickDuration(step int) time.Duration {
	base := time.Minute / time.Duration(s.Project.BPM*4)
	if s.Project.Swing > 0 {
		if step%2 == 1 {
			return base + time.Duration(float64(base)*s.Project.Swing)
		}
		return base - time.Duration(float64(base)*s.Project.Swing)
	}
	return base
}

// DuplicateTrack copies the sample assignment from src to the first empty track.
// Returns the destination track ID or -1 if no empty track is available.
func (s *Sequencer) DuplicateTrack(src TrackID) TrackID {
	srcT := &s.Project.Tracks[src]
	if srcT.Sample.Folder == "" {
		return -1
	}
	// Find first empty track.
	for i := range s.Project.Tracks {
		if s.Project.Tracks[i].Sample.Folder == "" {
			dst := &s.Project.Tracks[i]
			dst.Sample = srcT.Sample
			dst.Volume = srcT.Volume
			// Auto-name with number suffix based on count of same-folder tracks.
			count := 0
			for j := range s.Project.Tracks {
				if s.Project.Tracks[j].Sample.Folder == srcT.Sample.Folder {
					count++
				}
			}
			if count > 1 {
				dst.Name = fmt.Sprintf("%s %d", srcT.Sample.Folder, count)
			} else {
				dst.Name = srcT.Sample.Folder
			}
			return TrackID(i)
		}
	}
	return -1
}

// ClearTrack removes the sample assignment and resets all steps for a track.
func (s *Sequencer) ClearTrack(track TrackID) {
	if int(track) < 0 || int(track) >= len(s.Project.Tracks) {
		return
	}
	t := &s.Project.Tracks[track]
	t.Sample = SampleRef{}
	t.Name = fmt.Sprintf("Track %d", int(track)+1)
	for i := range t.Steps {
		t.Steps[i] = false
	}
	t.Muted = false
	t.Volume = 1.0
}

// SetSwing clamps and sets the swing amount (0.0 to 1.0).
func (s *Sequencer) SetSwing(swing float64) {
	s.Project.Swing = math.Max(0, math.Min(1, swing))
}
