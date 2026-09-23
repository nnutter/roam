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

func defaultZshOut(a *app) string {
	return filepath.Join(a.home, ".local/share/zsh/site-functions")
}

func expandOut(a *app, out string) string {
	if out == "~" {
		return a.home
	}
	if rest, ok := strings.CutPrefix(out, "~/"); ok {
		return filepath.Join(a.home, rest)
	}
	return out
}

func completionCommand(a *app) *cobra.Command {
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
	out := defaultZshOut(a)
	bash := &cobra.Command{
		Use:   "bash",
		Short: "Generate the autocompletion script for bash",
		Long: `Generate the autocompletion script for the bash shell.

This aliases roam completion to git completion, since roam wraps git.
It needs git's bash completion, usually provided by the 'bash-completion'
package. If it is not installed already, you can install it via your OS's
package manager.

To load completions in your current shell session:

	source <(roam completion bash)

To load completions for every new session, execute once:

#### Linux:

	roam completion bash > /etc/bash_completion.d/roam

#### macOS:

	roam completion bash > $(brew --prefix)/etc/bash_completion.d/roam

You will need to start a new shell for this setup to take effect.`,
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
			return completionZsh(force, expandOut(a, out))
		},
	}
	zsh.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing generated files")
	zsh.Flags().StringVarP(&out, "out", "o", out, "Output directory for generated files (~/.local/share/zsh/site-functions)")
	cmd.AddCommand(bash, zsh)
	return cmd
}

func completionZsh(force bool, out string) error {
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
