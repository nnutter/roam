package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed _roam
var zshCompletion []byte

func (a *app) defaultZshOut() string {
	return filepath.Join(a.home, ".local/share/zsh/site-functions")
}

func (a *app) generate(args []string) error {
	if len(args) == 0 {
		a.printGenerateHelp()
		return &exitError{code: gitUsageExit}
	}
	switch args[0] {
	case "-h", "--help":
		a.printGenerateHelp()
		return &exitError{code: gitUsageExit}
	case "zsh":
		return a.generateZsh(args[1:])
	default:
		return &exitError{code: 1, msg: "unknown generator: " + args[0]}
	}
}

func (a *app) generateZsh(args []string) error {
	force, out, err := a.parseGenerateZsh(args)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	path := filepath.Join(out, "_roam")
	_, err = os.Stat(path)
	switch {
	case err == nil:
		if !force {
			return &exitError{code: 1, msg: path + " already exists; use --force to overwrite"}
		}
	case !errors.Is(err, os.ErrNotExist):
		return err
	}
	return os.WriteFile(path, zshCompletion, 0o644)
}

func (a *app) parseGenerateZsh(args []string) (force bool, out string, err error) {
	out = a.defaultZshOut()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--":
			if i+1 < len(args) {
				return false, "", &exitError{code: 1, msg: "unexpected argument: " + args[i+1]}
			}
			return force, out, nil
		case "-h", "--help":
			a.printGenerateZshHelp()
			return false, "", &exitError{code: gitUsageExit}
		case "-f", "--force":
			force = true
		case "-o", "--out":
			val, err := optionValue(args, &i, arg)
			if err != nil {
				return false, "", err
			}
			out = a.expandOut(val)
		default:
			if val, ok := strings.CutPrefix(arg, "--out="); ok {
				out = a.expandOut(val)
				continue
			}
			return false, "", &exitError{code: 1, msg: "unknown option: " + arg}
		}
	}
	return force, out, nil
}

func (a *app) expandOut(out string) string {
	if out == "~" {
		return a.home
	}
	if rest, ok := strings.CutPrefix(out, "~/"); ok {
		return filepath.Join(a.home, rest)
	}
	return out
}

func (a *app) printGenerateHelp() {
	fmt.Fprint(a.stdout, `roam generate zsh

Generate shell completion files.
`)
}

func (a *app) printGenerateZshHelp() {
	fmt.Fprint(a.stdout, `roam generate zsh

Generate zsh completion for roam.

  FLAGS

    -f --force  Overwrite existing generated files
    -o --out    Output directory for generated files (~/.local/share/zsh/site-functions)
`)
}
