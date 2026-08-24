import XCTest
@testable import PlaylistMDCore

final class MarkdownExporterTests: XCTestCase {
    private let exporter = MarkdownExporter()

    func testRenderPlaylistIncludesTableAndMetadata() {
        let playlist = Playlist(
            id: "1",
            name: "Persona",
            tracks: [
                Track(
                    title: "Color Your Night",
                    artist: "Lyn Inaizumi",
                    album: "Persona 3 Reload",
                    year: 2024,
                    url: URL(string: "https://music.apple.com/song/1"),
                    position: 1
                ),
                Track(
                    title: "Full Moon Full Life",
                    artist: "Azumi Takahashi",
                    album: "Persona 3 Reload",
                    year: 2024,
                    url: nil,
                    position: 2
                ),
            ]
        )

        let markdown = exporter.renderPlaylist(playlist)

        XCTAssertTrue(markdown.hasPrefix("# Persona\n"))
        XCTAssertTrue(markdown.contains("2 tracks"))
        XCTAssertTrue(markdown.contains("| 1 | [Color Your Night](https://music.apple.com/song/1) | Lyn Inaizumi | Persona 3 Reload | 2024 |"))
        XCTAssertTrue(markdown.contains("| 2 | Full Moon Full Life | Azumi Takahashi | Persona 3 Reload | 2024 |"))
    }

    func testRenderPlaylistEscapesSpecialCharacters() {
        let playlist = Playlist(
            id: "1",
            name: "Mix | Special",
            tracks: [
                Track(
                    title: "Pipe | Song",
                    artist: "Artist *Name*",
                    album: "Album [Deluxe]",
                    year: nil,
                    url: nil,
                    position: 1
                ),
            ]
        )

        let markdown = exporter.renderPlaylist(playlist)

        XCTAssertTrue(markdown.contains("# Mix \\| Special"))
        XCTAssertTrue(markdown.contains("| 1 | Pipe \\| Song | Artist \\*Name\\* | Album \\[Deluxe\\] |  |"))
    }

    func testRenderIndexUsesRelativeLinks() {
        let exported = [
            ExportedPlaylist(
                playlist: Playlist(id: "2", name: "Gaming", tracks: []),
                relativePath: "playlists/gaming.md"
            ),
            ExportedPlaylist(
                playlist: Playlist(id: "1", name: "Chill", tracks: []),
                relativePath: "playlists/chill.md"
            ),
        ]

        let markdown = exporter.renderIndex(playlists: exported)

        XCTAssertEqual(
            markdown,
            """
            # Apple Music Playlists

            - [Chill](playlists/chill.md)
            - [Gaming](playlists/gaming.md)

            """
        )
    }

    func testDeterministicOutputForSameInput() {
        let playlist = Playlist(
            id: "1",
            name: "Repeat",
            tracks: [
                Track(title: "A", artist: "B", album: "C", year: 2020, url: nil, position: 1),
            ]
        )

        let first = exporter.renderPlaylist(playlist)
        let second = exporter.renderPlaylist(playlist)

        XCTAssertEqual(first, second)
    }
}
