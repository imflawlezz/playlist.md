import Foundation

struct FileSystemService: Sendable {
    func resolveOutputDirectory(_ path: String) throws -> URL {
        let expanded = (path as NSString).expandingTildeInPath
        let url = URL(fileURLWithPath: expanded, isDirectory: true).standardizedFileURL
        return url
    }

    func ensureDirectoryExists(at url: URL) throws {
        try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
    }

    func readManifest(at outputDirectory: URL) -> ExportManifest? {
        let manifestURL = outputDirectory.appendingPathComponent(ExportManifest.filename)
        guard let data = try? Data(contentsOf: manifestURL) else {
            return nil
        }
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        return try? decoder.decode(ExportManifest.self, from: data)
    }

    func writeManifest(_ manifest: ExportManifest, to outputDirectory: URL) throws {
        let manifestURL = outputDirectory.appendingPathComponent(ExportManifest.filename)
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        let data = try encoder.encode(manifest)
        try data.write(to: manifestURL, options: .atomic)
    }

    func writeUTF8(_ content: String, to url: URL) throws {
        guard let data = content.data(using: .utf8) else {
            throw AppError.fileSystemError("Failed to encode file as UTF-8: \(url.path)")
        }
        try data.write(to: url, options: .atomic)
    }

    /// Deletes previous-manifest paths that are absent from the current export; leaves other files alone.
    func removeStaleFiles(previousManifest: ExportManifest?, currentRelativePaths: Set<String>, in outputDirectory: URL) throws -> [String] {
        guard let previousManifest else { return [] }

        let current = currentRelativePaths
        var removed: [String] = []
        let fileManager = FileManager.default

        for entry in previousManifest.playlists {
            guard !current.contains(entry.relativePath) else { continue }

            let fileURL = outputDirectory.appendingPathComponent(entry.relativePath)
            guard fileManager.fileExists(atPath: fileURL.path) else { continue }

            try fileManager.removeItem(at: fileURL)
            removed.append(entry.relativePath)
        }

        return removed.sorted()
    }
}
