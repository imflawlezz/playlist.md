import XCTest
@testable import PlaylistMDCore

final class FilenameSanitizerTests: XCTestCase {
    func testSlugReplacesInvalidCharacters() {
        XCTAssertEqual(FilenameSanitizer.slug(from: "My Playlist / 2026"), "my-playlist-2026")
    }

    func testSlugTrimsWhitespace() {
        XCTAssertEqual(FilenameSanitizer.slug(from: "  Chill  "), "chill")
    }

    func testSlugPreservesUnicode() {
        XCTAssertEqual(FilenameSanitizer.slug(from: "Café Mix"), "café-mix")
    }

    func testSlugHandlesEmptyName() {
        XCTAssertEqual(FilenameSanitizer.slug(from: "   "), "untitled-playlist")
    }

    func testSlugPreventsPathTraversal() {
        XCTAssertEqual(FilenameSanitizer.slug(from: "../escape"), "..-escape")
        XCTAssertFalse(FilenameSanitizer.isSafeRelativeFilename("../escape.md"))
    }

    func testDuplicateFilenamesAreDeterministic() {
        let playlists = [
            PlaylistSummary(id: "2", name: "Mix"),
            PlaylistSummary(id: "1", name: "Mix"),
            PlaylistSummary(id: "3", name: "Other"),
        ]

        let filenames = FilenameSanitizer.assignFilenames(for: playlists)
        XCTAssertEqual(filenames["1"], "mix.md")
        XCTAssertEqual(filenames["2"], "mix-2.md")
        XCTAssertEqual(filenames["3"], "other.md")
    }

    func testDuplicateOrderingIsStableByNameThenID() {
        let playlists = [
            PlaylistSummary(id: "b", name: "Alpha"),
            PlaylistSummary(id: "a", name: "Alpha"),
        ]

        let first = FilenameSanitizer.assignFilenames(for: playlists)
        let second = FilenameSanitizer.assignFilenames(for: playlists.reversed())

        XCTAssertEqual(first, second)
        XCTAssertEqual(first["a"], "alpha.md")
        XCTAssertEqual(first["b"], "alpha-2.md")
    }
}
