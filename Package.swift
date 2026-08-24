// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "playlist-md",
    platforms: [
        .macOS(.v14),
    ],
    products: [
        .executable(name: "playlist-md-core", targets: ["playlist_md_core"]),
        .library(name: "PlaylistMDCore", targets: ["PlaylistMDCore"]),
    ],
    targets: [
        .target(
            name: "PlaylistMDCore",
            path: "Sources/PlaylistMDCore",
            exclude: ["Info.plist"]
        ),
        .executableTarget(
            name: "playlist_md_core",
            dependencies: ["PlaylistMDCore"],
            path: "Sources/playlist_md_core",
            linkerSettings: [
                .unsafeFlags([
                    "-Xlinker", "-sectcreate",
                    "-Xlinker", "__TEXT",
                    "-Xlinker", "__info_plist",
                    "-Xlinker", "Sources/PlaylistMDCore/Info.plist",
                ], .when(platforms: [.macOS])),
            ]
        ),
        .testTarget(
            name: "playlist_mdTests",
            dependencies: ["PlaylistMDCore"],
            path: "Tests/playlist_mdTests"
        ),
    ]
)
