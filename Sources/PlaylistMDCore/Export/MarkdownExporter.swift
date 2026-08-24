import Foundation

struct MarkdownExporter: Sendable {
    func renderIndex(playlists: [ExportedPlaylist]) -> String {
        var lines: [String] = [
            "# Apple Music Playlists",
            "",
        ]

        let sorted = playlists.sorted {
            $0.playlist.name.localizedCaseInsensitiveCompare($1.playlist.name) == .orderedAscending
        }

        for exported in sorted {
            let label = MarkdownEscaping.escapeLinkLabel(exported.playlist.name)
            lines.append("- [\(label)](\(exported.relativePath))")
        }

        lines.append("")
        return lines.joined(separator: "\n")
    }

    func renderPlaylist(_ playlist: Playlist) -> String {
        var lines: [String] = [
            "# \(MarkdownEscaping.escapeTableCell(playlist.name))",
            "",
            "\(playlist.tracks.count) tracks",
            "",
            "| # | Track | Artist | Album | Year |",
            "|---:|---|---|---|---:|",
        ]

        for track in playlist.tracks {
            let trackCell = MarkdownEscaping.formatTrackCell(title: track.title, url: track.url)
            let artist = MarkdownEscaping.escapeTableCell(track.artist)
            let album = MarkdownEscaping.escapeTableCell(track.album)
            let year = track.year.map(String.init) ?? ""

            lines.append("| \(track.position) | \(trackCell) | \(artist) | \(album) | \(year) |")
        }

        lines.append("")
        return lines.joined(separator: "\n")
    }
}
