package engine

import (
	"math/rand/v2"
	"testing"
)

// seededRNG returns a deterministic RNG for a given seed, so tests that need
// many independent draws stay reproducible.
func seededRNG(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed+1))
}

// setSteps activates the given steps on a track.
func setSteps(t *Track, steps ...int) {
	for _, step := range steps {
		t.Steps[step] = true
	}
}

// snapshot copies a track's steps so a fill's effect can be diffed.
func snapshot(t *Track) Steps {
	return t.Steps
}

// ---- FillLength tests ----

func TestFillLengthString(t *testing.T) {
	cases := map[FillLength]string{
		FillBeat:      "1 beat",
		FillHalfBar:   "1/2 bar",
		FillBar:       "1 bar",
		FillLength(7): "Unknown",
	}
	for length, want := range cases {
		if got := length.String(); got != want {
			t.Errorf("FillLength(%d).String() = %q, want %q", int(length), got, want)
		}
	}
}

func TestNextFillLengthCycles(t *testing.T) {
	want := []FillLength{FillHalfBar, FillBar, FillBeat}
	got := FillBeat
	for i, expected := range want {
		got = NextFillLength(got)
		if got != expected {
			t.Errorf("cycle step %d = %v, want %v", i, got, expected)
		}
	}
	if got := NextFillLength(FillLength(7)); got != FillBeat {
		t.Errorf("NextFillLength(unknown) = %v, want FillBeat", got)
	}
}

// ---- FillWindow tests ----

func TestFillWindow(t *testing.T) {
	cases := []struct {
		numSteps  int
		length    FillLength
		wantStart int
		wantCount int
	}{
		{16, FillBeat, 12, 4},
		{16, FillHalfBar, 8, 8},
		{16, FillBar, 0, 16},
		{64, FillBar, 48, 16},
		{32, FillHalfBar, 24, 8},
		{4, FillBar, 0, 4},  // fill longer than the pattern is clamped
		{0, FillBeat, 0, 1}, // degenerate pattern still yields one step
	}
	for _, c := range cases {
		start, count := FillWindow(c.numSteps, c.length)
		if start != c.wantStart || count != c.wantCount {
			t.Errorf("FillWindow(%d, %v) = (%d, %d), want (%d, %d)",
				c.numSteps, c.length, start, count, c.wantStart, c.wantCount)
		}
	}
}

// ---- FillPattern tests ----

func TestFillPatternLeavesEarlierStepsAlone(t *testing.T) {
	p := testProject(32, "kick", "snare", "closed-hat")
	GeneratePattern(p, GenreHipHop, fixedRNG())
	before := [3]Steps{snapshot(&p.Tracks[0]), snapshot(&p.Tracks[1]), snapshot(&p.Tracks[2])}

	FillPattern(p, GenreHipHop, FillHalfBar, fixedRNG())

	start, _ := FillWindow(p.NumSteps, FillHalfBar)
	for i := range before {
		for step := range start {
			if p.Tracks[i].Steps[step] != before[i][step] {
				t.Errorf("track %d step %d changed outside the fill window", i, step)
			}
		}
	}
}

func TestFillPatternPutsHitsInWindow(t *testing.T) {
	p := testProject(32, "snare")
	FillPattern(p, GenreHipHop, FillHalfBar, fixedRNG())

	start, count := FillWindow(p.NumSteps, FillHalfBar)
	hits := 0
	for step := start; step < start+count; step++ {
		if p.Tracks[0].Steps[step] {
			hits++
		}
	}
	if hits == 0 {
		t.Error("snare fill produced no hits in the fill window")
	}
}

func TestFillPatternRespectsWindowDensity(t *testing.T) {
	for _, genre := range []Genre{GenreHipHop, GenreBoomBap, GenreBreakcore} {
		for seed := range 20 {
			p := testProject(32, "kick", "snare", "closed-hat", "open-hat", "tom", "perc")
			FillPattern(p, genre, FillBar, seededRNG(uint64(seed)))

			start, count := FillWindow(p.NumSteps, FillBar)
			for i := range 6 {
				track := &p.Tracks[i]
				fp := fillProbsFor(getFillProfile(genre), FolderRole(track.Sample.Folder))
				hits := 0
				for step := start; step < start+count; step++ {
					if track.Steps[step] {
						hits++
					}
				}
				density := float64(hits) / float64(count)
				if density < fp.minDensity-0.01 || density > fp.maxDensity+0.01 {
					t.Errorf("%v %s fill density %.2f outside [%.2f, %.2f]",
						genre, track.Sample.Folder, density, fp.minDensity, fp.maxDensity)
				}
			}
		}
	}
}

