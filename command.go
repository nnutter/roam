package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func (a *app) command() *cobra.Command {
	root := &cobra.Command{
		Use:                "roam",
		Short:              "Manage dotfiles with a bare Git repository",
		Long:               "Use roam like git, with extra commands for a home-directory work tree.",
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || isHelpArg(args) {
				return cmd.Help()
			}
			return a.runGit(cmd.Context(), args...)
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})
	root.AddCommand(
		a.setupCommand(),
		a.updateCommand(),
		a.syncCommand(),
		a.generateCommand(),
	)
	return root
}

func isHelpArg(args []string) bool {
	return len(args) == 1 && (args[0] == "-h" || args[0] == "--help")
}

func (a *app) setupCommand() *cobra.Command {
	repo := a.defaultRepo()
	branch := defaultBranch
	cmd := &cobra.Command{
		Use:           "setup",
		Short:         "Initialize the dotfiles repository",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.setup(cmd.Context(), repo, branch)
		},
	}
	cmd.Flags().StringVarP(&repo, "repo", "r", repo, "repo name")
	cmd.Flags().StringVarP(&branch, "branch", "b", branch, "branch name")
	return cmd
}

func (a *app) updateCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "update",
		Short:         "Fetch and rebase with autostash",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.update(cmd.Context())
		},
	}
}

func (a *app) syncCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "sync",
		Short:         "Update and push",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.sync(cmd.Context())
		},
	}
}

func (a *app) generateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "generate",
		Short:         "Generate shell completion files",
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
	zsh := &cobra.Command{
		Use:           "zsh",
		Short:         "Generate zsh completion for roam",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.generateZsh(force, a.expandOut(out))
		},
	}
	zsh.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing generated files")
	zsh.Flags().StringVarP(&out, "out", "o", out, "Output directory for generated files (~/.local/share/zsh/site-functions)")
	cmd.AddCommand(zsh)
	return cmd
}

func (a *app) run(ctx context.Context, args []string) error {
	cmd := a.command()
	cmd.SetArgs(args)
	cmd.SetIn(a.stdin)
	cmd.SetOut(a.stdout)
	cmd.SetErr(a.stderr)
	return cmd.ExecuteContext(ctx)
}
