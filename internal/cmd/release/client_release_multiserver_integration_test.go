//go:build integration

package release

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
)

func TestReleaseCommands_ClientMode_MultiServer_E2E(t *testing.T) {
	ctx := context.Background()

	keyFile := "../../../testdata/ssh/client_rsa"
	hostPubKey := "../../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../../testdata/ssh/host_rsa"

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

		seedReleases(ctx, t, c, []string{"r1", "r2", "r3"}, "r3")
	}

	cfg := releaseSSHTestConfig(servers, []string{"web-01", "web-02"})

	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 2)

	t.Run("list", func(t *testing.T) {
		root, out := newReleaseRoot(nil, "plain")
		err := runClientReleaseList(root, "dev", merged, cfg, "test", "app", bifrostAgentBin)
		t.Logf("list output:\n%s", out.String())
		require.NoError(t, err)
		assert.Contains(t, out.String(), "web-01")
		assert.Contains(t, out.String(), "web-02")
		assert.Contains(t, out.String(), "r3  ← current")
	})

	t.Run("activate", func(t *testing.T) {
		root, out := newReleaseRoot(nil, "plain")
		err := runClientReleaseActivate(root, "dev", merged, cfg, "test", "app", "r2", bifrostAgentBin)
		t.Logf("activate output:\n%s", out.String())
		require.NoError(t, err)
		assert.Contains(t, out.String(), "web-01")
		assert.Contains(t, out.String(), "web-02")

		for name, c := range containers {
			res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
			require.NoError(t, err)
			assert.Equal(t, "/var/releases/r2", res.Output, "%s: current must point to the activated release", name)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		root, out := newReleaseRoot(nil, "plain")
		err := runClientReleaseRollback(root, "dev", merged, cfg, "test", "app", bifrostAgentBin)
		t.Logf("rollback output:\n%s", out.String())
		require.NoError(t, err)
		assert.Contains(t, out.String(), "web-01")
		assert.Contains(t, out.String(), "web-02")

		for name, c := range containers {
			res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
			require.NoError(t, err)
			assert.Equal(t, "/var/releases/r1", res.Output, "%s: rollback must move to the release preceding the activated one", name)
		}
	})
}
