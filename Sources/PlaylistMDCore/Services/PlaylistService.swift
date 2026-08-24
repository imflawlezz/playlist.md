import Foundation

struct PlaylistService: Sendable {
    let client: AppleMusicClient

    init(client: AppleMusicClient) {
        self.client = client
    }

    func authorizationStatus() -> AuthorizationStatus {
        client.authorizationStatus()
    }

    func requestAuthorization() async -> AuthorizationStatus {
        await client.requestAuthorization()
    }

    func fetchPlaylistSummaries() async throws -> [PlaylistSummary] {
        guard authorizationStatus() == .authorized else {
            throw authorizationError(for: authorizationStatus())
        }
        return try await client.fetchPlaylistSummaries()
    }

    func fetchPlaylist(_ summary: PlaylistSummary) async throws -> Playlist {
        guard authorizationStatus() == .authorized else {
            throw authorizationError(for: authorizationStatus())
        }

        let tracks = try await client.fetchTracks(for: summary.id)
        return Playlist(id: summary.id, name: summary.name, tracks: tracks)
    }

    private func authorizationError(for status: AuthorizationStatus) -> AppError {
        switch status {
        case .denied, .restricted:
            return .authorizationDenied
        case .notDetermined:
            return .authorizationRequired
        case .authorized:
            return .libraryUnavailable
        }
    }
}
