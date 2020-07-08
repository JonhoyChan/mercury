package x

import (
	"strconv"
	"strings"
	"unicode"
)

// Parses semantic version string in the following formats:
//  1.2, 1.2abc, 1.2.3, 1.2.3-abc, v0.12.34-rc5
func ParseVersion(version string) int {
	var major, minor, patch int
	// Remove optional "v" prefix.
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}
	// We can handle 3 parts only.
	parts := strings.SplitN(version, ".", 3)
	count := len(parts)
	if count > 0 {
		major = parseVersionPart(parts[0])
		if count > 1 {
			minor = parseVersionPart(parts[1])
			if count > 2 {
				patch = parseVersionPart(parts[2])
			}
		}
	}

	return (major << 16) | (minor << 8) | patch
}

// Parse one component of a semantic version string.
func parseVersionPart(version string) int {
	end := strings.IndexFunc(version, func(r rune) bool {
		return !unicode.IsDigit(r)
	})

	t := 0
	var err error
	if end > 0 {
		t, err = strconv.Atoi(version[:end])
	} else if len(version) > 0 {
		t, err = strconv.Atoi(version)
	}
	if err != nil || t > 0x1fff || t <= 0 {
		return 0
	}
	return t
}

// Returns > 0 if v1 > v2; zero if equal; < 0 if v1 < v2
// Only Major and Minor parts are compared, the trailer is ignored.
func VersionCompare(v1, v2 int) int {
	return (v1 >> 8) - (v2 >> 8)
}
