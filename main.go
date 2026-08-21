package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s VERSION CONSTRAINT\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "example: %s 1.2.3 \">=1.0.0 <2.0.0\"\n", os.Args[0])
		os.Exit(2)
	}

	v, err := ParseVersion(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cs, err := ParseConstraint(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if Satisfies(v, cs) {
		fmt.Println("true")
		os.Exit(0)
	}
	fmt.Println("false")
	os.Exit(1)
}
