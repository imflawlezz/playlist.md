import Foundation

extension AuthorizationStatus {
    var rawValue: String {
        switch self {
        case .authorized:
            return "authorized"
        case .denied:
            return "denied"
        case .restricted:
            return "restricted"
        case .notDetermined:
            return "not_determined"
        }
    }
}
