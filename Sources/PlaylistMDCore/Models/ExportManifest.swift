import Foundation

struct ExportManifest: Codable, Equatable {
    static let filename = ".playlist-md-manifest.json"
    static let currentVersion = 1

    let version: Int
    let generatedAt: Date
    let playlists: [Entry]

    struct Entry: Codable, Equatable {
        let id: String
        let name: String
        let relativePath: String
    }
}

struct ExportSummary: Equatable, Sendable {
    let exportedPlaylistCount: Int
    let exportedTrackCount: Int
    let outputDirectory: URL
    let removedStaleFiles: [String]
}

struct ExportedPlaylist: Equatable, Sendable {
    let playlist: Playlist
    let relativePath: String
}
