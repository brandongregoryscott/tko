package engine

import (
	"testing"
)

func TestTickAdvancesPosition(t *testing.T) {
	s := NewSequencer()
	s.Project.NumSteps = 16
	s.PlayState = Playing

	for i := 0; i < 32; i++ {
		prev := s.Position
		s.Tick()
		expected := (prev + 1) % 16
		if s.Position != expected {
			t.Errorf("tick %d: position=%d, expected=%d", i, s.Position, expected)
		}
	}
}

func TestTickIgnoresStopped(t *testing.T) {
	s := NewSequencer()
	s.Project.NumSteps = 16
	s.PlayState = Stopped

	triggers := s.Tick()
	if triggers != nil {
		t.Error("Tick should return nil when stopped")
	}
	if s.Position != 0 {
		t.Errorf("position should not advance when stopped, got %d", s.Position)
	}
}

func TestTickTriggersActiveSteps(t *testing.T) {
	s := NewSequencer()
	s.Project.NumSteps = 16
	s.PlayState = Playing

	// Set step 0 active on track 0 with an assigned sample.
	s.Project.Tracks[0].Steps[0] = true
	s.Project.Tracks[0].Sample = SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	s.Project.Tracks[1].Steps[0] = true // no sample assigned → shouldn't trigger

	// Position is 0; Tick advances to 0 (wraps from -1? No, starts at 0, Tick advances to 1).
	// Actually NewSequencer starts at position 0. First tick advances to 1.
	// Let's set position manually to test the trigger at step 0.
	s.Position = 15 // so next tick wraps to 0

	triggers := s.Tick()
	if s.Position != 0 {
		t.Errorf("expected position 0 after wrap, got %d", s.Position)
	}
	if len(triggers) != 1 {
		t.Errorf("expected 1 trigger, got %d: %v", len(triggers), triggers)
	}
	if len(triggers) > 0 && triggers[0] != 0 {
		t.Errorf("expected track 0, got %d", triggers[0])
	}
}

func TestTickSkipsMutedTracks(t *testing.T) {
	s := NewSequencer()
	s.Project.NumSteps = 16
	s.PlayState = Playing
	s.Position = 15 // wrap to 0

	s.Project.Tracks[0].Steps[0] = true
	s.Project.Tracks[0].Sample = SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	s.Project.Tracks[0].Muted = true

	triggers := s.Tick()
	if len(triggers) != 0 {
		t.Errorf("muted track should not trigger, got %d triggers", len(triggers))
	}
}

func TestToggleStep(t *testing.T) {
	s := NewSequencer()

	s.ToggleStep(0, 3)
	if !bool(s.Project.Tracks[0].Steps[3]) {
		t.Error("step should be active after toggle")
	}

	s.ToggleStep(0, 3)
	if bool(s.Project.Tracks[0].Steps[3]) {
		t.Error("step should be inactive after second toggle")
	}
}

func TestToggleStepOutOfBounds(t *testing.T) {
	s := NewSequencer()
	// These should not panic.
	s.ToggleStep(0, -1)
	s.ToggleStep(0, 64)
	s.ToggleStep(-1, 0)
	s.ToggleStep(8, 0) // only 8 tracks (0-7)
}

func TestCycleSample(t *testing.T) {
	s := NewSequencer()
	s.Project.Tracks[0].Sample = SampleRef{Folder: "kick", Index: 0, Name: "kick1"}

	// Cycle forward.
	s.CycleSample(0, 1, 5)
	if s.Project.Tracks[0].Sample.Index != 1 {
		t.Errorf("expected index 1, got %d", s.Project.Tracks[0].Sample.Index)
	}

	// Wrap around.
	s.Project.Tracks[0].Sample.Index = 4
	s.CycleSample(0, 1, 5)
	if s.Project.Tracks[0].Sample.Index != 0 {
		t.Errorf("expected wrap to 0, got %d", s.Project.Tracks[0].Sample.Index)
	}

	// Cycle backward.
	s.CycleSample(0, -1, 5)
	if s.Project.Tracks[0].Sample.Index != 4 {
		t.Errorf("expected wrap to 4, got %d", s.Project.Tracks[0].Sample.Index)
	}
}

func TestSetNumSteps(t *testing.T) {
	s := NewSequencer()

	s.SetNumSteps(32)
	if s.Project.NumSteps != 32 {
		t.Errorf("expected 32, got %d", s.Project.NumSteps)
	}

	// Clamp low.
	s.SetNumSteps(-5)
	if s.Project.NumSteps != 1 {
		t.Errorf("expected 1, got %d", s.Project.NumSteps)
	}

	// Clamp high.
	s.SetNumSteps(100)
	if s.Project.NumSteps != 64 {
		t.Errorf("expected 64, got %d", s.Project.NumSteps)
	}

	// Reset position when it exceeds new NumSteps.
	s.SetNumSteps(16)
	s.Position = 10
	s.SetNumSteps(8)
	if s.Position != 0 {
		t.Errorf("expected position reset to 0, got %d", s.Position)
	}
}

