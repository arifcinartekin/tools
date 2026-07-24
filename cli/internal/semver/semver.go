// Package semver implements just enough semantic version comparison for
// toolbox's update checks (MAJOR.MINOR.PATCH, optional leading "v").
package semver

import (
	"strconv"
	"strings"
)

// Compare returns -1, 0, or 1 depending on whether a is less than, equal
// to, or greater than b. Non-numeric or missing components are treated as 0.
func Compare(a, b string) int {
	pa := parse(a)
	pb := parse(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// LessThan reports whether a < b as semantic versions.
func LessThan(a, b string) bool {
	return Compare(a, b) < 0
}

func parse(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	// Drop any pre-release/build metadata suffix (e.g. "1.2.0-beta.1").
	if i := strings.IndexAny(v, "-+"); i != -1 {
		v = v[:i]
	}
	parts := strings.SplitN(v, ".", 3)
	var out [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			continue
		}
		out[i] = n
	}
	return out
}
