package main

import (
	"fmt"
	"os"

	"semver-satisfies/internal/semver"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "highest" {
		runHighest(os.Args[2:])
		return
	}
	runCheck(os.Args[1:])
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s VERSION CONSTRAINT\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s highest CONSTRAINT VERSION [VERSION...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "example: %s 1.2.3 \">=1.0.0 <2.0.0\"\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "example: %s highest \">=1.0.0 <2.0.0\" 1.2.3 1.9.9 2.0.0\n", os.Args[0])
}

func runCheck(args []string) {
	if len(args) != 2 {
		usage()
		os.Exit(2)
	}

	v, err := semver.ParseVersion(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cs, err := semver.ParseConstraint(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if semver.Satisfies(v, cs) {
		fmt.Println("true")
		os.Exit(0)
	}
	fmt.Println("false")
	os.Exit(1)
}

// runHighest prints the highest of a list of versions that satisfies a
// constraint, or "none" and a nonzero exit if no version does.
func runHighest(args []string) {
	if len(args) < 2 {
		usage()
		os.Exit(2)
	}

	cs, err := semver.ParseConstraint(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	versions := make([]semver.Version, len(args)-1)
	for i, a := range args[1:] {
		v, err := semver.ParseVersion(a)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		versions[i] = v
	}

	best, ok := semver.HighestMatch(versions, cs)
	if !ok {
		fmt.Println("none")
		os.Exit(1)
	}
	fmt.Println(best.String())
}
