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
