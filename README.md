# playlist-md

Export Apple Music library playlists to Markdown (don't ask why).

`playlist-md` is a macOS TUI. It embeds a Swift MusicKit engine (`playlist-md-core`) and extracts it to a local cache at runtime.

## Requirements

- macOS 14+
- Swift 5.9+ and Go 1.22+ to build
- Apple Music subscription with library playlists
- Media & Apple Music access for the terminal that runs `playlist-md`

## Install

Download `playlist-md-v<semver>.tar.gz` from [GitHub Releases](https://github.com/imflawlezz/playlist-md/releases), then:

```bash
tar xzf playlist-md-v1.3.0.tar.gz
./playlist-md-v1.3.0/playlist-md
```

The archive contains `README.md`, `LICENSE`, and a universal macOS binary named `playlist-md`. Move it onto your `PATH` if you like. Unsigned; Gatekeeper-clean distribution still needs signing and notarization.

## Build

```bash
make
make install   # PREFIX=/usr/local → $PREFIX/bin/playlist-md
make release-tar   # dist/playlist-md-v<semver>.tar.gz (same layout as releases)
make test
make clean
```

`make` builds a universal (arm64 + x86_64) binary at `dist/playlist-md-v<version-with-dashes>` (e.g. `1.3.0` → `v1-3-0`). `VERSION` in the Makefile controls that suffix.

Pushing a `v*` tag runs tests, builds the same `.tar.gz` via `make release-tar`, and publishes a [GitHub Release](https://github.com/imflawlezz/playlist-md/releases) plus a SHA-256. Release details: [docs/architecture.md#releases](docs/architecture.md#releases).

## Usage

```bash
playlist-md
```

Space toggles playlists; Enter inspects; `/` searches by playlist name or track title, artist, or album. Playlist export is **Export all playlists** or **Export selected playlists**; **Export music library** writes `library.md` separately (`--library` cannot combine with `--all` / `--ids`). Both write Markdown to the output folder from Settings.

Non-interactive:

```bash
playlist-md export --all --output ./music
playlist-md export --ids id1,id2 --output ./music
playlist-md export --library --output ./music
playlist-md export --all --output ./music --write-logs
playlist-md export --all --output ./music --no-write-logs
```

Forward to the engine:

```bash
playlist-md core status
playlist-md core list-playlists
playlist-md core export --all --output ~/Music/AppleMusic --write-logs
playlist-md core export --library --output ~/Music/AppleMusic
```

Core commands, JSON, and progress events: [docs/core-protocol.md](docs/core-protocol.md).

| Key | Action |
|-----|--------|
| ↑↓ / j k | move (search: **↑↓** only — **j** / **k** type in the query) |
| ←→ / h l | page / change value |
| tab | next section |
| space | toggle playlist |
| enter | inspect / confirm |
| / | search |
| a | select all |
| n / c | clear selection |
| o | open output folder |
| r | refresh library |
| s | settings |
| Ctrl+L | repair display |
| ? | keybindings |
| q | quit |

## Configuration

`~/.config/playlist-md/config.json`, or `$XDG_CONFIG_HOME/playlist-md/config.json`.

| Key | Default | Notes |
|-----|---------|--------|
| `output_dir` | `~/AppleMusicExports` | Enter in Settings to edit |
| `playlists_per_page` | `12` | Settings: **Items per page** — `8` / `12` / `16` / `24` / `32` / `40` (←→) |
| `write_export_log` | `true` | Settings: **Export log** — write `export.log` under the output folder (←→) |

## Authorization

First run prompts via the embedded Swift process. If denied: **System Settings → Privacy & Security → Media & Apple Music**.

## Output

```text
<output-directory>/
├── index.md
├── library.md                 ← after Export music library (kept on playlist re-export if present)
├── export.log                 ← when Export log is On (default)
├── .playlist-md-manifest.json
└── playlists/
    └── <slug>.md
```

Playlist export writes playlist Markdown only and keeps an existing `library.md` if present (still linked from `index.md`). **Export music library** / `export --library` writes or refreshes `library.md`. Playlist track totals in the result are a **sum** across playlists (overlap counts twice); library song count is the `library.md` row count. Re-exports delete only stale paths listed in the previous manifest.

Details: [docs/architecture.md#export-layout](docs/architecture.md#export-layout).

## Development

```bash
export PLAYLIST_MD_CORE="$PWD/.build/release/playlist-md-core"
playlist-md
```

Swift engine only:

```bash
swift build -c release --product playlist-md-core
swift test
.build/release/playlist-md-core status
```

## Architecture

```text
playlist-md              Go + Bubble Tea (what you run)
  └── playlist-md-core   Swift / MusicKit (embedded)
```

Details: [docs/architecture.md](docs/architecture.md).

## Limitations

- macOS only (MusicKit)
- Library content only (user playlists and library songs)—not catalog browse, radio, or non-library surfaces
- Unavailable tracks may be skipped or have partial metadata
- Large playlists and libraries are fetched in pages

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE) © 2026 imflawlezz
