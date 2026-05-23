package ui

import (
	"fmt"
	"math/rand/v2"

	"github.com/brandongregoryscott/tko/internal/engine"
)

func (m *Model) cycleSample(delta int) {
	t := &m.sequencer.Project.Tracks[m.cursorTrack]

	if t.Sample.Folder == "" {
		m.cycleTrackFolder()
		return
	}

	count := 0
	if m.audioLib != nil {
		count = m.audioLib.NumSamples(t.Sample.Bank, t.Sample.Folder)
	}
	if count == 0 {
		m.statusMsg = "No samples in " + t.Sample.Folder
		return
	}
	m.sequencer.CycleSample(t.ID, delta, count)
	if m.audioLib != nil {
		if name := m.audioLib.SampleName(t.Sample.Bank, t.Sample.Folder, t.Sample.Index); name != "" {
			t.Sample.Name = name
		}
	}
	m.statusMsg = t.Name + ": " + t.Sample.Name
}

// randomizeSample picks a random sample for the cursor track from the current bank.
// If the track has no folder assigned, it picks a random folder first.
func (m *Model) randomizeSample() {
	t := &m.sequencer.Project.Tracks[m.cursorTrack]
	if m.audioLib == nil {
		m.statusMsg = "No sample library loaded"
		return
	}

	// If no folder assigned, pick a random folder from the current bank.
	if t.Sample.Folder == "" {
		folders := m.audioLib.Folders(m.bank)
		if len(folders) == 0 {
			m.statusMsg = "No sample folders in bank " + m.bank
			return
		}
		folder := folders[rand.IntN(len(folders))]
		m.assignFolder(t, m.bank, folder)
	}

	count := m.audioLib.NumSamples(t.Sample.Bank, t.Sample.Folder)
	if count == 0 {
		m.statusMsg = "No samples in " + t.Sample.Folder
		return
	}
	t.Sample.Index = rand.IntN(count)
	if name := m.audioLib.SampleName(t.Sample.Bank, t.Sample.Folder, t.Sample.Index); name != "" {
		t.Sample.Name = name
	}
	m.statusMsg = t.Name + ": " + t.Sample.Name
}

// randomizeAllSamples picks a random sample for every track that has steps programmed.
func (m *Model) randomizeAllSamples() {
	if m.audioLib == nil {
		m.statusMsg = "No sample library loaded"
		return
	}
	folders := m.audioLib.Folders(m.bank)
	if len(folders) == 0 {
		m.statusMsg = "No sample folders in bank " + m.bank
		return
	}
	count := 0
	proj := m.sequencer.Project
	for i := range proj.Tracks {
		t := &proj.Tracks[i]
		if t.Sample.Folder == "" {
			continue
		}
		hasSteps := false
		for s := 0; s < proj.NumSteps; s++ {
			if t.Steps[s] {
				hasSteps = true
				break
			}
		}
		if !hasSteps {
			continue
		}
		n := m.audioLib.NumSamples(t.Sample.Bank, t.Sample.Folder)
		if n > 0 {
			t.Sample.Index = rand.IntN(n)
			if name := m.audioLib.SampleName(t.Sample.Bank, t.Sample.Folder, t.Sample.Index); name != "" {
				t.Sample.Name = name
			}
			count++
		}
	}
	if count == 0 {
		m.statusMsg = "No tracks with steps to randomize"
	} else {
		m.statusMsg = fmt.Sprintf("Randomized %d track(s)", count)
	}
}

// autoAssignSamples assigns the first sample from each folder in the first bank to tracks.
func (m *Model) autoAssignSamples() {
	if m.audioLib == nil {
		return
	}
	banks := m.audioLib.Banks()
	if len(banks) == 0 {
		return
	}
	bank := banks[0]
	folders := m.audioLib.Folders(bank)
	trackIdx := 0
	for _, folder := range folders {
		if trackIdx >= 8 {
			break
		}
		name := m.audioLib.SampleName(bank, folder, 0)
		m.sequencer.Project.Tracks[trackIdx].Sample = engine.SampleRef{
			Bank:   bank,
			Folder: folder,
			Index:  0,
			Name:   name,
		}
		m.sequencer.Project.Tracks[trackIdx].Name = folder
		trackIdx++
	}
}

