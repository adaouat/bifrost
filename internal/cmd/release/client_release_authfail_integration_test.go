//go:build integration

package release

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
)

// TestRunClientReleaseList_SSHAuthFailure_ExitCode3 proves stageAgentSession's
// connect failure both classifies as exit 3 (Runtime) and keeps the real SSH
// error reachable via errors.Unwrap for errors.Is/errors.As, instead of
// flattening it into a message string.
func TestRunClientReleaseList_SSHAuthFailure_ExitCode3(t *testing.T) {
	ctx := context.Background()

	hostPubKey := "../../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../../testdata/ssh/host_rsa"

	sshC := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)

	knownHostsPath, err := sshC.WriteKnownHosts(hostPubKey)
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)
	t.Setenv("SSH_AUTH_SOCK", "")

	// host_rsa is a valid private key, but its public half is the container's
	// host key, not an authorized client key — the connection fails at
	// user authentication specifically.
	unauthorizedKey, err := filepath.Abs(hostPrivKey)
	require.NoError(t, err)

	port, err := strconv.Atoi(sshC.Port)
	require.NoError(t, err)

	cfg := releaseSSHTestConfig(map[string]config.Server{
		"web-01": {Host: sshC.Host, Port: port, User: "deploy", KeyFile: unauthorizedKey, StagingDir: "/tmp"},
	}, []string{"web-01"})
	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)

	root, out := newReleaseRoot(nil, "plain")
	err = runClientReleaseList(root, "dev", merged, cfg, "test", "app", bifrostAgentBin)
	t.Logf("list output:\n%s", out.String())
	require.Error(t, err)

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, cmderr.Runtime, exitErr.Code, "auth failure must map to exit code 3 (Runtime)")
	assert.Contains(t, err.Error(), "connecting to web-01", "error must name the server it failed to reach")

	underlying := errors.Unwrap(errors.Unwrap(err))
	assert.NotNil(t, underlying, "the real SSH error must stay reachable via errors.Unwrap for errors.Is/As, not flattened into a message string")
}
