import Foundation

enum AppError: LocalizedError, Equatable {
    case authorizationRequired
    case authorizationDenied
    case libraryUnavailable
    case networkFailure(String)
    case playlistFetchFailed(String)
    case invalidOutputDirectory(String)
    case fileSystemError(String)
    case cancelled

    var errorDescription: String? {
        switch self {
        case .authorizationRequired:
            return "Apple Music authorization is required."
        case .authorizationDenied:
            return "Apple Music authorization was denied. Enable access in System Settings → Privacy & Security → Media & Apple Music."
        case .libraryUnavailable:
            return "Unable to access your Apple Music library."
        case .networkFailure(let context):
            return "Network error while fetching \(context)."
        case .playlistFetchFailed(let name):
            return "Unable to fetch playlist \"\(name)\"."
        case .invalidOutputDirectory(let path):
            return "Invalid output directory: \(path)"
        case .fileSystemError(let message):
            return message
        case .cancelled:
            return "Operation cancelled."
        }
    }
}
