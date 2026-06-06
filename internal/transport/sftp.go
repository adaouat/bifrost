package transport

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
)

// SFTP wraps an sftp.Client bound to an existing SSH connection.
type SFTP struct {
	client *sftp.Client
}

// NewSFTP opens an SFTP session over the client's SSH connection.
func (c *Client) NewSFTP() (*SFTP, error) {
	s, err := sftp.NewClient(c.conn)
	if err != nil {
		return nil, fmt.Errorf("opening sftp session: %w", err)
	}
	return &SFTP{client: s}, nil
}

// UploadReader copies r to remotePath on the server.
func (s *SFTP) UploadReader(r io.Reader, remotePath string) error {
	dst, err := s.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("creating remote %s: %w", remotePath, err)
	}
	defer func() { _ = dst.Close() }()
	if _, err := io.Copy(dst, r); err != nil {
		return fmt.Errorf("uploading to %s: %w", remotePath, err)
	}
	return nil
}

// Upload copies a local file to remotePath on the server.
func (s *SFTP) Upload(localPath, remotePath string) error {
	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", localPath, err)
	}
	defer func() { _ = src.Close() }()

	dst, err := s.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("creating remote %s: %w", remotePath, err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("uploading to %s: %w", remotePath, err)
	}
	return nil
}

// Mkdir creates remotePath and any missing parents.
func (s *SFTP) Mkdir(remotePath string) error {
	if err := s.client.MkdirAll(remotePath); err != nil {
		return fmt.Errorf("mkdir %s: %w", remotePath, err)
	}
	return nil
}

// Chmod sets the permission bits on remotePath.
func (s *SFTP) Chmod(remotePath string, mode os.FileMode) error {
	if err := s.client.Chmod(remotePath, mode); err != nil {
		return fmt.Errorf("chmod %s: %w", remotePath, err)
	}
	return nil
}

// Close terminates the SFTP session.
func (s *SFTP) Close() error {
	return s.client.Close()
}
