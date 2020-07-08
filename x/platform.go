package x

import "strings"

// Detect platform from the UserAgent string.
func PlatformFromUA(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "outgoing-js"):
		return "web"
	case strings.Contains(ua, "outgoing-android"):
		return "android"
	case strings.Contains(ua, "outgoing-ios"):
		return "ios"
	default:
		switch {
		case strings.Contains(ua, "iphone"),
			strings.Contains(ua, "ipad"):
			return "ios"
		case strings.Contains(ua, "android"):
			return "android"
		}
		return ""
	}
}
