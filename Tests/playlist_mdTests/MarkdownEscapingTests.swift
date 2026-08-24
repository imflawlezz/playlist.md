import XCTest
@testable import PlaylistMDCore

final class MarkdownEscapingTests: XCTestCase {
    func testEscapesTablePipeCharacters() {
        XCTAssertEqual(MarkdownEscaping.escapeTableCell("A | B"), "A \\| B")
    }

    func testEscapesTableBackslashes() {
        XCTAssertEqual(MarkdownEscaping.escapeTableCell("A \\ B"), "A \\\\ B")
    }

    func testEscapesLinkBrackets() {
        XCTAssertEqual(MarkdownEscaping.escapeLinkLabel("Track [Live]"), "Track \\[Live\\]")
    }

    func testFormatTrackCellWithoutURL() {
        XCTAssertEqual(MarkdownEscaping.formatTrackCell(title: "Song *Title*", url: nil), "Song \\*Title\\*")
    }

    func testFormatTrackCellWithURL() {
        let result = MarkdownEscaping.formatTrackCell(
            title: "Color Your Night",
            url: URL(string: "https://music.apple.com/song/1")
        )
        XCTAssertEqual(result, "[Color Your Night](https://music.apple.com/song/1)")
    }

    func testFormatTrackCellEscapesLinkLabel() {
        let result = MarkdownEscaping.formatTrackCell(
            title: "Track [Remix]",
            url: URL(string: "https://music.apple.com/song/2")
        )
        XCTAssertEqual(result, "[Track \\[Remix\\]](https://music.apple.com/song/2)")
    }
}
