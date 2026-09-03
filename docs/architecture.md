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

Home includes **Export all playlists** / **Export selected playlists**, **Export music library**, and **Refresh library** (**r**). Refresh re-fetches playlists and re-indexes tracks for search; library export does not require a prior playlist load.

Export progress shows **Exporting** with a percentage on the right and a checklist of finished playlist names (no progress bar). Library export uses the `library` progress phase instead of the playlist checklist. Inspect track rows keep the **year** and truncate the **album** when space is tight.

## Releases

CI (`.github/workflows/ci.yml`) runs on push to `main`/`master` and on pull requests: Swift tests, Go tests, `go vet`, and a syntax/smoke check of `scripts/release-notes.sh`.

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
├── library.md                 ← after Export music library / export --library (kept on playlist re-export if present)
├── export.log                 ← when write_export_log / --write-logs is on
├── .playlist-md-manifest.json
└── playlists/
    └── <slug>.md
```

Playlist export and library export are separate. Playlist export writes playlist Markdown and keeps an existing `library.md` when present; library export (`export --library` / **Export music library**) writes or refreshes `library.md`. `library.md` lists library songs deduped by MusicKit song id, sorted by artist then title (identical metadata can still appear twice when MusicKit returns distinct ids). `exported_tracks` is the sum of tracks across exported playlists (a song in two playlists counts twice); `exported_library_tracks` is the library dump row count (playlist-only runs report `0`).

When enabled (`write_export_log` / `--write-logs`), `export.log` records INFO/WARNING/ERROR lines for the run (and is still written if the export fails after the output directory is known). Stale cleanup only removes paths listed in the previous manifest (`playlists` entries and `files`, e.g. `library.md`) and absent from the current export. Playlist export preserves managed files that still exist on disk; library export preserves playlist entries from the previous manifest. Other files in the output directory—including `export.log`—are left alone.

## Config

TUI settings live at `$XDG_CONFIG_HOME/playlist-md/config.json`, or `~/.config/playlist-md/config.json` when unset:

- `output_dir` (default `~/AppleMusicExports`)
- `playlists_per_page` (Settings label: **Items per page**; options `8` / `12` / `16` / `24` / `32` / `40`)
- `write_export_log` (Settings label: **Export log**; default on — writes `export.log`)
