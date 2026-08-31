package engine

import (
	"math"
	"math/rand/v2"
	"strings"
)

// DrumRole classifies a track by the type of drum its sample folder represents.
type DrumRole int

const (
	RoleUnknown DrumRole = iota
	RoleKick
	RoleSnare
	RoleClap
	RoleClosedHat
	RoleOpenHat
	RoleRide
	RoleCrash
	RoleTom
	RolePercussion
	RoleShaker
	RoleRimshot
)

// Genre selects a generation profile.
type Genre int

const (
	GenreHipHop Genre = iota
	GenreBoomBap
	GenreBreakcore
)

// String returns a human-readable genre name.
func (g Genre) String() string {
	switch g {
	case GenreHipHop:
		return "HipHop"
	case GenreBoomBap:
		return "BoomBap"
	case GenreBreakcore:
		return "Breakcore"
	default:
		return "Unknown"
	}
}

// FolderRole returns the drum role inferred from a folder name using substring matching.
// Unknown folder names default to RolePercussion so every assigned track gets a pattern.
func FolderRole(folder string) DrumRole {
	lower := strings.ToLower(folder)
	switch {
	case strings.Contains(lower, "kick") || strings.Contains(lower, "bd") || strings.Contains(lower, "bassdrum"):
		return RoleKick
	case strings.Contains(lower, "snare") || strings.Contains(lower, "sd") || strings.Contains(lower, "snr"):
		return RoleSnare
	case strings.Contains(lower, "clap"):
		return RoleClap
	case strings.Contains(lower, "closed") && (strings.Contains(lower, "hat") || strings.Contains(lower, "hh")):
		return RoleClosedHat
	case strings.Contains(lower, "open") && (strings.Contains(lower, "hat") || strings.Contains(lower, "oh")):
		return RoleOpenHat
	case strings.Contains(lower, "ride"):
		return RoleRide
	case strings.Contains(lower, "crash") || strings.Contains(lower, "cymbal"):
		return RoleCrash
	case strings.Contains(lower, "tom"):
		return RoleTom
	case strings.Contains(lower, "shaker") || strings.Contains(lower, "tamb"):
		return RoleShaker
	case strings.Contains(lower, "rim") || strings.Contains(lower, "sidestick"):
		return RoleRimshot
	case strings.Contains(lower, "perc"):
		return RolePercussion
	default:
		return RolePercussion // fallback: treat unknown folders as percussion
	}
}

// ---- internal profile types ----

type roleProbs struct {
	steps      [64]float64 // hit probability per step (0.0 to 1.0)
	minDensity float64     // minimum fraction of steps that should hit
	maxDensity float64     // maximum fraction of steps that should hit
}

type genreProfile struct {
	name  string
	probs map[DrumRole]*roleProbs
}

// stepCandidate is used internally by density enforcement for weighted random selection.
type stepCandidate struct {
	step   int
	weight float64
}

// stepRange is a contiguous slice of a pattern paired with the hit probability
// for each of its steps. Rolling and density enforcement work on a range so
// that fills can target the tail of a pattern without disturbing the steps
// before it.
type stepRange struct {
	start      int
	weights    []float64 // hit probability per step, indexed from start
	minDensity float64
	maxDensity float64
}

// rangeFor returns the range covering count steps from start, carrying the
// role's probabilities and density bounds.
func (rp *roleProbs) rangeFor(start, count int) stepRange {
	weights := make([]float64, count)
	copy(weights, rp.steps[start:start+count])
	return stepRange{
		start:      start,
		weights:    weights,
		minDensity: rp.minDensity,
		maxDensity: rp.maxDensity,
	}
}

// ---- probability table helpers ----

// expand16 repeats a 16-step seed four times to fill 64 steps.
func expand16(seed [16]float64) [64]float64 {
	var out [64]float64
	for i := range out {
		out[i] = seed[i%16]
	}
	return out
}

