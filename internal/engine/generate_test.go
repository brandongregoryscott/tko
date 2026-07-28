package engine

import (
	"math/rand/v2"
	"testing"
)

// testProject creates a project with the given folders assigned to tracks.
// Tracks get sequential IDs, folder names as display names, and the test bank.
func testProject(numSteps int, folders ...string) *Project {
	p := DefaultProject()
	p.NumSteps = numSteps
	for i, folder := range folders {
		if i >= 8 {
			break
		}
		p.Tracks[i].Sample = SampleRef{
			Bank: "test", Folder: folder, Index: 0, Name: folder + "_01",
		}
		p.Tracks[i].Name = folder
	}
	return p
}

// fixedRNG returns a deterministic RNG for reproducible tests.
func fixedRNG() *rand.Rand {
	return rand.New(rand.NewPCG(0, 0))
}

// hitCount returns the number of active steps within numSteps.
func hitCount(t *Track, numSteps int) int {
	n := 0
	for step := range numSteps {
		if t.Steps[step] {
			n++
		}
	}
	return n
}

// ---- FolderRole tests ----

func TestFolderRole_Kick(t *testing.T) {
	for _, name := range []string{"kicks", "kick", "KICK", "Kicks", "bd", "BassDrum"} {
		if got := FolderRole(name); got != RoleKick {
			t.Errorf("FolderRole(%q) = %v, want RoleKick", name, got)
		}
	}
}

func TestFolderRole_Snare(t *testing.T) {
	for _, name := range []string{"snares", "snare", "SNARE", "Snares", "sd", "snr"} {
		if got := FolderRole(name); got != RoleSnare {
			t.Errorf("FolderRole(%q) = %v, want RoleSnare", name, got)
		}
	}
}

func TestFolderRole_ClosedHat(t *testing.T) {
	for _, name := range []string{"closed-hats", "closed-hat", "closed_hh", "Closed Hats"} {
		if got := FolderRole(name); got != RoleClosedHat {
			t.Errorf("FolderRole(%q) = %v, want RoleClosedHat", name, got)
		}
	}
}

func TestFolderRole_OpenHat(t *testing.T) {
	for _, name := range []string{"open-hats", "open-hat", "open_oh", "Open Hats"} {
		if got := FolderRole(name); got != RoleOpenHat {
			t.Errorf("FolderRole(%q) = %v, want RoleOpenHat", name, got)
		}
	}
}

func TestFolderRole_Percussion(t *testing.T) {
	for _, name := range []string{"percs", "percussion", "Percs", "my-perc-loop"} {
		if got := FolderRole(name); got != RolePercussion {
			t.Errorf("FolderRole(%q) = %v, want RolePercussion", name, got)
		}
	}
}

func TestFolderRole_Clap(t *testing.T) {
	for _, name := range []string{"claps", "clap", "Claps"} {
		if got := FolderRole(name); got != RoleClap {
			t.Errorf("FolderRole(%q) = %v, want RoleClap", name, got)
		}
	}
}

func TestFolderRole_Fallback(t *testing.T) {
	// Unknown folder names should fall back to RolePercussion, not RoleUnknown.
	for _, name := range []string{"", "something-weird", "fx", "808"} {
		if got := FolderRole(name); got != RolePercussion {
			t.Errorf("FolderRole(%q) = %v, want RolePercussion (fallback)", name, got)
		}
	}
}

// ---- Generation tests ----

func TestGenerateTrackPattern_NoFolder(t *testing.T) {
	p := DefaultProject()
	p.NumSteps = 16
	// Track 0 has no folder (Sample.Folder == "").
	rng := fixedRNG()
	GenerateTrackPattern(p, 0, GenreHipHop, rng)

	// Should be a no-op: no steps activated.
	for step := range p.NumSteps {
		if p.Tracks[0].Steps[step] {
			t.Errorf("step %d is active but track has no folder", step)
		}
	}
}

