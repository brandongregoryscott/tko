package engine

import (
	"math/rand/v2"
)

// FillLength is how many steps at the end of a pattern a fill occupies.
// The value is the step count.
type FillLength int

const (
	FillBeat    FillLength = 4  // last beat
	FillHalfBar FillLength = 8  // last half bar
	FillBar     FillLength = 16 // last bar
)

// fillLengths lists the selectable fill lengths in cycle order.
var fillLengths = []FillLength{FillBeat, FillHalfBar, FillBar}

// String returns a human-readable fill length name.
func (f FillLength) String() string {
	switch f {
	case FillBeat:
		return "1 beat"
	case FillHalfBar:
		return "1/2 bar"
	case FillBar:
		return "1 bar"
	default:
		return "Unknown"
	}
}

// NextFillLength returns the next fill length in cycle order, wrapping around.
// An unrecognized length cycles back to the shortest.
func NextFillLength(f FillLength) FillLength {
	for i, l := range fillLengths {
		if l == f {
			return fillLengths[(i+1)%len(fillLengths)]
		}
	}
	return fillLengths[0]
}

// ---- internal fill profile types ----

// fillProbs holds hit probabilities for the last 16 steps of a pattern.
// Fills shorter than 16 steps use the tail of the table, which is where the
// tables are busiest, so a 1-beat fill is the climax of a 1-bar fill.
type fillProbs struct {
	steps      [16]float64
	minDensity float64
	maxDensity float64
}

type fillProfile struct {
	name  string
	probs map[DrumRole]*fillProbs
}

// rangeFor returns the step range covering count steps from start, weighted by
// the tail of the fill table.
func (fp *fillProbs) rangeFor(start, count int) stepRange {
	count = min(count, len(fp.steps))
	weights := make([]float64, count)
	copy(weights, fp.steps[len(fp.steps)-count:])
	return stepRange{
		start:      start,
		weights:    weights,
		minDensity: fp.minDensity,
		maxDensity: fp.maxDensity,
	}
}

// ---- hip-hop fill tables ----

// hipHopKickFill16 — kick thins out as the fill takes over.
var hipHopKickFill16 = fillProbs{
	steps: [16]float64{
		0.85, 0.05, 0.10, 0.05,
		0.25, 0.05, 0.10, 0.40,
		0.20, 0.10, 0.30, 0.10,
		0.15, 0.05, 0.20, 0.05,
	},
	minDensity: 0.10, maxDensity: 0.35,
}

// hipHopSnareFill16 — backbeat then a roll that builds into the downbeat.
var hipHopSnareFill16 = fillProbs{
	steps: [16]float64{
		0.10, 0.05, 0.10, 0.05,
		0.85, 0.10, 0.20, 0.15,
		0.30, 0.25, 0.50, 0.30,
		0.55, 0.50, 0.75, 0.65,
	},
	minDensity: 0.30, maxDensity: 0.60,
}

// hipHopClosedHatFill16 — hats step aside for the snare roll.
var hipHopClosedHatFill16 = fillProbs{
	steps: [16]float64{
		0.80, 0.15, 0.80, 0.15,
		0.70, 0.15, 0.70, 0.20,
		0.55, 0.20, 0.50, 0.20,
		0.35, 0.15, 0.25, 0.10,
	},
	minDensity: 0.25, maxDensity: 0.55,
}

// hipHopOpenHatFill16 — accent right at the end of the fill.
var hipHopOpenHatFill16 = fillProbs{
	steps: [16]float64{
		0.05, 0.02, 0.05, 0.02,
		0.10, 0.05, 0.05, 0.05,
		0.10, 0.05, 0.15, 0.05,
		0.20, 0.05, 0.30, 0.35,
	},
	minDensity: 0.05, maxDensity: 0.30,
}

// hipHopTomFill16 — tom run through the back half.
var hipHopTomFill16 = fillProbs{
	steps: [16]float64{
		0.05, 0.05, 0.10, 0.05,
		0.15, 0.10, 0.20, 0.10,
		0.40, 0.30, 0.45, 0.35,
		0.50, 0.45, 0.60, 0.50,
	},
	minDensity: 0.25, maxDensity: 0.55,
}

// hipHopPercFill16 — percussion ramps up steadily.
var hipHopPercFill16 = fillProbs{
	steps: [16]float64{
		0.15, 0.25, 0.15, 0.30,
		0.20, 0.30, 0.20, 0.35,
		0.30, 0.40, 0.30, 0.45,
		0.40, 0.50, 0.45, 0.55,
	},
	minDensity: 0.25, maxDensity: 0.55,
}

// ---- boom-bap fill tables ----

