import Foundation

struct ExportProgress: Sendable {
    enum Phase: Sendable {
        case fetchingPlaylist(name: String, index: Int, total: Int)
        case fetchingLibrary
        case writingFiles
        case cleaningStaleFiles
    }

    let phase: Phase
}

struct ExportService: Sendable {
    let playlistService: PlaylistService
    let fileSystem: FileSystemService
    let markdownExporter: MarkdownExporter

    init(
        playlistService: PlaylistService,
        fileSystem: FileSystemService,
        markdownExporter: MarkdownExporter
    ) {
        self.playlistService = playlistService
        self.fileSystem = fileSystem
        self.markdownExporter = markdownExporter
    }

    func exportPlaylists(
        summaries: [PlaylistSummary],
        to outputDirectory: URL,
        writeLogs: Bool = false,
        onProgress: (@Sendable (ExportProgress) -> Void)? = nil
    ) async throws -> ExportSummary {
        let log = ExportLogger()
        await log.info("Playlist export started (\(summaries.count) playlist\(summaries.count == 1 ? "" : "s"))")

        do {
            try fileSystem.ensureDirectoryExists(at: outputDirectory)

            let playlistsDirectory = outputDirectory.appendingPathComponent("playlists", isDirectory: true)
            try fileSystem.ensureDirectoryExists(at: playlistsDirectory)

            let filenameMap = FilenameSanitizer.assignFilenames(for: summaries)
            var exported: [ExportedPlaylist] = []
            var totalTracks = 0

            for (index, summary) in summaries.enumerated() {
                onProgress?(.init(phase: .fetchingPlaylist(name: summary.name, index: index + 1, total: summaries.count)))
                await log.info("Fetching playlist \(index + 1)/\(summaries.count): \(summary.name)")

                let playlist = try await playlistService.fetchPlaylist(summary)
                totalTracks += playlist.tracks.count

                if playlist.tracks.isEmpty {
                    await log.warning("Playlist \"\(summary.name)\" has no tracks")
                }
                for track in playlist.tracks where track.title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                    await log.warning("Playlist \"\(summary.name)\" track #\(track.position) has an empty title")
                }

                guard let filename = filenameMap[summary.id],
                      FilenameSanitizer.isSafeRelativeFilename(filename) else {
                    throw AppError.fileSystemError("Unsafe filename generated for playlist \"\(summary.name)\".")
                }

                let baseSlug = FilenameSanitizer.slug(from: summary.name)
                if filename != "\(baseSlug).md" {
                    await log.warning("Duplicate playlist slug; wrote playlists/\(filename) for \"\(summary.name)\"")
                }

                let relativePath = "playlists/\(filename)"
                exported.append(ExportedPlaylist(playlist: playlist, relativePath: relativePath))
            }

            onProgress?(.init(phase: .writingFiles))
            await log.info("Writing Markdown files")

            let sortedExported = exported.sorted {
                $0.playlist.name.localizedCaseInsensitiveCompare($1.playlist.name) == .orderedAscending
            }

            for item in sortedExported {
                let markdown = markdownExporter.renderPlaylist(item.playlist)
                let fileURL = outputDirectory.appendingPathComponent(item.relativePath)
                try fileSystem.writeUTF8(markdown, to: fileURL)
            }

            let libraryExists = fileSystem.fileExists(
                ExportManifest.libraryRelativePath,
                in: outputDirectory
            )
            let indexMarkdown = markdownExporter.renderIndex(
                playlists: sortedExported,
                includeLibrary: libraryExists
            )
            try fileSystem.writeUTF8(indexMarkdown, to: outputDirectory.appendingPathComponent("index.md"))

            onProgress?(.init(phase: .cleaningStaleFiles))
            await log.info("Cleaning stale exports")

            let previousManifest = fileSystem.readManifest(at: outputDirectory)
            let managedFiles = preservedManagedFiles(from: previousManifest, in: outputDirectory)
            let currentPaths = Set(sortedExported.map(\.relativePath) + managedFiles)
            let removed = try fileSystem.removeStaleFiles(
                previousManifest: previousManifest,
                currentRelativePaths: currentPaths,
                in: outputDirectory
            )
            if !removed.isEmpty {
                await log.info("Removed \(removed.count) stale file\(removed.count == 1 ? "" : "s")")
            }

            let manifest = ExportManifest(
                version: ExportManifest.currentVersion,
                generatedAt: Date(),
                playlists: sortedExported.map { item in
                    ExportManifest.Entry(
                        id: item.playlist.id,
                        name: item.playlist.name,
                        relativePath: item.relativePath
                    )
                },
                files: managedFiles
            )
            try fileSystem.writeManifest(manifest, to: outputDirectory)

            await log.info("Playlist export complete: \(sortedExported.count) playlists, \(totalTracks) tracks")

            let logPath = try await writeLogIfNeeded(log, enabled: writeLogs, to: outputDirectory)

            return ExportSummary(
                exportedPlaylistCount: sortedExported.count,
                exportedTrackCount: totalTracks,
                exportedLibraryTrackCount: 0,
                outputDirectory: outputDirectory,
                removedStaleFiles: removed,
                logPath: logPath
            )
        } catch {
            await log.error(errorMessage(error))
            _ = try? await writeLogIfNeeded(log, enabled: writeLogs, to: outputDirectory)
            throw error
        }
    }

    func exportLibrary(
        to outputDirectory: URL,
        writeLogs: Bool = false,
        onProgress: (@Sendable (ExportProgress) -> Void)? = nil
    ) async throws -> ExportSummary {
        let log = ExportLogger()
        await log.info("Library export started")

        do {
            try fileSystem.ensureDirectoryExists(at: outputDirectory)

            onProgress?(.init(phase: .fetchingLibrary))
            await log.info("Fetching library songs")
            let libraryTracks = try await playlistService.fetchLibrarySongs()
            if libraryTracks.isEmpty {
                await log.warning("Library song list is empty")
            } else {
                await log.info("Fetched \(libraryTracks.count) library song\(libraryTracks.count == 1 ? "" : "s")")
            }
            for track in libraryTracks where track.title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                await log.warning("Library song #\(track.position) has an empty title")
            }

            onProgress?(.init(phase: .writingFiles))
            await log.info("Writing library.md")

            let libraryMarkdown = markdownExporter.renderLibrary(libraryTracks)
            try fileSystem.writeUTF8(
                libraryMarkdown,
                to: outputDirectory.appendingPathComponent(ExportManifest.libraryRelativePath)
            )

            let previousManifest = fileSystem.readManifest(at: outputDirectory)
            let indexPlaylists = indexPlaylists(from: previousManifest)
            let indexMarkdown = markdownExporter.renderIndex(playlists: indexPlaylists, includeLibrary: true)
            try fileSystem.writeUTF8(indexMarkdown, to: outputDirectory.appendingPathComponent("index.md"))

            onProgress?(.init(phase: .cleaningStaleFiles))

            let managedFiles = [ExportManifest.libraryRelativePath]
            let playlistPaths = Set((previousManifest?.playlists ?? []).map(\.relativePath))
            let currentPaths = playlistPaths.union(managedFiles)
            let removed = try fileSystem.removeStaleFiles(
                previousManifest: previousManifest,
                currentRelativePaths: currentPaths,
                in: outputDirectory
            )

            let manifest = ExportManifest(
                version: ExportManifest.currentVersion,
                generatedAt: Date(),
                playlists: previousManifest?.playlists ?? [],
                files: managedFiles
            )
            try fileSystem.writeManifest(manifest, to: outputDirectory)

            await log.info("Library export complete: \(libraryTracks.count) songs")

            let logPath = try await writeLogIfNeeded(log, enabled: writeLogs, to: outputDirectory)

            return ExportSummary(
                exportedPlaylistCount: 0,
                exportedTrackCount: 0,
                exportedLibraryTrackCount: libraryTracks.count,
                outputDirectory: outputDirectory,
                removedStaleFiles: removed,
                logPath: logPath
            )
        } catch {
            await log.error(errorMessage(error))
            _ = try? await writeLogIfNeeded(log, enabled: writeLogs, to: outputDirectory)
            throw error
        }
    }

    private func preservedManagedFiles(
        from previous: ExportManifest?,
        in outputDirectory: URL
    ) -> [String] {
        guard let previous else { return [] }
        return previous.files.filter { fileSystem.fileExists($0, in: outputDirectory) }
    }

    private func indexPlaylists(from previous: ExportManifest?) -> [ExportedPlaylist] {
        (previous?.playlists ?? []).map { entry in
            ExportedPlaylist(
                playlist: Playlist(id: entry.id, name: entry.name, tracks: []),
                relativePath: entry.relativePath
            )
        }
    }

    private func errorMessage(_ error: Error) -> String {
        if let appError = error as? AppError {
            return appError.localizedDescription
        }
        return error.localizedDescription
    }

    private func writeLogIfNeeded(
        _ log: ExportLogger,
        enabled: Bool,
        to outputDirectory: URL
    ) async throws -> String? {
        guard enabled else { return nil }
        let text = await log.rendered()
        guard !text.isEmpty else { return nil }
        try fileSystem.ensureDirectoryExists(at: outputDirectory)
        try fileSystem.writeUTF8(
            text,
            to: outputDirectory.appendingPathComponent(ExportManifest.exportLogRelativePath)
        )
        return ExportManifest.exportLogRelativePath
    }
}
