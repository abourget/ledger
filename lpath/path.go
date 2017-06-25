// Package lpath provides support for manipulating colon-separated account names.
package lpath

import "strings"

var Separator = ":"

func Base(path string) string {
	if path == "" {
		return ""
	}

	strings.TrimRight(path, Separator)
	i := strings.LastIndex(path, Separator)
	if i < 0 {
		return ""
	}
	return path[:i]
}

func HasBase(path, base string) bool {
	return strings.HasPrefix(path+Separator, base+Separator)
}