// boomBapKickFill16 — sparse, leaves space for the snare.
var boomBapKickFill16 = fillProbs{
	steps: [16]float64{
		0.80, 0.02, 0.05, 0.05,
		0.15, 0.02, 0.05, 0.35,
		0.10, 0.05, 0.20, 0.05,
		0.10, 0.02, 0.10, 0.05,
	},
	minDensity: 0.08, maxDensity: 0.28,
}

// boomBapSnareFill16 — looser roll than hip-hop, more gaps.
var boomBapSnareFill16 = fillProbs{
	steps: [16]float64{
		0.10, 0.05, 0.10, 0.05,
		0.80, 0.05, 0.15, 0.10,
		0.25, 0.15, 0.40, 0.20,
		0.45, 0.40, 0.65, 0.55,
	},
	minDensity: 0.25, maxDensity: 0.50,
}

// boomBapClosedHatFill16 — sparse hats thinning toward the downbeat.
var boomBapClosedHatFill16 = fillProbs{
	steps: [16]float64{
		0.60, 0.05, 0.60, 0.05,
		0.55, 0.05, 0.55, 0.10,
		0.45, 0.10, 0.40, 0.10,
		0.25, 0.10, 0.20, 0.05,
	},
	minDensity: 0.20, maxDensity: 0.45,
}

// boomBapOpenHatFill16 — a single accent at the end.
var boomBapOpenHatFill16 = fillProbs{
	steps: [16]float64{
		0.05, 0.02, 0.05, 0.02,
		0.10, 0.02, 0.05, 0.02,
		0.10, 0.02, 0.10, 0.05,
		0.15, 0.05, 0.25, 0.30,
	},
	minDensity: 0.05, maxDensity: 0.25,
}

// boomBapTomFill16 — classic tom run, sparser than hip-hop.
var boomBapTomFill16 = fillProbs{
	steps: [16]float64{
		0.05, 0.02, 0.05, 0.02,
		0.10, 0.05, 0.15, 0.05,
		0.35, 0.20, 0.40, 0.25,
		0.45, 0.35, 0.55, 0.45,
	},
	minDensity: 0.20, maxDensity: 0.50,
}

// boomBapPercFill16 — restrained percussion build.
var boomBapPercFill16 = fillProbs{
	steps: [16]float64{
		0.10, 0.15, 0.10, 0.20,
		0.15, 0.20, 0.15, 0.25,
		0.20, 0.30, 0.25, 0.35,
		0.30, 0.40, 0.35, 0.45,
	},
	minDensity: 0.15, maxDensity: 0.45,
}

// ---- breakcore fill tables ----

// breakcoreKickFill16 — kicks stay dense through the fill.
var breakcoreKickFill16 = fillProbs{
	steps: [16]float64{
		0.90, 0.20, 0.30, 0.25,
		0.60, 0.25, 0.40, 0.30,
		0.50, 0.30, 0.45, 0.35,
		0.40, 0.35, 0.50, 0.40,
	},
	minDensity: 0.35, maxDensity: 0.60,
}

// breakcoreSnareFill16 — amen-style snare rush into the downbeat.
var breakcoreSnareFill16 = fillProbs{
	steps: [16]float64{
		0.30, 0.35, 0.40, 0.45,
		0.85, 0.45, 0.55, 0.50,
		0.70, 0.60, 0.75, 0.65,
		0.85, 0.80, 0.90, 0.85,
	},
	minDensity: 0.55, maxDensity: 0.85,
}

// breakcoreClosedHatFill16 — hats back off as the snare rush takes over.
var breakcoreClosedHatFill16 = fillProbs{
	steps: [16]float64{
		0.80, 0.60, 0.80, 0.65,
		0.75, 0.55, 0.70, 0.60,
		0.60, 0.50, 0.55, 0.45,
		0.40, 0.35, 0.35, 0.30,
	},
	minDensity: 0.40, maxDensity: 0.75,
}

// breakcoreOpenHatFill16 — open hats splash through the fill.
var breakcoreOpenHatFill16 = fillProbs{
	steps: [16]float64{
		0.20, 0.10, 0.20, 0.10,
		0.25, 0.15, 0.20, 0.15,
		0.30, 0.15, 0.30, 0.20,
		0.35, 0.25, 0.45, 0.50,
	},
	minDensity: 0.20, maxDensity: 0.50,
}

// breakcoreTomFill16 — relentless tom run.
var breakcoreTomFill16 = fillProbs{
	steps: [16]float64{
		0.20, 0.15, 0.25, 0.20,
		0.30, 0.20, 0.35, 0.25,
		0.50, 0.40, 0.55, 0.45,
		0.65, 0.55, 0.70, 0.60,
	},
	minDensity: 0.40, maxDensity: 0.70,
}

// breakcorePercFill16 — chaotic percussion throughout.
var breakcorePercFill16 = fillProbs{
	steps: [16]float64{
		0.35, 0.45, 0.35, 0.50,
		0.40, 0.50, 0.40, 0.55,
		0.45, 0.55, 0.50, 0.60,
		0.55, 0.65, 0.60, 0.70,
	},
	minDensity: 0.40, maxDensity: 0.75,
}

