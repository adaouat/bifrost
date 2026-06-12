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

func TestRunClientDeploy_AgentFailsMidDeploy_CleansStaging(t *testing.T) {
	ctx := context.Background()

	keyFile := "../../testdata/ssh/client_rsa"
	hostPubKey := "../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../testdata/ssh/host_rsa"

	sshC := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)

	knownHostsPath, err := sshC.WriteKnownHosts(hostPubKey)
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	absKeyFile, err := filepath.Abs(keyFile)
	require.NoError(t, err)

	port, err := strconv.Atoi(sshC.Port)
	require.NoError(t, err)
	servers := map[string]config.Server{
		"web-01": {Host: sshC.Host, Port: port, User: "deploy", KeyFile: absKeyFile, StagingDir: "/tmp"},
	}

	for _, dir := range []string{"/var/releases", "/var/shared"} {
		res, err := sshC.Exec(ctx, []string{"sh", "-c", fmt.Sprintf("mkdir -p %s && chown deploy:deploy %s", dir, dir)})
		require.NoError(t, err)
		require.Equal(t, 0, res.ExitCode)
	}

	// A post_extract hook that fails makes the remote agent exit non-zero after
	// it has already created and extracted the release — a mid-deploy failure.
	cfg := multiServerTestConfig(servers, []string{"web-01"})
	prio := 10
	env := cfg.Environments["test"]
	app := env.Applications["app"]
	app.Hooks.PostExtract = []config.HookEntry{{Cmd: "exit 1", Priority: &prio}}
	env.Applications["app"] = app

	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)

	waitForSSHReady(t, merged.Servers[0])

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "plain", "")
	var out bytes.Buffer
	root.SetOut(&out)

	err = runClientDeploy(root, "dev", merged, cfg, "test", "app",
		"../../testdata/release.tar.gz", "cleanup-r1", testutil.BinaryPath)
	t.Logf("deploy output:\n%s", out.String())

	var exitErr *ExitError
	require.ErrorAs(t, err, &exitErr, "deploy must fail when the agent exits non-zero")
	assert.NotEqual(t, 0, exitErr.Code)

	// The release dir was extracted before the hook failed: the agent reached
	// mid-deploy rather than failing on connect or config parsing.
	res, err := sshC.Exec(ctx, []string{"test", "-d", "/var/releases/cleanup-r1"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "release dir must exist — agent must have reached extraction")

	// Despite the failure, the deferred staging cleanup must have removed the
	// staging dir; no bifrost-* directory may leak under the staging base.
	res, err = sshC.Exec(ctx, []string{"sh", "-c", "find /tmp -maxdepth 1 -name 'bifrost-*'"})
	require.NoError(t, err)
	assert.Empty(t, res.Output, "staging dir must be cleaned up after the agent fails")
}
