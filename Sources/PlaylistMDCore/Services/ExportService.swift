import Foundation

struct ExportOptions: Sendable {
    let outputDirectory: URL
    let playlists: [PlaylistSummary]
    let exportAll: Bool
}

struct ExportProgress: Sendable {
    enum Phase: Sendable {
        case fetchingPlaylist(name: String, index: Int, total: Int)
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

    func export(
        summaries: [PlaylistSummary],
        to outputDirectory: URL,
        onProgress: (@Sendable (ExportProgress) -> Void)? = nil
    ) async throws -> ExportSummary {
        try fileSystem.ensureDirectoryExists(at: outputDirectory)

        let playlistsDirectory = outputDirectory.appendingPathComponent("playlists", isDirectory: true)
        try fileSystem.ensureDirectoryExists(at: playlistsDirectory)

        let filenameMap = FilenameSanitizer.assignFilenames(for: summaries)
        var exported: [ExportedPlaylist] = []
        var totalTracks = 0

        for (index, summary) in summaries.enumerated() {
            onProgress?(.init(phase: .fetchingPlaylist(name: summary.name, index: index + 1, total: summaries.count)))

            let playlist = try await playlistService.fetchPlaylist(summary)
            totalTracks += playlist.tracks.count

            guard let filename = filenameMap[summary.id],
                  FilenameSanitizer.isSafeRelativeFilename(filename) else {
                throw AppError.fileSystemError("Unsafe filename generated for playlist \"\(summary.name)\".")
            }

            let relativePath = "playlists/\(filename)"
            exported.append(ExportedPlaylist(playlist: playlist, relativePath: relativePath))
        }

        onProgress?(.init(phase: .writingFiles))

        let sortedExported = exported.sorted {
            $0.playlist.name.localizedCaseInsensitiveCompare($1.playlist.name) == .orderedAscending
        }

        for item in sortedExported {
            let markdown = markdownExporter.renderPlaylist(item.playlist)
            let fileURL = outputDirectory.appendingPathComponent(item.relativePath)
            try fileSystem.writeUTF8(markdown, to: fileURL)
        }

        let indexMarkdown = markdownExporter.renderIndex(playlists: sortedExported)
        try fileSystem.writeUTF8(indexMarkdown, to: outputDirectory.appendingPathComponent("index.md"))

        onProgress?(.init(phase: .cleaningStaleFiles))

        let previousManifest = fileSystem.readManifest(at: outputDirectory)
        let currentPaths = Set(sortedExported.map(\.relativePath))
        let removed = try fileSystem.removeStaleFiles(
            previousManifest: previousManifest,
            currentRelativePaths: currentPaths,
            in: outputDirectory
        )

        let manifest = ExportManifest(
            version: ExportManifest.currentVersion,
            generatedAt: Date(),
            playlists: sortedExported.map { item in
                ExportManifest.Entry(
                    id: item.playlist.id,
                    name: item.playlist.name,
                    relativePath: item.relativePath
                )
            }
        )
        try fileSystem.writeManifest(manifest, to: outputDirectory)

        return ExportSummary(
            exportedPlaylistCount: sortedExported.count,
            exportedTrackCount: totalTracks,
            outputDirectory: outputDirectory,
            removedStaleFiles: removed
        )
    }
}
