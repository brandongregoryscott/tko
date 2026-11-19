# tko

A terminal-based step sequencer for constructing beats. 8 tracks, 64 steps, real-time sample playback, MIDI + WAV export.

## Install

```bash
go install github.com/brandongregoryscott/tko/cmd/tko@latest
```

Or download a pre-built binary from the [releases page](https://github.com/brandongregoryscott/tko/releases).

## Quick Start

```bash
tko
```

Place WAV samples in subdirectories under `samples/`:

```
samples/
  kick/
    kick-01.wav
    kick-02.wav
  snare/
    snare-01.wav
  closed-hat/
    hat-01.wav
```

Each subdirectory becomes a sample type. The first sample from each folder is auto-assigned to a track on startup.

Samples can also be grouped into banks, which can be easily swapped:

```
samples/
  lofi-kit/
    kick/
      kick-01.wav
    snare/
      snare-01.wav
  boombap-kit/
    kick/
      kick-01.wav
    snare/
      snare-01.wav
```

## Key Bindings

<!-- BEGIN GENERATED KEY BINDINGS -->
<!-- Generated from controls.json; run go generate ./... to update. -->

### Navigation

| Key | Action |
|---|---|
| <code>↑</code> / <code>k</code> | Cursor up |
| <code>↓</code> / <code>j</code> | Cursor down |
| <code>←</code> / <code>h</code> | Cursor left |
| <code>→</code> / <code>l</code> | Cursor right |
| <code>Ctrl+A</code> / <code>Home</code> | Jump to first step |
| <code>End</code> | Jump to last step |

### Editing

| Key | Action |
|---|---|
| <code>Space</code> | Toggle step on/off |
| <code>]</code> | Next sample on current track |
| <code>[</code> | Previous sample on current track |
| <code>}</code> | Move forward 10 samples, wrapping within the folder |
| <code>{</code> | Move backward 10 samples, wrapping within the folder |
| <code>f</code> / <code>F</code> | Cycle sample folder for current track |
| <code>r</code> | Randomize sample on current track |
| <code>R</code> | Randomize samples on all tracks with steps |
| <code>m</code> / <code>x</code> | Mute/unmute current track |
| <code>+</code> / <code>=</code> | Volume +10% |
| <code>-</code> / <code>_</code> | Volume −10% |
| <code>d</code> | Duplicate track to next empty track |
| <code>Backspace</code> / <code>Delete</code> | Clear current track |
| <code>1</code> | Set 16 steps |
| <code>2</code> | Set 32 steps |
| <code>3</code> | Set 48 steps |
| <code>4</code> | Set 64 steps |
| <code>b</code> | Next bank |
| <code>B</code> | Previous bank |

### Pattern Generation

New sessions start with the boom-bap genre at 80 BPM. Randomizing changes samples and keeps the steps; generating rewrites step patterns and keeps the samples.

| Key | Action |
|---|---|
| <code>t</code> | Cycle genre (hip-hop / boom-bap / breakcore) |
| <code>g</code> | Generate pattern on current track |
| <code>G</code> | Generate pattern on all tracks |
| <code>y</code> | Remix all tracks (partial re-roll) |
| <code>T</code> | Cycle fill length (1 beat / 1/2 bar / 1 bar) |
| <code>i</code> | Fill current track |
| <code>I</code> | Fill all tracks |

### Transport

| Key | Action |
|---|---|
| <code>Enter</code> | Play / Pause |
| <code>0</code> | Stop and reset to step 1 |
| <code>&gt;</code> | BPM +5 |
| <code>&lt;</code> | BPM −5 |
| <code>.</code> | BPM +1 |
| <code>,</code> | BPM −1 |
| <code>s</code> | Swing +10% (0–100%) |
| <code>S</code> | Swing −10% (0–100%) |

Swing is shown beside BPM, including when set to 0% (straight timing), and is saved with the project.

### File

| Key | Action |
|---|---|
| <code>Ctrl+S</code> | Save project |
| <code>Ctrl+L</code> | Load project |
| <code>Ctrl+E</code> | Export MIDI + combined WAV + track WAVs to exports/ |

### General

| Key | Action |
|---|---|
| <code>?</code> | Show/hide key bindings |
| <code>q</code> / <code>Ctrl+C</code> | Quit |

<!-- END GENERATED KEY BINDINGS -->

## Export

Export writes the current pattern as a Standard MIDI File (format 1), a combined WAV loop, and individual track WAVs. Arrange the combined loop in your DAW, then bring in the track WAVs to mix individual drums.

Exports are grouped by project under `exports/`, separate from saved project JSON files in `projects/`. Each export creates its own folder using the existing `YYYYMMDD_HHMMSS` timestamp. Repeated exports within the same second get a `_2`, `_3`, etc. suffix so earlier exports are preserved. Unsaved projects use `tko` as the project name.

For a project named `foo`, all files live directly in the export folder:

```text
exports/foo/foo_20261104_150405/
  foo_bounce.wav
  foo.mid
  foo_track_01-Kick.wav
  foo_track_02-Snare.wav
  ...
```

All eight track WAVs start at the same point and span the full pattern, preserving swing and track volume. Muted, empty, or unavailable-sample tracks render as silence. Place all track WAVs at the same position as the combined loop to keep them aligned.

Each active track becomes a labeled MIDI track with notes mapped to distinct keys (C3 through C4) for easy identification in a DAW piano roll. The WAV is rendered exactly one pattern long (16-bit stereo, 44.1kHz) so it loops cleanly, and muted tracks are excluded from the audio, matching playback.

## Fills

Fills rewrite the last few steps of the current track (or every track) as a genre-appropriate fill, leaving the rest of the pattern untouched. The fill-length control cycles how much of the pattern the fill covers — the last beat, the last half bar, or the last bar — and a fill longer than the pattern is clamped to the whole pattern.

Fills are built from the same genre profiles as generation, with tables that intensify toward the end of the window, so a 1-beat fill is the climax of a 1-bar fill. Each drum role behaves differently: snares roll and build, hats and kicks thin out to make room, toms run through the back half, and percussion ramps up. Claps and rimshots roll with the snare, rides with the closed hat, and unrecognized folders are treated as percussion. Crash tracks are cleared through the fill and hit step 1 instead, so the fill lands on a cymbal when the pattern loops.

Fill intensity follows the genre: boom-bap fills stay sparse, hip-hop fills sit in the middle, and breakcore fills are an amen-style rush.

## Samples

Supports WAV files in PCM and IEEE float formats. Samples are resampled to 44.1kHz on load. Mono handled automatically.

Samples cycle in natural filename order: `Snare (7)`, `Snare (69)`, `Snare (70)`, etc. Numbers are compared numerically, including in names with zero padding. Saved projects resolve samples by name so changes in library ordering preserve their sample choices.

## Projects

Projects are saved as JSON in `projects/`. A project stores BPM, swing, step count, and per-track: name, sample assignment, 64 step states, volume, and mute state. Saving opens a filename prompt; `Enter` confirms, `Esc` cancels. You can view some example patterns in the [`projects/`](./projects/) directory in the repo.

## Monome Grid Hardware

Optional support for the [Monome Grid](https://monome.org/docs/grid/) via [serialosc](https://github.com/monome/serialosc). The app works (more or less) the same without hardware.

The grid's 16×8 button layout maps to the step sequencer — rows 0–5 are the step grid (6 tracks visible, scroll with ▼/▲), row 6 has controls (step pages, randomize, folder/sample cycling), and row 7 has transport (play/pause, stop, mute, BPM). Press any step-grid button to toggle that step. Modifier buttons (B, b, F, S) held with ▲/▼ perform BPM, bank, folder, or sample adjustments.

Connection is automatic via serialosc discovery on startup. Disconnected grids are re-polled every 5 seconds.

```
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15
    ┌──────────────────────────────────────────────────┐
  0 │  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  │  Track T+0
  1 │  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  │  Track T+1
  2 │  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  │  Track T+2
  3 │  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  │  Track T+3
  4 │  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  │  Track T+4
  5 │  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  ●  │  Track T+5
  6 │  1  2  3  4  ·  ·  ·  ·  ·  ·  ·  r  b  F  ▲  S  │  Controls
  7 │  ▶  ■  M  ·  ·  ·  ·  ·  ·  ·  ·  R  B  ◄  ▼  ►  │  Transport
    └──────────────────────────────────────────────────┘
```

## Disclaimer

This project is almost entirely vibe-coded. I don't know much Go, so I can't speak to the code quality. There are likely bugs or unexpected behaviors, or intended behaviors that just don't make sense to you, and that's okay. I built this as a way to quickly prototype drum patterns with different sample packs in a workflow similar to the PO-33.

## Updating keyboard controls

Edit [`controls.json`](./controls.json), then run `go generate ./...` (or `make generate`). This generates the Go key map, in-app help, and the key-binding tables above. Keys use Bubble Tea names (for example `ctrl+s`, `up`, or a literal space); `help_key` and `help` are compact labels for in-app help, and `description` is the README action text. Section order controls the README; `short_help` and `full_help` control the in-app help layout. Action IDs match the handlers in `internal/ui/keyboard.go`; adding a new action also requires a handler.

Commit the config and generated files together. `make generate-check` verifies they are in sync without modifying files and runs in CI. The config defines built-in sequencer keyboard defaults; it is not loaded from the working directory at runtime. File-dialog keys and Monome Grid controls are defined separately.
