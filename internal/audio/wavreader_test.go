package audio

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopxl/beep/v2"
)

func TestOpenWAVValid(t *testing.T) {
	path := filepath.Join("testdata", "open-hat.wav")
	r, err := openWAV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if r.format.SampleRate != 44100 {
		t.Errorf("SampleRate: got %d, want 44100", r.format.SampleRate)
	}
	if r.format.NumChannels != 1 {
		t.Errorf("NumChannels: got %d, want 1", r.format.NumChannels)
	}
	if r.format.Precision != 2 {
		t.Errorf("Precision: got %d, want 2 (16-bit)", r.format.Precision)
	}
	if r.floatFmt {
		t.Error("should be PCM, not float")
	}
	if r.Len() <= 0 {
		t.Errorf("Len should be > 0, got %d", r.Len())
	}
}

func TestOpenWAVNonexistent(t *testing.T) {
	_, err := openWAV("nonexistent.wav")
	if err == nil {
		t.Error("should return error for nonexistent file")
	}
}

func TestOpenWAVNotAWAV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-wav.bin")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := openWAV(path)
	if err == nil {
		t.Error("should return error for non-WAV file")
	}
}

func TestWAVPositionAndSeek(t *testing.T) {
	path := filepath.Join("testdata", "open-hat.wav")
	r, err := openWAV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if r.Position() != 0 {
		t.Errorf("initial position should be 0, got %d", r.Position())
	}

	// Seek to sample 100.
	if err := r.Seek(100); err != nil {
		t.Fatal(err)
	}
	if r.Position() != 100 {
		t.Errorf("position after seek: got %d, want 100", r.Position())
	}

	// Seek to 0.
	if err := r.Seek(0); err != nil {
		t.Fatal(err)
	}
	if r.Position() != 0 {
		t.Errorf("position after seek to 0: got %d", r.Position())
	}
}

func TestWAVErr(t *testing.T) {
	path := filepath.Join("testdata", "open-hat.wav")
	r, err := openWAV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if r.Err() != nil {
		t.Error("Err should always return nil")
	}
}

func TestWAVStreamPCM(t *testing.T) {
	path := filepath.Join("testdata", "open-hat.wav")
	r, err := openWAV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	// Read 512 samples.
	var samples [512][2]float64
	n, ok := r.Stream(samples[:])
	if !ok {
		t.Error("Stream should return ok=true for first read")
	}
	if n != 512 {
		t.Errorf("expected 512 samples, got %d", n)
	}

	// Samples should be within [-1, 1] for 16-bit PCM normalized.
	for i := 0; i < n; i++ {
		if samples[i][0] < -1.0 || samples[i][0] > 1.0 {
			t.Errorf("sample %d left channel out of range: %f", i, samples[i][0])
		}
		// Mono file: both channels should be equal.
		if samples[i][0] != samples[i][1] {
			t.Errorf("sample %d: mono file should have equal channels, got [%f, %f]",
				i, samples[i][0], samples[i][1])
		}
	}
}

func TestWAVStreamExhausted(t *testing.T) {
	path := filepath.Join("testdata", "open-hat.wav")
	r, err := openWAV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	total := r.Len()
	// Read all samples in chunks.
	chunk := 1024
	var samples [1024][2]float64
	read := 0
	for {
		n, ok := r.Stream(samples[:])
		read += n
		if !ok {
			break
		}
		if n < chunk {
			break
		}
	}
	// After reading all samples, Stream should return ok=false.
	if read < total-chunk {
		t.Errorf("only read %d of %d samples", read, total)
	}
	// One more read should return 0, false.
	n, ok := r.Stream(samples[:1])
	if ok || n != 0 {
		t.Errorf("after exhaustion: got n=%d ok=%v, want 0 false", n, ok)
	}
}

