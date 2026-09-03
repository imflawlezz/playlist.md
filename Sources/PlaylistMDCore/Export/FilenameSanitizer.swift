import Foundation

enum FilenameSanitizer {
    private static let invalidCharacterSet: CharacterSet = {
        var set = CharacterSet()
        set.formUnion(.controlCharacters)
        // Dots are separators too: trailing "..." and "../" must not leave ".." in the slug.
        set.insert(charactersIn: "/\\:?*\"<>|.")
        set.insert(charactersIn: "\u{0000}")
        return set
    }()

    static func slug(from name: String) -> String {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            return "untitled-playlist"
        }

        var result = ""
        var previousWasSeparator = false

        for scalar in trimmed.unicodeScalars {
            if invalidCharacterSet.contains(scalar) || scalar == " " {
                if !previousWasSeparator {
                    result.append("-")
                    previousWasSeparator = true
                }
                continue
            }

            result.unicodeScalars.append(scalar)
            previousWasSeparator = false
        }

        let collapsed = result
            .trimmingCharacters(in: CharacterSet(charactersIn: "-"))
            .lowercased()

        if collapsed.isEmpty {
            return "untitled-playlist"
        }

        return collapsed
    }

    /// Duplicate slugs get `-2`, `-3`, … in name-then-id order so reruns stay stable.
    static func assignFilenames(for playlists: [PlaylistSummary]) -> [String: String] {
        var slugCounts: [String: Int] = [:]
        var filenames: [String: String] = [:]

        let sorted = playlists.sorted {
            if $0.name != $1.name {
                return $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending
            }
            return $0.id < $1.id
        }

        for playlist in sorted {
            let baseSlug = slug(from: playlist.name)
            let count = slugCounts[baseSlug, default: 0] + 1
            slugCounts[baseSlug] = count

            let filename: String
            if count == 1 {
                filename = "\(baseSlug).md"
            } else {
                filename = "\(baseSlug)-\(count).md"
            }

            filenames[playlist.id] = filename
        }

        return filenames
    }

    /// Rejects traversal and absolute paths so a generated name cannot leave the output directory.
    static func isSafeRelativeFilename(_ filename: String) -> Bool {
        guard !filename.isEmpty else { return false }
        guard !filename.contains("..") else { return false }
        guard !filename.hasPrefix("/") else { return false }
        guard !filename.hasPrefix(".") else { return false }
        return filename.allSatisfy { char in
            !char.isNewline && char != "\0"
        }
    }
}