// expand32 repeats a 32-step seed twice to fill 64 steps.
func expand32(seed [32]float64) [64]float64 {
	var out [64]float64
	for i := range out {
		out[i] = seed[i%32]
	}
	return out
}

// ---- genre profile definitions ----

// hipHopKicks16 is the 16-step seed for hip-hop kick patterns.
// Strong on downbeat (0), lazy syncopation at 7, varied elsewhere.
var hipHopKicks16 = [16]float64{
	1.00, 0.05, 0.15, 0.10, // beat 1
	0.20, 0.05, 0.15, 0.50, // beat 2 (note: step 7 is the classic hip-hop kick)
	0.30, 0.20, 0.20, 0.40, // beat 3
	0.15, 0.10, 0.25, 0.15, // beat 4
}

// hipHopSnares16 — strict backbeat on steps 4 and 12.
var hipHopSnares16 = [16]float64{
	0.02, 0.02, 0.02, 0.02,
	1.00, 0.02, 0.02, 0.02,
	0.02, 0.02, 0.02, 0.02,
	1.00, 0.02, 0.02, 0.02,
}

// hipHopClosedHats16 — steady eighth notes with occasional ghost notes.
var hipHopClosedHats16 = [16]float64{
	0.85, 0.10, 0.85, 0.10,
	0.85, 0.10, 0.85, 0.10,
	0.85, 0.10, 0.85, 0.10,
	0.85, 0.10, 0.85, 0.10,
}

// hipHopOpenHats16 — accent hits every half bar with occasional extras.
var hipHopOpenHats16 = [16]float64{
	0.05, 0.02, 0.02, 0.02,
	0.10, 0.40, 0.02, 0.02,
	0.10, 0.02, 0.02, 0.02,
	0.10, 0.40, 0.02, 0.02,
}

// hipHopPerc16 — off-beat emphasis for percussion/shakers.
var hipHopPerc16 = [16]float64{
	0.10, 0.25, 0.10, 0.25,
	0.10, 0.30, 0.10, 0.25,
	0.10, 0.25, 0.10, 0.25,
	0.10, 0.30, 0.10, 0.25,
}

// boomBapKicks16 — sparser than hip-hop, fewer hits, punchier feel.
var boomBapKicks16 = [16]float64{
	1.00, 0.02, 0.05, 0.05,
	0.10, 0.02, 0.05, 0.60,
	0.10, 0.05, 0.10, 0.30,
	0.10, 0.02, 0.10, 0.05,
}

// boomBapSnares16 — same strict backbeat as hip-hop.
var boomBapSnares16 = [16]float64{
	0.02, 0.02, 0.02, 0.02,
	1.00, 0.02, 0.02, 0.02,
	0.02, 0.02, 0.02, 0.02,
	1.00, 0.02, 0.02, 0.02,
}

// boomBapClosedHats16 — sparser hats, ~35% density instead of ~50%.
var boomBapClosedHats16 = [16]float64{
	0.60, 0.05, 0.60, 0.05,
	0.60, 0.05, 0.60, 0.05,
	0.60, 0.05, 0.60, 0.05,
	0.60, 0.05, 0.60, 0.05,
}

// boomBapOpenHats16 — similar to hip-hop open hats.
var boomBapOpenHats16 = [16]float64{
	0.05, 0.02, 0.02, 0.02,
	0.10, 0.35, 0.02, 0.02,
	0.10, 0.02, 0.02, 0.02,
	0.10, 0.35, 0.02, 0.02,
}

// boomBapPerc16 — sparser percussion.
var boomBapPerc16 = [16]float64{
	0.05, 0.15, 0.05, 0.15,
	0.05, 0.20, 0.05, 0.20,
	0.05, 0.15, 0.05, 0.15,
	0.05, 0.20, 0.05, 0.15,
}

