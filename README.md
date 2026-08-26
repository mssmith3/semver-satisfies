# semver-satisfies

A command that answers one question: does this version satisfy this
constraint?

## why

Shell scripts and CI configs end up doing version checks with plain string
comparison — `[[ "$VERSION" > "1.2.0" ]]` — which breaks the moment a major
version hits double digits (`1.10.0` sorts before `1.9.0` as strings) and has
no idea that `2.0.0-rc.1` is *older* than `2.0.0`, not newer. The rules for
comparing semantic versions correctly are short but easy to get wrong by
hand, especially around prerelease precedence, so it's worth having one
small tool that does it once and gets tested.

## usage

```
$ go run . 1.2.3 ">=1.0.0 <2.0.0"
true

$ go run . 2.0.0 ">=1.0.0 <2.0.0"
false

$ go run . 1.0.0-beta ">=1.0.0"
false
```

The exit code mirrors the answer: `0` if the version satisfies the
constraint, `1` if it doesn't, `2` if either argument fails to parse.

### constraint syntax

A constraint is one or more comparator terms separated by spaces, all of
which must hold (AND). Supported comparators: `=`, `>`, `>=`, `<`, `<=`. A
bare version with no operator is shorthand for `=`. Groups of AND terms can
be separated by `||`, and the constraint holds if any group does (OR).

```
1.2.3                          exactly 1.2.3
>=1.2.3                        1.2.3 or newer
>=1.2.3 <2.0.0                 1.2.3 up to, but not including, 2.0.0
<1.0.0 || >=2.0.0              anything except the 1.x line
```

`^` and `~` are shorthand for an `>=` / `<` pair, expanded before matching.
`^1.2.3` means "compatible with 1.2.3" — it allows anything up to the next
major version, except below `1.0.0` where a `0.x` release is treated as
unstable and the allowed range narrows to match (`^0.2.3` stays within
`0.2.x`, `^0.0.3` pins the patch too). `~1.2.3` is narrower and always pins
major and minor, allowing only patch-level changes.

```
^1.2.3                         >=1.2.3 <2.0.0
^0.2.3                         >=0.2.3 <0.3.0
^0.0.3                         >=0.0.3 <0.0.4
~1.2.3                         >=1.2.3 <1.3.0
```

Both require a full `MAJOR.MINOR.PATCH` version after the prefix; partial
versions like `^1.2` aren't supported.

### version syntax

Versions follow [semver.org 2.0.0](https://semver.org/). A leading `v` or
`V` is stripped before parsing, since that's how most git tags are written.
Build metadata (`+build.1`) is parsed but never affects comparison or
ordering, per spec.

## development

Standard library only, no build step beyond `go build`. Run the tests with:

```
go test ./...
```

The test suite is table-driven and leans on semver.org's own precedence
example (`1.0.0-alpha < 1.0.0-alpha.1 < ... < 1.0.0`) plus the cases that
are easy to get wrong: leading zeros, numeric-vs-alphanumeric prerelease
identifiers, and build metadata that should be ignored entirely.
