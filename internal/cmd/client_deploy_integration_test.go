//go:build integration

package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/testutil"
	"github.com/adaouat/bifrost/internal/transport"
	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
)

// sshTestConfig builds an in-memory bifrost config for the SSH container test.
// The actual SSH connection parameters (host, port) are stored in the server map;
// the env/app structure mirrors bifrost-deploy-int-test.yml.
func sshTestConfig(host, port string, keyFile string) *config.Config {
	p, _ := strconv.Atoi(port)
	return &config.Config{
		Strategy: "atomic",
		Servers: map[string]config.ServerConfig{
			"ssh-test": {
				Host:       host,
				Port:       p,
				User:       "deploy",
				KeyFile:    keyFile,
				StagingDir: "/tmp",
			},
		},
		Paths: config.Paths{
			ReleasesRoot: "/var/releases",
			SharedRoot:   "/var/shared",
		},
		Settings: config.Settings{ReleasesToKeep: 5},
		Environments: map[string]config.Environment{
			"test": {
				Servers: []string{"ssh-test"},
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

func TestDeployCmd_ClientMode_E2E(t *testing.T) {
	ctx := context.Background()

	keyFile := "../../testdata/ssh/client_rsa"
	hostPubKey := "../../testdata/ssh/host_rsa.pub"
	clientPubKey := "../../testdata/ssh/client_rsa.pub"
	hostPrivKey := "../../testdata/ssh/host_rsa"

	// Start SSH server container configured with the test keys.
	sshC := testutil.NewSSHContainer(ctx, t, clientPubKey, hostPrivKey)

	// Write a known_hosts file and set env var so bifrost trusts the container's host key.
	knownHostsPath, err := sshC.WriteKnownHosts(hostPubKey)
	require.NoError(t, err)
	t.Setenv("BIFROST_KNOWN_HOSTS", knownHostsPath)

	// Resolve the SSH container address.
	absKeyFile, err := filepath.Abs(keyFile)
	require.NoError(t, err)

	cfg := sshTestConfig(sshC.Host, sshC.Port, absKeyFile)

	// Merge config to get the resolved server list.
	merged, err := config.Merge(cfg, "test", "app")
	require.NoError(t, err)
	require.Len(t, merged.Servers, 1)

	srv := merged.Servers[0]

	// Connect to the SSH server.
	client, err := transport.Connect(srv)
	require.NoError(t, err, "SSH connection failed")
	defer func() { _ = client.Close() }()

	// Create staging directory on the remote.
	staging, err := transport.NewStaging(client, srv.StagingDir)
	require.NoError(t, err)
	defer func() {
		if cleanErr := staging.Cleanup(); cleanErr != nil {
			t.Logf("cleanup warning: %v", cleanErr)
		}
	}()

	// Upload the linux agent binary.
	remoteAgent, err := staging.Upload(testutil.BinaryPath, "bifrost-agent")
	require.NoError(t, err, "uploading agent binary")

	// Generate and upload the flat config.
	var flatBuf bytes.Buffer
	require.NoError(t, config.GenerateFlatConfig(cfg, "test", "app", &flatBuf))
	remoteConfig, err := staging.UploadBytes(flatBuf.Bytes(), "bifrost.yml")
	require.NoError(t, err, "uploading flat config")

	// Upload the test artifact.
	remoteArtifact, err := staging.Upload("../../testdata/release.tar.gz", "release.tar.gz")
	require.NoError(t, err, "uploading artifact")

	// Make the agent executable.
	chmodRes, err := client.Exec("chmod +x " + remoteAgent)
	require.NoError(t, err)
	require.Equal(t, 0, chmodRes.ExitCode)

	// Build the directories referenced in the config before the deploy runs.
	// The container's setup script creates /var/releases and /var/shared, but
	// the flat config uses a different layout — ensure they exist.
	for _, dir := range []string{"/var/releases", "/var/shared"} {
		res, err := client.Exec(fmt.Sprintf("mkdir -p %s && chown deploy:deploy %s", dir, dir))
		require.NoError(t, err)
		require.Equal(t, 0, res.ExitCode)
	}

	// Exec the agent and stream its JSON events.
	agentCmd := fmt.Sprintf(
		"%s deploy --output json --config %s --artifact %s --release-name test-r1",
		remoteAgent, remoteConfig, remoteArtifact,
	)

	pr, pw := io.Pipe()
	var execResult *transport.ExecResult
	var execErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() { _ = pw.Close(); wg.Done() }()
		execResult, execErr = client.ExecStream(agentCmd, pw)
	}()

	var outBuf bytes.Buffer
	_, streamErr := tui.ForwardStream(pr, srv.Name, forgeui.Plain, &outBuf)
	_ = pr.Close()
	wg.Wait()

	t.Logf("deploy output:\n%s", outBuf.String())

	require.NoError(t, execErr, "agent exec error")
	require.Equal(t, 0, execResult.ExitCode, "agent exited non-zero:\nstderr: %s", execResult.Stderr.String())
	require.NoError(t, streamErr, "event stream error")

	// Assert the release directory was created.
	res, err := sshC.Exec(ctx, []string{"test", "-d", "/var/releases/test-r1"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "release directory must exist")

	// Assert the extracted file is present.
	res, err = sshC.Exec(ctx, []string{"test", "-f", "/var/releases/test-r1/public/index.html"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "extracted file must exist in release dir")

	// Assert the current symlink points to the release.
	res, err = sshC.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/test-r1", res.Output, "current symlink target")

	// Assert shared dir is symlinked.
	res, err = sshC.Exec(ctx, []string{"readlink", "/var/releases/test-r1/var/log"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/var/log", res.Output, "shared dir symlink")

	// Assert shared file is symlinked.
	res, err = sshC.Exec(ctx, []string{"readlink", "/var/releases/test-r1/.env"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/.env", res.Output, "shared file symlink")
}
