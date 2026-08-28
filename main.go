// Command roam manages dotfiles with a bare Git repository.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	app, err := newApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := app.run(context.Background(), os.Args[1:]); err != nil {
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	if ee, ok := errors.AsType[*exec.ExitError](err); ok {
		return ee.ExitCode()
	}
	if ee, ok := errors.AsType[*exitError](err); ok {
		if ee.msg != "" {
			fmt.Fprintln(os.Stderr, ee.msg)
		}
		return ee.code
	}
	fmt.Fprintln(os.Stderr, err)
	return 1
}
