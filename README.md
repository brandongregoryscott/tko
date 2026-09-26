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

### Navigation

| Key | Action |
|---|---|
| `↑` / `k` | Cursor up |
| `↓` / `j` | Cursor down |
| `←` / `h` | Cursor left |
| `→` / `l` | Cursor right |
| `Ctrl+A` / `Home` | Jump to first step |
| `Ctrl+E` / `End` | Jump to last step |

### Editing

| Key | Action |
|---|---|
| `Space` | Toggle step on/off |
| `/` | Next sample on current track |
| `U` | Previous sample on current track |
| `f` | Cycle sample folder for current track |
| `r` | Randomize sample on current track |
| `R` | Randomize all track samples |
| `m` / `x` | Mute/unmute current track |
| `+` / `=` | Volume +10% |
| `-` / `_` | Volume −10% |
| `d` | Duplicate track to next empty track |
| `Backspace` / `Delete` | Clear current track |
| `1` / `2` / `3` / `4` | Set 16 / 32 / 48 / 64 steps |
| `b` / `B` | Next / previous bank |

### Pattern Generation

| Key | Action |
|---|---|
| `t` | Cycle genre (hip-hop / boom-bap / breakcore) |
| `g` | Generate pattern on current track |
| `G` | Generate pattern on all tracks |
| `y` | Remix all tracks (partial re-roll) |
| `T` | Cycle fill length (1 beat / 1/2 bar / 1 bar) |
| `i` | Fill current track |
| `I` | Fill all tracks |

### Transport

| Key | Action |
|---|---|
| `Enter` | Play / Pause |
| `0` | Stop and reset to step 1 |
| `>` / `<` | BPM ±5 |
| `.` / `,` | BPM ±1 |
| `Shift+>` / `Shift+<` | Swing ±10% |

### File

| Key | Action |
|---|---|
| `Ctrl+S` | Save project |
| `Ctrl+L` | Load project |
| `Ctrl+E` | Export MIDI + WAV to `projects/` |

### General

| Key | Action |
|---|---|
| `?` | Show/hide key bindings |
| `q` / `Ctrl+C` | Quit |

## Export

`Ctrl+E` exports the current pattern twice, sharing a timestamped base name: a Standard MIDI File (format 1) and a rendered audio loop. Drop the WAV straight into a DAW while exploring ideas, then reference the matching MIDI file when you're ready to map the individual drum sounds.

Exports are named after the project when it has been saved or loaded — `projects/hip-hop-4b_TIMESTAMP.mid` + `.wav` for `hip-hop-4b.json` — and fall back to `projects/tko_TIMESTAMP.mid` + `.wav` for an unsaved project.

Each active track becomes a labeled MIDI track with notes mapped to distinct keys (C3 through C4) for easy identification in a DAW piano roll. The WAV is rendered exactly one pattern long (16-bit stereo, 44.1kHz) so it loops cleanly, and muted tracks are excluded from the audio, matching playback.

## Fills

`i` / `I` rewrite the last few steps of the current track (or every track) as a genre-appropriate fill, leaving the rest of the pattern untouched. `T` cycles how much of the pattern the fill covers — the last beat, the last half bar, or the last bar — and a fill longer than the pattern is clamped to the whole pattern.

Fills are built from the same genre profiles as generation, with tables that intensify toward the end of the window, so a 1-beat fill is the climax of a 1-bar fill. Each drum role behaves differently: snares roll and build, hats and kicks thin out to make room, toms run through the back half, and percussion ramps up. Claps and rimshots roll with the snare, rides with the closed hat, and unrecognized folders are treated as percussion. Crash tracks are cleared through the fill and hit step 1 instead, so the fill lands on a cymbal when the pattern loops.

Fill intensity follows the genre: boom-bap fills stay sparse, hip-hop fills sit in the middle, and breakcore fills are an amen-style rush.

## Samples

Supports WAV files in PCM and IEEE float formats. Samples are resampled to 44.1kHz on load. Mono handled automatically.

## Projects

Projects are saved as JSON in `projects/`. A project stores BPM, step count, and per-track: name, sample assignment, 64 step states, volume, and mute state. `Ctrl+S` opens a filename prompt; `Enter` confirms, `Esc` cancels. You can view some example patterns in the [`projects/`](./projects/) directory in the repo.

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