func TestReadPCM8Bit(t *testing.T) {
	// 8-bit PCM: -128..127 range, unsigned offset of 128 is not applied here.
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int8(42))
	v, err := readPCM(&buf, 1)
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}

	buf.Reset()
	binary.Write(&buf, binary.LittleEndian, int8(-100))
	v, err = readPCM(&buf, 1)
	if err != nil {
		t.Fatal(err)
	}
	if v != -100 {
		t.Errorf("got %d, want -100", v)
	}
}

func TestReadPCM16Bit(t *testing.T) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int16(12345))
	v, err := readPCM(&buf, 2)
	if err != nil {
		t.Fatal(err)
	}
	if v != 12345 {
		t.Errorf("got %d, want 12345", v)
	}

	buf.Reset()
	binary.Write(&buf, binary.LittleEndian, int16(-12345))
	v, err = readPCM(&buf, 2)
	if err != nil {
		t.Fatal(err)
	}
	if v != -12345 {
		t.Errorf("got %d, want -12345", v)
	}
}

func TestReadPCM24Bit(t *testing.T) {
	// 24-bit positive: 0x010203 → 66051
	var buf bytes.Buffer
	buf.Write([]byte{0x03, 0x02, 0x01})
	v, err := readPCM(&buf, 3)
	if err != nil {
		t.Fatal(err)
	}
	expected := int64(0x010203)
	if v != expected {
		t.Errorf("got %d, want %d", v, expected)
	}

	// 24-bit negative (sign-extended): 0xFFFFFF → -1
	buf.Reset()
	buf.Write([]byte{0xFF, 0xFF, 0xFF})
	v, err = readPCM(&buf, 3)
	if err != nil {
		t.Fatal(err)
	}
	if v != -1 {
		t.Errorf("sign-extended: got %d, want -1", v)
	}

	// 24-bit: 0x800000 → sign-extended to -8388608
	buf.Reset()
	buf.Write([]byte{0x00, 0x00, 0x80})
	v, err = readPCM(&buf, 3)
	if err != nil {
		t.Fatal(err)
	}
	if v != -8388608 {
		t.Errorf("sign-extended max negative: got %d, want -8388608", v)
	}
}

func TestReadPCM32Bit(t *testing.T) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int32(1000000))
	v, err := readPCM(&buf, 4)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1000000 {
		t.Errorf("got %d, want 1000000", v)
	}
}

func TestLoadWAV(t *testing.T) {
	path := filepath.Join("testdata", "open-hat.wav")
	buf, err := loadWAV(path, beep.SampleRate(44100))
	if err != nil {
		t.Fatal(err)
	}
	if buf == nil {
		t.Fatal("buffer should not be nil")
	}
	if buf.Len() <= 0 {
		t.Error("buffer should have samples")
	}
}

func TestWAVStreamFloat(t *testing.T) {
	r, err := openWAV(filepath.Join("testdata", "float-stereo.wav"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if !r.floatFmt {
		t.Fatal("expected float format")
	}
	if r.format.NumChannels != 2 {
		t.Fatalf("expected stereo, got %d ch", r.format.NumChannels)
	}

	// Read all samples — float WAVs have no [-1,1] bound.
	var samples [512][2]float64
	read := 0
	for {
		n, ok := r.Stream(samples[:])
		read += n
		if !ok {
			break
		}
	}

	if read != r.Len() {
		t.Errorf("read %d samples, Len()=%d", read, r.Len())
	}
}

func TestWAVStreamFloatExhausted(t *testing.T) {
	r, err := openWAV(filepath.Join("testdata", "float-stereo.wav"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	total := r.Len()
	var samples [1024][2]float64
	read := 0
	for {
		n, ok := r.Stream(samples[:])
		read += n
		if !ok || n < 1024 {
			break
		}
	}
	if read < total-1024 {
		t.Errorf("only read %d of %d samples", read, total)
	}
	// Stream after exhaustion.
	n, ok := r.Stream(samples[:1])
	if ok || n != 0 {
		t.Errorf("after exhaustion: got n=%d ok=%v, want 0 false", n, ok)
	}
}
