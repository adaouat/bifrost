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

func TestRunClientDeploy_AgentBinaryFlag_SkipsDownload(t *testing.T) {
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

	cfg := multiServerTestConfig(servers, []string{"web-01"})
	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "plain", "")
	var out bytes.Buffer
	root.SetOut(&out)

	// A version that maps to no published GitHub release. If --agent-binary were
	// ignored, deployToServer would fall through to ResolveAgentBinary(version, …)
	// and fail trying to download this tag. A successful deploy therefore proves
	// the supplied binary was used and the download path was never taken.
	const unresolvableVersion = "v0.0.0-no-such-release"

	err = runClientDeploy(root, unresolvableVersion, merged, cfg, "test", "app",
		"../../testdata/release.tar.gz", "agentflag-r1", testutil.BinaryPath)
	t.Logf("deploy output:\n%s", out.String())
	require.NoError(t, err)

	res, err := sshC.Exec(ctx, []string{"test", "-d", "/var/releases/agentflag-r1"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "release directory must exist")

	res, err = sshC.Exec(ctx, []string{"test", "-f", "/var/releases/agentflag-r1/public/index.html"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "extracted file must exist in release dir")

	res, err = sshC.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/agentflag-r1", res.Output, "current symlink target")
}
