package transport

import (
	"bytes"
	"fmt"
	"os"

	"github.com/adaouat/bifrost/internal/config"
)

// AgentSession is a connected client with the agent binary and flat config
// already staged on the remote, ready to run bifrost-agent commands.
type AgentSession struct {
	Client       *Client
	Staging      *Staging
	RemoteAgent  string
	RemoteConfig string
	serverName   string
}

// Close tears down the staging directory and the SSH connection. A cleanup
// failure is reported as a warning on stderr — it never affects the result of
// the command that used the session.
func (s *AgentSession) Close() {
	if err := s.Staging.Cleanup(); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: cleanup staging on %s: %v\n", s.serverName, err)
	}
	_ = s.Client.Close()
}

// StageAgentSession connects to srv, resolves and uploads the agent binary (or
// uses agentBinary if given), generates and uploads env/app's flat config
// directly from memory, and chmods the agent executable. It is the shared
// connect+stage sequence for deploy, release list, release activate, and
// release rollback's client-mode paths — previously each reimplemented it
// independently and had already drifted.
func StageAgentSession(version string, cfg *config.Config, env, app, agentBinary string, srv config.ResolvedServer) (*AgentSession, error) {
	client, err := Connect(srv)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", srv.Name, err)
	}

	localAgentPath := agentBinary
	if localAgentPath == "" {
		p, err := DetectPlatform(client)
		if err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("detecting platform on %s: %w", srv.Name, err)
		}
		localAgentPath, err = ResolveAgentBinary(version, p)
		if err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("resolving agent binary for %s: %w", srv.Name, err)
		}
	}

	staging, err := NewStaging(client, srv.StagingDir)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("creating staging dir on %s: %w", srv.Name, err)
	}

	remoteAgent, err := staging.Upload(localAgentPath, "bifrost-agent")
	if err != nil {
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, fmt.Errorf("uploading agent to %s: %w", srv.Name, err)
	}

	var flat bytes.Buffer
	if err := config.GenerateFlatConfig(cfg, env, app, &flat); err != nil {
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, fmt.Errorf("generating flat config for %s: %w", srv.Name, err)
	}
	remoteConfig, err := staging.UploadBytes(flat.Bytes(), "bifrost.yml")
	if err != nil {
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, fmt.Errorf("uploading config to %s: %w", srv.Name, err)
	}

	if err := staging.Chmod(remoteAgent, 0o755); err != nil {
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, fmt.Errorf("chmod agent on %s: %w", srv.Name, err)
	}

	return &AgentSession{
		Client:       client,
		Staging:      staging,
		RemoteAgent:  remoteAgent,
		RemoteConfig: remoteConfig,
		serverName:   srv.Name,
	}, nil
}
