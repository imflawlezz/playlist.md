import Foundation

actor ExportLogger {
    enum Level: String {
        case info = "INFO"
        case warning = "WARNING"
        case error = "ERROR"
    }

    private var lines: [String] = []
    private let formatter: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    func info(_ message: String) {
        append(level: .info, message: message)
    }

    func warning(_ message: String) {
        append(level: .warning, message: message)
    }

    func error(_ message: String) {
        append(level: .error, message: message)
    }

    func rendered() -> String {
        if lines.isEmpty {
            return ""
        }
        return lines.joined(separator: "\n") + "\n"
    }

    var isEmpty: Bool {
        lines.isEmpty
    }

    private func append(level: Level, message: String) {
        let stamp = formatter.string(from: Date())
        lines.append("\(stamp) \(level.rawValue) \(message)")
    }
}
