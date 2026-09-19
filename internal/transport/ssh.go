package transport

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/adaouat/bifrost/internal/config"
)

// Client is an SSH connection to a single remote server.
type Client struct {
	conn      *ssh.Client
	agentConn net.Conn // the SSH-agent socket dialed for auth, if any; closed alongside conn
}

// ExecResult captures the output and exit status of a remote command.
// A non-zero ExitCode is a normal result, not a Go error — only connection or
// session failures are returned as errors.
type ExecResult struct {
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	ExitCode int
}

// Connect opens an SSH connection to srv using the configured auth chain and
// strict host-key verification against the system known_hosts file.
func Connect(srv config.ResolvedServer) (*Client, error) {
	auth, agentConn, err := authMethods(srv)
	if err != nil {
		return nil, err
	}
	hostKey, err := knownHostsCallback()
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return nil, err
	}

	port := srv.Port
	if port == 0 {
		port = 22
	}
	clientCfg := &ssh.ClientConfig{
		User:            srv.User,
		Auth:            auth,
		HostKeyCallback: hostKey,
		Timeout:         15 * time.Second,
	}

	addr := net.JoinHostPort(srv.Host, strconv.Itoa(port))
	conn, err := ssh.Dial("tcp", addr, clientCfg)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}
	return &Client{conn: conn, agentConn: agentConn}, nil
}

// Exec runs cmd on the remote host and returns its captured output and exit code.
func (c *Client) Exec(cmd string) (*ExecResult, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("opening ssh session: %w", err)
	}
	defer func() { _ = session.Close() }()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	runErr := session.Run(cmd)
	res := &ExecResult{Stdout: &stdout, Stderr: &stderr}
	if runErr == nil {
		return res, nil
	}
	if exitErr, ok := errors.AsType[*ssh.ExitError](runErr); ok {
		res.ExitCode = exitErr.ExitStatus()
		return res, nil
	}
	return res, fmt.Errorf("running %q: %w", cmd, runErr)
}

// ExecStream runs cmd on the remote host, streaming stdout to out as data
// arrives. Stderr is captured and returned in the ExecResult. ExecStream
// blocks until the remote command exits.
func (c *Client) ExecStream(cmd string, stdout io.Writer) (*ExecResult, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("opening ssh session: %w", err)
	}
	defer func() { _ = session.Close() }()

	var stderr bytes.Buffer
	session.Stdout = stdout
	session.Stderr = &stderr

	runErr := session.Run(cmd)
	res := &ExecResult{Stdout: &bytes.Buffer{}, Stderr: &stderr}
	if runErr == nil {
		return res, nil
	}
	if exitErr, ok := errors.AsType[*ssh.ExitError](runErr); ok {
		res.ExitCode = exitErr.ExitStatus()
		return res, nil
	}
	return res, fmt.Errorf("running %q: %w", cmd, runErr)
}

// Close terminates the underlying SSH connection and the SSH-agent socket
// dialed for auth, if any.
func (c *Client) Close() error {
	if c.agentConn != nil {
		_ = c.agentConn.Close()
	}
	return c.conn.Close()
}

// dialAgentSocket dials the SSH-agent unix socket, replaced in tests.
var dialAgentSocket = func(sock string) (net.Conn, error) {
	return net.Dial("unix", sock)
}

// authMethods builds the auth chain in priority order: explicit key file, then
// the SSH agent, then the BIFROST_SSH_PASSWORD env var. The ssh library tries
// each in order and uses the first that authenticates. The second return
// value is the SSH-agent socket connection, if one was dialed — the caller
// must close it alongside the SSH connection, or it leaks.
func authMethods(srv config.ResolvedServer) ([]ssh.AuthMethod, net.Conn, error) {
	var methods []ssh.AuthMethod
	var agentConn net.Conn

	if srv.KeyFile != "" {
		signer, err := loadKey(srv.KeyFile)
		if err != nil {
			return nil, nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if conn, err := dialAgentSocket(sock); err == nil {
			agentConn = conn
			ag := agent.NewClient(conn)
			methods = append(methods, ssh.PublicKeysCallback(ag.Signers))
		}
	}

	if pw := os.Getenv("BIFROST_SSH_PASSWORD"); pw != "" {
		methods = append(methods, ssh.Password(pw))
	}

	if len(methods) == 0 {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return nil, nil, fmt.Errorf("no SSH authentication available: set key_file, start an ssh-agent, or set BIFROST_SSH_PASSWORD")
	}
	return methods, agentConn, nil
}

func loadKey(path string) (ssh.Signer, error) {
	expanded, err := expandHome(path)
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile(expanded)
	if err != nil {
		return nil, fmt.Errorf("reading key file %s: %w", expanded, err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parsing key file %s: %w", expanded, err)
	}
	return signer, nil
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	path := os.Getenv("BIFROST_KNOWN_HOSTS")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".ssh", "known_hosts")
	}
	cb, err := knownhosts.New(path)
	if err != nil {
		return nil, fmt.Errorf("loading known_hosts %s: %w", path, err)
	}
	return cb, nil
}

func expandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}
