import Foundation

enum JSONOutput {
    private static let encoder: JSONEncoder = {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.sortedKeys]
        return encoder
    }()

    static func write<T: Encodable>(_ value: T, to handle: FileHandle = .standardOutput) {
        guard let data = try? encoder.encode(value),
              var line = String(data: data, encoding: .utf8) else {
            return
        }
        line.append("\n")
        if let encoded = line.data(using: .utf8) {
            handle.write(encoded)
        }
    }

    static func writeError(_ message: String) {
        write(ErrorResponse(error: message), to: .standardError)
    }
}

struct ErrorResponse: Encodable {
    let error: String
}

struct StatusResponse: Encodable {
    let status: String
}

struct VersionResponse: Encodable {
    let name: String
    let version: String
}

struct PlaylistsResponse: Encodable {
    let playlists: [PlaylistDTO]
}

struct PlaylistDTO: Encodable {
    let id: String
    let name: String
}

struct PlaylistDetailDTO: Encodable {
    let id: String
    let name: String
    let tracks: [TrackDTO]
}

struct TrackDTO: Encodable {
    let title: String
    let artist: String
    let album: String
    let year: Int?
    let position: Int
}

struct ExportResultResponse: Encodable {
    let exportedPlaylists: Int
    let exportedTracks: Int
    let output: String
    let removedStaleFiles: [String]

    enum CodingKeys: String, CodingKey {
        case exportedPlaylists = "exported_playlists"
        case exportedTracks = "exported_tracks"
        case output
        case removedStaleFiles = "removed_stale_files"
    }
}

struct ProgressEvent: Encodable {
    let type: String
    let phase: String
    let name: String?
    let index: Int?
    let total: Int?

    static func fetching(name: String, index: Int, total: Int) -> ProgressEvent {
        ProgressEvent(type: "progress", phase: "fetching", name: name, index: index, total: total)
    }

    static func writing() -> ProgressEvent {
        ProgressEvent(type: "progress", phase: "writing", name: nil, index: nil, total: nil)
    }

    static func cleaning() -> ProgressEvent {
        ProgressEvent(type: "progress", phase: "cleaning", name: nil, index: nil, total: nil)
    }
}
