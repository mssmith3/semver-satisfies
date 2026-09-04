package semver

import (
	"fmt"
	"strings"
)

// Comparator is a single "OP VERSION" term, e.g. ">=1.2.3".
type Comparator struct {
	Op  string // one of "", "=", ">", ">=", "<", "<="
	Ver Version
}

// Constraint is a set of comparator groups joined by OR, where each group's
// comparators are joined by AND. It's satisfied by a version that satisfies
// at least one group.
type Constraint [][]Comparator

var comparatorOps = []string{">=", "<=", ">", "<", "="}

// ParseConstraint parses one or more "||" separated groups of whitespace
// separated comparators. Within a group all comparators must hold (AND);
// across groups only one needs to (OR). "1.2.3" is shorthand for "=1.2.3".
// "^1.2.3" and "~1.2.3" are shorthand for a ">= / <" pair; see expandRange.
func ParseConstraint(s string) (Constraint, error) {
	groups := strings.Split(s, "||")
	out := make(Constraint, 0, len(groups))
	for _, g := range groups {
		fields := strings.Fields(g)
		if len(fields) == 0 {
			return nil, fmt.Errorf("empty constraint group in %q", s)
		}
		cs := make([]Comparator, 0, len(fields))
		for _, f := range fields {
			if f[0] == '^' || f[0] == '~' {
				v, err := ParseVersion(f[1:])
				if err != nil {
					return nil, fmt.Errorf("invalid constraint term %q: %w", f, err)
				}
				lo, hi := expandRange(f[0], v)
				cs = append(cs, lo, hi)
				continue
			}
			op, rest := splitOp(f)
			v, err := ParseVersion(rest)
			if err != nil {
				return nil, fmt.Errorf("invalid constraint term %q: %w", f, err)
			}
			cs = append(cs, Comparator{Op: op, Ver: v})
		}
		out = append(out, cs)
	}
	return out, nil
}

// expandRange turns a caret ('^') or tilde ('~') shorthand into an explicit
// ">= lower" / "< upper" pair. Caret allows changes that don't touch the
// leftmost nonzero component (so a 0.x release is treated as unstable and
// gets a narrower range than a 1.x one); tilde always pins major and minor.
func expandRange(kind byte, v Version) (lo, hi Comparator) {
	lo = Comparator{Op: ">=", Ver: v}
	var upper Version
	switch {
	case kind == '~':
		upper = Version{Major: v.Major, Minor: v.Minor + 1}
	case v.Major > 0:
		upper = Version{Major: v.Major + 1}
	case v.Minor > 0:
		upper = Version{Minor: v.Minor + 1}
	default:
		upper = Version{Patch: v.Patch + 1}
	}
	hi = Comparator{Op: "<", Ver: upper}
	return lo, hi
}

func splitOp(f string) (op, rest string) {
	for _, op := range comparatorOps {
		if strings.HasPrefix(f, op) {
			return op, f[len(op):]
		}
	}
	return "=", f
}

// Satisfies reports whether v meets at least one group of c, where meeting a
// group means matching every comparator in it.
func Satisfies(v Version, c Constraint) bool {
	for _, group := range c {
		if satisfiesGroup(v, group) {
			return true
		}
	}
	return false
}

// HighestMatch returns the highest of versions that satisfies c. ok is false
// if none of them do, in which case the returned Version is meaningless.
func HighestMatch(versions []Version, c Constraint) (best Version, ok bool) {
	for _, v := range versions {
		if !Satisfies(v, c) {
			continue
		}
		if !ok || Compare(v, best) > 0 {
			best = v
			ok = true
		}
	}
	return best, ok
}

func satisfiesGroup(v Version, cs []Comparator) bool {
	for _, c := range cs {
		cmp := Compare(v, c.Ver)
		var ok bool
		switch c.Op {
		case "=", "":
			ok = cmp == 0
		case ">":
			ok = cmp > 0
		case ">=":
			ok = cmp >= 0
		case "<":
			ok = cmp < 0
		case "<=":
			ok = cmp <= 0
		}
		if !ok {
			return false
		}
	}
	return true
}
