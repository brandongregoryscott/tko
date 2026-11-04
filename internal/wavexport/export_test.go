package wavexport

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/brandongregoryscott/tko/internal/audio"
	"github.com/brandongregoryscott/tko/internal/engine"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/wav"
)

const testSampleRate = beep.SampleRate(44100)

// writeSample writes a constant-value stereo WAV sample into root/bank/folder/name.wav.
func writeSample(t *testing.T, root, bank, folder, name string, value float64, numSamples int) {
	t.Helper()
	dir := filepath.Join(root, bank, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name+".wav")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	remaining := numSamples
	s := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if remaining <= 0 {
			return 0, false
		}
		n = min(remaining, len(samples))
		for i := range n {
			samples[i][0] = value
			samples[i][1] = value
		}
		remaining -= n
		return n, true
	})
	format := beep.Format{SampleRate: testSampleRate, NumChannels: 2, Precision: 2}
	if err := wav.Encode(f, s, format); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

// readWAV parses the 44-byte header wav.Encode produces and decodes the frames
// to float64 in [-1, 1].
func readWAV(t *testing.T, path string) (channels, sampleRate, bits int, frames [][2]float64) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 44 {
		t.Fatal("file too short for WAV header")
	}
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		t.Fatal("missing RIFF/WAVE header")
	}
	channels = int(binary.LittleEndian.Uint16(data[22:24]))
	sampleRate = int(binary.LittleEndian.Uint32(data[24:28]))
	bits = int(binary.LittleEndian.Uint16(data[34:36]))
	dataSize := int(binary.LittleEndian.Uint32(data[40:44]))
	if len(data) < 44+dataSize {
		t.Fatalf("data size mismatch: header says %d, file has %d", dataSize, len(data)-44)
	}
	for p := 44; p+4 <= 44+dataSize; p += 4 {
		l := float64(int16(binary.LittleEndian.Uint16(data[p:p+2]))) / 32768
		r := float64(int16(binary.LittleEndian.Uint16(data[p+2:p+4]))) / 32768
		frames = append(frames, [2]float64{l, r})
	}
	return channels, sampleRate, bits, frames
}

// testProject returns a 16-step, 120 BPM project (one step = 125ms = 5512.5 samples,
// one pass = 2s = 88200 samples).
func testProject() *engine.Project {
	proj := engine.DefaultProject()
	proj.BPM = 120
	proj.NumSteps = 16
	return proj
}

func newLibrary(t *testing.T, root string) *audio.Library {
	t.Helper()
	lib, err := audio.NewLibrary(root, testSampleRate)
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

func stepFrames(step int) int {
	d := time.Duration(step) * time.Minute / (120 * 4)
	return testSampleRate.N(d)
}

func TestExportWritesValidWAV(t *testing.T) {
	root := t.TempDir()
	writeSample(t, root, "kit", "kick", "kick_01", 0.5, 200)
	lib := newLibrary(t, root)

	proj := testProject()
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "kit", Folder: "kick", Index: 0, Name: "kick_01"}
	proj.Tracks[0].Steps[0] = true

	path := filepath.Join(t.TempDir(), "out.wav")
	if err := Export(proj, lib, path, testSampleRate); err != nil {
		t.Fatalf("export error: %v", err)
	}

	channels, sampleRate, bits, frames := readWAV(t, path)
	if channels != 2 {
		t.Errorf("expected 2 channels, got %d", channels)
	}
	if sampleRate != 44100 {
		t.Errorf("expected 44100 sample rate, got %d", sampleRate)
	}
	if bits != 16 {
		t.Errorf("expected 16 bits, got %d", bits)
	}
	// 16 steps at 120 BPM = 2 seconds.
	if len(frames) != 88200 {
		t.Errorf("expected 88200 frames (2s), got %d", len(frames))
	}
}

