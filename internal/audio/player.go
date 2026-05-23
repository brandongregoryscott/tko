package audio

import (
	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// Player manages real-time sample playback via the beep speaker.
type Player struct {
	lib *Library
}

// NewPlayer creates a new player. speaker.Init must have been called first.
func NewPlayer(lib *Library) *Player {
	return &Player{lib: lib}
}

// Trigger plays samples for the given track IDs immediately.
func (p *Player) Trigger(triggers []engine.TrackID, proj *engine.Project) {
	if len(triggers) == 0 {
		return
	}

	for _, id := range triggers {
		t := &proj.Tracks[id]
		buf := p.lib.Buffer(t.Sample.Bank, t.Sample.Folder, t.Sample.Index)
		if buf == nil {
			continue
		}
		var s beep.Streamer = buf.Streamer(0, buf.Len())
		if t.Volume < 0.99 {
			s = Gain(s, t.Volume)
		}
		speaker.Play(s)
	}
}

// Close cleans up audio resources.
func (p *Player) Close() error {
	speaker.Close()
	return nil
}
