import Foundation

struct Track: Equatable, Sendable {
    let title: String
    let artist: String
    let album: String
    let year: Int?
    let url: URL?
    let position: Int
}