// breakcoreKicks32 — dense, syncopated kicks over 32 steps (2 bars).
var breakcoreKicks32 = [32]float64{
	// Bar 1
	1.00, 0.10, 0.10, 0.20,
	0.70, 0.05, 0.30, 0.10,
	0.50, 0.20, 0.30, 0.10,
	0.30, 0.15, 0.20, 0.15,
	// Bar 2
	1.00, 0.15, 0.10, 0.20,
	0.30, 0.10, 0.40, 0.15,
	0.20, 0.20, 0.30, 0.15,
	0.20, 0.20, 0.30, 0.20,
}

// breakcoreSnares32 — syncopated, not just backbeat.
var breakcoreSnares32 = [32]float64{
	// Bar 1
	0.02, 0.02, 0.02, 0.02,
	0.90, 0.02, 0.30, 0.02,
	0.20, 0.02, 0.30, 0.02,
	0.90, 0.02, 0.20, 0.02,
	// Bar 2
	0.02, 0.10, 0.02, 0.10,
	0.90, 0.05, 0.30, 0.02,
	0.10, 0.15, 0.30, 0.05,
	0.80, 0.15, 0.20, 0.20,
}

// breakcoreHats32 — very dense, near-continuous ride/hat patterns.
var breakcoreHats32 = [32]float64{
	// Bar 1
	0.80, 0.50, 0.80, 0.70,
	0.85, 0.40, 0.80, 0.60,
	0.75, 0.55, 0.80, 0.65,
	0.80, 0.45, 0.80, 0.70,
	// Bar 2
	0.80, 0.50, 0.80, 0.70,
	0.85, 0.40, 0.80, 0.60,
	0.75, 0.55, 0.80, 0.65,
	0.80, 0.45, 0.80, 0.70,
}

// breakcorePerc32 — very active percussion.
var breakcorePerc32 = [32]float64{
	0.20, 0.50, 0.20, 0.40,
	0.15, 0.55, 0.30, 0.45,
	0.20, 0.50, 0.25, 0.40,
	0.15, 0.60, 0.30, 0.45,
	0.20, 0.50, 0.20, 0.40,
	0.15, 0.55, 0.30, 0.45,
	0.20, 0.50, 0.25, 0.40,
	0.15, 0.60, 0.30, 0.45,
}

// ---- profile lookup ----

func getProfile(g Genre) *genreProfile {
	switch g {
	case GenreHipHop:
		return &hipHopProfile
	case GenreBoomBap:
		return &boomBapProfile
	case GenreBreakcore:
		return &breakcoreProfile
	default:
		return &hipHopProfile
	}
}

var hipHopProfile = genreProfile{
	name: "Hip-Hop",
	probs: map[DrumRole]*roleProbs{
		RoleKick:       {steps: expand16(hipHopKicks16), minDensity: 0.20, maxDensity: 0.40},
		RoleSnare:      {steps: expand16(hipHopSnares16), minDensity: 0.10, maxDensity: 0.18},
		RoleClosedHat:  {steps: expand16(hipHopClosedHats16), minDensity: 0.38, maxDensity: 0.55},
		RoleOpenHat:    {steps: expand16(hipHopOpenHats16), minDensity: 0.05, maxDensity: 0.20},
		RolePercussion: {steps: expand16(hipHopPerc16), minDensity: 0.15, maxDensity: 0.35},
	},
}

var boomBapProfile = genreProfile{
	name: "Boom-Bap",
	probs: map[DrumRole]*roleProbs{
		RoleKick:       {steps: expand16(boomBapKicks16), minDensity: 0.12, maxDensity: 0.28},
		RoleSnare:      {steps: expand16(boomBapSnares16), minDensity: 0.10, maxDensity: 0.18},
		RoleClosedHat:  {steps: expand16(boomBapClosedHats16), minDensity: 0.28, maxDensity: 0.42},
		RoleOpenHat:    {steps: expand16(boomBapOpenHats16), minDensity: 0.05, maxDensity: 0.18},
		RolePercussion: {steps: expand16(boomBapPerc16), minDensity: 0.10, maxDensity: 0.25},
	},
}

