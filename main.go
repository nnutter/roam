// Command roam manages dotfiles with a bare Git repository.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"charm.land/fang/v2"
)

func main() {
	app, err := newApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = fang.Execute(
		context.Background(),
		app.command(),
		fang.WithoutCompletions(),
		fang.WithoutVersion(),
		fang.WithErrorHandler(errorHandler),
		fang.WithNotifySignal(os.Interrupt),
	)
	if err != nil {
		os.Exit(exitCode(err))
	}
}

func errorHandler(w io.Writer, styles fang.Styles, err error) {
	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		return
	}
	fang.DefaultErrorHandler(w, styles, err)
}

func exitCode(err error) int {
	if ee, ok := errors.AsType[*exec.ExitError](err); ok {
		return ee.ExitCode()
	}
	return 1
}
