package semver

import "testing"

func mustParse(t *testing.T, s string) Version {
	t.Helper()
	v, err := ParseVersion(s)
	if err != nil {
		t.Fatalf("ParseVersion(%q) unexpected error: %v", s, err)
	}
	return v
}

func TestParseVersionValid(t *testing.T) {
	cases := []struct {
		in   string
		want Version
	}{
		{"1.2.3", Version{Major: 1, Minor: 2, Patch: 3}},
		{"v1.2.3", Version{Major: 1, Minor: 2, Patch: 3}},
		{"V1.2.3", Version{Major: 1, Minor: 2, Patch: 3}},
		{"0.0.0", Version{Major: 0, Minor: 0, Patch: 0}},
		{"1.0.0-alpha", Version{Major: 1, Patch: 0, Pre: []string{"alpha"}}},
		{"1.0.0-alpha.1", Version{Major: 1, Pre: []string{"alpha", "1"}}},
		{"1.0.0-0.3.7", Version{Major: 1, Pre: []string{"0", "3", "7"}}},
		{"1.0.0-x.7.z.92", Version{Major: 1, Pre: []string{"x", "7", "z", "92"}}},
		{"1.0.0+20130313144700", Version{Major: 1, Build: []string{"20130313144700"}}},
		{"1.0.0-beta+exp.sha.5114f85", Version{Major: 1, Pre: []string{"beta"}, Build: []string{"exp", "sha", "5114f85"}}},
		{"1.0.0+21AF26D3---117B344092BD", Version{Major: 1, Build: []string{"21AF26D3---117B344092BD"}}},
	}
	for _, c := range cases {
		got, err := ParseVersion(c.in)
		if err != nil {
			t.Errorf("ParseVersion(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got.String() != c.want.String() {
			t.Errorf("ParseVersion(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestParseVersionInvalid(t *testing.T) {
	cases := []string{
		"",
		"1",
		"1.2",
		"1.2.3.4",
		"01.2.3",  // leading zero, major
		"1.02.3",  // leading zero, minor
		"1.2.03",  // leading zero, patch
		"1.2.3-",  // empty prerelease
		"1.2.3-01",         // numeric prerelease identifier with leading zero
		"1.2.3-alpha_beta", // underscore isn't allowed
		"1.2.3-alpha..1",   // empty identifier between dots
		"1.2.3+",           // empty build metadata
		"-1.2.3",           // negative
		"1.a.3",            // non-numeric core field
		"abc",
	}
	for _, in := range cases {
		if _, err := ParseVersion(in); err == nil {
			t.Errorf("ParseVersion(%q) got no error, want one", in)
		}
	}
}

func TestCompareSpecOrder(t *testing.T) {
	// Ascending precedence order taken from semver.org's own worked example.
	ordered := []string{
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11",
		"1.0.0-rc.1",
		"1.0.0",
	}
	for i := 0; i < len(ordered)-1; i++ {
		lo := mustParse(t, ordered[i])
		hi := mustParse(t, ordered[i+1])
		if c := Compare(lo, hi); c != -1 {
			t.Errorf("Compare(%s, %s) = %d, want -1", ordered[i], ordered[i+1], c)
		}
		if c := Compare(hi, lo); c != 1 {
			t.Errorf("Compare(%s, %s) = %d, want 1", ordered[i+1], ordered[i], c)
		}
		if c := Compare(lo, lo); c != 0 {
			t.Errorf("Compare(%s, %s) = %d, want 0", ordered[i], ordered[i], c)
		}
	}
}

func TestCompareIgnoresBuildMetadata(t *testing.T) {
	a := mustParse(t, "1.0.0+build1")
	b := mustParse(t, "1.0.0+build2")
	if c := Compare(a, b); c != 0 {
		t.Errorf("Compare with differing build metadata = %d, want 0", c)
	}
}

func TestCompareNumericFieldsNotLexical(t *testing.T) {
	// 1.10.0 must outrank 1.9.0 numerically; a naive string compare gets
	// this backwards.
	a := mustParse(t, "1.9.0")
	b := mustParse(t, "1.10.0")
	if c := Compare(a, b); c != -1 {
		t.Errorf("Compare(1.9.0, 1.10.0) = %d, want -1", c)
	}
}

func TestSatisfies(t *testing.T) {
	cases := []struct {
		version    string
		constraint string
		want       bool
	}{
		{"1.2.3", ">=1.0.0 <2.0.0", true},
		{"1.0.0", ">=1.0.0 <2.0.0", true},  // inclusive lower bound
		{"2.0.0", ">=1.0.0 <2.0.0", false}, // exclusive upper bound
		{"1.2.3", "1.2.3", true},           // bare version means "="
		{"1.2.3", "=1.2.3", true},
		{"1.2.4", "=1.2.3", false},
		{"1.2.3+build", "=1.2.3", true},     // build metadata ignored
		{"1.0.0-beta", ">=1.0.0", false},    // prerelease ranks below the release it precedes
		{"1.0.0-beta", ">=1.0.0-alpha", true},
		{"1.0.0", ">1.0.0-rc.1", true},
		{"0.9.0", "<1.0.0 || >=2.0.0", true},
		{"1.5.0", "<1.0.0 || >=2.0.0", false},
		{"2.5.0", "<1.0.0 || >=2.0.0", true},
		{"1.2.3", "=1.2.3 || =4.5.6", true},
		{"4.5.6", "=1.2.3 || =4.5.6", true},
		{"7.0.0", "=1.2.3 || =4.5.6", false},
		{"1.2.3", "^1.2.3", true},
		{"1.9.9", "^1.2.3", true},
		{"1.2.2", "^1.2.3", false}, // caret has an inclusive lower bound
		{"2.0.0", "^1.2.3", false}, // caret never crosses a major bump
		{"0.2.5", "^0.2.3", true},  // 0.x: caret pins the minor instead
		{"0.3.0", "^0.2.3", false},
		{"0.0.4", "^0.0.3", false}, // 0.0.x: caret pins the patch too
		{"0.0.3", "^0.0.3", true},
		{"1.2.4", "~1.2.3", true},
		{"1.3.0", "~1.2.3", false}, // tilde never crosses a minor bump
		{"1.2.2", "~1.2.3", false},
	}
	for _, c := range cases {
		v := mustParse(t, c.version)
		cs, err := ParseConstraint(c.constraint)
		if err != nil {
			t.Fatalf("ParseConstraint(%q) unexpected error: %v", c.constraint, err)
		}
		got := Satisfies(v, cs)
		if got != c.want {
			t.Errorf("Satisfies(%s, %q) = %v, want %v", c.version, c.constraint, got, c.want)
		}
	}
}

func TestHighestMatch(t *testing.T) {
	in := []string{"1.0.0", "1.2.3", "1.9.9", "2.0.0", "1.5.0-beta"}
	versions := make([]Version, len(in))
	for i, s := range in {
		versions[i] = mustParse(t, s)
	}
	cs, err := ParseConstraint("^1.0.0")
	if err != nil {
		t.Fatalf("ParseConstraint unexpected error: %v", err)
	}

	best, ok := HighestMatch(versions, cs)
	if !ok {
		t.Fatal("HighestMatch found no match, want one")
	}
	if want := "1.9.9"; best.String() != want {
		t.Errorf("HighestMatch = %s, want %s", best, want)
	}
}

func TestHighestMatchNoMatch(t *testing.T) {
	versions := []Version{mustParse(t, "1.0.0"), mustParse(t, "1.5.0")}
	cs, err := ParseConstraint(">=2.0.0")
	if err != nil {
		t.Fatalf("ParseConstraint unexpected error: %v", err)
	}
	if _, ok := HighestMatch(versions, cs); ok {
		t.Error("HighestMatch found a match, want none")
	}
}

func TestHighestMatchEmptyList(t *testing.T) {
	cs, err := ParseConstraint(">=1.0.0")
	if err != nil {
		t.Fatalf("ParseConstraint unexpected error: %v", err)
	}
	if _, ok := HighestMatch(nil, cs); ok {
		t.Error("HighestMatch on empty list found a match, want none")
	}
}

// FuzzParseVersion checks that ParseVersion never panics and that any
// version it accepts round-trips: String() must reparse to an equal
// version. That catches formatting bugs (wrong separators, dropped
// identifiers) that the table tests above wouldn't stumble onto.
func FuzzParseVersion(f *testing.F) {
	seeds := []string{
		"1.2.3",
		"v1.2.3",
		"V1.2.3",
		"0.0.0",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-0.3.7",
		"1.0.0+20130313144700",
		"1.0.0-beta+exp.sha.5114f85",
		"1.0.0+21AF26D3---117B344092BD",
		"",
		"1",
		"1.2",
		"1.2.3.4",
		"01.2.3",
		"1.2.3-",
		"1.2.3-01",
		"1.2.3-alpha_beta",
		"1.2.3-alpha..1",
		"1.2.3+",
		"-1.2.3",
		"1.a.3",
		"abc",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		v, err := ParseVersion(s)
		if err != nil {
			return
		}

		out := v.String()
		v2, err2 := ParseVersion(out)
		if err2 != nil {
			t.Fatalf("ParseVersion(%q) = %+v, but reparsing its String() %q failed: %v", s, v, out, err2)
		}
		if Compare(v, v2) != 0 {
			t.Fatalf("round-trip changed precedence: ParseVersion(%q) = %+v, String() = %q, reparsed = %+v", s, v, out, v2)
		}
		if out != v2.String() {
			t.Fatalf("round-trip not stable: %q -> %q -> %q", s, out, v2.String())
		}
	})
}

func TestParseConstraintInvalid(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"not-a-version",
		">=",
		">=1.2.3-",
		"||",
		"1.2.3 ||",
		"|| 1.2.3",
		"1.2.3 || not-a-version",
		"^not-a-version",
		"~1.2",
	}
	for _, in := range cases {
		if _, err := ParseConstraint(in); err == nil {
			t.Errorf("ParseConstraint(%q) got no error, want one", in)
		}
	}
}
