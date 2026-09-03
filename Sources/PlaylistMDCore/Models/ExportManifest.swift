import Foundation

struct ExportManifest: Codable, Equatable {
    static let filename = ".playlist-md-manifest.json"
    static let currentVersion = 2
    static let libraryRelativePath = "library.md"
    static let exportLogRelativePath = "export.log"

    let version: Int
    let generatedAt: Date
    let playlists: [Entry]
    /// Extra root files managed by export (e.g. `library.md`).
    let files: [String]

    struct Entry: Codable, Equatable {
        let id: String
        let name: String
        let relativePath: String
    }

    init(version: Int, generatedAt: Date, playlists: [Entry], files: [String] = []) {
        self.version = version
        self.generatedAt = generatedAt
        self.playlists = playlists
        self.files = files
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        version = try container.decode(Int.self, forKey: .version)
        generatedAt = try container.decode(Date.self, forKey: .generatedAt)
        playlists = try container.decode([Entry].self, forKey: .playlists)
        files = try container.decodeIfPresent([String].self, forKey: .files) ?? []
    }

    var managedRelativePaths: [String] {
        playlists.map(\.relativePath) + files
    }
}

struct ExportSummary: Equatable, Sendable {
    let exportedPlaylistCount: Int
    let exportedTrackCount: Int
    let exportedLibraryTrackCount: Int
    let outputDirectory: URL
    let removedStaleFiles: [String]
    let logPath: String?
}

struct ExportedPlaylist: Equatable, Sendable {
    let playlist: Playlist
    let relativePath: String
}