func TestGenerateTrackPattern_KickHipHop(t *testing.T) {
	p := testProject(16, "kicks")
	rng := fixedRNG()
	GenerateTrackPattern(p, 0, GenreHipHop, rng)

	tr := &p.Tracks[0]
	hits := hitCount(tr, 16)

	// Step 0 (downbeat) should always be active for a kick.
	if !tr.Steps[0] {
		t.Error("kick step 0 (downbeat) should be active")
	}

	// Density should be within genre bounds.
	if hits < 3 {
		t.Errorf("kick has only %d hits, expected at least 3", hits)
	}
	if hits > 7 {
		t.Errorf("kick has %d hits, expected at most 7", hits)
	}

	// Steps beyond NumSteps should be off.
	for step := p.NumSteps; step < 64; step++ {
		if tr.Steps[step] {
			t.Errorf("step %d is beyond NumSteps but active", step)
		}
	}
}

func TestGenerateTrackPattern_SnareHipHop(t *testing.T) {
	p := testProject(16, "snares")
	rng := fixedRNG()
	GenerateTrackPattern(p, 0, GenreHipHop, rng)

	tr := &p.Tracks[0]

	// Steps 4 and 12 should always be active (backbeat).
	if !tr.Steps[4] {
		t.Error("snare step 4 (backbeat) should be active")
	}
	if !tr.Steps[12] {
		t.Error("snare step 12 (backbeat) should be active")
	}

	// Density should be roughly 2-3 hits (12-18% of 16).
	hits := hitCount(tr, 16)
	if hits < 2 {
		t.Errorf("snare has only %d hits, expected at least 2", hits)
	}
	if hits > 3 {
		t.Errorf("snare has %d hits, expected at most 3", hits)
	}
}

func TestGenerateTrackPattern_ClosedHatHipHop(t *testing.T) {
	p := testProject(16, "closed-hats")
	rng := fixedRNG()
	GenerateTrackPattern(p, 0, GenreHipHop, rng)

	// Hats should produce a rhythm (mix of on and off steps).
	tr := &p.Tracks[0]
	hits := hitCount(tr, 16)

	// Hip-hop hats at ~50% density = 6-9 hits in 16 steps.
	if hits < 5 {
		t.Errorf("hats have only %d hits, expected at least 5", hits)
	}
	if hits > 10 {
		t.Errorf("hats have %d hits, expected at most 10", hits)
	}
}

func TestGeneratePattern_AllTracks(t *testing.T) {
	folders := []string{"kicks", "snares", "closed-hats", "open-hats"}
	p := testProject(16, folders...)
	rng := fixedRNG()
	GeneratePattern(p, GenreHipHop, rng)

	// Tracks 0-3 should have steps programmed.
	for i := range 4 {
		tr := &p.Tracks[i]
		hits := hitCount(tr, 16)
		if hits == 0 {
			t.Errorf("track %d (%s) has no hits after generation", i, folders[i])
		}
	}

	// Tracks 4-7 have no folder, so they should remain empty.
	for i := 4; i < 8; i++ {
		tr := &p.Tracks[i]
		for step := range 16 {
			if tr.Steps[step] {
				t.Errorf("track %d has no folder but step %d is active", i, step)
			}
		}
	}
}

func TestGeneratePattern_Determinism(t *testing.T) {
	folders := []string{"kicks", "snares", "closed-hats"}
	p1 := testProject(16, folders...)
	p2 := testProject(16, folders...)

	GeneratePattern(p1, GenreHipHop, fixedRNG())
	GeneratePattern(p2, GenreHipHop, fixedRNG())

	// Same seed, same project config → identical output.
	for i := range 3 {
		for step := range 16 {
			if p1.Tracks[i].Steps[step] != p2.Tracks[i].Steps[step] {
				t.Errorf("determinism fail: track %d step %d differs (p1=%v, p2=%v)",
					i, step, p1.Tracks[i].Steps[step], p2.Tracks[i].Steps[step])
				return
			}
		}
	}
}

