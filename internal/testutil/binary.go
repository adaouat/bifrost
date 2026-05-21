//go:build integration

package testutil

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// BinaryPath is where the cross-compiled bifrost binary is written.
const BinaryPath = "/tmp/bifrost-linux-amd64"

// BuildBifrostBinary cross-compiles the bifrost binary for linux/amd64 and writes it to BinaryPath.
// Returns the path and any error. Intended for use in TestMain where testing.TB is not available.
func BuildBifrostBinary() (string, error) {
	root, err := moduleRoot()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-o", BinaryPath, "./cmd/bifrost/")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")

	if out, err := cmd.CombinedOutput(); err != nil {
		return "", errors.New(string(out))
	}

	return BinaryPath, nil
}

// BuildBifrost cross-compiles the bifrost binary and calls t.Fatal on error.
// Intended to be called once from TestMain; the result is shared across all tests in the package.
func BuildBifrost(t testing.TB) string {
	t.Helper()

	path, err := BuildBifrostBinary()
	if err != nil {
		t.Fatalf("BuildBifrost: %v", err)
	}

	return path
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}
