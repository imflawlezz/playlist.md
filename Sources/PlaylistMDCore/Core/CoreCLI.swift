import Foundation

public enum CoreCLI {
    public static func run(arguments: [String]) async -> Int32 {
        guard arguments.count > 1 else {
            JSONOutput.writeError("Usage: playlist-md-core <command>")
            return 1
        }

        let services = ServiceContainer()

        switch arguments[1] {
        case "version":
            JSONOutput.write(VersionResponse(name: AppVersion.coreName, version: AppVersion.version))
            return 0

        case "status":
            JSONOutput.write(StatusResponse(status: services.playlistService.authorizationStatus().rawValue))
            return 0

        case "authorize":
            let status = await services.playlistService.requestAuthorization()
            JSONOutput.write(StatusResponse(status: status.rawValue))
            return status == .authorized ? 0 : 1

        case "list-playlists":
            do {
                let playlists = try await services.playlistService.fetchPlaylistSummaries()
                JSONOutput.write(PlaylistsResponse(playlists: playlists.map {
                    PlaylistDTO(id: $0.id, name: $0.name)
                }))
                return 0
            } catch {
                return fail(error)
            }

        case "get-playlist":
            return await runGetPlaylist(arguments: Array(arguments.dropFirst(2)), services: services)

        case "index-tracks":
            return await runIndexTracks(services: services)

        case "export":
            return await runExport(arguments: Array(arguments.dropFirst(2)), services: services)

        default:
            JSONOutput.writeError("Unknown command: \(arguments[1])")
            return 1
        }
    }

    private static func runGetPlaylist(arguments: [String], services: ServiceContainer) async -> Int32 {
        var id: String?
        var index = 0
        while index < arguments.count {
            let arg = arguments[index]
            switch arg {
            case "--id":
                guard index + 1 < arguments.count else {
                    JSONOutput.writeError("--id requires a playlist id")
                    return 1
                }
                id = arguments[index + 1]
                index += 2
            default:
                JSONOutput.writeError("Unknown get-playlist flag: \(arg)")
                return 1
            }
        }

        guard let playlistID = id, !playlistID.isEmpty else {
            JSONOutput.writeError("--id is required")
            return 1
        }

        do {
            let summaries = try await services.playlistService.fetchPlaylistSummaries()
            guard let summary = summaries.first(where: { $0.id == playlistID }) else {
                JSONOutput.writeError("Playlist not found")
                return 1
            }
            let playlist = try await services.playlistService.fetchPlaylist(summary)
            JSONOutput.write(detailDTO(playlist))
            return 0
        } catch {
            return fail(error)
        }
    }

    private static func runIndexTracks(services: ServiceContainer) async -> Int32 {
        do {
            let summaries = try await services.playlistService.fetchPlaylistSummaries()
            for summary in summaries {
                let playlist = try await services.playlistService.fetchPlaylist(summary)
                JSONOutput.write(detailDTO(playlist))
            }
            return 0
        } catch {
            return fail(error)
        }
    }

    private static func detailDTO(_ playlist: Playlist) -> PlaylistDetailDTO {
        PlaylistDetailDTO(
            id: playlist.id,
            name: playlist.name,
            tracks: playlist.tracks.map {
                TrackDTO(
                    title: $0.title,
                    artist: $0.artist,
                    album: $0.album,
                    year: $0.year,
                    position: $0.position
                )
            }
        )
    }

    private static func runExport(arguments: [String], services: ServiceContainer) async -> Int32 {
        var output: String?
        var ids: [String] = []
        var exportAll = false
        var exportLibraryOnly = false
        var writeLogs = false

        var index = 0
        while index < arguments.count {
            let arg = arguments[index]
            switch arg {
            case "--output":
                guard index + 1 < arguments.count else {
                    JSONOutput.writeError("--output requires a path")
                    return 1
                }
                output = arguments[index + 1]
                index += 2
            case "--ids":
                guard index + 1 < arguments.count else {
                    JSONOutput.writeError("--ids requires a comma-separated list")
                    return 1
                }
                ids = arguments[index + 1]
                    .split(separator: ",")
                    .map { String($0.trimmingCharacters(in: .whitespaces)) }
                    .filter { !$0.isEmpty }
                index += 2
            case "--all":
                exportAll = true
                index += 1
            case "--library":
                exportLibraryOnly = true
                index += 1
            case "--write-logs":
                writeLogs = true
                index += 1
            case "--no-write-logs":
                writeLogs = false
                index += 1
            default:
                JSONOutput.writeError("Unknown export flag: \(arg)")
                return 1
            }
        }

        guard let outputPath = output else {
            JSONOutput.writeError("--output is required")
            return 1
        }

        if exportLibraryOnly && (exportAll || !ids.isEmpty) {
            JSONOutput.writeError("--library cannot be combined with --all or --ids")
            return 1
        }

        do {
            let outputDirectory = try services.fileSystem.resolveOutputDirectory(outputPath)

            let summary: ExportSummary
            if exportLibraryOnly {
                summary = try await services.exportService.exportLibrary(
                    to: outputDirectory,
                    writeLogs: writeLogs
                ) { progress in
                    emitProgress(progress)
                }
            } else {
                let summaries = try await services.playlistService.fetchPlaylistSummaries()

                let selected: [PlaylistSummary]
                if exportAll {
                    selected = summaries
                } else if !ids.isEmpty {
                    let idSet = Set(ids)
                    selected = summaries.filter { idSet.contains($0.id) }
                } else {
                    JSONOutput.writeError("Specify --all, --ids, or --library")
                    return 1
                }

                guard !selected.isEmpty else {
                    JSONOutput.writeError("No playlists matched the export request")
                    return 1
                }

                summary = try await services.exportService.exportPlaylists(
                    summaries: selected,
                    to: outputDirectory,
                    writeLogs: writeLogs
                ) { progress in
                    emitProgress(progress)
                }
            }

            JSONOutput.write(ExportResultResponse(
                exportedPlaylists: summary.exportedPlaylistCount,
                exportedTracks: summary.exportedTrackCount,
                exportedLibraryTracks: summary.exportedLibraryTrackCount,
                output: summary.outputDirectory.path,
                removedStaleFiles: summary.removedStaleFiles,
                logPath: summary.logPath
            ))
            return 0
        } catch {
            return fail(error)
        }
    }

    private static func emitProgress(_ progress: ExportProgress) {
        let event: ProgressEvent
        switch progress.phase {
        case .fetchingPlaylist(let name, let index, let total):
            event = .fetching(name: name, index: index, total: total)
        case .fetchingLibrary:
            event = .library()
        case .writingFiles:
            event = .writing()
        case .cleaningStaleFiles:
            event = .cleaning()
        }
        JSONOutput.write(event, to: .standardError)
    }

    private static func fail(_ error: Error) -> Int32 {
        if let appError = error as? AppError {
            JSONOutput.writeError(appError.localizedDescription)
        } else {
            JSONOutput.writeError(error.localizedDescription)
        }
        return 1
    }
}

private struct ServiceContainer {
    let playlistService: PlaylistService
    let fileSystem: FileSystemService
    let exportService: ExportService

    init() {
        let client = MusicKitAppleMusicClient()
        playlistService = PlaylistService(client: client)
        fileSystem = FileSystemService()
        exportService = ExportService(
            playlistService: playlistService,
            fileSystem: fileSystem,
            markdownExporter: MarkdownExporter()
        )
    }
}
