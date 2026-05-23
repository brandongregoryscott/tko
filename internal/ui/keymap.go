package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all key bindings for the sequencer.
type KeyMap struct {
	CursorUp      key.Binding
	CursorDown    key.Binding
	CursorLeft    key.Binding
	CursorRight   key.Binding
	FirstStep     key.Binding
	LastStep      key.Binding
	ToggleStep    key.Binding
	CycleNext     key.Binding
	CyclePrev     key.Binding
	MuteTrack     key.Binding
	VolUp         key.Binding
	VolDown       key.Binding
	PlayPause     key.Binding
	BPMUp         key.Binding
	BPMDown       key.Binding
	BPMUpFine     key.Binding
	BPMDownFine   key.Binding
	ResetPos      key.Binding
	Save          key.Binding
	Load          key.Binding
	ExportMIDI    key.Binding
	HelpToggle    key.Binding
	Quit          key.Binding
	Tab           key.Binding
	Steps16       key.Binding
	Steps32       key.Binding
	Steps48       key.Binding
	Steps64       key.Binding
	DupTrack      key.Binding
	ClearTrack    key.Binding
	CycleFolder   key.Binding
	CycleBankNext key.Binding
	CycleBankPrev key.Binding
	RandomSample  key.Binding
	RandomizeAll  key.Binding
	SwingUp       key.Binding
	SwingDown     key.Binding
}

// DefaultKeyMap returns the standard key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		CursorUp:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		CursorDown:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		CursorLeft:    key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		CursorRight:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		FirstStep:     key.NewBinding(key.WithKeys("ctrl+a", "home"), key.WithHelp("^A", "first")),
		LastStep:      key.NewBinding(key.WithKeys("ctrl+e", "end"), key.WithHelp("^E", "last")),
		ToggleStep:    key.NewBinding(key.WithKeys(" "), key.WithHelp("␣", "toggle")),
		CycleNext:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "next sample")),
		CyclePrev:     key.NewBinding(key.WithKeys("?", "U"), key.WithHelp("?/U", "prev sample")),
		MuteTrack:     key.NewBinding(key.WithKeys("m", "x"), key.WithHelp("m/x", "mute")),
		VolUp:         key.NewBinding(key.WithKeys("+", "="), key.WithHelp("+", "vol ↑")),
		VolDown:       key.NewBinding(key.WithKeys("-", "_"), key.WithHelp("-", "vol ↓")),
		PlayPause:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("⏎", "play/pause")),
		BPMUp:         key.NewBinding(key.WithKeys(">"), key.WithHelp(">", "bpm↑5")),
		BPMDown:       key.NewBinding(key.WithKeys("<"), key.WithHelp("<", "bpm↓5")),
		BPMUpFine:     key.NewBinding(key.WithKeys("."), key.WithHelp(".", "bpm↑1")),
		BPMDownFine:   key.NewBinding(key.WithKeys(","), key.WithHelp(",", "bpm↓1")),
		ResetPos:      key.NewBinding(key.WithKeys("0"), key.WithHelp("0", "stop/reset")),
		Save:          key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("^S", "save")),
		Load:          key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("^L", "load")),
		ExportMIDI:    key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("^E", "export")),
		HelpToggle:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:          key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Tab:           key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "focus")),
		Steps16:       key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "16 steps")),
		Steps32:       key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "32 steps")),
		Steps48:       key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "48 steps")),
		Steps64:       key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "64 steps")),
		DupTrack:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "dup track")),
		ClearTrack:    key.NewBinding(key.WithKeys("backspace", "delete"), key.WithHelp("⌫", "clear")),
		CycleFolder:   key.NewBinding(key.WithKeys("f", "F"), key.WithHelp("f", "folder")),
		CycleBankNext: key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "next bank")),
		CycleBankPrev: key.NewBinding(key.WithKeys("B"), key.WithHelp("B", "prev bank")),
		RandomSample:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "randomize")),
		RandomizeAll:  key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "randomize all")),
		SwingUp:       key.NewBinding(key.WithKeys("shift+>"), key.WithHelp("⇧>", "swing↑")),
		SwingDown:     key.NewBinding(key.WithKeys("shift+<"), key.WithHelp("⇧<", "swing↓")),
	}
}

// ShortHelp returns key bindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.CursorUp, k.CursorDown, k.CursorLeft, k.CursorRight,
		k.ToggleStep, k.PlayPause, k.Quit, k.HelpToggle,
	}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PlayPause, k.ResetPos, k.BPMUp, k.BPMDown, k.BPMUpFine, k.BPMDownFine},
		{k.CursorUp, k.CursorDown, k.CursorLeft, k.CursorRight, k.FirstStep, k.LastStep},
		{k.ToggleStep, k.CycleNext, k.CyclePrev, k.CycleFolder, k.RandomSample, k.RandomizeAll},
		{k.CycleBankNext, k.CycleBankPrev, k.MuteTrack, k.VolUp, k.VolDown},
		{k.ClearTrack, k.DupTrack, k.Save, k.Load, k.ExportMIDI, k.HelpToggle, k.Quit},
	}
}
