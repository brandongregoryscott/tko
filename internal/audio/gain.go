package audio

import "github.com/gopxl/beep/v2"

// gainStreamer wraps a Streamer and applies a volume multiplier.
type gainStreamer struct {
	streamer beep.Streamer
	gain     float64
}

// Gain wraps a Streamer with a volume multiplier (0.0 to 1.0).
func Gain(s beep.Streamer, vol float64) beep.Streamer {
	return &gainStreamer{streamer: s, gain: vol * vol} // square for perceptual loudness
}

func (g *gainStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = g.streamer.Stream(samples)
	for i := range n {
		samples[i][0] *= g.gain
		samples[i][1] *= g.gain
	}
	return n, ok
}

func (g *gainStreamer) Err() error {
	return nil
}
