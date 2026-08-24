import Foundation
import PlaylistMDCore

@main
struct PlaylistMDCore {
    static func main() async {
        let exitCode = await CoreCLI.run(arguments: CommandLine.arguments)
        exit(exitCode)
    }
}
