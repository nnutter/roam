package main

import "slices"

var helpArgs = []string{"-h", "--help"}

func isHelpArg(args []string) bool {
	if len(args) != 1 {
		return false
	}
	return slices.Contains(helpArgs, args[0])
}
