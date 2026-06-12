//go:build integration

package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
	"github.com/adaouat/bifrost/internal/transport"
)

// waitForSSHReady dials srv until the handshake succeeds. wait.ForListeningPort
// only proves the port is open, leaving a brief window where sshd accepts TCP
// but drops the handshake with EOF; this test needs the connection to succeed
// so the failure it asserts comes from platform detection, not the dial.
func waitForSSHReady(t *testing.T, srv config.ResolvedServer) {
	t.Helper()
	var lastErr error
	for range 20 {
		client, err := transport.Connect(srv)
		if err == nil {
			_ = client.Close()
			return
		}
		lastErr = err
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("ssh never became ready: %v", lastErr)
}

// fakeUnameI386 replaces the remote uname with one reporting Linux/i386, an
// arch absent from mapPlatform's table, so DetectPlatform rejects it. The
// heredoc delimiter is quoted so $1 reaches the file unexpanded.
const fakeUnameI386 = `cat > /usr/bin/uname <<"EOF"
#!/bin/sh
case "$1" in
  -s) echo Linux ;;
  -m) echo i386 ;;
  *) echo unknown ;;
esac
EOF
chmod 755 /usr/bin/uname`

func TestRunClientDeploy_UnknownRemoteArch_ExitCode3(t *testing.T) {
	ctx := context.Background()

	keyFile := "../../testdata/ssh/client_rsa"
	hostPubKey := "../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../testdata/ssh/host_rsa"

	sshC := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)

	knownHostsPath, err := sshC.WriteKnownHosts(hostPubKey)
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	res, err := sshC.Exec(ctx, []string{"sh", "-c", fakeUnameI386})
	require.NoError(t, err)
	require.Equal(t, 0, res.ExitCode, "installing fake uname: %s", res.Stderr)

	absKeyFile, err := filepath.Abs(keyFile)
	require.NoError(t, err)

	port, err := strconv.Atoi(sshC.Port)
	require.NoError(t, err)
	servers := map[string]config.Server{
		"web-01": {Host: sshC.Host, Port: port, User: "deploy", KeyFile: absKeyFile, StagingDir: "/tmp"},
	}

	cfg := multiServerTestConfig(servers, []string{"web-01"})
	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)

	waitForSSHReady(t, merged.Servers[0])

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "plain", "")
	var out bytes.Buffer
	root.SetOut(&out)

	// Empty agentBinary forces the download path, which begins with remote
	// platform detection — the step that must reject the unsupported arch.
	err = runClientDeploy(root, "dev", merged, cfg, "test", "app",
		"../../testdata/release.tar.gz", "arch-r1", "")
	require.Error(t, err)

	var exitErr *ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, Runtime, exitErr.Code, "unknown arch must map to exit code 3 (Runtime)")
	assert.Contains(t, err.Error(), "detecting platform on web-01")
	assert.Contains(t, err.Error(), "Linux/i386")
}
