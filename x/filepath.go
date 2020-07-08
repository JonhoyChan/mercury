package x

import "path/filepath"

// Convert relative filepath to absolute.
func ToAbsolutePath(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Clean(filepath.Join(base, path))
}