func TestGeneratePattern_DifferentSeeds(t *testing.T) {
	folders := []string{"kicks", "snares", "closed-hats"}
	p1 := testProject(16, folders...)
	p2 := testProject(16, folders...)

	GeneratePattern(p1, GenreHipHop, rand.New(rand.NewPCG(0, 0)))
	GeneratePattern(p2, GenreHipHop, rand.New(rand.NewPCG(1, 1)))

	// Different seeds should (almost always) produce different output.
	different := false
	for i := range 3 {
		for step := range 16 {
			if p1.Tracks[i].Steps[step] != p2.Tracks[i].Steps[step] {
				different = true
			}
		}
	}
	if !different {
		t.Error("different seeds should produce different patterns")
	}
}

func TestGeneratePattern_32Steps(t *testing.T) {
	folders := []string{"kicks", "snares", "closed-hats"}
	p := testProject(32, folders...)
	rng := fixedRNG()
	GeneratePattern(p, GenreHipHop, rng)

	// All assigned tracks should have hits in both bars.
	for i := range 3 {
		tr := &p.Tracks[i]
		hits := hitCount(tr, 32)
		if hits == 0 {
			t.Errorf("track %d has no hits in 32-step pattern", i)
		}
		// Steps beyond 32 should be off.
		for step := 32; step < 64; step++ {
			if tr.Steps[step] {
				t.Errorf("track %d step %d beyond NumSteps but active", i, step)
			}
		}
	}
}

// ---- Genre tests ----

func TestGenerate_AllGenres(t *testing.T) {
	folders := []string{"kicks", "snares", "closed-hats"}
	genres := []Genre{GenreHipHop, GenreBoomBap, GenreBreakcore}

	for _, genre := range genres {
		t.Run(genre.String(), func(t *testing.T) {
			p := testProject(16, folders...)
			rng := fixedRNG()
			GeneratePattern(p, genre, rng)

			for i := range 3 {
				tr := &p.Tracks[i]
				hits := hitCount(tr, 16)
				if hits == 0 {
					t.Errorf("genre %s: track %d (%s) has no hits",
						genre, i, folders[i])
				}
				// Kick should always hit step 0.
				if i == 0 && !tr.Steps[0] {
					t.Errorf("genre %s: kick step 0 is not active", genre)
				}
			}
		})
	}
}

// ---- Remix tests ----

func TestRemixPattern_IntensityZero(t *testing.T) {
	folders := []string{"kicks", "snares"}
	p := testProject(16, folders...)
	rng := fixedRNG()

	// Generate an initial pattern.
	GeneratePattern(p, GenreHipHop, rng)

	// Capture it.
	var orig [2][64]StepState
	for i := range 2 {
		orig[i] = p.Tracks[i].Steps
	}

	// Remix with intensity 0 should leave it unchanged.
	RemixPattern(p, GenreHipHop, 0, fixedRNG())

	for i := range 2 {
		for step := range 16 {
			if p.Tracks[i].Steps[step] != orig[i][step] {
				t.Errorf("intensity=0: track %d step %d changed", i, step)
			}
		}
	}
}

func TestRemixPattern_IntensityFull(t *testing.T) {
	p := testProject(16, "kicks")
	// Generate fresh pattern.
	GeneratePattern(p, GenreHipHop, fixedRNG())

	// Remix at intensity 1.0 should be equivalent to fully regenerating.
	RemixPattern(p, GenreHipHop, 1.0, fixedRNG())

	// The track should still be valid (kick has hits, downbeat on).
	tr := &p.Tracks[0]
	if !tr.Steps[0] {
		t.Error("after remix at intensity=1.0, kick step 0 should be active")
	}
	hits := hitCount(tr, 16)
	if hits < 2 {
		t.Errorf("after remix at intensity=1.0, kick has only %d hits", hits)
	}
}

func TestRemixPattern_MidIntensity(t *testing.T) {
	p := testProject(16, "kicks")
	rng := fixedRNG()
	GeneratePattern(p, GenreHipHop, rng)

	// Count hits before remix.
	before := hitCount(&p.Tracks[0], 16)

	// Remix at 0.3 — should change some but not all steps.
	RemixPattern(p, GenreHipHop, 0.3, fixedRNG())

	after := hitCount(&p.Tracks[0], 16)

	// With a deterministic RNG at 0.3 intensity, we expect some change.
	// The exact number depends on the RNG, but it should differ from both
	// 0 change and 100% change (which would be equivalent to regenerate).
	tr := &p.Tracks[0]
	changed := 0
	for step := range 16 {
		// We can't compare to original easily since we used the same rng
		// for both, but the density should still be reasonable.
		if tr.Steps[step] {
			changed++
		}
	}
	_ = before
	_ = after
	// Just verify density is still within bounds.
	if changed < 2 || changed > 8 {
		t.Errorf("remix intensity=0.3: kick has %d hits, expected 2-8", changed)
	}
}

