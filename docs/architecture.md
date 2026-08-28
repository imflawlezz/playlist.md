# Architecture

`playlist-md` is a macOS tool with two processes packaged as one binary:

```text
playlist-md                 Go + Bubble Tea (UI / CLI wrapper)
  └── playlist-md-core      Swift + MusicKit (auth, fetch, Markdown export)
```

## Responsibilities

| Component | Role |
|-----------|------|
| `go/cmd/playlistmd` | Entry point: interactive TUI, `export` CLI, `core` passthrough |
| `go/internal/ui` | Bubble Tea screens, framed layout, selection, search, settings, terminal sizing |
| `go/internal/engine` | Locates/extracts the core binary and speaks its JSON protocol |
| `Sources/PlaylistMDCore` | MusicKit access, export pipeline, filename/manifest rules |
| `Sources/playlist_md_core` | Thin `@main` that runs `CoreCLI` |

Go never talks to MusicKit. Swift never draws the TUI. The boundary is process spawn plus line-oriented JSON (see [core-protocol.md](core-protocol.md)).

## Packaging

1. `make` builds `playlist-md-core` for arm64 and x86_64, then `lipo`s them into `go/assets/playlist-md-core`.
2. Go embeds that file (`//go:embed`) and builds a universal launcher at `dist/playlist-md-v<version-with-dashes>` (dots in `VERSION` become dashes in the filename).
3. `make release-tar` stages `README.md`, `LICENSE`, and the launcher renamed to `playlist-md` into `dist/playlist-md-v<semver>.tar.gz` (dots preserved in the archive name).
4. At runtime, the launcher writes the embedded core to a content-addressed path under the user cache (`~/Library/Caches/playlist-md/<hash>/playlist-md-core` on macOS) unless `PLAYLIST_MD_CORE` points at a local binary.

The core embeds `Sources/PlaylistMDCore/Info.plist` via linker `__info_plist` so MusicKit can associate the process with a bundle ID and usage description.

`VERSION` in the Makefile controls both the internal build output and the release archive name. CI passes `make VERSION=<semver> release-tar` when building from a tag (see [Releases](#releases)).

## TUI

The launcher runs Bubble Tea in an alternate screen with a bordered frame. Key hints render on the bottom row inside the frame; **?** opens grouped keybindings for the current screen.

The title bar includes an OSC 8 hyperlink on the author name. Terminals that support OSC 8 can open the URL; on macOS a left-click also invokes `open(1)`.

Unless the user has manually resized the window, the TUI grows terminal height to fit the current view (minimum 72×22 columns×rows). **Ctrl+L** clears the screen, forgets manual resize, and re-queries terminal size.

Home includes **Refresh library** (**r**), which re-fetches playlists and re-indexes tracks for search.

## Releases

CI (`.github/workflows/ci.yml`) runs on push and pull requests to `main`: Swift tests, Go tests, `go vet`, and a syntax/smoke check of `scripts/release-notes.sh`.

Pushing a `v*` tag triggers `.github/workflows/release.yml`:

1. Same test steps as CI
2. Release notes from `CHANGELOG.md` via `sh scripts/release-notes.sh <tag>`
3. `make VERSION=<semver> release-tar` (tag without the `v` prefix)
4. SHA-256 checksum of `dist/playlist-md-v<semver>.tar.gz`
5. GitHub Release with the archive and checksum

The `.tar.gz` contains `README.md`, `LICENSE`, and a universal macOS binary named `playlist-md` (no version suffix).

The changelog section for the tagged version must exist before the tag is pushed; the release workflow fails if `release-notes.sh` cannot find it.

## Export layout

```text
<output>/
├── index.md
├── .playlist-md-manifest.json
└── playlists/
    └── <slug>.md
```

Stale cleanup only removes paths listed in the previous manifest and absent from the current export. Other files in the output directory are left alone.

## Config

TUI settings live at `$XDG_CONFIG_HOME/playlist-md/config.json`, or `~/.config/playlist-md/config.json` when unset:

- `output_dir` (default `~/AppleMusicExports`)
- `playlists_per_page` (Settings label: **Items per page**; options `8` / `12` / `16` / `24` / `32` / `40`)
