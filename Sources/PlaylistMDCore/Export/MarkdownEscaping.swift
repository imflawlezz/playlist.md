import Foundation

enum MarkdownEscaping {
    static func escapeTableCell(_ text: String) -> String {
        text
            .replacingOccurrences(of: "\\", with: "\\\\")
            .replacingOccurrences(of: "|", with: "\\|")
            .replacingOccurrences(of: "[", with: "\\[")
            .replacingOccurrences(of: "]", with: "\\]")
            .replacingOccurrences(of: "*", with: "\\*")
            .replacingOccurrences(of: "_", with: "\\_")
    }

    static func escapeLinkLabel(_ text: String) -> String {
        text
            .replacingOccurrences(of: "\\", with: "\\\\")
            .replacingOccurrences(of: "[", with: "\\[")
            .replacingOccurrences(of: "]", with: "\\]")
    }

    static func formatTrackCell(title: String, url: URL?) -> String {
        let escapedTitle = escapeTableCell(title)
        guard let url else {
            return escapedTitle
        }
        return "[\(escapeLinkLabel(title))](\(url.absoluteString))"
    }
}
