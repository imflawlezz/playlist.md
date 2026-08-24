import Foundation
import MusicKit

struct MusicKitAppleMusicClient: AppleMusicClient {
    init() {}
    func authorizationStatus() -> AuthorizationStatus {
        mapStatus(MusicAuthorization.currentStatus)
    }

    func requestAuthorization() async -> AuthorizationStatus {
        let status = await MusicAuthorization.request()
        return mapStatus(status)
    }

    func fetchPlaylistSummaries() async throws -> [PlaylistSummary] {
        let request = MusicLibraryRequest<MusicKit.Playlist>()
        let response: MusicLibraryResponse<MusicKit.Playlist>

        do {
            response = try await request.response()
        } catch {
            throw mapError(error, context: "your playlists")
        }

        return response.items.map { playlist in
            PlaylistSummary(id: playlist.id.rawValue, name: playlist.name)
        }
        .sorted {
            if $0.name != $1.name {
                return $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending
            }
            return $0.id < $1.id
        }
    }

    func fetchTracks(for playlistID: String) async throws -> [Track] {
        let itemID = MusicItemID(playlistID)

        let request = MusicLibraryRequest<MusicKit.Playlist>()

        let response: MusicLibraryResponse<MusicKit.Playlist>
        do {
            response = try await request.response()
        } catch {
            throw mapError(error, context: "playlist tracks")
        }

        guard let playlist = response.items.first(where: { $0.id == itemID }) else {
            return []
        }

        let detailedPlaylist: MusicKit.Playlist
        do {
            detailedPlaylist = try await playlist.with([.tracks])
        } catch {
            throw mapError(error, context: "playlist \"\(playlist.name)\"")
        }

        guard var trackCollection = detailedPlaylist.tracks else {
            return []
        }

        var musicTracks: [MusicKit.Track] = []
        musicTracks.append(contentsOf: trackCollection)

        while trackCollection.hasNextBatch {
            do {
                guard let nextBatch = try await trackCollection.nextBatch(limit: 300) else {
                    break
                }
                musicTracks.append(contentsOf: nextBatch)
                trackCollection = nextBatch
            } catch {
                throw mapError(error, context: "playlist \"\(playlist.name)\"")
            }
        }

        return musicTracks.enumerated().map { index, musicTrack in
            normalizeTrack(musicTrack, position: index + 1)
        }
    }

    private func normalizeTrack(_ musicTrack: MusicKit.Track, position: Int) -> Track {
        Track(
            title: musicTrack.title,
            artist: musicTrack.artistName,
            album: musicTrack.albumTitle ?? "",
            year: musicTrack.releaseDate.map { Calendar.current.component(.year, from: $0) },
            url: musicTrack.url,
            position: position
        )
    }

    private func mapStatus(_ status: MusicAuthorization.Status) -> AuthorizationStatus {
        switch status {
        case .authorized:
            return .authorized
        case .denied:
            return .denied
        case .restricted:
            return .restricted
        case .notDetermined:
            return .notDetermined
        @unknown default:
            return .notDetermined
        }
    }

    private func mapError(_ error: Error, context: String) -> AppError {
        if let urlError = error as? URLError {
            return .networkFailure(context + " (\(urlError.localizedDescription))")
        }
        if context.contains("playlist \"") {
            let name = context
                .replacingOccurrences(of: "playlist \"", with: "")
                .replacingOccurrences(of: "\"", with: "")
            return .playlistFetchFailed(name)
        }
        return .libraryUnavailable
    }
}
