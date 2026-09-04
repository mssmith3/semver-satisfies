// Package semver parses semantic versions and constraints and compares them
// per semver.org 2.0.0 precedence rules.
package semver

import (
	"fmt"
	"strings"
)

// Version is a parsed semantic version per semver.org 2.0.0.
// Build metadata is kept for round-tripping but never affects Compare.
type Version struct {
	Major, Minor, Patch uint64
	Pre                 []string
	Build               []string
}

func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Pre) > 0 {
		s += "-" + strings.Join(v.Pre, ".")
	}
	if len(v.Build) > 0 {
		s += "+" + strings.Join(v.Build, ".")
	}
	return s
}

// ParseVersion parses a semantic version string. A single leading "v" or "V"
// is tolerated since it's how most tags in the wild are written.
func ParseVersion(s string) (Version, error) {
	orig := s
	if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		s = s[1:]
	}

	var build string
	hasBuild := false
	if i := strings.IndexByte(s, '+'); i >= 0 {
		hasBuild = true
		build = s[i+1:]
		s = s[:i]
	}

	var pre string
	hasPre := false
	if i := strings.IndexByte(s, '-'); i >= 0 {
		hasPre = true
		pre = s[i+1:]
		s = s[:i]
	}

	core := strings.Split(s, ".")
	if len(core) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: expected MAJOR.MINOR.PATCH", orig)
	}
	major, err := parseNumericField(core[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version %q: major: %w", orig, err)
	}
	minor, err := parseNumericField(core[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version %q: minor: %w", orig, err)
	}
	patch, err := parseNumericField(core[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version %q: patch: %w", orig, err)
	}

	var preIDs []string
	if hasPre {
		if pre == "" {
			return Version{}, fmt.Errorf("invalid version %q: empty prerelease", orig)
		}
		preIDs = strings.Split(pre, ".")
		for _, id := range preIDs {
			if err := validatePreID(id); err != nil {
				return Version{}, fmt.Errorf("invalid version %q: prerelease: %w", orig, err)
			}
		}
	}

	var buildIDs []string
	if hasBuild {
		if build == "" {
			return Version{}, fmt.Errorf("invalid version %q: empty build metadata", orig)
		}
		buildIDs = strings.Split(build, ".")
		for _, id := range buildIDs {
			if id == "" || !isAlnumHyphen(id) {
				return Version{}, fmt.Errorf("invalid version %q: build metadata identifier %q", orig, id)
			}
		}
	}

	return Version{Major: major, Minor: minor, Patch: patch, Pre: preIDs, Build: buildIDs}, nil
}

func parseNumericField(s string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty numeric field")
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("leading zero in %q", s)
	}
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("non-numeric field %q", s)
		}
		n = n*10 + uint64(c-'0')
	}
	return n, nil
}

func validatePreID(id string) error {
	if id == "" {
		return fmt.Errorf("empty identifier")
	}
	if isNumericID(id) {
		if len(id) > 1 && id[0] == '0' {
			return fmt.Errorf("leading zero in numeric identifier %q", id)
		}
		return nil
	}
	if !isAlnumHyphen(id) {
		return fmt.Errorf("invalid identifier %q", id)
	}
	return nil
}

func isNumericID(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isAlnumHyphen(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c == '-':
		default:
			return false
		}
	}
	return true
}

// Compare returns -1, 0, or 1 as v is less than, equal to, or greater than
// other, following semver.org section 11 precedence rules.
func Compare(v, other Version) int {
	if v.Major != other.Major {
		return cmpUint(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return cmpUint(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return cmpUint(v.Patch, other.Patch)
	}
	return comparePre(v.Pre, other.Pre)
}

func cmpUint(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// comparePre implements the rule that a version with no prerelease outranks
// one that has any, and otherwise compares identifiers left to right.
func comparePre(a, b []string) int {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	if len(a) == 0 {
		return 1
	}
	if len(b) == 0 {
		return -1
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if c := compareIdentifier(a[i], b[i]); c != 0 {
			return c
		}
	}
	return cmpInt(len(a), len(b))
}

func compareIdentifier(a, b string) int {
	aNum, bNum := isNumericID(a), isNumericID(b)
	switch {
	case aNum && bNum:
		if len(a) != len(b) {
			return cmpInt(len(a), len(b))
		}
		return strings.Compare(a, b)
	case aNum && !bNum:
		return -1
	case !aNum && bNum:
		return 1
	default:
		return strings.Compare(a, b)
	}
}