func TestExportTriggersSamplesAtSteps(t *testing.T) {
	root := t.TempDir()
	writeSample(t, root, "kit", "kick", "kick_01", 0.5, 200)
	lib := newLibrary(t, root)

	proj := testProject()
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "kit", Folder: "kick", Index: 0, Name: "kick_01"}
	proj.Tracks[0].Steps[0] = true
	proj.Tracks[0].Steps[2] = true

	path := filepath.Join(t.TempDir(), "out.wav")
	if err := Export(proj, lib, path, testSampleRate); err != nil {
		t.Fatalf("export error: %v", err)
	}

	_, _, _, frames := readWAV(t, path)

	const tol = 0.01
	assertFrame := func(i int, want float64) {
		t.Helper()
		if math.Abs(frames[i][0]-want) > tol || math.Abs(frames[i][1]-want) > tol {
			t.Errorf("frame %d: got [%f, %f], want ~%f", i, frames[i][0], frames[i][1], want)
		}
	}

	// First hit starts at frame 0 and plays for the sample's 200 frames.
	assertFrame(0, 0.5)
	assertFrame(199, 0.5)
	// Silence between the hits (step 1 is inactive).
	assertFrame(stepFrames(1)+50, 0)
	// Second hit starts two steps in.
	assertFrame(stepFrames(2)+100, 0.5)
}

func TestExportSkipsMutedTracks(t *testing.T) {
	root := t.TempDir()
	writeSample(t, root, "kit", "kick", "kick_01", 0.5, 200)
	lib := newLibrary(t, root)

	proj := testProject()
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "kit", Folder: "kick", Index: 0, Name: "kick_01"}
	proj.Tracks[0].Steps[0] = true
	proj.Tracks[0].Muted = true

	path := filepath.Join(t.TempDir(), "out.wav")
	if err := Export(proj, lib, path, testSampleRate); err != nil {
		t.Fatalf("export error: %v", err)
	}

	_, _, _, frames := readWAV(t, path)
	for i, fr := range frames {
		if fr[0] != 0 || fr[1] != 0 {
			t.Fatalf("frame %d: expected silence, got [%f, %f]", i, fr[0], fr[1])
		}
	}
}

func TestExportAppliesTrackVolume(t *testing.T) {
	root := t.TempDir()
	writeSample(t, root, "kit", "kick", "kick_01", 0.5, 200)
	lib := newLibrary(t, root)

	proj := testProject()
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "kit", Folder: "kick", Index: 0, Name: "kick_01"}
	proj.Tracks[0].Steps[0] = true
	proj.Tracks[0].Volume = 0.5

	path := filepath.Join(t.TempDir(), "out.wav")
	if err := Export(proj, lib, path, testSampleRate); err != nil {
		t.Fatalf("export error: %v", err)
	}

	_, _, _, frames := readWAV(t, path)
	// Gain squares the volume (perceptual loudness): 0.5^2 = 0.25 applied to 0.5.
	want := 0.5 * 0.25
	if math.Abs(frames[0][0]-want) > 0.01 {
		t.Errorf("frame 0: got %f, want ~%f", frames[0][0], want)
	}
}

func TestExportHonorsSwing(t *testing.T) {
	root := t.TempDir()
	writeSample(t, root, "kit", "kick", "kick_01", 0.5, 200)
	lib := newLibrary(t, root)

	proj := testProject()
	proj.Swing = 0.5
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "kit", Folder: "kick", Index: 0, Name: "kick_01"}
	proj.Tracks[0].Steps[1] = true

	path := filepath.Join(t.TempDir(), "out.wav")
	if err := Export(proj, lib, path, testSampleRate); err != nil {
		t.Fatalf("export error: %v", err)
	}

	_, _, _, frames := readWAV(t, path)

	// With 50% swing the even step 0 shortens to 62.5ms, so the hit on step 1
	// lands earlier than the straight 125ms position.
	swung := testSampleRate.N(62500 * time.Microsecond)
	if frames[swung-100][0] != 0 {
		t.Errorf("frame %d: expected silence before swung onset, got %f", swung-100, frames[swung-100][0])
	}
	if math.Abs(frames[swung+100][0]-0.5) > 0.01 {
		t.Errorf("frame %d: got %f, want ~0.5 (hit on swung step 1)", swung+100, frames[swung+100][0])
	}
}

func TestExportSkipsMissingSamples(t *testing.T) {
	root := t.TempDir()
	writeSample(t, root, "kit", "kick", "kick_01", 0.5, 200)
	lib := newLibrary(t, root)

	proj := testProject()
	// References a sample that doesn't exist in the library.
	proj.Tracks[0].Sample = engine.SampleRef{Bank: "kit", Folder: "snare", Index: 0, Name: "snare_01"}
	proj.Tracks[0].Steps[0] = true

	path := filepath.Join(t.TempDir(), "out.wav")
	if err := Export(proj, lib, path, testSampleRate); err != nil {
		t.Fatalf("export error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to be written: %v", err)
	}
}

