package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupDefaultFlags(t *testing.T) {
	app, logPath := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"setup"}))
	got := gitCalls(t, logPath)
	home := app.home
	gitDir := filepath.Join(home, ".dotfiles.git")
	want := [][]string{
		{"--git-dir", gitDir, "--work-tree", home, "init"},
		{"--git-dir", gitDir, "--work-tree", home, "config", "status.showUntrackedFiles", "no"},
		{"--git-dir", gitDir, "--work-tree", home, "remote", "add", "-f", "origin", "git@github.com:tester/dotfiles.git"},
		{"--git-dir", gitDir, "--work-tree", home, "checkout", "master"},
	}
	require.Equal(t, want, got)
}

func TestSetupCustomFlags(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		repo   string
		branch string
	}{
		{
			name:   "short flags",
			args:   []string{"setup", "-r", "alice/dots", "-b", "main"},
			repo:   "alice/dots",
			branch: "main",
		},
		{
			name:   "long flags",
			args:   []string{"setup", "--repo", "bob/dots", "--branch", "trunk"},
			repo:   "bob/dots",
			branch: "trunk",
		},
		{
			name:   "equals flags",
			args:   []string{"setup", "--repo=carol/dots", "--branch=dev"},
			repo:   "carol/dots",
			branch: "dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, logPath := testApp(t)
			require.NoError(t, app.run(t.Context(), tt.args))
			got := gitCalls(t, logPath)
			origin := "git@github.com:" + tt.repo + ".git"
			last := got[len(got)-2]
			require.Equal(t, origin, last[len(last)-1])
			last = got[len(got)-1]
			require.Equal(t, tt.branch, last[len(last)-1])
		})
	}
}

func TestSetupHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	require.NoError(t, app.run(t.Context(), []string{"setup", "-h"}))
	require.Contains(t, stdout.String(), "roam setup")
}

func TestSetupUnknownOption(t *testing.T) {
	app, logPath := testApp(t)
	require.Error(t, app.run(t.Context(), []string{"setup", "--bogus"}))
	_, statErr := os.Stat(logPath)
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestUpdate(t *testing.T) {
	app, logPath := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"update"}))
	got := gitCalls(t, logPath)
	home := app.home
	gitDir := filepath.Join(home, ".dotfiles.git")
	want := [][]string{
		{"--git-dir", gitDir, "--work-tree", home, "fetch"},
		{"--git-dir", gitDir, "--work-tree", home, "rebase", "--autostash"},
	}
	require.Equal(t, want, got)
}

func TestSync(t *testing.T) {
	app, logPath := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"sync"}))
	got := gitCalls(t, logPath)
	home := app.home
	gitDir := filepath.Join(home, ".dotfiles.git")
	want := [][]string{
		{"--git-dir", gitDir, "--work-tree", home, "fetch"},
		{"--git-dir", gitDir, "--work-tree", home, "rebase", "--autostash"},
		{"--git-dir", gitDir, "--work-tree", home, "push"},
	}
	require.Equal(t, want, got)
}

func TestGitPassthrough(t *testing.T) {
	app, logPath := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"status", "--short"}))
	got := gitCalls(t, logPath)
	want := [][]string{
		{"--git-dir", filepath.Join(app.home, ".dotfiles.git"), "--work-tree", app.home, "status", "--short"},
	}
	require.Equal(t, want, got)
}

func TestGitHelpPassthrough(t *testing.T) {
	app, logPath := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"help", "status"}))
	got := gitCalls(t, logPath)
	want := [][]string{
		{"--git-dir", filepath.Join(app.home, ".dotfiles.git"), "--work-tree", app.home, "help", "status"},
	}
	require.Equal(t, want, got)
}

func TestGitFailureStopsChain(t *testing.T) {
	app, logPath := testApp(t)
	failGit(t, app, "fetch")
	err := app.run(t.Context(), []string{"sync"})
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	got := gitCalls(t, logPath)
	require.Len(t, got, 1)
	require.Equal(t, "fetch", got[0][len(got[0])-1])
}

func TestExitCode(t *testing.T) {
	require.Equal(t, 1, exitCode(errors.New("boom")))
}

func testApp(t *testing.T) (*app, string) {
	t.Helper()
	home := t.TempDir()
	bin := t.TempDir()
	logPath := filepath.Join(bin, "git.log")
	failPath := filepath.Join(bin, "fail")
	gitPath := filepath.Join(bin, "git")
	script := "#!/bin/sh\n" +
		"log=" + shellQuote(logPath) + "\n" +
		"failfile=" + shellQuote(failPath) + "\n" +
		`last=
for arg in "$@"; do
  printf '%s\n' "$arg" >>"$log"
  last=$arg
done
printf '\n' >>"$log"
if [ -f "$failfile" ]; then
  read fail <"$failfile"
  if [ "$last" = "$fail" ]; then
    exit 1
  fi
fi
`
	require.NoError(t, os.WriteFile(gitPath, []byte(script), 0o755))
	return &app{
		home:   home,
		git:    gitPath,
		user:   "tester",
		stdin:  bytes.NewReader(nil),
		stdout: io.Discard,
		stderr: io.Discard,
	}, logPath
}

func failGit(t *testing.T, a *app, subcmd string) {
	t.Helper()
	failPath := filepath.Join(filepath.Dir(a.git), "fail")
	require.NoError(t, os.WriteFile(failPath, []byte(subcmd+"\n"), 0o644))
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func gitCalls(t *testing.T, path string) [][]string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var calls [][]string
	var cur []string
	for line := range strings.SplitSeq(string(data), "\n") {
		if line == "" {
			if cur != nil {
				calls = append(calls, cur)
				cur = nil
			}
			continue
		}
		cur = append(cur, line)
	}
	return calls
}
