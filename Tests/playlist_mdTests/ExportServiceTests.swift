import XCTest
@testable import PlaylistMDCore

final class ExportServiceTests: XCTestCase {
    func testExportPlaylistsWritesIndexPlaylistAndManifest() async throws {
        let tempDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString, isDirectory: true)
        defer { try? FileManager.default.removeItem(at: tempDirectory) }

        let mockClient = MockAppleMusicClient()
        mockClient.playlists = [
            PlaylistSummary(id: "1", name: "Chill"),
        ]
        mockClient.tracksByPlaylistID = [
            "1": [
                Track(title: "Song", artist: "Artist", album: "Album", year: 2024, url: nil, position: 1),
            ],
        ]
        mockClient.librarySongs = [
            Track(title: "Song", artist: "Artist", album: "Album", year: 2024, url: nil, position: 1),
            Track(title: "Other", artist: "Band", album: "LP", year: 2023, url: nil, position: 2),
        ]

        let playlistService = PlaylistService(client: mockClient)
        let fileSystem = FileSystemService()
        let exportService = ExportService(
            playlistService: playlistService,
            fileSystem: fileSystem,
            markdownExporter: MarkdownExporter()
        )

        let summary = try await exportService.exportPlaylists(
            summaries: mockClient.playlists,
            to: tempDirectory,
            writeLogs: true
        )

        XCTAssertEqual(summary.exportedPlaylistCount, 1)
        XCTAssertEqual(summary.exportedTrackCount, 1)
        XCTAssertEqual(summary.exportedLibraryTrackCount, 0)
        XCTAssertEqual(summary.logPath, ExportManifest.exportLogRelativePath)

        let indexURL = tempDirectory.appendingPathComponent("index.md")
        let libraryURL = tempDirectory.appendingPathComponent(ExportManifest.libraryRelativePath)
        let logURL = tempDirectory.appendingPathComponent(ExportManifest.exportLogRelativePath)
        let playlistURL = tempDirectory.appendingPathComponent("playlists/chill.md")
        let manifestURL = tempDirectory.appendingPathComponent(ExportManifest.filename)

        XCTAssertTrue(FileManager.default.fileExists(atPath: indexURL.path))
        XCTAssertFalse(FileManager.default.fileExists(atPath: libraryURL.path))
        XCTAssertTrue(FileManager.default.fileExists(atPath: logURL.path))
        XCTAssertTrue(FileManager.default.fileExists(atPath: playlistURL.path))
        XCTAssertTrue(FileManager.default.fileExists(atPath: manifestURL.path))

        let log = try String(contentsOf: logURL, encoding: .utf8)
        XCTAssertTrue(log.contains(" INFO "))
        XCTAssertTrue(log.contains("Playlist export complete"))

        let index = try String(contentsOf: indexURL, encoding: .utf8)
        XCTAssertFalse(index.contains("- [Library](library.md)"))
        XCTAssertTrue(index.contains("- [Chill](playlists/chill.md)"))

