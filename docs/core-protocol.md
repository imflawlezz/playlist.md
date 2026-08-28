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
| `export --output <path> (--all \| --ids a,b,…)` | Write Markdown + manifest |

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

```json
{
  "exported_playlists": 1,
  "exported_tracks": 12,
  "output": "/Users/…/AppleMusicExports",
  "removed_stale_files": []
}
```

## Progress (stderr)

Export emits progress events on stderr:

```json
{"type":"progress","phase":"fetching","name":"Chill","index":1,"total":3}
{"type":"progress","phase":"writing"}
{"type":"progress","phase":"cleaning"}
```

The Go client parses these for the TUI; the final export result remains on stdout.
