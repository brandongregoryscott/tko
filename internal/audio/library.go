package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/gopxl/beep/v2"
)

// Library manages loaded samples organized by bank then folder.
type Library struct {
	buffers    map[string]map[string][]*beep.Buffer // bank -> folder -> buffers
	names      map[string]map[string][]string       // bank -> folder -> names
	sampleRate beep.SampleRate
}

// LoadProgress reports sample loading progress.
type LoadProgress struct {
	Total   int
	Loaded  int
	Current string
}

// LoadProgressFunc is called as each sample finishes loading.
type LoadProgressFunc func(LoadProgress)

type loadJob struct {
	bank, folder, name, path string
	idx                      int
}

// NewLibrary scans root for banks (subdirs), each containing folders of WAV files,
// and loads all samples in parallel.
func NewLibrary(root string, targetSR beep.SampleRate) (*Library, error) {
	return NewLibraryWithProgress(root, targetSR, nil)
}

// NewLibraryWithProgress loads samples in parallel and calls progress for each file.
func NewLibraryWithProgress(root string, targetSR beep.SampleRate, progress LoadProgressFunc) (*Library, error) {
	lib := &Library{
		buffers:    make(map[string]map[string][]*beep.Buffer),
		names:      make(map[string]map[string][]string),
		sampleRate: targetSR,
	}

	var jobs []loadJob

	bankEntries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read samples dir %s: %w", root, err)
	}

	for _, bankEntry := range bankEntries {
		if !bankEntry.IsDir() || strings.HasPrefix(bankEntry.Name(), ".") {
			continue
		}
		bank := bankEntry.Name()
		bankPath := filepath.Join(root, bank)

		folderEntries, err := os.ReadDir(bankPath)
		if err != nil {
			continue
		}

		lib.buffers[bank] = make(map[string][]*beep.Buffer)
		lib.names[bank] = make(map[string][]string)

		for _, folderEntry := range folderEntries {
			if !folderEntry.IsDir() || strings.HasPrefix(folderEntry.Name(), ".") {
				continue
			}
			folder := folderEntry.Name()
			folderPath := filepath.Join(bankPath, folder)

			files, err := os.ReadDir(folderPath)
			if err != nil {
				continue
			}

			sort.Slice(files, func(i, j int) bool {
				return files[i].Name() < files[j].Name()
			})

			// Count WAV files to pre-allocate slices.
			wavCount := 0
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".wav") {
					wavCount++
				}
			}
			lib.buffers[bank][folder] = make([]*beep.Buffer, wavCount)
			lib.names[bank][folder] = make([]string, wavCount)

			idx := 0
			for _, file := range files {
				if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".wav") {
					continue
				}
				name := strings.TrimSuffix(file.Name(), ".wav")
				name = strings.TrimSuffix(name, ".WAV")

				jobs = append(jobs, loadJob{
					bank:   bank,
					folder: folder,
					name:   name,
					path:   filepath.Join(folderPath, file.Name()),
					idx:    idx,
				})
				idx++
			}
		}
	}

	total := len(jobs)
	if total == 0 {
		return lib, nil
	}

	// Worker pool for parallel WAV loading.
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	if numWorkers > total {
		numWorkers = total
	}

	jobCh := make(chan loadJob, total)
	resultCh := make(chan loadResult, total)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				buf, err := loadWAV(job.path, targetSR)
				resultCh <- loadResult{job: job, buf: buf, err: err}
			}
		}()
	}

	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	loaded := 0
	for result := range resultCh {
		loaded++
		if result.err != nil {
			fmt.Fprintf(os.Stderr, "audio: skip %s: %v\n", result.job.name, result.err)
			continue
		}
		lib.buffers[result.job.bank][result.job.folder][result.job.idx] = result.buf
		lib.names[result.job.bank][result.job.folder][result.job.idx] = result.job.name

		if progress != nil {
			progress(LoadProgress{
				Total:   total,
				Loaded:  loaded,
				Current: result.job.bank + "/" + result.job.folder + "/" + result.job.name,
			})
		}
	}

	return lib, nil
}

type loadResult struct {
	job loadJob
	buf *beep.Buffer
	err error
}

func loadWAV(path string, targetSR beep.SampleRate) (*beep.Buffer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := openWAV(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var s beep.Streamer = r

	if r.format.SampleRate != targetSR {
		s = beep.Resample(4, r.format.SampleRate, targetSR, s)
	}

	buf := beep.NewBuffer(beep.Format{
		SampleRate:  targetSR,
		NumChannels: 2,
		Precision:   4,
	})
	buf.Append(s)
	return buf, nil
}

// Banks returns a sorted list of bank names.
func (l *Library) Banks() []string {
	var banks []string
	for b := range l.buffers {
		banks = append(banks, b)
	}
	sort.Strings(banks)
	return banks
}

// Folders returns a sorted list of folder names within a bank.
func (l *Library) Folders(bank string) []string {
	bankData, ok := l.buffers[bank]
	if !ok {
		return nil
	}
	var folders []string
	for f := range bankData {
		folders = append(folders, f)
	}
	sort.Strings(folders)
	return folders
}

// NumSamples returns the number of samples in a bank/folder.
func (l *Library) NumSamples(bank, folder string) int {
	if bankData, ok := l.buffers[bank]; ok {
		return len(bankData[folder])
	}
	return 0
}

// SampleName returns the display name for a sample.
func (l *Library) SampleName(bank, folder string, idx int) string {
	if bankData, ok := l.names[bank]; ok {
		if names, ok := bankData[folder]; ok {
			if idx >= 0 && idx < len(names) {
				return names[idx]
			}
		}
	}
	return ""
}

// Counts returns a map of folder name to sample count for a bank.
func (l *Library) Counts(bank string) map[string]int {
	result := make(map[string]int)
	if bankData, ok := l.buffers[bank]; ok {
		for folder, bufs := range bankData {
			result[folder] = len(bufs)
		}
	}
	return result
}

// Buffer returns the audio buffer for a sample.
func (l *Library) Buffer(bank, folder string, idx int) *beep.Buffer {
	if bankData, ok := l.buffers[bank]; ok {
		if bufs, ok := bankData[folder]; ok {
			if idx >= 0 && idx < len(bufs) {
				return bufs[idx]
			}
		}
	}
	return nil
}
