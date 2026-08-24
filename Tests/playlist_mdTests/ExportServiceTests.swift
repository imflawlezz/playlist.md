import XCTest
@testable import PlaylistMDCore

final class ExportServiceTests: XCTestCase {
    func testExportWritesIndexPlaylistAndManifest() async throws {
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

        let playlistService = PlaylistService(client: mockClient)
        let fileSystem = FileSystemService()
        let exportService = ExportService(
            playlistService: playlistService,
            fileSystem: fileSystem,
            markdownExporter: MarkdownExporter()
        )

        let summary = try await exportService.export(
            summaries: mockClient.playlists,
            to: tempDirectory
        )

        XCTAssertEqual(summary.exportedPlaylistCount, 1)
        XCTAssertEqual(summary.exportedTrackCount, 1)

        let indexURL = tempDirectory.appendingPathComponent("index.md")
        let playlistURL = tempDirectory.appendingPathComponent("playlists/chill.md")
        let manifestURL = tempDirectory.appendingPathComponent(ExportManifest.filename)

        XCTAssertTrue(FileManager.default.fileExists(atPath: indexURL.path))
        XCTAssertTrue(FileManager.default.fileExists(atPath: playlistURL.path))
        XCTAssertTrue(FileManager.default.fileExists(atPath: manifestURL.path))

        let index = try String(contentsOf: indexURL, encoding: .utf8)
        XCTAssertTrue(index.contains("- [Chill](playlists/chill.md)"))
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
            ]
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

        let summary = try await exportService.export(
            summaries: mockClient.playlists,
            to: tempDirectory
        )

        XCTAssertEqual(summary.removedStaleFiles, ["playlists/old.md"])
        XCTAssertFalse(FileManager.default.fileExists(atPath: staleURL.path))
        XCTAssertTrue(FileManager.default.fileExists(atPath: untouchedURL.path))
    }
}
