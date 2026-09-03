# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.0] — 2026-09-03

### Added

- Separate **Export music library** action / `export --library` writes **`library.md`**; `index.md` links to it when present
- Optional **`export.log`** (INFO / WARNING / ERROR) under the output folder; Settings **Export log** / `--write-logs` / `--no-write-logs` (default on)
- Export progress shows **Exporting** with a percentage and a checklist of finished playlists (progress bar removed)
- Export result includes `exported_library_tracks` and optional `log_path`; progress phase `library` while fetching songs

### Changed

- Playlist export actions labeled **Export all playlists** / **Export selected playlists** (no longer bundled with the library dump)

### Fixed

- Home cursor resets to the first playlist when the library loads (no longer stuck on a mid-list row)
- Inspect track rows keep the **year** intact and truncate the **album** when space is tight
- Playlist names with trailing or internal dots (e.g. `…`) no longer fail export as unsafe filenames
- Export errors from the Swift core surface their real message instead of a generic `playlist-md-core export failed`

## [1.1.2] — 2026-08-28

### Changed

- GitHub Releases attach `playlist-md-v<semver>.tar.gz` (README, LICENSE, and a `playlist-md` binary inside) instead of a bare versioned binary

## [1.1.1] — 2026-08-28

### Added

- **Refresh library** action on home (and **r**); re-fetches playlists and re-indexes tracks for search

### Changed

- Settings row renamed to **Items per page**; page-size options are now `8` / `12` / `16` / `24` / `32` / `40`
- Search screen drops the page indicator; the match list still scrolls with **↑↓**
- Playlist and search rows keep full title and hint text when they fit; otherwise the longer side truncates more, with match hints right-aligned

### Fixed

- **j** and **k** type into the search field instead of moving the match list
- Search input no longer shows a stray trailing **…** from frame clipping

## [1.1.0] — 2026-08-27

### Added

- Bordered TUI frame with the help bar pinned to the bottom of the window
- Clickable author GitHub link in the title (OSC 8; a click opens the page in supporting terminals)
- Terminal auto-resize to fit content; **Ctrl+L** repairs the display after a manual resize or garbled redraw
- GitHub Actions CI on push/PR and tag (`v*`) releases of the universal macOS binary (SHA-256 checksum)

### Changed

- Settings uses labeled rows and segmented controls for playlists per page; output folder is edited inline (**Enter** on the row)
- Home, search, inspect, and keybindings use numbered lists, **Playlists:** / **Actions** sections, a sparse footer, and grouped **?** help; selected playlists show **✓** on the right
- Home shows Apple Music authorization status again; export / open-folder actions are separated from Settings, Repair TUI, Keybindings, and Quit
- Release binary name includes the version with dashes (`dist/playlist-md-v1-1-0`)

### Fixed

- Inspect pages matching tracks only when a search is active, so paging no longer leaves phantom rows or a split cursor

## [1.0.0] — 2026-08-24

### Added

- Interactive Bubble Tea TUI (`playlist-md`) with playlist selection, inspect, search, settings, and export progress.
- Embedded Swift MusicKit engine (`playlist-md-core`) for Apple Music auth, library fetch, and Markdown export.
- Non-interactive `playlist-md export` and `playlist-md core` passthrough commands.
- Export output layout: `index.md`, per-playlist Markdown under `playlists/`, and `.playlist-md-manifest.json` with stale-file cleanup.
- Block-style export progress bar with percentage; playlist name on the Exporting line; completed bar retained on the export-complete screen.
- Search results as a fixed-height scrolling list (**↑↓** / **j k**); match hints show truncated track title and artist.
- Authorization status with ✓ / ✗ markers and capitalized labels (Authorized, Denied, Restricted, Not authorized).
- Universal (arm64 + x86_64) `make` / `make install` build of `dist/playlist-md`.
- TUI config at `~/.config/playlist-md/config.json` (`output_dir`, `playlists_per_page`).
- `PLAYLIST_MD_CORE` override for a local engine binary during development.

[Unreleased]: https://github.com/imflawlezz/playlist-md/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/imflawlezz/playlist-md/releases/tag/v1.3.0
[1.1.2]: https://github.com/imflawlezz/playlist-md/releases/tag/v1.1.2
[1.1.1]: https://github.com/imflawlezz/playlist-md/releases/tag/v1.1.1
[1.1.0]: https://github.com/imflawlezz/playlist-md/releases/tag/v1.1.0
[1.0.0]: https://github.com/imflawlezz/playlist-md/releases/tag/v1.0.0
