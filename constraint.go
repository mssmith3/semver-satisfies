package main

import (
	"fmt"
	"strings"
)

// Comparator is a single "OP VERSION" term, e.g. ">=1.2.3".
type Comparator struct {
	Op  string // one of "", "=", ">", ">=", "<", "<="
	Ver Version
}

var comparatorOps = []string{">=", "<=", ">", "<", "="}

// ParseConstraint parses a whitespace separated list of comparators that
// must all hold (AND semantics). "1.2.3" is shorthand for "=1.2.3".
func ParseConstraint(s string) ([]Comparator, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty constraint")
	}
	out := make([]Comparator, 0, len(fields))
	for _, f := range fields {
		op, rest := splitOp(f)
		v, err := ParseVersion(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid constraint term %q: %w", f, err)
		}
		out = append(out, Comparator{Op: op, Ver: v})
	}
	return out, nil
}

func splitOp(f string) (op, rest string) {
	for _, op := range comparatorOps {
		if strings.HasPrefix(f, op) {
			return op, f[len(op):]
		}
	}
	return "=", f
}

// Satisfies reports whether v meets every comparator in cs.
func Satisfies(v Version, cs []Comparator) bool {
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