func TestFillPatternIntensifiesTowardTheEnd(t *testing.T) {
	// Averaged over many rolls, the back half of a snare fill should be busier
	// than the front half.
	front, back := 0, 0
	for seed := range 50 {
		p := testProject(16, "snare")
		FillPattern(p, GenreHipHop, FillBar, seededRNG(uint64(seed)))
		for step := range 8 {
			if p.Tracks[0].Steps[step] {
				front++
			}
			if p.Tracks[0].Steps[step+8] {
				back++
			}
		}
	}
	if back <= front {
		t.Errorf("snare fill back half (%d hits) not busier than front half (%d hits)", back, front)
	}
}

func TestFillPatternSkipsUnassignedTracks(t *testing.T) {
	p := testProject(16, "kick")
	FillPattern(p, GenreHipHop, FillBeat, fixedRNG())

	for i := 1; i < len(p.Tracks); i++ {
		if hitCount(&p.Tracks[i], 16) != 0 {
			t.Errorf("track %d without a folder was modified", i)
		}
	}
}

func TestFillPatternCrashMarksDownbeat(t *testing.T) {
	p := testProject(32, "crash")
	setSteps(&p.Tracks[0], 26, 30)

	FillPattern(p, GenreHipHop, FillHalfBar, fixedRNG())

	if !p.Tracks[0].Steps[0] {
		t.Error("crash fill did not place a hit on the downbeat")
	}
	start, count := FillWindow(p.NumSteps, FillHalfBar)
	for step := start; step < start+count; step++ {
		if p.Tracks[0].Steps[step] {
			t.Errorf("crash fill left a hit at step %d inside the fill window", step)
		}
	}
}

func TestFillPatternGenreDensityOrdering(t *testing.T) {
	// Breakcore fills should be busier than boom-bap fills for the same kit.
	total := func(genre Genre) int {
		n := 0
		for seed := range 30 {
			p := testProject(32, "kick", "snare", "closed-hat", "perc")
			FillPattern(p, genre, FillBar, seededRNG(uint64(seed)))
			start, count := FillWindow(p.NumSteps, FillBar)
			for i := range 4 {
				for step := start; step < start+count; step++ {
					if p.Tracks[i].Steps[step] {
						n++
					}
				}
			}
		}
		return n
	}
	if breakcore, boomBap := total(GenreBreakcore), total(GenreBoomBap); breakcore <= boomBap {
		t.Errorf("breakcore fill hits (%d) not greater than boom-bap fill hits (%d)", breakcore, boomBap)
	}
}

// ---- FillTrackPattern tests ----

func TestFillTrackPatternOnlyTouchesOneTrack(t *testing.T) {
	p := testProject(16, "kick", "snare")
	GeneratePattern(p, GenreHipHop, fixedRNG())
	before := snapshot(&p.Tracks[0])

	FillTrackPattern(p, 1, GenreHipHop, FillBeat, fixedRNG())

	if p.Tracks[0].Steps != before {
		t.Error("FillTrackPattern modified a track it was not given")
	}
}

func TestFillTrackPatternNoFolderIsNoOp(t *testing.T) {
	p := testProject(16)
	setSteps(&p.Tracks[0], 0, 4, 8, 12)
	before := snapshot(&p.Tracks[0])

	FillTrackPattern(p, 0, GenreHipHop, FillBeat, fixedRNG())

	if p.Tracks[0].Steps != before {
		t.Error("FillTrackPattern modified a track with no sample folder")
	}
}

func TestFillTrackPatternOutOfRangeTrack(t *testing.T) {
	p := testProject(16, "kick")
	// Must not panic.
	FillTrackPattern(p, -1, GenreHipHop, FillBeat, fixedRNG())
	FillTrackPattern(p, 99, GenreHipHop, FillBeat, fixedRNG())
}

func TestFillTrackPatternClampsToPattern(t *testing.T) {
	p := testProject(16, "snare")
	FillTrackPattern(p, 0, GenreHipHop, FillBar, fixedRNG())

	for step := 16; step < 64; step++ {
		if p.Tracks[0].Steps[step] {
			t.Errorf("fill wrote step %d beyond the pattern length", step)
		}
	}
}

// ---- role aliasing ----

func TestFillRoleAliases(t *testing.T) {
	cases := map[DrumRole]DrumRole{
		RoleClap:       RoleSnare,
		RoleRimshot:    RoleSnare,
		RoleRide:       RoleClosedHat,
		RoleKick:       RoleKick,
		RolePercussion: RolePercussion,
	}
	for role, want := range cases {
		if got := fillRole(role); got != want {
			t.Errorf("fillRole(%v) = %v, want %v", role, got, want)
		}
	}
}

func TestFillProbsForUnmappedRoleFallsBack(t *testing.T) {
	profile := getFillProfile(GenreHipHop)
	if got := fillProbsFor(profile, RoleShaker); got != profile.probs[RolePercussion] {
		t.Error("unmapped role did not fall back to percussion fill probabilities")
	}
}
