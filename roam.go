package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
