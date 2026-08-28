package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSetupDefaultFlags(t *testing.T) {
	app, logPath := testApp(t)
	if err := app.run(t.Context(), []string{"setup"}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	got := gitCalls(t, logPath)
	home := app.home
	gitDir := filepath.Join(home, ".dotfiles.git")
	want := [][]string{
		{"--git-dir", gitDir, "--work-tree", home, "init"},
		{"--git-dir", gitDir, "--work-tree", home, "config", "status.showUntrackedFiles", "no"},
		{"--git-dir", gitDir, "--work-tree", home, "remote", "add", "-f", "origin", "git@github.com:tester/dotfiles.git"},
		{"--git-dir", gitDir, "--work-tree", home, "checkout", "master"},
	}
	assertCalls(t, got, want)
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
			if err := app.run(t.Context(), tt.args); err != nil {
				t.Fatalf("setup: %v", err)
			}
			got := gitCalls(t, logPath)
			origin := "git@github.com:" + tt.repo + ".git"
			if last := got[len(got)-2]; last[len(last)-1] != origin {
				t.Fatalf("origin = %q, want %q", last[len(last)-1], origin)
			}
			if last := got[len(got)-1]; last[len(last)-1] != tt.branch {
				t.Fatalf("branch = %q, want %q", last[len(last)-1], tt.branch)
			}
		})
	}
}

func TestSetupHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	if err := app.run(t.Context(), []string{"setup", "-h"}); err != nil {
		t.Fatalf("setup -h: %v", err)
	}
	if !strings.Contains(stdout.String(), "roam setup") {
		t.Fatalf("help = %q, want it to mention roam setup", stdout.String())
	}
}

func TestSetupUnknownOption(t *testing.T) {
	app, logPath := testApp(t)
	err := app.run(t.Context(), []string{"setup", "--bogus"})
	if err == nil {
		t.Fatal("expected an error")
	}
	if _, statErr := os.Stat(logPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("git should not have been invoked: %v", statErr)
	}
}

func TestUpdate(t *testing.T) {
	app, logPath := testApp(t)
	if err := app.run(t.Context(), []string{"update"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	got := gitCalls(t, logPath)
	home := app.home
	gitDir := filepath.Join(home, ".dotfiles.git")
	want := [][]string{
		{"--git-dir", gitDir, "--work-tree", home, "fetch"},
		{"--git-dir", gitDir, "--work-tree", home, "rebase", "--autostash"},
	}
	assertCalls(t, got, want)
}

func TestSync(t *testing.T) {
	app, logPath := testApp(t)
	if err := app.run(t.Context(), []string{"sync"}); err != nil {
		t.Fatalf("sync: %v", err)
	}
	got := gitCalls(t, logPath)
	home := app.home
	gitDir := filepath.Join(home, ".dotfiles.git")
	want := [][]string{
		{"--git-dir", gitDir, "--work-tree", home, "fetch"},
		{"--git-dir", gitDir, "--work-tree", home, "rebase", "--autostash"},
		{"--git-dir", gitDir, "--work-tree", home, "push"},
	}
	assertCalls(t, got, want)
}

func TestGitPassthrough(t *testing.T) {
	app, logPath := testApp(t)
	if err := app.run(t.Context(), []string{"status", "--short"}); err != nil {
		t.Fatalf("status: %v", err)
	}
	got := gitCalls(t, logPath)
	want := [][]string{
		{"--git-dir", filepath.Join(app.home, ".dotfiles.git"), "--work-tree", app.home, "status", "--short"},
	}
	assertCalls(t, got, want)
}

func TestGitHelpPassthrough(t *testing.T) {
	app, logPath := testApp(t)
	if err := app.run(t.Context(), []string{"help", "status"}); err != nil {
		t.Fatalf("help status: %v", err)
	}
	got := gitCalls(t, logPath)
	want := [][]string{
		{"--git-dir", filepath.Join(app.home, ".dotfiles.git"), "--work-tree", app.home, "help", "status"},
	}
	assertCalls(t, got, want)
}

func TestGitFailureStopsChain(t *testing.T) {
	app, logPath := testApp(t)
	failGit(t, app, "fetch")
	err := app.run(t.Context(), []string{"sync"})
	if _, ok := errors.AsType[*exec.ExitError](err); !ok {
		t.Fatalf("error = %v, want exec.ExitError", err)
	}
	got := gitCalls(t, logPath)
	if len(got) != 1 || got[0][len(got[0])-1] != "fetch" {
		t.Fatalf("calls = %v, want only fetch", got)
	}
}

func TestExitCode(t *testing.T) {
	if code := exitCode(errors.New("boom")); code != 1 {
		t.Fatalf("exitCode(plain) = %d", code)
	}
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
	if err := os.WriteFile(gitPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
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
	if err := os.WriteFile(failPath, []byte(subcmd+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func gitCalls(t *testing.T, path string) [][]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
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

func assertCalls(t *testing.T, got, want [][]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d git calls, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if !slices.Equal(got[i], want[i]) {
			t.Fatalf("call %d: got %v, want %v", i, got[i], want[i])
		}
	}
}
