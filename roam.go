package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const defaultBranch = "master"

type app struct {
	home   string
	git    string
	user   string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func newApp() (*app, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	username := os.Getenv("USER")
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		}
	}
	return &app{
		home:   home,
		git:    git,
		user:   username,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}, nil
}

func (a *app) gitDir() string {
	return filepath.Join(a.home, ".dotfiles.git")
}

func (a *app) defaultRepo() string {
	return a.user + "/dotfiles"
}

func (a *app) runGit(ctx context.Context, args ...string) error {
	cmdArgs := make([]string, 0, 4+len(args))
	cmdArgs = append(cmdArgs, "--git-dir", a.gitDir(), "--work-tree", a.home)
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.CommandContext(ctx, a.git, cmdArgs...)
	cmd.Stdin = a.stdin
	cmd.Stdout = a.stdout
	cmd.Stderr = a.stderr
	return cmd.Run()
}

func (a *app) setup(ctx context.Context, repo, branch string) error {
	if err := a.runGit(ctx, "init"); err != nil {
		return err
	}
	if err := a.runGit(ctx, "config", "status.showUntrackedFiles", "no"); err != nil {
		return err
	}
	origin := "git@github.com:" + repo + ".git"
	if err := a.runGit(ctx, "remote", "add", "-f", "origin", origin); err != nil {
		return err
	}
	return a.runGit(ctx, "checkout", branch)
}

func (a *app) update(ctx context.Context) error {
	if err := a.runGit(ctx, "fetch"); err != nil {
		return err
	}
	return a.runGit(ctx, "rebase", "--autostash")
}

func (a *app) sync(ctx context.Context) error {
	if err := a.update(ctx); err != nil {
		return err
	}
	return a.runGit(ctx, "push")
}

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
