//go:build integration

package testutil

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHContainer wraps a container running an OpenSSH server for integration tests.
type SSHContainer struct {
	inner testcontainers.Container
	Host  string
	Port  string
}

// NewSSHContainer starts a Linux container running OpenSSH, configured with
// the given client public key (for authorized_keys) and host private key.
// Both paths must point to pre-generated keys committed in testdata/.
// The container is terminated when the test ends.
func NewSSHContainer(ctx context.Context, t testing.TB, clientPubKeyPath, hostPrivKeyPath string) *SSHContainer {
	t.Helper()

	clientPub, err := os.ReadFile(clientPubKeyPath)
	if err != nil {
		t.Fatalf("NewSSHContainer: reading client public key: %v", err)
	}
	hostPriv, err := os.ReadFile(hostPrivKeyPath)
	if err != nil {
		t.Fatalf("NewSSHContainer: reading host private key: %v", err)
	}

	// Base64-encode keys to avoid quoting issues in shell scripts.
	clientPubB64 := base64.StdEncoding.EncodeToString(bytes.TrimSpace(clientPub))
	hostPrivB64 := base64.StdEncoding.EncodeToString(hostPriv)

	setup := fmt.Sprintf(`#!/bin/sh
set -e
apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq openssh-server
useradd -m -s /bin/sh deploy
mkdir -p /home/deploy/.ssh
printf '%%s' '%s' | base64 -d > /home/deploy/.ssh/authorized_keys
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh
mkdir -p /var/releases /var/shared
chown deploy:deploy /var/releases /var/shared
mkdir -p /run/sshd
printf '%%s' '%s' | base64 -d > /etc/ssh/ssh_host_rsa_key
chmod 600 /etc/ssh/ssh_host_rsa_key
touch /tmp/.setup-complete
exec /usr/sbin/sshd -D \
  -o HostKey=/etc/ssh/ssh_host_rsa_key \
  -o PubkeyAuthentication=yes \
  -o PasswordAuthentication=no \
  -o PermitRootLogin=no \
  -o StrictModes=no
`, clientPubB64, hostPrivB64)

	req := testcontainers.ContainerRequest{
		Image:        containerImage,
		ExposedPorts: []string{"22/tcp"},
		Cmd:          []string{"sh", "-c", setup},
		// ForListeningPort can pass before the setup script finishes — Docker's
		// port proxy binds the mapped port early, so tests would run container
		// setup (e.g. `chown deploy:deploy`) before `useradd deploy` exists. Gate
		// on a marker the script writes only after all setup, just before sshd.
		WaitingFor: wait.ForExec([]string{"test", "-f", "/tmp/.setup-complete"}).WithStartupTimeout(3 * time.Minute),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("NewSSHContainer: start container: %v", err)
	}

	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	mappedPort, err := c.MappedPort(ctx, "22/tcp")
	if err != nil {
		t.Fatalf("NewSSHContainer: getting mapped port: %v", err)
	}
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("NewSSHContainer: getting host: %v", err)
	}

	return &SSHContainer{
		inner: c,
		Host:  host,
		Port:  mappedPort.Port(),
	}
}

// Exec runs a command inside the SSH container.
func (s *SSHContainer) Exec(ctx context.Context, cmd []string) (ExecResult, error) {
	code, reader, err := s.inner.Exec(ctx, cmd)
	if err != nil {
		return ExecResult{}, err
	}
	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return ExecResult{}, err
	}
	return ExecResult{
		ExitCode: code,
		Output:   strings.TrimRight(stdout.String(), "\n"),
		Stderr:   strings.TrimRight(stderr.String(), "\n"),
	}, nil
}

// WriteKnownHosts generates a known_hosts file for this container using the
// given host public key and writes it to /tmp. Returns the file path.
func (s *SSHContainer) WriteKnownHosts(hostPubKeyPath string) (string, error) {
	return WriteKnownHosts([]*SSHContainer{s}, []string{hostPubKeyPath})
}

// WriteKnownHosts builds a known_hosts file covering each given SSH
// container, matched by index with hostPubKeyPaths, and writes it to /tmp.
// Returns the file path. Used by tests connecting to multiple containers
// over a single BIFROST_KNOWN_HOSTS file.
func WriteKnownHosts(containers []*SSHContainer, hostPubKeyPaths []string) (string, error) {
	if len(containers) != len(hostPubKeyPaths) {
		return "", fmt.Errorf("containers and host public keys must have the same length")
	}

	var buf bytes.Buffer
	for i, c := range containers {
		pubKeyBytes, err := os.ReadFile(hostPubKeyPaths[i])
		if err != nil {
			return "", fmt.Errorf("reading host public key: %w", err)
		}

		pk, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyBytes)
		if err != nil {
			return "", fmt.Errorf("parsing host public key: %w", err)
		}

		addr := fmt.Sprintf("[%s]:%s", c.Host, c.Port)
		buf.WriteString(knownhosts.Line([]string{addr}, pk) + "\n")
	}

	// Unique per call: the CMD and RELEASE test packages run in parallel and would
	// otherwise clobber each other's known_hosts at a shared path, causing
	// "knownhosts: key is unknown" host-key failures on loaded CI runners.
	path := filepath.Join("/tmp", "bifrost-known-hosts-"+uuid.NewString())
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return "", fmt.Errorf("writing known_hosts: %w", err)
	}
	return path, nil
}