var breakcoreProfile = genreProfile{
	name: "Breakcore",
	probs: map[DrumRole]*roleProbs{
		RoleKick:       {steps: expand32(breakcoreKicks32), minDensity: 0.30, maxDensity: 0.50},
		RoleSnare:      {steps: expand32(breakcoreSnares32), minDensity: 0.15, maxDensity: 0.35},
		RoleClosedHat:  {steps: expand32(breakcoreHats32), minDensity: 0.55, maxDensity: 0.80},
		RoleOpenHat:    {steps: expand32(breakcoreHats32), minDensity: 0.30, maxDensity: 0.55},
		RolePercussion: {steps: expand32(breakcorePerc32), minDensity: 0.30, maxDensity: 0.55},
	},
}

// ---- generation algorithms ----

// GeneratePattern fills all tracks in a project with genre-consistent step patterns.
// Tracks without a folder assignment are left unchanged.
// Project.NumSteps determines how many steps are generated.
func GeneratePattern(proj *Project, genre Genre, rng *rand.Rand) {
	profile := getProfile(genre)
	for i := range proj.Tracks {
		generateTrack(proj, TrackID(i), genre, rng, profile)
	}
}

// GenerateTrackPattern fills a single track with a genre-consistent step pattern.
// If the track has no folder assignment, this is a no-op.
func GenerateTrackPattern(proj *Project, track TrackID, genre Genre, rng *rand.Rand) {
	generateTrack(proj, track, genre, rng, nil)
}

// generateTrack is the internal implementation that accepts an optional profile.
func generateTrack(proj *Project, track TrackID, genre Genre, rng *rand.Rand, profile *genreProfile) {
	if int(track) < 0 || int(track) >= len(proj.Tracks) {
		return
	}
	t := &proj.Tracks[track]
	if t.Sample.Folder == "" {
		return
	}
	if profile == nil {
		profile = getProfile(genre)
	}

	// Clear existing steps first.
	for i := range t.Steps {
		t.Steps[i] = false
	}

	generateSteps(t, proj.NumSteps, profile, rng)
}

func generateSteps(t *Track, numSteps int, profile *genreProfile, rng *rand.Rand) {
	rp := roleProbsFor(profile, FolderRole(t.Sample.Folder))
	if rp == nil {
		return
	}

	numSteps = max(numSteps, 1)
	numSteps = min(numSteps, 64)

	// Generate candidate hits by rolling against step probabilities.
	r := rp.rangeFor(0, numSteps)
	rollRange(t, r, rng)
	enforceDensity(t, r, rng)
}

// roleProbsFor looks up a role's probabilities, falling back to percussion for
// roles the profile does not define.
func roleProbsFor(profile *genreProfile, role DrumRole) *roleProbs {
	rp, ok := profile.probs[role]
	if !ok {
		rp = profile.probs[RolePercussion]
	}
	return rp
}

// rollRange re-rolls every step in the range against its probability,
// overwriting whatever was there before.
func rollRange(t *Track, r stepRange, rng *rand.Rand) {
	for i, weight := range r.weights {
		t.Steps[r.start+i] = StepState(rng.Float64() < weight)
	}
}

// enforceDensity adjusts hits up or down to meet the range's min/max density
// targets. Steps with higher genre probability are preferred when adding hits;
// steps with lower genre probability are preferred when removing.
func enforceDensity(t *Track, r stepRange, rng *rand.Rand) {
	if len(r.weights) == 0 {
		return
	}
	hits := 0
	for i := range r.weights {
		if t.Steps[r.start+i] {
			hits++
		}
	}
	density := float64(hits) / float64(len(r.weights))

	if density < r.minDensity {
		addHits(t, r, rng, hits)
	} else if density > r.maxDensity {
		removeHits(t, r, rng, hits)
	}
}

