//go:build integration

package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
)

func TestRunClientDeploy_SSHAuthFailure_ExitCode3(t *testing.T) {
	ctx := context.Background()

	hostPubKey := "../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../testdata/ssh/host_rsa"

	sshC := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)

	knownHostsPath, err := sshC.WriteKnownHosts(hostPubKey)
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	// Disable the ssh-agent fallback so the only credential offered is the key
	// file below — the test must fail at authentication, not succeed via some
	// key happening to be loaded in the developer's agent.
	t.Setenv("SSH_AUTH_SOCK", "")

	// host_rsa is a valid private key, but its public half is the container's
	// host key — it is not in authorized_keys (only client_rsa.pub is). The
	// known_hosts file lets host verification pass, so the connection fails
	// specifically at user authentication.
	unauthorizedKey, err := filepath.Abs(hostPrivKey)
	require.NoError(t, err)

	port, err := strconv.Atoi(sshC.Port)
	require.NoError(t, err)
	servers := map[string]config.ServerConfig{
		"web-01": {Host: sshC.Host, Port: port, User: "deploy", KeyFile: unauthorizedKey, StagingDir: "/tmp"},
	}

	cfg := multiServerTestConfig(servers, []string{"web-01"})
	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "plain", "")
	var out bytes.Buffer
	root.SetOut(&out)

	err = runClientDeploy(root, "dev", merged, cfg, "test", "app",
		"../../testdata/release.tar.gz", "authfail-r1", testutil.BinaryPath)
	require.Error(t, err)

	var exitErr *ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, Runtime, exitErr.Code, "auth failure must map to exit code 3 (Runtime)")
	assert.Contains(t, err.Error(), "connecting to web-01", "error must name the server it failed to reach")
}