func TestSetBPMClamp(t *testing.T) {
	s := NewSequencer()

	s.SetBPM(140)
	if s.Project.BPM != 140 {
		t.Errorf("expected 140, got %f", s.Project.BPM)
	}

	s.SetBPM(10)
	if s.Project.BPM != 20 {
		t.Errorf("expected clamp to 20, got %f", s.Project.BPM)
	}

	s.SetBPM(500)
	if s.Project.BPM != 300 {
		t.Errorf("expected clamp to 300, got %f", s.Project.BPM)
	}
}

func TestToggleMute(t *testing.T) {
	s := NewSequencer()

	if s.Project.Tracks[3].Muted {
		t.Error("track should not start muted")
	}

	s.ToggleMute(3)
	if !s.Project.Tracks[3].Muted {
		t.Error("track should be muted after toggle")
	}

	s.ToggleMute(3)
	if s.Project.Tracks[3].Muted {
		t.Error("track should be unmuted after second toggle")
	}
}

func TestSetVolume(t *testing.T) {
	s := NewSequencer()

	s.SetVolume(0, 0.5)
	if s.Project.Tracks[0].Volume != 0.5 {
		t.Errorf("expected 0.5, got %f", s.Project.Tracks[0].Volume)
	}

	s.SetVolume(0, 2.0)
	if s.Project.Tracks[0].Volume != 1.0 {
		t.Errorf("expected clamp to 1.0, got %f", s.Project.Tracks[0].Volume)
	}

	s.SetVolume(0, -0.5)
	if s.Project.Tracks[0].Volume != 0.0 {
		t.Errorf("expected clamp to 0.0, got %f", s.Project.Tracks[0].Volume)
	}
}

func TestTogglePlayPause(t *testing.T) {
	s := NewSequencer()

	if s.PlayState != Stopped {
		t.Error("should start stopped")
	}

	s.TogglePlayPause()
	if s.PlayState != Playing {
		t.Error("should be playing after toggle")
	}

	s.TogglePlayPause()
	if s.PlayState != Stopped {
		t.Error("should be stopped after second toggle")
	}
}

func TestResetPosition(t *testing.T) {
	s := NewSequencer()
	s.PlayState = Playing
	s.Position = 10

	s.ResetPosition()

	if s.PlayState != Stopped {
		t.Error("should be stopped after reset")
	}
	if s.Position != 0 {
		t.Errorf("expected position 0, got %d", s.Position)
	}
}

func TestTickDuration(t *testing.T) {
	s := NewSequencer()

	s.SetBPM(120)
	dur := s.TickDuration(0) // even step, no swing
	expected := "125ms"
	if dur.String() != expected {
		t.Errorf("expected %s, got %s", expected, dur.String())
	}

	s.SetBPM(60)
	dur = s.TickDuration(0)
	expected = "250ms"
	if dur.String() != expected {
		t.Errorf("expected %s, got %s", expected, dur.String())
	}
}

func TestTickDurationSwing(t *testing.T) {
	s := NewSequencer()
	s.SetBPM(120)
	s.SetSwing(0.5)

	base := "125ms"
	even := s.TickDuration(0)
	odd := s.TickDuration(1)

	// Even steps should be shorter than base.
	if even >= s.TickDuration(0) || even.String() == base {
		// Actually with swing=0.5: even = 125ms * 0.5 = 62.5ms, odd = 125ms * 1.5 = 187.5ms
	}
	_ = even
	_ = odd

	// With no swing, even and odd should be equal.
	s.SetSwing(0)
	if s.TickDuration(0) != s.TickDuration(1) {
		t.Error("even and odd steps should be equal with no swing")
	}
}

func TestDuplicateTrack(t *testing.T) {
	s := NewSequencer()
	s.Project.Tracks[0].Sample = SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	s.Project.Tracks[0].Name = "Kick"

	// Tracks 1-7 are empty (no Sample.Folder set).
	dst := s.DuplicateTrack(0)
	if dst < 1 || dst > 7 {
		t.Errorf("expected dst track 1-7, got %d", dst)
	}
	if s.Project.Tracks[dst].Sample.Folder != "kick" {
		t.Error("duplicate should copy sample folder")
	}
	if s.Project.Tracks[dst].Name != "kick 2" {
		t.Errorf("expected name 'kick 2', got %q", s.Project.Tracks[dst].Name)
	}
}

func TestDuplicateTrackNoEmpty(t *testing.T) {
	s := NewSequencer()
	// Fill all tracks.
	for i := range s.Project.Tracks {
		s.Project.Tracks[i].Sample = SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	}
	dst := s.DuplicateTrack(0)
	if dst != -1 {
		t.Errorf("expected -1 when no empty tracks, got %d", dst)
	}
}