// cycleTrackFolder advances the cursor track to the next sample folder within the current bank.
func (m *Model) cycleTrackFolder() {
	t := &m.sequencer.Project.Tracks[m.cursorTrack]
	if m.audioLib == nil {
		m.statusMsg = "No sample library loaded"
		return
	}
	folders := m.audioLib.Folders(m.bank)
	if len(folders) == 0 {
		m.statusMsg = "No sample folders in bank " + m.bank
		return
	}

	cur := -1
	for i, f := range folders {
		if f == t.Sample.Folder {
			cur = i
			break
		}
	}
	next := cur + 1
	if next >= len(folders) {
		next = 0
	}
	m.assignFolder(t, m.bank, folders[next])
}

func (m *Model) assignFolder(t *engine.Track, bank, folder string) {
	name := m.audioLib.SampleName(bank, folder, 0)
	t.Sample = engine.SampleRef{Bank: bank, Folder: folder, Index: 0, Name: name}
	t.Name = folder
	m.statusMsg = t.Name + ": " + name
}

// cycleBank switches all tracks to the next/previous bank, preserving folder assignments
// where the same folder name exists in the new bank. Tracks without a matching folder
// get the next available folder from the new bank (round-robin), so every track has a sound.
func (m *Model) cycleBank(delta int) {
	if m.audioLib == nil {
		return
	}
	banks := m.audioLib.Banks()
	if len(banks) == 0 {
		return
	}

	idx := -1
	for i, b := range banks {
		if b == m.bank {
			idx = i
			break
		}
	}
	idx += delta
	if idx < 0 {
		idx = len(banks) - 1
	} else if idx >= len(banks) {
		idx = 0
	}
	m.bank = banks[idx]

	folders := m.audioLib.Folders(m.bank)
	fallback := 0
	remapped := 0
	for i := range m.sequencer.Project.Tracks {
		t := &m.sequencer.Project.Tracks[i]
		if t.Sample.Folder == "" {
			continue
		}
		if m.audioLib.NumSamples(m.bank, t.Sample.Folder) > 0 {
			name := m.audioLib.SampleName(m.bank, t.Sample.Folder, 0)
			t.Sample.Bank = m.bank
			t.Sample.Index = 0
			t.Sample.Name = name
		} else if len(folders) > 0 {
			f := folders[fallback%len(folders)]
			fallback++
			name := m.audioLib.SampleName(m.bank, f, 0)
			t.Sample = engine.SampleRef{Bank: m.bank, Folder: f, Index: 0, Name: name}
			t.Name = f
			remapped++
		}
	}

	msg := "Bank: " + m.bank
	if remapped > 0 {
		msg += fmt.Sprintf(" (%d remapped)", remapped)
	}
	m.statusMsg = msg
}

// cycleTrackFolderDelta cycles the cursor track's folder by delta (-1 or 1).
func (m *Model) cycleTrackFolderDelta(delta int) {
	t := &m.sequencer.Project.Tracks[m.cursorTrack]
	if m.audioLib == nil {
		m.statusMsg = "No sample library loaded"
		return
	}
	folders := m.audioLib.Folders(m.bank)
	if len(folders) == 0 {
		m.statusMsg = "No sample folders in bank " + m.bank
		return
	}

	cur := -1
	for i, f := range folders {
		if f == t.Sample.Folder {
			cur = i
			break
		}
	}
	next := cur + delta
	if next >= len(folders) {
		next = 0
	} else if next < 0 {
		next = len(folders) - 1
	}
	m.assignFolder(t, m.bank, folders[next])
}
