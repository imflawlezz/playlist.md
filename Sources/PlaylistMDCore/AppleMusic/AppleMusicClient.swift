import Foundation

enum AuthorizationStatus: Equatable, Sendable {
    case notDetermined
    case denied
    case restricted
    case authorized
}

protocol AppleMusicClient: Sendable {
    func authorizationStatus() -> AuthorizationStatus
    func requestAuthorization() async -> AuthorizationStatus
    func fetchPlaylistSummaries() async throws -> [PlaylistSummary]
    func fetchTracks(for playlistID: String) async throws -> [Track]
}
