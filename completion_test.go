package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletionBash(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	require.NoError(t, app.run(t.Context(), []string{"completion", "bash"}))
	require.Equal(t, "complete -F _git roam\n", stdout.String())
}

func TestCompletionZsh(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	require.NoError(t, app.run(t.Context(), []string{"completion", "zsh", "--out", out}))
	got, err := os.ReadFile(filepath.Join(out, "_roam"))
	require.NoError(t, err)
	require.Equal(t, string(zshCompletion), string(got))
}

func TestCompletionZshDefaultOut(t *testing.T) {
	app, _ := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"completion", "zsh"}))
	path := filepath.Join(app.home, ".local/share/zsh/site-functions", "_roam")
	require.FileExists(t, path)
}

func TestCompletionZshExistsWithoutForce(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	path := filepath.Join(out, "_roam")
	require.NoError(t, os.MkdirAll(out, 0o755))
	require.NoError(t, os.WriteFile(path, []byte("old\n"), 0o644))
	err := app.run(t.Context(), []string{"completion", "zsh", "-o", out})
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "old\n", string(got))
}

func TestCompletionZshForce(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	path := filepath.Join(out, "_roam")
	require.NoError(t, os.MkdirAll(out, 0o755))
	require.NoError(t, os.WriteFile(path, []byte("old\n"), 0o644))
	require.NoError(t, app.run(t.Context(), []string{"completion", "zsh", "--force", "--out=" + out}))
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(zshCompletion), string(got))
}

func TestCompletionZshHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	require.NoError(t, app.run(t.Context(), []string{"completion", "zsh", "--help"}))
	help := stdout.String()
	for _, want := range []string{"--force", "--out", "~/.local/share/zsh/site-functions"} {
		require.Contains(t, help, want)
	}
}

func TestCompletionHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	require.NoError(t, app.run(t.Context(), []string{"completion", "-h"}))
	help := stdout.String()
	require.Contains(t, help, "bash")
	require.Contains(t, help, "zsh")
}

func TestCompletionZshTildeOut(t *testing.T) {
	app, _ := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"completion", "zsh", "-o", "~/custom"}))
	path := filepath.Join(app.home, "custom", "_roam")
	require.FileExists(t, path)
}

func TestCompletionUnknown(t *testing.T) {
	app, _ := testApp(t)
	require.Error(t, app.run(t.Context(), []string{"completion", "fish"}))
}
