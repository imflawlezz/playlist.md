import Foundation

struct MarkdownExporter: Sendable {
    func renderIndex(playlists: [ExportedPlaylist], includeLibrary: Bool) -> String {
        var lines: [String] = [
            "# Apple Music Playlists",
            "",
        ]

        if includeLibrary {
            lines.append("- [Library](\(ExportManifest.libraryRelativePath))")
            lines.append("")
        }

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
            lines.append(trackRow(track))
        }

        lines.append("")
        return lines.joined(separator: "\n")
    }

    func renderLibrary(_ tracks: [Track]) -> String {
        var lines: [String] = [
            "# Apple Music Library",
            "",
            "\(tracks.count) songs",
            "",
            "| # | Track | Artist | Album | Year |",
            "|---:|---|---|---|---:|",
        ]

        for track in tracks {
            lines.append(trackRow(track))
        }

        lines.append("")
        return lines.joined(separator: "\n")
    }

    private func trackRow(_ track: Track) -> String {
        let trackCell = MarkdownEscaping.formatTrackCell(title: track.title, url: track.url)
        let artist = MarkdownEscaping.escapeTableCell(track.artist)
        let album = MarkdownEscaping.escapeTableCell(track.album)
        let year = track.year.map(String.init) ?? ""
        return "| \(track.position) | \(trackCell) | \(artist) | \(album) | \(year) |"
    }
}
