package transport

import (
	"bytes"
	"fmt"
	"path"

	"github.com/google/uuid"
)

// Staging is a unique per-invocation directory on a remote server. It holds the
// agent binary, the flat config, and (for deploy) the artifact, and is always
// removed via Cleanup after the agent exits.
type Staging struct {
	client *Client
	sftp   *SFTP
	Dir    string
}

// NewStaging creates a fresh staging directory on the remote server and returns
// a handle for uploading files into it and cleaning it up afterwards. base is
// the server's configured staging_dir; an empty base defaults to /tmp.
func NewStaging(client *Client, base string) (*Staging, error) {
	dir := stagingPath(base, uuid.NewString())
	sftpClient, err := client.NewSFTP()
	if err != nil {
		return nil, err
	}
	if err := sftpClient.Mkdir(dir); err != nil {
		_ = sftpClient.Close()
		return nil, err
	}
	return &Staging{client: client, sftp: sftpClient, Dir: dir}, nil
}

// Upload places a local file into the staging directory under remoteName and
// returns its full remote path.
func (s *Staging) Upload(localPath, remoteName string) (string, error) {
	remotePath := path.Join(s.Dir, remoteName)
	if err := s.sftp.Upload(localPath, remotePath); err != nil {
		return "", err
	}
	return remotePath, nil
}

// UploadBytes writes data into the staging directory under remoteName without
// requiring a local file on disk. Returns the full remote path.
func (s *Staging) UploadBytes(data []byte, remoteName string) (string, error) {
	remotePath := path.Join(s.Dir, remoteName)
	if err := s.sftp.UploadReader(bytes.NewReader(data), remotePath); err != nil {
		return "", err
	}
	return remotePath, nil
}

// Cleanup removes the staging directory and closes the SFTP session. It is meant
// to run after the agent exits, regardless of the deploy outcome.
func (s *Staging) Cleanup() error {
	_ = s.sftp.Close()
	res, err := s.client.Exec("rm -rf " + ShellQuote(s.Dir))
	if err != nil {
		return fmt.Errorf("cleanup staging dir %s: %w", s.Dir, err)
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("cleanup staging dir %s: exit %d: %s", s.Dir, res.ExitCode, res.Stderr.String())
	}
	return nil
}

// stagingPath builds the remote staging directory path. Remote paths are always
// POSIX, so path.Join is used rather than filepath.Join (which would emit
// backslashes when the client runs on Windows).
func stagingPath(base, id string) string {
	if base == "" {
		base = "/tmp"
	}
	return path.Join(base, "bifrost-"+id)
}
