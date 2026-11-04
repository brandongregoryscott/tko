// Package wavexport renders the current pattern to a WAV audio file.
package wavexport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/brandongregoryscott/tko/internal/audio"
	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/generators"
	"github.com/gopxl/beep/v2/wav"
)

// Export renders one pass of the project's pattern through the sample library
// and writes it as a 16-bit stereo WAV file to the given path. The render is
// exactly one pattern long so the file loops cleanly in a DAW. Muted tracks
// are skipped, matching what playback sounds like.
func Export(proj *engine.Project, lib *audio.Library, path string, sampleRate beep.SampleRate) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Reuse the sequencer's swing-aware step timing.
	seq := &engine.Sequencer{Project: proj}

	var mixer beep.Mixer
	var total time.Duration
	for step := 0; step < proj.NumSteps; step++ {
		offset := sampleRate.N(total)
		for i := range proj.Tracks {
			t := &proj.Tracks[i]
			if t.Muted || !bool(t.Steps[step]) || t.Sample.Folder == "" {
				continue
			}
			buf := lib.Buffer(t.Sample.Bank, t.Sample.Folder, t.Sample.Index)
			if buf == nil {
				continue
			}
			var s beep.Streamer = buf.Streamer(0, buf.Len())
			if t.Volume < 0.99 {
				s = audio.Gain(s, t.Volume)
			}
			mixer.Add(beep.Seq(generators.Silence(offset), s))
		}
		total += seq.TickDuration(step)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Take bounds the render to the pattern length — the mixer otherwise
	// streams silence forever once its streamers have drained.
	format := beep.Format{SampleRate: sampleRate, NumChannels: 2, Precision: 2}
	return wav.Encode(f, beep.Take(sampleRate.N(total), &mixer), format)
}

// ExportTracks writes basePath_track_01-Name.wav etc., preserving the mix's timing,
// volume and mute settings. Every file spans the whole pattern, including silent
// tracks, so all files can be placed at the same position in a DAW.
// On failure, paths contains the files that were successfully exported.
func ExportTracks(proj *engine.Project, lib *audio.Library, basePath string, sampleRate beep.SampleRate) (paths []string, err error) {
	for i, track := range proj.Tracks {
		isolated := *proj
		for j := range isolated.Tracks {
			if j != i {
				isolated.Tracks[j].Muted = true
			}
		}
		name := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, track.Name)
		name = strings.Trim(name, "-_")
		if name == "" {
			name = "track"
		}
		path := fmt.Sprintf("%s_track_%02d-%s.wav", basePath, i+1, name)
		if err := Export(&isolated, lib, path, sampleRate); err != nil {
			return paths, fmt.Errorf("track %d: %w", i+1, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}
