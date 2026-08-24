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
| `go/internal/ui` | Bubble Tea screens, selection, search, settings |
| `go/internal/engine` | Locates/extracts the core binary and speaks its JSON protocol |
| `Sources/PlaylistMDCore` | MusicKit access, export pipeline, filename/manifest rules |
| `Sources/playlist_md_core` | Thin `@main` that runs `CoreCLI` |

Go never talks to MusicKit. Swift never draws the TUI. The boundary is process spawn plus line-oriented JSON (see [core-protocol.md](core-protocol.md)).

## Packaging

1. `make` builds `playlist-md-core` for arm64 and x86_64, then `lipo`s them into `go/assets/playlist-md-core`.
2. Go embeds that file (`//go:embed`) and builds a universal `dist/playlist-md` the same way.
3. At runtime, the launcher writes the embedded core to a content-addressed path under the user cache (`~/Library/Caches/playlist-md/<hash>/playlist-md-core` on macOS) unless `PLAYLIST_MD_CORE` points at a local binary.

The core embeds `Sources/PlaylistMDCore/Info.plist` via linker `__info_plist` so MusicKit can associate the process with a bundle ID and usage description.

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
- `playlists_per_page` (`8` / `12` / `16` / `24`)
