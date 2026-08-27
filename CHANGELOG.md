# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/imflawlezz/playlist-md/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/imflawlezz/playlist-md/releases/tag/v1.0.0
