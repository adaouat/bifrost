package atomic

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

// Extract decompresses and extracts an archive at artifactPath into destDir.
// onProgress, if non-nil, is called with the number of bytes written per chunk.
func Extract(ctx context.Context, artifactPath, destDir string, onProgress func(n int64)) error {
	f, err := os.Open(artifactPath)
	if err != nil {
		return fmt.Errorf("opening artifact: %w", err)
	}
	defer func() { _ = f.Close() }()

	format, stream, err := archives.Identify(ctx, artifactPath, f)
	if err != nil {
		return fmt.Errorf("identifying archive format: %w", err)
	}

	ex, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("format %T does not support extraction", format)
	}

	extractStream := stream
	if onProgress != nil {
		extractStream = &progressReader{r: stream, fn: onProgress}
	}

	if err := ex.Extract(ctx, extractStream, extractHandler(destDir)); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}

	return nil
}

func extractHandler(destDir string) archives.FileHandler {
	return func(ctx context.Context, fi archives.FileInfo) error {
		name := path.Clean(fi.NameInArchive)
		if name == "." {
			return nil
		}
		if path.IsAbs(name) || strings.HasPrefix(name, "../") || name == ".." {
			return fmt.Errorf("rejected unsafe path in archive: %s", fi.NameInArchive)
		}

		destPath := filepath.Join(destDir, filepath.FromSlash(name))

		if fi.IsDir() {
			return os.MkdirAll(destPath, fi.Mode().Perm())
		}

		if fi.Mode()&fs.ModeSymlink != 0 {
			_ = os.Remove(destPath)
			return os.Symlink(fi.LinkTarget, destPath)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", name, err)
		}

		src, err := fi.Open()
		if err != nil {
			return fmt.Errorf("opening archive entry %s: %w", fi.NameInArchive, err)
		}
		defer func() { _ = src.Close() }()

		out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode().Perm())
		if err != nil {
			return fmt.Errorf("creating %s: %w", name, err)
		}

		_, copyErr := io.Copy(out, src)
		closeErr := out.Close()
		if copyErr != nil {
			return fmt.Errorf("writing %s: %w", name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("closing %s: %w", name, closeErr)
		}

		return nil
	}
}

type progressReader struct {
	r  io.Reader
	fn func(n int64)
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.fn(int64(n))
	}
	return n, err
}
