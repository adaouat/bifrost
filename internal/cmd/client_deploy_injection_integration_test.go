//go:build integration

package cmd_test

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
	"github.com/adaouat/bifrost/internal/transport"
)

// TestClientDeploy_ShellQuotePreventsInjection proves that transport.ShellQuote
// — applied by every client agent-command builder — neutralises shell
// metacharacters against a real remote /bin/sh: a malicious value is passed as
// one literal argument and the command it tries to inject never runs.
func TestClientDeploy_ShellQuotePreventsInjection(t *testing.T) {
	ctx := context.Background()

	sshC := testutil.NewSSHContainer(ctx, t,
		"../../testdata/ssh/client_rsa.pub", "../../testdata/ssh/host_rsa")

	knownHostsPath, err := sshC.WriteKnownHosts("../../testdata/ssh/host_rsa.pub")
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	absKey, err := filepath.Abs("../../testdata/ssh/client_rsa")
	require.NoError(t, err)
	port, err := strconv.Atoi(sshC.Port)
	require.NoError(t, err)

	client, err := transport.Connect(config.ResolvedServer{
		Name: "ssh-test", Host: sshC.Host, Port: port, User: "deploy", KeyFile: absKey,
	})
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// If interpolated unquoted, this value would run `touch /tmp/INJECTED`.
	payload := "r1; touch /tmp/INJECTED"
	res, err := client.Exec("echo " + transport.ShellQuote(payload))
	require.NoError(t, err)
	require.Equal(t, 0, res.ExitCode)

	// The payload survives as one literal argument...
	assert.Equal(t, payload, strings.TrimRight(res.Stdout.String(), "\n"))

	// ...and the injected command never ran.
	check, err := sshC.Exec(ctx, []string{"test", "-e", "/tmp/INJECTED"})
	require.NoError(t, err)
	assert.Equal(t, 1, check.ExitCode, "/tmp/INJECTED exists — injection executed")
}