// ---- fill profile lookup ----

var hipHopFillProfile = fillProfile{
	name: "Hip-Hop",
	probs: map[DrumRole]*fillProbs{
		RoleKick:       &hipHopKickFill16,
		RoleSnare:      &hipHopSnareFill16,
		RoleClosedHat:  &hipHopClosedHatFill16,
		RoleOpenHat:    &hipHopOpenHatFill16,
		RoleTom:        &hipHopTomFill16,
		RolePercussion: &hipHopPercFill16,
	},
}

var boomBapFillProfile = fillProfile{
	name: "Boom-Bap",
	probs: map[DrumRole]*fillProbs{
		RoleKick:       &boomBapKickFill16,
		RoleSnare:      &boomBapSnareFill16,
		RoleClosedHat:  &boomBapClosedHatFill16,
		RoleOpenHat:    &boomBapOpenHatFill16,
		RoleTom:        &boomBapTomFill16,
		RolePercussion: &boomBapPercFill16,
	},
}

var breakcoreFillProfile = fillProfile{
	name: "Breakcore",
	probs: map[DrumRole]*fillProbs{
		RoleKick:       &breakcoreKickFill16,
		RoleSnare:      &breakcoreSnareFill16,
		RoleClosedHat:  &breakcoreClosedHatFill16,
		RoleOpenHat:    &breakcoreOpenHatFill16,
		RoleTom:        &breakcoreTomFill16,
		RolePercussion: &breakcorePercFill16,
	},
}

func getFillProfile(g Genre) *fillProfile {
	switch g {
	case GenreHipHop:
		return &hipHopFillProfile
	case GenreBoomBap:
		return &boomBapFillProfile
	case GenreBreakcore:
		return &breakcoreFillProfile
	default:
		return &hipHopFillProfile
	}
}

// fillRole maps a drum role onto the role whose fill table it should use.
// Claps and rimshots roll with the snare, rides with the closed hat, and
// anything else falls back to percussion.
func fillRole(role DrumRole) DrumRole {
	switch role {
	case RoleClap, RoleRimshot:
		return RoleSnare
	case RoleRide:
		return RoleClosedHat
	default:
		return role
	}
}

// fillProbsFor looks up a role's fill probabilities, falling back to percussion
// for roles the profile does not define.
func fillProbsFor(profile *fillProfile, role DrumRole) *fillProbs {
	fp, ok := profile.probs[fillRole(role)]
	if !ok {
		fp = profile.probs[RolePercussion]
	}
	return fp
}

// ---- fill algorithms ----

// FillWindow returns the first step of the fill and how many steps it covers
// for a pattern of numSteps steps. A fill longer than the pattern is clamped to
// the whole pattern.
func FillWindow(numSteps int, length FillLength) (start, count int) {
	numSteps = max(numSteps, 1)
	numSteps = min(numSteps, 64)
	count = min(int(length), numSteps)
	count = max(count, 1)
	return numSteps - count, count
}

// FillPattern writes a genre-appropriate fill over the last steps of every
// track, leaving the rest of the pattern intact. Tracks without a folder
// assignment are left unchanged.
func FillPattern(proj *Project, genre Genre, length FillLength, rng *rand.Rand) {
	profile := getFillProfile(genre)
	for i := range proj.Tracks {
		fillTrack(proj, TrackID(i), genre, length, rng, profile)
	}
}

// FillTrackPattern writes a genre-appropriate fill over the last steps of a
// single track. If the track has no folder assignment, this is a no-op.
func FillTrackPattern(proj *Project, track TrackID, genre Genre, length FillLength, rng *rand.Rand) {
	fillTrack(proj, track, genre, length, rng, nil)
}

// fillTrack is the internal implementation that accepts an optional profile.
func fillTrack(proj *Project, track TrackID, genre Genre, length FillLength, rng *rand.Rand, profile *fillProfile) {
	if int(track) < 0 || int(track) >= len(proj.Tracks) {
		return
	}
	t := &proj.Tracks[track]
	if t.Sample.Folder == "" {
		return
	}
	if profile == nil {
		profile = getFillProfile(genre)
	}

	start, count := FillWindow(proj.NumSteps, length)
	role := FolderRole(t.Sample.Folder)

	// Crashes mark the downbeat the fill lands on rather than playing through it.
	if role == RoleCrash {
		for step := start; step < start+count; step++ {
			t.Steps[step] = false
		}
		t.Steps[0] = true
		return
	}

	fp := fillProbsFor(profile, role)
	if fp == nil {
		return
	}

	r := fp.rangeFor(start, count)
	rollRange(t, r, rng)
	enforceDensity(t, r, rng)
}