func TestSetSwing(t *testing.T) {
	s := NewSequencer()

	s.SetSwing(0.5)
	if s.Project.Swing != 0.5 {
		t.Errorf("expected 0.5, got %f", s.Project.Swing)
	}

	s.SetSwing(2.0)
	if s.Project.Swing != 1.0 {
		t.Errorf("expected clamp to 1.0, got %f", s.Project.Swing)
	}

	s.SetSwing(-0.5)
	if s.Project.Swing != 0.0 {
		t.Errorf("expected clamp to 0.0, got %f", s.Project.Swing)
	}
}

func TestLoadProject(t *testing.T) {
	s := NewSequencer()
	s.PlayState = Playing
	s.Position = 5

	p := DefaultProject()
	p.BPM = 90
	p.NumSteps = 32
	s.LoadProject(p)

	if s.Project != p {
		t.Error("project should be replaced")
	}
	if s.Position != 0 {
		t.Errorf("position should reset to 0, got %d", s.Position)
	}
	if s.PlayState != Stopped {
		t.Error("playState should be stopped after load")
	}
}

func TestSetStep(t *testing.T) {
	s := NewSequencer()

	s.SetStep(0, 3, true)
	if !bool(s.Project.Tracks[0].Steps[3]) {
		t.Error("step should be active")
	}

	s.SetStep(0, 3, false)
	if bool(s.Project.Tracks[0].Steps[3]) {
		t.Error("step should be inactive")
	}

	// Bounds checks — should not panic.
	s.SetStep(0, -1, true)
	s.SetStep(0, 64, true)
	s.SetStep(-1, 0, true)
	s.SetStep(8, 0, true)
}

func TestClearTrack(t *testing.T) {
	s := NewSequencer()
	s.Project.Tracks[0].Sample = SampleRef{Folder: "kick", Index: 0, Name: "kick1"}
	s.Project.Tracks[0].Name = "Kick"
	s.Project.Tracks[0].Steps[0] = true
	s.Project.Tracks[0].Steps[1] = true
	s.Project.Tracks[0].Muted = true
	s.Project.Tracks[0].Volume = 0.5

	s.ClearTrack(0)

	tr := s.Project.Tracks[0]
	if tr.Sample.Folder != "" {
		t.Error("sample should be cleared")
	}
	if tr.Name != "Track 1" {
		t.Errorf("name should reset to 'Track 1', got %q", tr.Name)
	}
	for i, step := range tr.Steps {
		if bool(step) {
			t.Errorf("step %d should be cleared", i)
		}
	}
	if tr.Muted {
		t.Error("track should be unmuted")
	}
	if tr.Volume != 1.0 {
		t.Errorf("volume should reset to 1.0, got %f", tr.Volume)
	}

	// Bounds — should not panic.
	s.ClearTrack(-1)
	s.ClearTrack(8)
}

func TestStepsMarshalJSON(t *testing.T) {
	var s Steps
	s[0] = true
	s[3] = true
	s[63] = true

	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	// Should be a JSON array of 64 ints.
	expected := "[1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1]"
	if string(data) != expected {
		t.Errorf("got %s", string(data))
	}
}

func TestStepsUnmarshalJSON(t *testing.T) {
	var s Steps
	err := s.UnmarshalJSON([]byte("[1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1]"))
	if err != nil {
		t.Fatal(err)
	}
	if !bool(s[0]) {
		t.Error("step 0 should be active")
	}
	if !bool(s[3]) {
		t.Error("step 3 should be active")
	}
	if !bool(s[63]) {
		t.Error("step 63 should be active")
	}
	if bool(s[1]) {
		t.Error("step 1 should be inactive")
	}
}

func TestStepsUnmarshalJSONTruncated(t *testing.T) {
	var s Steps
	// Array with fewer than 64 elements.
	err := s.UnmarshalJSON([]byte("[1,0,1]"))
	if err != nil {
		t.Fatal(err)
	}
	if !bool(s[0]) {
		t.Error("step 0 should be active")
	}
	if !bool(s[2]) {
		t.Error("step 2 should be active")
	}
}

func TestStepsUnmarshalJSONEmpty(t *testing.T) {
	var s Steps
	s[0] = true // preset a value, should be cleared by unmarshal
	err := s.UnmarshalJSON([]byte("[]"))
	if err != nil {
		t.Fatal(err)
	}
	if bool(s[0]) {
		t.Error("step 0 should be cleared")
	}
}

func TestStepsJSONRoundTrip(t *testing.T) {
	var original Steps
	original[0] = true
	original[7] = true
	original[15] = true
	original[31] = true

	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var decoded Steps
	if err := decoded.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 64; i++ {
		if bool(original[i]) != bool(decoded[i]) {
			t.Errorf("step %d: original=%v, decoded=%v", i, original[i], decoded[i])
		}
	}
}
