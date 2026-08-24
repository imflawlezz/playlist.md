# playlist-md

Export Apple Music library playlists to Markdown (don't ask why).

The user-facing binary is **`playlist-md`**. It embeds a Swift MusicKit engine (`playlist-md-core`) and extracts it to a local cache at runtime.

## Architecture

```text
playlist-md              Go + Bubble Tea TUI (what you run)
  └── playlist-md-core   Swift/MusicKit engine (embedded)
```

Details: [docs/architecture.md](docs/architecture.md). Core JSON CLI: [docs/core-protocol.md](docs/core-protocol.md).

## Requirements

- macOS 14+
- Swift 5.9+ and Go 1.22+ (to build)
- Apple Music subscription with library playlists
- Media & Apple Music access for the terminal running `playlist-md`

## Build

```bash
make
```

Produces a universal (arm64 + x86_64) binary at `dist/playlist-md`.

```bash
make install   # PREFIX=/usr/local by default → $PREFIX/bin/playlist-md
make test
make clean
```

## Run

Interactive TUI:

```bash
playlist-md
```

Non-interactive export:

```bash
playlist-md export --all --output ./music
playlist-md export --ids id1,id2 --output ./music
```

Forward to the Swift engine:

```bash
playlist-md core status
playlist-md core list-playlists
playlist-md core export --all --output ~/Music/AppleMusic
```

Local core during development:

```bash
export PLAYLIST_MD_CORE="$PWD/.build/release/playlist-md-core"
playlist-md
```

## TUI

Home shows playlists plus actions (export, clear selection, open folder, settings). Auth status uses ✓ / ✗ with capitalized labels (Authorized, Denied, Restricted, Not authorized).

Search (`/`) filters by playlist name or track title / artist / album. Results use a fixed-height window (same size as playlists per page); **↑↓** / **j k** scroll the list. Match hints truncate title and artist to fit.

Export shows a block progress bar with percentage and the current playlist name; the completed bar remains on the export-complete screen.

| Key | Action |
|-----|--------|
| ↑↓ / j k | navigate |
| space | toggle playlist |
| enter | inspect playlist |
| / | search |
| a / n c | select all / clear |
| ←→ / h l | page home list (keeps cursor row) |
| tab | list ↔ actions |
| o | open output folder |
| r | reload |
| s | settings |
| ? | keybindings |
| q | quit |

Settings persist to `~/.config/playlist-md/config.json` (or `$XDG_CONFIG_HOME/playlist-md/config.json`):

- `output_dir` — default `~/AppleMusicExports`
- `playlists_per_page` — `8` / `12` / `16` / `24`

## Authorization

First use prompts for Apple Music access via the embedded Swift process.

If denied: **System Settings → Privacy & Security → Media & Apple Music**.

For distribution, code-sign and notarize `playlist-md` (and the embedded core).

## Output

```text
<output-directory>/
├── index.md
├── .playlist-md-manifest.json
└── playlists/
    ├── chill.md
    └── gaming.md
```

Re-exports remove only stale paths recorded in the previous manifest.

## Swift-only development

```bash
swift build -c release --product playlist-md-core
swift test
.build/release/playlist-md-core status
```

## Layout

```text
go/                         Bubble Tea launcher
  cmd/playlistmd/
  internal/engine/          embed + exec Swift core
  internal/ui/
Sources/
  PlaylistMDCore/           Swift library
  playlist_md_core/         Swift engine entry
Tests/
Makefile                    builds dist/playlist-md
docs/                       architecture, protocol, ADRs
LICENSE                     MIT
```

## Limitations

- macOS only (MusicKit)
- Library playlists only
- Unavailable tracks may be skipped or have partial metadata
- Large playlists are fetched in paginated batches

## License

[MIT](LICENSE) © 2026 imflawlezz
