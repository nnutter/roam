package main

import "slices"

var helpArgs = []string{"-h", "--help"}

func isHelpArg(args []string) bool {
	if len(args) != 1 {
		return false
	}
	return slices.Contains(helpArgs, args[0])
}

var versionArgs = []string{"-v", "--version"}

func isVersionArg(args []string) bool {
	if len(args) != 1 {
		return false
	}
	return slices.Contains(versionArgs, args[0])
}

func isGitVersionArg(args []string) bool {
	return len(args) == 1 && args[0] == "--git-version"
}