        let manifest = try XCTUnwrap(fileSystem.readManifest(at: tempDirectory))
        XCTAssertEqual(manifest.files, [])
    }

    func testExportLibraryWritesLibraryAndIndexLink() async throws {
        let tempDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString, isDirectory: true)
        defer { try? FileManager.default.removeItem(at: tempDirectory) }

        let fileSystem = FileSystemService()
        try fileSystem.ensureDirectoryExists(at: tempDirectory.appendingPathComponent("playlists"))
        try fileSystem.writeManifest(
            ExportManifest(
                version: 1,
                generatedAt: Date(),
                playlists: [
                    .init(id: "1", name: "Chill", relativePath: "playlists/chill.md"),
                ],
                files: []
            ),
            to: tempDirectory
        )

        let mockClient = MockAppleMusicClient()
        mockClient.librarySongs = [
            Track(title: "Song", artist: "Artist", album: "Album", year: 2024, url: nil, position: 1),
            Track(title: "Other", artist: "Band", album: "LP", year: 2023, url: nil, position: 2),
        ]

        let exportService = ExportService(
            playlistService: PlaylistService(client: mockClient),
            fileSystem: fileSystem,
            markdownExporter: MarkdownExporter()
        )

        let summary = try await exportService.exportLibrary(
            to: tempDirectory,
            writeLogs: true
        )

        XCTAssertEqual(summary.exportedPlaylistCount, 0)
        XCTAssertEqual(summary.exportedTrackCount, 0)
        XCTAssertEqual(summary.exportedLibraryTrackCount, 2)
        XCTAssertEqual(summary.logPath, ExportManifest.exportLogRelativePath)

        let libraryURL = tempDirectory.appendingPathComponent(ExportManifest.libraryRelativePath)
        let indexURL = tempDirectory.appendingPathComponent("index.md")
        XCTAssertTrue(FileManager.default.fileExists(atPath: libraryURL.path))

        let library = try String(contentsOf: libraryURL, encoding: .utf8)
        XCTAssertTrue(library.contains("# Apple Music Library"))
        XCTAssertTrue(library.contains("2 songs"))

        let index = try String(contentsOf: indexURL, encoding: .utf8)
        XCTAssertTrue(index.contains("- [Library](library.md)"))
        XCTAssertTrue(index.contains("- [Chill](playlists/chill.md)"))

        let manifest = try XCTUnwrap(fileSystem.readManifest(at: tempDirectory))
        XCTAssertEqual(manifest.files, [ExportManifest.libraryRelativePath])
        XCTAssertEqual(manifest.playlists.map(\.id), ["1"])
    }

    func testExportPlaylistsPreservesExistingLibrary() async throws {
        let tempDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString, isDirectory: true)
        defer { try? FileManager.default.removeItem(at: tempDirectory) }

        let fileSystem = FileSystemService()
        try fileSystem.ensureDirectoryExists(at: tempDirectory.appendingPathComponent("playlists"))
        try "library".write(
            to: tempDirectory.appendingPathComponent(ExportManifest.libraryRelativePath),
            atomically: true,
            encoding: .utf8
        )
        try fileSystem.writeManifest(
            ExportManifest(
                version: 1,
                generatedAt: Date(),
                playlists: [
                    .init(id: "old", name: "Old", relativePath: "playlists/old.md"),
                ],
                files: [ExportManifest.libraryRelativePath]
            ),
            to: tempDirectory
        )

        let mockClient = MockAppleMusicClient()
        mockClient.playlists = [PlaylistSummary(id: "1", name: "New")]
        mockClient.tracksByPlaylistID = ["1": []]

        let exportService = ExportService(
            playlistService: PlaylistService(client: mockClient),
            fileSystem: fileSystem,
            markdownExporter: MarkdownExporter()
        )

        let summary = try await exportService.exportPlaylists(
            summaries: mockClient.playlists,
            to: tempDirectory
        )

        XCTAssertEqual(summary.exportedLibraryTrackCount, 0)
        XCTAssertTrue(
            FileManager.default.fileExists(
                atPath: tempDirectory.appendingPathComponent(ExportManifest.libraryRelativePath).path
            )
        )
        let index = try String(
            contentsOf: tempDirectory.appendingPathComponent("index.md"),
            encoding: .utf8
        )
        XCTAssertTrue(index.contains("- [Library](library.md)"))

        let manifest = try XCTUnwrap(fileSystem.readManifest(at: tempDirectory))
        XCTAssertEqual(manifest.files, [ExportManifest.libraryRelativePath])
    }

    func testExportRemovesStaleManifestFilesOnly() async throws {
        let tempDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString, isDirectory: true)
        defer { try? FileManager.default.removeItem(at: tempDirectory) }

        let fileSystem = FileSystemService()
        try fileSystem.ensureDirectoryExists(at: tempDirectory.appendingPathComponent("playlists"))

        let staleURL = tempDirectory.appendingPathComponent("playlists/old.md")
        try "stale".write(to: staleURL, atomically: true, encoding: .utf8)

        let untouchedURL = tempDirectory.appendingPathComponent("playlists/manual.md")
        try "manual".write(to: untouchedURL, atomically: true, encoding: .utf8)

        let previousManifest = ExportManifest(
            version: 1,
            generatedAt: Date(),
            playlists: [
                .init(id: "old", name: "Old", relativePath: "playlists/old.md"),
            ],
            files: ["library.md"]
        )
        try fileSystem.writeManifest(previousManifest, to: tempDirectory)

        let mockClient = MockAppleMusicClient()
        mockClient.playlists = [PlaylistSummary(id: "1", name: "New")]
        mockClient.tracksByPlaylistID = ["1": []]

        let exportService = ExportService(
            playlistService: PlaylistService(client: mockClient),
            fileSystem: fileSystem,
            markdownExporter: MarkdownExporter()
        )

        let summary = try await exportService.exportPlaylists(
            summaries: mockClient.playlists,
            to: tempDirectory
        )

        XCTAssertEqual(summary.removedStaleFiles, ["playlists/old.md"])
        XCTAssertFalse(FileManager.default.fileExists(atPath: staleURL.path))
        XCTAssertTrue(FileManager.default.fileExists(atPath: untouchedURL.path))
        XCTAssertNil(summary.logPath)
    }

    func testExportWarnsOnEmptyPlaylistInLog() async throws {
        let tempDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString, isDirectory: true)
        defer { try? FileManager.default.removeItem(at: tempDirectory) }

        let mockClient = MockAppleMusicClient()
        mockClient.playlists = [PlaylistSummary(id: "1", name: "Empty")]
        mockClient.tracksByPlaylistID = ["1": []]
        mockClient.librarySongs = []

        let exportService = ExportService(
            playlistService: PlaylistService(client: mockClient),
            fileSystem: FileSystemService(),
            markdownExporter: MarkdownExporter()
        )
        let summary = try await exportService.exportPlaylists(
            summaries: mockClient.playlists,
            to: tempDirectory,
            writeLogs: true
        )
        XCTAssertEqual(summary.logPath, "export.log")
        let log = try String(
            contentsOf: tempDirectory.appendingPathComponent("export.log"),
            encoding: .utf8
        )
        XCTAssertTrue(log.contains("WARNING Playlist \"Empty\" has no tracks"))
        XCTAssertFalse(log.contains("WARNING Library song list is empty"))
    }
}
