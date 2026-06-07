//go:build integration

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
)

// multiServerTestConfig builds an in-memory bifrost config referencing the
// given named servers, mirroring sshTestConfig in client_deploy_integration_test.go
// but for a fleet of more than one server.
func multiServerTestConfig(servers map[string]config.ServerConfig, order []string) *config.Config {
	return &config.Config{
		Strategy: "atomic",
		Servers:  servers,
		Paths: config.Paths{
			ReleasesRoot: "/var/releases",
			SharedRoot:   "/var/shared",
		},
		Settings: config.Settings{ReleasesToKeep: 5},
		Environments: map[string]config.Environment{
			"test": {
				Servers: order,
				Applications: map[string]config.Application{
					"app": {
						Paths: config.Paths{
							Shared: config.SharedPaths{
								Directories: []string{"var/log"},
								Files:       []string{".env"},
							},
						},
					},
				},
			},
		},
	}
}

func TestRunClientDeploy_MultiServer_E2E(t *testing.T) {
	ctx := context.Background()

	keyFile := "../../testdata/ssh/client_rsa"
	hostPubKey := "../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../testdata/ssh/host_rsa"

	sshA := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)
	sshB := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)
	containers := map[string]*testutil.SSHContainer{"web-01": sshA, "web-02": sshB}

	knownHostsPath, err := testutil.WriteKnownHosts([]*testutil.SSHContainer{sshA, sshB}, []string{hostPubKey, hostPubKey})
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	absKeyFile, err := filepath.Abs(keyFile)
	require.NoError(t, err)

	servers := make(map[string]config.ServerConfig, len(containers))
	for name, c := range containers {
		port, err := strconv.Atoi(c.Port)
		require.NoError(t, err)
		servers[name] = config.ServerConfig{
			Host:       c.Host,
			Port:       port,
			User:       "deploy",
			KeyFile:    absKeyFile,
			StagingDir: "/tmp",
		}

		// Ensure the release/shared roots exist with the right ownership before deploying.
		for _, dir := range []string{"/var/releases", "/var/shared"} {
			res, err := c.Exec(ctx, []string{"sh", "-c", fmt.Sprintf("mkdir -p %s && chown deploy:deploy %s", dir, dir)})
			require.NoError(t, err)
			require.Equal(t, 0, res.ExitCode)
		}
	}

	cfg := multiServerTestConfig(servers, []string{"web-01", "web-02"})

	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 2)

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "plain", "")
	var out bytes.Buffer
	root.SetOut(&out)

	err = runClientDeploy(root, "dev", merged, cfg, "test", "app", "../../testdata/release.tar.gz", "multi-r1", testutil.BinaryPath)
	t.Logf("deploy output:\n%s", out.String())
	require.NoError(t, err)

	// Both servers must get a header line introducing their output section,
	// and both must reach the same release name (spec 08 — sequential loop).
	assert.Contains(t, out.String(), "web-01")
	assert.Contains(t, out.String(), "web-02")

	for name, c := range containers {
		res, err := c.Exec(ctx, []string{"test", "-d", "/var/releases/multi-r1"})
		require.NoError(t, err)
		assert.Equal(t, 0, res.ExitCode, "%s: release directory must exist", name)

		res, err = c.Exec(ctx, []string{"test", "-f", "/var/releases/multi-r1/public/index.html"})
		require.NoError(t, err)
		assert.Equal(t, 0, res.ExitCode, "%s: extracted file must exist in release dir", name)

		res, err = c.Exec(ctx, []string{"readlink", "/var/releases/current"})
		require.NoError(t, err)
		assert.Equal(t, "/var/releases/multi-r1", res.Output, "%s: current symlink target", name)

		res, err = c.Exec(ctx, []string{"readlink", "/var/releases/multi-r1/var/log"})
		require.NoError(t, err)
		assert.Equal(t, "/var/shared/var/log", res.Output, "%s: shared dir symlink", name)

		res, err = c.Exec(ctx, []string{"readlink", "/var/releases/multi-r1/.env"})
		require.NoError(t, err)
		assert.Equal(t, "/var/shared/.env", res.Output, "%s: shared file symlink", name)
	}
}