func TestExportTracksAlignWithMix(t *testing.T) {
	root := t.TempDir()
	writeSample(t, root, "kit", "kick", "kick", 0.5, 200)
	writeSample(t, root, "kit", "snare", "snare", 0.25, 200)
	lib := newLibrary(t, root)
	proj := testProject()
	proj.Swing = 0.5
	for i := range 4 {
		proj.Tracks[i].Sample = engine.SampleRef{Bank: "kit", Folder: "kick"}
		proj.Tracks[i].Steps[0] = true
	}
	proj.Tracks[0].Volume = 0.5
	proj.Tracks[1].Sample.Folder = "snare"
	proj.Tracks[1].Steps[1] = true
	proj.Tracks[2].Muted = true
	proj.Tracks[3].Sample.Folder = "missing"
	// Duplicate and unsafe names must still produce distinct files in dir.
	proj.Tracks[0].Name = "../Kick\\:*?"
	proj.Tracks[1].Name = proj.Tracks[0].Name
	proj.Tracks[2].Name = ""
	before := *proj
	dir := filepath.Join(t.TempDir(), "stems")
	paths, err := ExportTracks(proj, lib, filepath.Join(dir, "foo"), testSampleRate)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 8 {
		t.Fatalf("got %d stems, want 8", len(paths))
	}
	if *proj != before {
		t.Fatal("export modified the project")
	}
	if filepath.Base(paths[0]) != "foo_track_01-Kick.wav" || filepath.Base(paths[1]) != "foo_track_02-Kick.wav" {
		t.Fatalf("unexpected sanitized names: %v", paths[:2])
	}
	mixPath := filepath.Join(t.TempDir(), "mix.wav")
	if err := Export(proj, lib, mixPath, testSampleRate); err != nil {
		t.Fatal(err)
	}
	_, _, _, mix := readWAV(t, mixPath)
	sum := make([][2]float64, len(mix))
	onset := testSampleRate.N(62500 * time.Microsecond)
	for i, path := range paths {
		if filepath.Dir(path) != dir {
			t.Fatalf("track escaped output directory: %s", path)
		}
		channels, rate, bits, frames := readWAV(t, path)
		if channels != 2 || rate != int(testSampleRate) || bits != 16 || len(frames) != len(mix) {
			t.Fatalf("track %d format or length differs from mix", i+1)
		}
		for j, frame := range frames {
			if i >= 2 && frame != [2]float64{} {
				t.Fatalf("silent track %d has audio at frame %d", i+1, j)
			}
			for ch := range 2 {
				sum[j][ch] += frame[ch]
			}
		}
		if i == 0 && (math.Abs(frames[0][0]-0.125) > 0.001 || frames[onset][0] != 0) {
			t.Fatal("kick stem has incorrect volume or contains snare audio")
		}
		if i == 1 && (math.Abs(frames[onset][0]-0.25) > 0.001 || frames[onset-1][0] != 0) {
			t.Fatal("snare stem does not preserve swung onset")
		}
	}
	for i, frame := range mix {
		for ch := range 2 {
			if math.Abs(sum[i][ch]-frame[ch]) > 2.0/32768 {
				t.Fatalf("stems do not sum to mix at frame %d, channel %d", i, ch)
			}
		}
	}
}

func TestExportTracksReportsPartialFailure(t *testing.T) {
	dir := t.TempDir()
	// Block the second track's destination after the first can be exported.
	if err := os.Mkdir(filepath.Join(dir, "foo_track_02-Snare.wav"), 0755); err != nil {
		t.Fatal(err)
	}
	lib := newLibrary(t, t.TempDir())
	paths, err := ExportTracks(testProject(), lib, filepath.Join(dir, "foo"), testSampleRate)
	if err == nil {
		t.Fatal("expected an error for an unwritable track path")
	}
	if len(paths) != 1 {
		t.Fatalf("got %d successful paths, want 1", len(paths))
	}
	if _, err := os.Stat(paths[0]); err != nil {
		t.Fatal(err)
	}
}
