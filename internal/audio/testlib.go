package audio

import "github.com/gopxl/beep/v2"

// NewTestLibrary creates a Library with fake data for testing.
// banks maps bank name → sorted folder names.
// samples maps bank→folder → sample names.
func NewTestLibrary(banks map[string][]string, samples map[string]map[string][]string) *Library {
	lib := &Library{
		buffers:    make(map[string]map[string][]*beep.Buffer),
		names:      make(map[string]map[string][]string),
		sampleRate: 44100,
	}
	for bank, folders := range banks {
		lib.buffers[bank] = make(map[string][]*beep.Buffer)
		lib.names[bank] = make(map[string][]string)
		for _, folder := range folders {
			names := samples[bank][folder]
			lib.buffers[bank][folder] = make([]*beep.Buffer, len(names))
			lib.names[bank][folder] = names
		}
	}
	return lib
}