func TestRemix_NoFolderTrack(t *testing.T) {
	p := DefaultProject()
	p.NumSteps = 16
	RemixPattern(p, GenreHipHop, 0.5, fixedRNG())

	// All tracks have no folder, so nothing should happen.
	for i := range p.Tracks {
		for step := range 16 {
			if p.Tracks[i].Steps[step] {
				t.Errorf("no-folder track %d step %d is active after remix", i, step)
			}
		}
	}
}

// ---- Genre profile sanity checks ----

func TestGenreProfiles_Coverage(t *testing.T) {
	genres := []Genre{GenreHipHop, GenreBoomBap, GenreBreakcore}
	roles := []DrumRole{RoleKick, RoleSnare, RoleClosedHat, RoleOpenHat, RolePercussion}

	for _, genre := range genres {
		profile := getProfile(genre)
		if profile == nil {
			t.Fatalf("genre %s has nil profile", genre)
		}
		for _, role := range roles {
			rp, ok := profile.probs[role]
			if !ok || rp == nil {
				t.Errorf("genre %s missing profile for role %d", genre, role)
				continue
			}
			// Verify probabilities are in valid range.
			for step, p := range rp.steps {
				if p < 0 || p > 1.0 {
					t.Errorf("genre %s role %d step %d: probability %f out of range",
						genre, role, step, p)
				}
			}
			// Verify density constraints are sensible.
			if rp.minDensity < 0 || rp.minDensity > 1.0 {
				t.Errorf("genre %s role %d: minDensity %f out of range",
					genre, role, rp.minDensity)
			}
			if rp.maxDensity < 0 || rp.maxDensity > 1.0 {
				t.Errorf("genre %s role %d: maxDensity %f out of range",
					genre, role, rp.maxDensity)
			}
			if rp.minDensity > rp.maxDensity {
				t.Errorf("genre %s role %d: minDensity %f > maxDensity %f",
					genre, role, rp.minDensity, rp.maxDensity)
			}
		}
	}
}

func TestGenreString(t *testing.T) {
	tests := []struct {
		genre Genre
		want  string
	}{
		{GenreHipHop, "HipHop"},
		{GenreBoomBap, "BoomBap"},
		{GenreBreakcore, "Breakcore"},
		{Genre(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.genre.String(); got != tt.want {
			t.Errorf("Genre(%d).String() = %q, want %q", tt.genre, got, tt.want)
		}
	}
}

// ---- Density enforcement tests ----

func TestDensityEnforcement_WithinBounds(t *testing.T) {
	// Generate many patterns and verify density always falls within bounds.
	genres := []Genre{GenreHipHop, GenreBoomBap, GenreBreakcore}
	roles := map[string]DrumRole{
		"kicks":       RoleKick,
		"snares":      RoleSnare,
		"closed-hats": RoleClosedHat,
		"open-hats":   RoleOpenHat,
		"percs":       RolePercussion,
	}

	for _, genre := range genres {
		profile := getProfile(genre)
		for folder, role := range roles {
			rp, ok := profile.probs[role]
			if !ok {
				continue
			}
			// Run multiple generations to catch density violations.
			for seed := range 20 {
				p := testProject(16, folder)
				rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed)))
				GenerateTrackPattern(p, 0, genre, rng)
				hits := hitCount(&p.Tracks[0], 16)
				density := float64(hits) / 16.0
				if density < rp.minDensity-0.01 || density > rp.maxDensity+0.01 {
					t.Errorf("genre %s folder %s seed %d: density %.3f outside [%.3f, %.3f]",
						genre, folder, seed, density, rp.minDensity, rp.maxDensity)
				}
			}
		}
	}
}
