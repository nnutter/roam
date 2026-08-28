package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateZsh(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	if err := app.run(t.Context(), []string{"generate", "zsh", "--out", out}); err != nil {
		t.Fatalf("generate zsh: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(out, "_roam"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(zshCompletion) {
		t.Fatalf("generated completion does not match embedded _roam")
	}
}

func TestGenerateZshDefaultOut(t *testing.T) {
	app, _ := testApp(t)
	if err := app.run(t.Context(), []string{"generate", "zsh"}); err != nil {
		t.Fatalf("generate zsh: %v", err)
	}
	path := filepath.Join(app.home, ".local/share/zsh/site-functions", "_roam")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("default output missing: %v", err)
	}
}

func TestGenerateZshExistsWithoutForce(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	path := filepath.Join(out, "_roam")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := app.run(t.Context(), []string{"generate", "zsh", "-o", out})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %q, want already exists", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old\n" {
		t.Fatalf("file was overwritten without --force")
	}
}

func TestGenerateZshForce(t *testing.T) {
	app, _ := testApp(t)
	out := filepath.Join(app.home, "completions")
	path := filepath.Join(out, "_roam")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := app.run(t.Context(), []string{"generate", "zsh", "--force", "--out=" + out}); err != nil {
		t.Fatalf("generate zsh --force: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(zshCompletion) {
		t.Fatalf("generated completion does not match embedded _roam")
	}
}

func TestGenerateZshHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	if err := app.run(t.Context(), []string{"generate", "zsh", "--help"}); err != nil {
		t.Fatalf("generate zsh --help: %v", err)
	}
	help := stdout.String()
	for _, want := range []string{"--force", "--out", "~/.local/share/zsh/site-functions"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help = %q, want it to contain %q", help, want)
		}
	}
}

func TestGenerateHelp(t *testing.T) {
	app, _ := testApp(t)
	var stdout bytes.Buffer
	app.stdout = &stdout
	if err := app.run(t.Context(), []string{"generate", "-h"}); err != nil {
		t.Fatalf("generate -h: %v", err)
	}
	if !strings.Contains(stdout.String(), "zsh") {
		t.Fatalf("help = %q, want zsh", stdout.String())
	}
}

func TestGenerateZshTildeOut(t *testing.T) {
	app, _ := testApp(t)
	if err := app.run(t.Context(), []string{"generate", "zsh", "-o", "~/custom"}); err != nil {
		t.Fatalf("generate zsh: %v", err)
	}
	path := filepath.Join(app.home, "custom", "_roam")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("tilde output missing: %v", err)
	}
}

func TestGenerateUnknown(t *testing.T) {
	app, _ := testApp(t)
	err := app.run(t.Context(), []string{"generate", "fish"})
	if err == nil {
		t.Fatal("expected an error")
	}
}
