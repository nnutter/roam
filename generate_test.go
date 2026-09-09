package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateZsh(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	require.NoError(t, app.run(t.Context(), []string{"generate", "zsh", "--out", out}))
	got, err := os.ReadFile(filepath.Join(out, "_roam"))
	require.NoError(t, err)
	require.Equal(t, string(zshCompletion), string(got))
}

func TestGenerateZshDefaultOut(t *testing.T) {
	app, _ := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"generate", "zsh"}))
	path := filepath.Join(app.home, ".local/share/zsh/site-functions", "_roam")
	require.FileExists(t, path)
}

func TestGenerateZshExistsWithoutForce(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	path := filepath.Join(out, "_roam")
	require.NoError(t, os.MkdirAll(out, 0o755))
	require.NoError(t, os.WriteFile(path, []byte("old\n"), 0o644))
	err := app.run(t.Context(), []string{"generate", "zsh", "-o", out})
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "old\n", string(got))
}

func TestGenerateZshForce(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	path := filepath.Join(out, "_roam")
	require.NoError(t, os.MkdirAll(out, 0o755))
	require.NoError(t, os.WriteFile(path, []byte("old\n"), 0o644))
	require.NoError(t, app.run(t.Context(), []string{"generate", "zsh", "--force", "--out=" + out}))
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(zshCompletion), string(got))
}

func TestGenerateZshHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	require.NoError(t, app.run(t.Context(), []string{"generate", "zsh", "--help"}))
	help := stdout.String()
	for _, want := range []string{"--force", "--out", "~/.local/share/zsh/site-functions"} {
		require.Contains(t, help, want)
	}
}

func TestGenerateHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	require.NoError(t, app.run(t.Context(), []string{"generate", "-h"}))
	require.Contains(t, stdout.String(), "zsh")
}

func TestGenerateZshTildeOut(t *testing.T) {
	app, _ := testApp(t)
	require.NoError(t, app.run(t.Context(), []string{"generate", "zsh", "-o", "~/custom"}))
	path := filepath.Join(app.home, "custom", "_roam")
	require.FileExists(t, path)
}

func TestGenerateUnknown(t *testing.T) {
	app, _ := testApp(t)
	require.Error(t, app.run(t.Context(), []string{"generate", "fish"}))
}