func addHits(t *Track, r stepRange, rng *rand.Rand, currentHits int) {
	target := int(math.Ceil(r.minDensity * float64(len(r.weights))))
	if target <= currentHits {
		return
	}

	// Collect inactive steps with their probability weights.
	var candidates []stepCandidate
	for i, weight := range r.weights {
		if step := r.start + i; !t.Steps[step] {
			candidates = append(candidates, stepCandidate{step, weight + 0.01})
		}
	}

	// Randomly activate steps, weighted by probability, until target met.
	for attempts := 0; attempts < 200 && currentHits < target && len(candidates) > 0; attempts++ {
		idx := weightedPick(candidates, rng)
		if idx < 0 {
			break
		}
		step := candidates[idx].step
		t.Steps[step] = StepState(true)
		currentHits++
		candidates = append(candidates[:idx], candidates[idx+1:]...)
	}
}

func removeHits(t *Track, r stepRange, rng *rand.Rand, currentHits int) {
	target := int(r.maxDensity * float64(len(r.weights)))
	if target >= currentHits {
		return
	}

	// Collect active steps with inverse probability weights (prefer removing low-prob steps).
	var candidates []stepCandidate
	for i, weight := range r.weights {
		if step := r.start + i; t.Steps[step] {
			candidates = append(candidates, stepCandidate{step, 1.0 - weight + 0.01})
		}
	}

	// Randomly deactivate steps, weighted by inverse probability.
	for attempts := 0; attempts < 200 && currentHits > target && len(candidates) > 0; attempts++ {
		idx := weightedPick(candidates, rng)
		if idx < 0 {
			break
		}
		step := candidates[idx].step
		t.Steps[step] = StepState(false)
		currentHits--
		candidates = append(candidates[:idx], candidates[idx+1:]...)
	}
}

// weightedPick selects a random index from candidates weighted by their weight field.
// Returns -1 if the slice is empty.
func weightedPick(candidates []stepCandidate, rng *rand.Rand) int {
	if len(candidates) == 0 {
		return -1
	}
	total := 0.0
	for _, c := range candidates {
		total += c.weight
	}
	r := rng.Float64() * total
	cumulative := 0.0
	for i, c := range candidates {
		cumulative += c.weight
		if r < cumulative {
			return i
		}
	}
	return len(candidates) - 1
}

// ---- remix algorithms ----

// RemixPattern mutates existing step patterns with genre-aware randomization.
// For each step, with probability equal to intensity, the step is re-rolled
// against the genre profile. intensity=0 means no change, intensity=1 means
// full regeneration. Tracks without a folder assignment are left unchanged.
func RemixPattern(proj *Project, genre Genre, intensity float64, rng *rand.Rand) {
	if intensity <= 0 {
		return
	}
	profile := getProfile(genre)
	for i := range proj.Tracks {
		remixTrack(proj, TrackID(i), genre, intensity, rng, profile)
	}
}

// RemixTrackPattern mutates a single track's pattern with genre-aware randomization.
func RemixTrackPattern(proj *Project, track TrackID, genre Genre, intensity float64, rng *rand.Rand) {
	remixTrack(proj, track, genre, intensity, rng, nil)
}

// remixTrack is the internal implementation that accepts an optional profile.
func remixTrack(proj *Project, track TrackID, genre Genre, intensity float64, rng *rand.Rand, profile *genreProfile) {
	if intensity <= 0 {
		return
	}
	if int(track) < 0 || int(track) >= len(proj.Tracks) {
		return
	}
	t := &proj.Tracks[track]
	if t.Sample.Folder == "" {
		return
	}

	// intensity >= 1.0 is equivalent to full regeneration.
	if intensity >= 1.0 {
		generateTrack(proj, track, genre, rng, profile)
		return
	}

	if profile == nil {
		profile = getProfile(genre)
	}

	rp := roleProbsFor(profile, FolderRole(t.Sample.Folder))
	if rp == nil {
		return
	}

	numSteps := max(proj.NumSteps, 1)
	numSteps = min(numSteps, 64)

	for step := range numSteps {
		if rng.Float64() < intensity {
			t.Steps[step] = StepState(rng.Float64() < rp.steps[step])
		}
	}

	// Re-enforce density after remix (partial re-rolls can drift).
	enforceDensity(t, rp.rangeFor(0, numSteps), rng)
}
