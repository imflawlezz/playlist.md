# playlist-md

Export Apple Music library playlists to Markdown (don't ask why).

`playlist-md` is a macOS TUI. It embeds a Swift MusicKit engine (`playlist-md-core`) and extracts it to a local cache at runtime.

## Requirements

- macOS 14+
- Swift 5.9+ and Go 1.22+ to build
- Apple Music subscription with library playlists
- Media & Apple Music access for the terminal that runs `playlist-md`

## Build

```bash
make
make install   # PREFIX=/usr/local → $PREFIX/bin/playlist-md
make test
make clean
```

`make` builds a universal (arm64 + x86_64) binary at `dist/playlist-md-v1-1-0`. `VERSION` in the Makefile controls that suffix (dots become dashes).

Pushing a `v*` tag runs tests, builds that binary, and publishes a [GitHub Release](https://github.com/imflawlezz/playlist-md/releases) plus a SHA-256. Unsigned; Gatekeeper-clean distribution still needs signing and notarization. Release details: [docs/architecture.md#releases](docs/architecture.md#releases).

## Usage

```bash
playlist-md
```

Space toggles playlists; Enter inspects; `/` searches by playlist name or track title, artist, or album. Export writes Markdown to the output folder from Settings.

Non-interactive:

```bash
playlist-md export --all --output ./music
playlist-md export --ids id1,id2 --output ./music
```

Forward to the engine:

```bash
playlist-md core status
playlist-md core list-playlists
playlist-md core export --all --output ~/Music/AppleMusic
```

Core commands, JSON, and progress events: [docs/core-protocol.md](docs/core-protocol.md).

| Key | Action |
|-----|--------|
| ↑↓ / j k | move |
| ←→ / h l | page / change value |
| tab | next section |
| space | toggle playlist |
| enter | inspect / confirm |
| / | search |
| a / n c | select all / clear |
| o | open output folder |
| r | reload |
| s | settings |
| Ctrl+L | repair display |
| ? | keybindings |
| q | quit |

## Configuration

`~/.config/playlist-md/config.json`, or `$XDG_CONFIG_HOME/playlist-md/config.json`.

| Key | Default | Notes |
|-----|---------|--------|
| `output_dir` | `~/AppleMusicExports` | Enter in Settings to edit |
| `playlists_per_page` | `12` | `8` / `12` / `16` / `24` (←→) |

## Authorization

First run prompts via the embedded Swift process. If denied: **System Settings → Privacy & Security → Media & Apple Music**.

## Output

```text
<output-directory>/
├── index.md
├── .playlist-md-manifest.json
└── playlists/
    └── <slug>.md
```

Re-exports delete only stale paths listed in the previous manifest.

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
- Library playlists only
- Unavailable tracks may be skipped or have partial metadata
- Large playlists are fetched in pages

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE) © 2026 imflawlezz
