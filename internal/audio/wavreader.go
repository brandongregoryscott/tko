package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/gopxl/beep/v2"
)

// wavReader decodes WAV files (PCM and IEEE float).
type wavReader struct {
	f        *os.File
	dataSize int64
	dataPos  int64
	format   beep.Format
	floatFmt bool
}

// openWAV opens a WAV file and reads the header.
func openWAV(path string) (*wavReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Read RIFF header.
	var riff [12]byte
	if _, err := io.ReadFull(f, riff[:]); err != nil {
		f.Close()
		return nil, fmt.Errorf("not a WAV file")
	}
	if string(riff[0:4]) != "RIFF" || string(riff[8:12]) != "WAVE" {
		f.Close()
		return nil, fmt.Errorf("not a WAV file")
	}

	var (
		audioFormat   uint16
		numChannels   uint16
		sampleRate    uint32
		bitsPerSample uint16
		dataSize      uint32
		foundFmt      bool
		foundData     bool
	)

	// Scan chunks.
	for !foundFmt || !foundData {
		var chunkID [4]byte
		var chunkSize uint32
		if err := binary.Read(f, binary.LittleEndian, &chunkID); err != nil {
			f.Close()
			return nil, fmt.Errorf("truncated WAV")
		}
		if err := binary.Read(f, binary.LittleEndian, &chunkSize); err != nil {
			f.Close()
			return nil, fmt.Errorf("truncated WAV")
		}

		switch string(chunkID[:]) {
		case "fmt ":
			if err := binary.Read(f, binary.LittleEndian, &audioFormat); err != nil {
				f.Close()
				return nil, err
			}
			if err := binary.Read(f, binary.LittleEndian, &numChannels); err != nil {
				f.Close()
				return nil, err
			}
			if err := binary.Read(f, binary.LittleEndian, &sampleRate); err != nil {
				f.Close()
				return nil, err
			}
			// Skip byteRate (4) + blockAlign (2).
			if _, err := f.Seek(6, io.SeekCurrent); err != nil {
				f.Close()
				return nil, err
			}
			if err := binary.Read(f, binary.LittleEndian, &bitsPerSample); err != nil {
				f.Close()
				return nil, err
			}
			// Skip remaining fmt bytes if chunkSize > 16.
			if chunkSize > 16 {
				if _, err := f.Seek(int64(chunkSize-16), io.SeekCurrent); err != nil {
					f.Close()
					return nil, err
				}
			}
			foundFmt = true

		case "data":
			dataSize = chunkSize
			foundData = true

		default:
			// Skip unknown chunk.
			if _, err := f.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				f.Close()
				return nil, err
			}
		}
	}

	dataPos, _ := f.Seek(0, io.SeekCurrent)

	if audioFormat != 1 && audioFormat != 3 {
		f.Close()
		return nil, fmt.Errorf("unsupported format type - %d", audioFormat)
	}

	precision := bitsPerSample / 8
	if precision < 1 {
		precision = 1
	}
	if precision > 4 {
		precision = 4
	}

	return &wavReader{
		f:        f,
		dataSize: int64(dataSize),
		dataPos:  dataPos,
		floatFmt: audioFormat == 3,
		format: beep.Format{
			SampleRate:  beep.SampleRate(sampleRate),
			NumChannels: int(numChannels),
			Precision:   int(precision),
		},
	}, nil
}

func (w *wavReader) Stream(samples [][2]float64) (n int, ok bool) {
	if w.floatFmt {
		return w.streamFloat(samples)
	}
	return w.streamPCM(samples)
}

func (w *wavReader) streamFloat(samples [][2]float64) (n int, ok bool) {
	ch := w.format.NumChannels
	var buf [2]float32

	for i := range samples {
		curPos, _ := w.f.Seek(0, io.SeekCurrent)
		if curPos-w.dataPos >= w.dataSize {
			return n, false
		}

		if ch == 1 {
			var v float32
			if err := binary.Read(w.f, binary.LittleEndian, &v); err != nil {
				return n, false
			}
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				v = 0
			}
			samples[i][0] = float64(v)
			samples[i][1] = float64(v)
		} else {
			if err := binary.Read(w.f, binary.LittleEndian, &buf); err != nil {
				return n, false
			}
			if math.IsNaN(float64(buf[0])) || math.IsInf(float64(buf[0]), 0) {
				buf[0] = 0
			}
			if math.IsNaN(float64(buf[1])) || math.IsInf(float64(buf[1]), 0) {
				buf[1] = 0
			}
			samples[i][0] = float64(buf[0])
			samples[i][1] = float64(buf[1])
		}
		n++
	}
	return n, true
}

func (w *wavReader) streamPCM(samples [][2]float64) (n int, ok bool) {
	ch := w.format.NumChannels
	bytesPerSample := w.format.Precision
	scale := 1.0 / float64(int(1)<<(bytesPerSample*8-1))

	for i := range samples {
		curPos, _ := w.f.Seek(0, io.SeekCurrent)
		if curPos-w.dataPos >= w.dataSize {
			return n, false
		}

		if ch == 1 {
			v, err := readPCM(w.f, bytesPerSample)
			if err != nil {
				return n, false
			}
			samples[i][0] = float64(v) * scale
			samples[i][1] = float64(v) * scale
		} else {
			v0, err := readPCM(w.f, bytesPerSample)
			if err != nil {
				return n, false
			}
			v1, err := readPCM(w.f, bytesPerSample)
			if err != nil {
				return n, false
			}
			samples[i][0] = float64(v0) * scale
			samples[i][1] = float64(v1) * scale
		}
		n++
	}
	return n, true
}

func readPCM(r io.Reader, bytesPerSample int) (int64, error) {
	var (
		v8  int8
		v16 int16
		v24 [3]byte
		v32 int32
	)
	switch bytesPerSample {
	case 1:
		if err := binary.Read(r, binary.LittleEndian, &v8); err != nil {
			return 0, err
		}
		return int64(v8), nil
	case 2:
		if err := binary.Read(r, binary.LittleEndian, &v16); err != nil {
			return 0, err
		}
		return int64(v16), nil
	case 3:
		if _, err := io.ReadFull(r, v24[:]); err != nil {
			return 0, err
		}
		// Sign-extend 24-bit.
		v := int32(v24[0]) | int32(v24[1])<<8 | int32(v24[2])<<16
		if v&0x800000 != 0 {
			v |= ^0xFFFFFF
		}
		return int64(v), nil
	case 4:
		if err := binary.Read(r, binary.LittleEndian, &v32); err != nil {
			return 0, err
		}
		return int64(v32), nil
	default:
		return 0, fmt.Errorf("unsupported precision: %d", bytesPerSample)
	}
}

func (w *wavReader) Len() int {
	totalSamples := w.dataSize / int64(w.format.NumChannels*w.format.Precision)
	return int(totalSamples)
}

func (w *wavReader) Position() int {
	cur, _ := w.f.Seek(0, io.SeekCurrent)
	return int((cur - w.dataPos) / int64(w.format.NumChannels*w.format.Precision))
}

func (w *wavReader) Seek(p int) error {
	_, err := w.f.Seek(w.dataPos+int64(p*w.format.NumChannels*w.format.Precision), io.SeekStart)
	return err
}

func (w *wavReader) Close() error {
	return w.f.Close()
}

func (w *wavReader) Err() error {
	return nil
}
