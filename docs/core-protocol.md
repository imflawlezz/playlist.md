# Core protocol

`playlist-md-core` is a JSON CLI. Successful payloads go to stdout; errors and progress go to stderr. One JSON object per line.

Exit code `0` means success. Non-zero means failure (stderr usually carries `{"error":"..."}`).

## Commands

| Command | Purpose |
|---------|---------|
| `version` | Core name and version |
| `status` | Current Music authorization |
| `authorize` | Request Music authorization (exit `1` if not authorized) |
| `list-playlists` | Library playlist id/name list |
| `get-playlist --id <id>` | One playlist with tracks |
| `index-tracks` | Stream one playlist detail JSON line per playlist |
| `export --output <path> (--all \| --ids a,b,… \| --library) [--write-logs \| --no-write-logs]` | Playlist Markdown **or** `library.md` (`--library` exclusive with `--all`/`--ids`); + manifest (+ optional `export.log`) |

Forwarded from the launcher as:

```bash
playlist-md core <command> …
```

Development override:

```bash
export PLAYLIST_MD_CORE="$PWD/.build/release/playlist-md-core"
```

## Authorization status values

`authorized` · `denied` · `restricted` · `not_determined`

## Representative responses

The `version` field matches `AppVersion.version` in the Swift core.

```json
{"name":"playlist-md-core","version":"<semver>"}
```

```json
{"status":"authorized"}
```

```json
{"playlists":[{"id":"…","name":"Chill"}]}
```

```json
{
  "id": "…",
  "name": "Chill",
  "tracks": [
    {"title":"…","artist":"…","album":"…","year":2024,"position":1}
  ]
}
```

Playlist export:

```json
{
  "exported_playlists": 1,
  "exported_tracks": 12,
  "exported_library_tracks": 0,
  "output": "/Users/…/AppleMusicExports",
  "removed_stale_files": [],
  "log_path": "export.log"
}
```

Library export (`--library`):

```json
{
  "exported_playlists": 0,
  "exported_tracks": 0,
  "exported_library_tracks": 240,
  "output": "/Users/…/AppleMusicExports",
  "removed_stale_files": [],
  "log_path": "export.log"
}
```

`log_path` is omitted when logging is disabled.

`--library` is mutually exclusive with `--all` / `--ids`. Playlist export leaves `exported_library_tracks` at `0` (and preserves an existing `library.md` if present). Library export leaves playlist counts at `0` and sets `exported_library_tracks` to the rows written to `library.md`.

`exported_tracks` sums tracks across the exported playlists.

If neither `--write-logs` nor `--no-write-logs` is passed to `playlist-md-core`, logging is off. The Go `playlist-md export` CLI defaults to on and always passes one of those flags.

## Progress (stderr)

Export emits progress events on stderr:

```json
{"type":"progress","phase":"fetching","name":"Chill","index":1,"total":3}
{"type":"progress","phase":"library"}
{"type":"progress","phase":"writing"}
{"type":"progress","phase":"cleaning"}
```

Playlist export uses `fetching` / `writing` / `cleaning`. Library export uses `library` / `writing` / `cleaning`. The Go client parses these for the TUI; the final export result remains on stdout.
