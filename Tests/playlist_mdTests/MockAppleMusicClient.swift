import Foundation
@testable import PlaylistMDCore

final class MockAppleMusicClient: AppleMusicClient, @unchecked Sendable {
    var status: AuthorizationStatus = .authorized
    var playlists: [PlaylistSummary] = []
    var tracksByPlaylistID: [String: [Track]] = [:]
    var librarySongs: [Track] = []
    var shouldThrowOnFetchPlaylists = false
    var shouldThrowOnFetchTracks = false
    var shouldThrowOnFetchLibrary = false

    func authorizationStatus() -> AuthorizationStatus {
        status
    }

    func requestAuthorization() async -> AuthorizationStatus {
        status = .authorized
        return status
    }

    func fetchPlaylistSummaries() async throws -> [PlaylistSummary] {
        if shouldThrowOnFetchPlaylists {
            throw AppError.libraryUnavailable
        }
        return playlists
    }

    func fetchTracks(for playlistID: String) async throws -> [Track] {
        if shouldThrowOnFetchTracks {
            throw AppError.playlistFetchFailed("Test")
        }
        return tracksByPlaylistID[playlistID] ?? []
    }

    func fetchLibrarySongs() async throws -> [Track] {
        if shouldThrowOnFetchLibrary {
            throw AppError.libraryUnavailable
        }
        return librarySongs
    }
}
