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

func (a *app) generateZsh(force bool, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	path := filepath.Join(out, "_roam")
	_, err := os.Stat(path)
	switch {
	case err == nil:
		if !force {
			return fmt.Errorf("%s already exists; use --force to overwrite", path)
		}
	case !errors.Is(err, os.ErrNotExist):
		return err
	}
	return os.WriteFile(path, zshCompletion, 0o644)
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
