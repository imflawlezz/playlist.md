# Go TUI with embedded Swift core

Status: Accepted  
Date: 2026-08-24

## Context

Exporting Apple Music library playlists requires MusicKit on macOS. A usable tool also needs a terminal UI for selection, search, and progress. MusicKit is Swift-only; a polished TUI is more practical in Go (Bubble Tea) than in a home-grown Swift terminal stack.

Shipping two separate binaries would complicate install and version skew. Putting MusicKit behind a network service would add latency and attack surface for a local personal tool.

## Decision

Keep MusicKit in a Swift executable (`playlist-md-core`) that exposes a small JSON CLI. Build a Go launcher (`playlist-md`) that embeds that core, extracts it to a content-addressed cache at runtime, and drives the TUI (or a thin `export` / `core` CLI) over stdin/stdout/stderr.

## Consequences

- Single user-facing binary; architecture and auth stay on the platform API that requires them.
- Universal builds need dual-arch Swift and Go artifacts plus `lipo`.
- The JSON protocol is a stable contract; either side can change internals without forcing a rewrite of the other.
- Auth prompts and Media & Apple Music permissions attach to the extracted core process and its embedded Info.plist, not to Go itself.
