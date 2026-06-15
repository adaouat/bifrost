//go:build integration

package cmd_test

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
	forgeui "github.com/adaouat/forge/ui"
)

// sshTestConfig builds an in-memory bifrost config for the SSH container test.
// The SSH connection parameters live in the server map; the env/app structure
// mirrors bifrost-deploy-int-test.yml.
func sshTestConfig(host, port, keyFile string) *config.Config {
	p, _ := strconv.Atoi(port)
	return &config.Config{
		Strategy: "atomic",
		Servers: map[string]config.Server{
			"ssh-test": {Host: host, Port: p, User: "deploy", KeyFile: keyFile, StagingDir: "/tmp"},
		},
		Paths:    config.Paths{ReleasesRoot: "/var/releases", SharedRoot: "/var/shared"},
		Settings: config.Settings{ReleasesToKeep: 5},
		Environments: map[string]config.Environment{
			"test": {
				Servers: []string{"ssh-test"},
				Applications: map[string]config.Application{
					"app": {Paths: config.Paths{Shared: config.SharedPaths{
						Directories: []string{"var/log"},
						Files:       []string{".env"},
					}}},
				},
			},
		},
	}
}

func newSSHTestServer(ctx context.Context, t *testing.T) (*testutil.SSHContainer, config.ResolvedServer, *config.Config) {
	t.Helper()
	sshC := testutil.NewSSHContainer(ctx, t, "../../testdata/ssh/client_rsa.pub", "../../testdata/ssh/host_rsa")
	knownHostsPath, err := sshC.WriteKnownHosts("../../testdata/ssh/host_rsa.pub")
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	absKeyFile, err := filepath.Abs("../../testdata/ssh/client_rsa")
	require.NoError(t, err)

	cfg := sshTestConfig(sshC.Host, sshC.Port, absKeyFile)
	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)
	return sshC, merged.Servers[0], cfg
}

func TestDeployCmd_ClientMode_E2E(t *testing.T) {
	ctx := context.Background()
	sshC, srv, cfg := newSSHTestServer(ctx, t)

	out := testutil.DeployOverSSH(ctx, t, srv, cfg, "test", "app", "../../testdata/release.tar.gz", "test-r1", forgeui.Plain)
	t.Logf("deploy output:\n%s", out)

	res, err := sshC.Exec(ctx, []string{"test", "-d", "/var/releases/test-r1"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "release directory must exist")

	res, err = sshC.Exec(ctx, []string{"test", "-f", "/var/releases/test-r1/public/index.html"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "extracted file must exist in release dir")

	res, err = sshC.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/test-r1", res.Output, "current symlink target")

	res, err = sshC.Exec(ctx, []string{"readlink", "/var/releases/test-r1/var/log"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/var/log", res.Output, "shared dir symlink")

	res, err = sshC.Exec(ctx, []string{"readlink", "/var/releases/test-r1/.env"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/.env", res.Output, "shared file symlink")
}

func TestDeployCmd_HookOutputVisibleOverSSH(t *testing.T) {
	ctx := context.Background()
	_, srv, cfg := newSSHTestServer(ctx, t)

	env := cfg.Environments["test"]
	app := env.Applications["app"]
	app.Hooks.PostExtract = []config.HookEntry{{Cmd: "echo HOOK-OUTPUT-MARKER"}}
	env.Applications["app"] = app
	cfg.Environments["test"] = env

	out := testutil.DeployOverSSH(ctx, t, srv, cfg, "test", "app", "../../testdata/release.tar.gz", "hook-r1", forgeui.Plain)
	t.Logf("deploy output:\n%s", out)

	assert.Contains(t, out, "HOOK-OUTPUT-MARKER", "hook stdout must be visible over SSH")
}
