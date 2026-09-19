package transport

import (
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
)

// TestAuthMethods_ReturnsAgentConnForCleanup proves authMethods hands back the
// SSH-agent socket connection it dials, so the caller (Client.Close) can close
// it — without this, every Connect against a server with SSH_AUTH_SOCK set
// leaks one fd/connection per deploy.
func TestAuthMethods_ReturnsAgentConnForCleanup(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", "/fake/agent.sock")

	clientEnd, serverEnd := net.Pipe()
	defer func() { _ = serverEnd.Close() }()

	orig := dialAgentSocket
	dialAgentSocket = func(sock string) (net.Conn, error) {
		assert.Equal(t, "/fake/agent.sock", sock)
		return clientEnd, nil
	}
	defer func() { dialAgentSocket = orig }()

	_, agentConn, err := authMethods(config.ResolvedServer{User: "deploy"})
	require.NoError(t, err)
	require.NotNil(t, agentConn, "authMethods must return the dialed agent conn for the caller to close")

	require.NoError(t, agentConn.Close())
	_, writeErr := agentConn.Write([]byte("x"))
	assert.ErrorIs(t, writeErr, io.ErrClosedPipe, "the returned conn must be the real dialed conn, not a copy")
}

// TestAuthMethods_NoAgentSocket_ReturnsNilConn proves authMethods returns a
// nil conn (nothing to clean up) when SSH_AUTH_SOCK isn't set.
func TestAuthMethods_NoAgentSocket_ReturnsNilConn(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", "")
	t.Setenv("BIFROST_SSH_PASSWORD", "hunter2")

	_, agentConn, err := authMethods(config.ResolvedServer{User: "deploy"})
	require.NoError(t, err)
	assert.Nil(t, agentConn)
}
