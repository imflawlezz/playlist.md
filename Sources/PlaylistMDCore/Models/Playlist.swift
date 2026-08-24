import Foundation

struct PlaylistSummary: Identifiable, Equatable, Sendable {
    let id: String
    let name: String
}

struct Playlist: Identifiable, Equatable, Sendable {
    let id: String
    let name: String
    let tracks: [Track]
}
