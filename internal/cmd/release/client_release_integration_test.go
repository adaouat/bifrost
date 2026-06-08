//go:build integration

package release

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
)

var bifrostAgentBin string

func TestMain(m *testing.M) {
	path, err := testutil.BuildBifrostBinary()
	if err != nil {
		panic("BuildBifrostBinary: " + err.Error())
	}
	bifrostAgentBin = path
	os.Exit(m.Run())
}

// releaseSSHTestConfig builds an in-memory bifrost config for SSH-based release
// command tests: a single named server, no shared dirs/files (kept minimal —
// these tests exercise release management, not shared resource linking).
func releaseSSHTestConfig(servers map[string]config.ServerConfig, order []string) *config.Config {
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
				Servers:      order,
				Applications: map[string]config.Application{"app": {}},
			},
		},
	}
}

// seedReleases creates the given release directories under /var/releases on c,
// points "current" at currentRelease, and fixes ownership for the deploy user.
func seedReleases(ctx context.Context, t *testing.T, c *testutil.SSHContainer, releases []string, currentRelease string) {
	t.Helper()

	mkdirArgs := make([]string, len(releases))
	for i, r := range releases {
		mkdirArgs[i] = "/var/releases/" + r
	}
	script := fmt.Sprintf("mkdir -p %s && ln -sfn /var/releases/%s /var/releases/current && chown -R deploy:deploy /var/releases",
		joinArgs(mkdirArgs), currentRelease)

	res, err := c.Exec(ctx, []string{"sh", "-c", script})
	require.NoError(t, err)
	require.Equal(t, 0, res.ExitCode, "seeding releases: %s", res.Output)
}

func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		out += a
	}
	return out
}

func newReleaseSSHContainer(ctx context.Context, t *testing.T) (*testutil.SSHContainer, string) {
	t.Helper()

	keyFile := "../../../testdata/ssh/client_rsa"
	hostPubKey := "../../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../../testdata/ssh/host_rsa"

	c := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)

	knownHostsPath, err := c.WriteKnownHosts(hostPubKey)
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	absKeyFile, err := filepath.Abs(keyFile)
	require.NoError(t, err)

	return c, absKeyFile
}

func newReleaseRoot(cmd *cobra.Command, mode string) (*cobra.Command, *bytes.Buffer) {
	root := cmd
	if root == nil {
		root = &cobra.Command{Use: "root"}
	}
	root.PersistentFlags().String("output", mode, "")
	root.PersistentFlags().String("config", "", "")
	var out bytes.Buffer
	root.SetOut(&out)
	return root, &out
}

func TestReleaseCommands_ClientMode_SingleServer_E2E(t *testing.T) {
	ctx := context.Background()

	c, absKeyFile := newReleaseSSHContainer(ctx, t)

	port, err := strconv.Atoi(c.Port)
	require.NoError(t, err)

	cfg := releaseSSHTestConfig(map[string]config.ServerConfig{
		"ssh-test": {Host: c.Host, Port: port, User: "deploy", KeyFile: absKeyFile, StagingDir: "/tmp"},
	}, []string{"ssh-test"})

	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)

	seedReleases(ctx, t, c, []string{"r1", "r2", "r3"}, "r3")

	t.Run("list", func(t *testing.T) {
		root, out := newReleaseRoot(nil, "plain")
		err := runClientReleaseList(root, "dev", merged, cfg, "test", "app", bifrostAgentBin)
		t.Logf("list output:\n%s", out.String())
		require.NoError(t, err)
		assert.Contains(t, out.String(), "r1")
		assert.Contains(t, out.String(), "r3  ← current")
	})

	t.Run("activate", func(t *testing.T) {
		root, out := newReleaseRoot(nil, "plain")
		err := runClientReleaseActivate(root, "dev", merged, cfg, "test", "app", "r2", bifrostAgentBin)
		t.Logf("activate output:\n%s", out.String())
		require.NoError(t, err)
		assert.Contains(t, out.String(), "Activated r2")

		res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
		require.NoError(t, err)
		assert.Equal(t, "/var/releases/r2", res.Output, "current must point to the activated release")
	})

	t.Run("rollback", func(t *testing.T) {
		root, out := newReleaseRoot(nil, "plain")
		err := runClientReleaseRollback(root, "dev", merged, cfg, "test", "app", bifrostAgentBin)
		t.Logf("rollback output:\n%s", out.String())
		require.NoError(t, err)
		assert.Contains(t, out.String(), "Rolled back to r1")

		res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
		require.NoError(t, err)
		assert.Equal(t, "/var/releases/r1", res.Output, "rollback must move to the release preceding the activated one")
	})
}
