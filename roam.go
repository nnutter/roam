package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	defaultBranch = "master"
	gitUsageExit  = 129
)

type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return fmt.Sprintf("exit status %d", e.code)
}

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

func (a *app) run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return a.runGit(ctx, args...)
	}
	switch args[0] {
	case "setup":
		return a.setup(ctx, args[1:])
	case "update":
		return a.update(ctx)
	case "sync":
		return a.sync(ctx)
	default:
		return a.runGit(ctx, args...)
	}
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

func (a *app) setup(ctx context.Context, args []string) error {
	repo, branch, err := a.parseSetup(args)
	if err != nil {
		return err
	}
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

func (a *app) parseSetup(args []string) (repo, branch string, err error) {
	repo = a.defaultRepo()
	branch = defaultBranch
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--":
			return repo, branch, nil
		case "-h", "--help":
			a.printSetupHelp()
			return "", "", &exitError{code: gitUsageExit}
		case "-r", "--repo":
			val, err := optionValue(args, &i, arg)
			if err != nil {
				return "", "", err
			}
			repo = val
		case "-b", "--branch":
			val, err := optionValue(args, &i, arg)
			if err != nil {
				return "", "", err
			}
			branch = val
		default:
			if val, ok := strings.CutPrefix(arg, "--repo="); ok {
				repo = val
				continue
			}
			if val, ok := strings.CutPrefix(arg, "--branch="); ok {
				branch = val
				continue
			}
			return "", "", &exitError{code: 1, msg: "unknown option: " + arg}
		}
	}
	return repo, branch, nil
}

func optionValue(args []string, i *int, opt string) (string, error) {
	if *i+1 >= len(args) {
		return "", &exitError{code: 1, msg: "option " + opt + " requires a value"}
	}
	*i++
	return args[*i], nil
}

func (a *app) printSetupHelp() {
	fmt.Fprintf(a.stdout, `roam setup [-r <repo>] [-b <branch>]

Otherwise use roam like git.

    -h, --help     show the help
    -r, --repo     repo name (default: %s)
    -b, --branch   branch name (default: %s)
`, a.defaultRepo(), defaultBranch)
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
