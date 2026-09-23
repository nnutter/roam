package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed _roam
var zshCompletion []byte

func (a *app) defaultZshOut() string {
	return filepath.Join(a.home, ".local/share/zsh/site-functions")
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

func (a *app) completionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "completion",
		Short:         "Generate the autocompletion script for the specified shell",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			return fmt.Errorf("unknown generator: %s", args[0])
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	var force bool
	out := a.defaultZshOut()
	bash := &cobra.Command{
		Use:           "bash",
		Short:         "Generate the autocompletion script for bash",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "complete -F _git roam")
			return err
		},
	}
	zsh := &cobra.Command{
		Use:           "zsh",
		Short:         "Generate the autocompletion script for zsh",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.completionZsh(force, a.expandOut(out))
		},
	}
	zsh.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing generated files")
	zsh.Flags().StringVarP(&out, "out", "o", out, "Output directory for generated files (~/.local/share/zsh/site-functions)")
	cmd.AddCommand(bash, zsh)
	return cmd
}

func (a *app) completionZsh(force bool, out string) error {
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
