package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupEndToEnd(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found in PATH")
	}
	home := t.TempDir()
	origin := t.TempDir()
	config := fmt.Sprintf(`[user]
	name = Tester
	email = tester@example.com
[url %q]
	insteadOf = git@github.com:tester/dotfiles.git
`, origin)
	configPath := filepath.Join(t.TempDir(), "gitconfig")
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0o644))
	unsetGitEnv(t)
	t.Setenv("GIT_CONFIG_GLOBAL", configPath)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command(git, args...)
		cmd.Dir = origin
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	runGit("init", "--initial-branch", "master")
	require.NoError(t, os.WriteFile(filepath.Join(origin, ".testrc"), []byte("hello\n"), 0o644))
	runGit("add", ".testrc")
	runGit("commit", "-m", "Add .testrc")

	app := &app{
		home:   home,
		git:    git,
		user:   "tester",
		stdin:  bytes.NewReader(nil),
		stdout: io.Discard,
		stderr: io.Discard,
	}
	require.NoError(t, app.run(t.Context(), []string{"setup"}))

	got, err := os.ReadFile(filepath.Join(home, ".testrc"))
	require.NoError(t, err)
	require.Equal(t, "hello\n", string(got))

	var stdout bytes.Buffer
	app.stdout = &stdout
	require.NoError(t, app.run(t.Context(), []string{"status", "--short"}))
	require.Empty(t, stdout.String())
}

// unsetGitEnv removes all GIT_* environment variables so git invocations in
// the test are not polluted by the developer's environment, for example when
// the tests run from a hook or during an interactive rebase, which exports
// variables such as GIT_DIR and GIT_INDEX_FILE.
func unsetGitEnv(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		name, _, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(name, "GIT_") {
			continue
		}
		// t.Setenv registers restoration on cleanup and forbids t.Parallel.
		t.Setenv(name, "")
		require.NoError(t, os.Unsetenv(name))
	}
}
