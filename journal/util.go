package journal

import "strings"

func BaseAccount(path string) string {
	if path == "" {
		return ""
	}
	// Strip trailing slashes.
	for len(path) > 0 && path[len(path)-1] == ':' {
		path = path[0 : len(path)-1]
	}
	// Find the last element
	i := strings.LastIndex(path, ":")
	if i < 0 {
		return ""
	}
	return path[:i]
}
