//go:build integration

package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// BinaryPath is where the cross-compiled bifrost binary is written.
const BinaryPath = "/tmp/bifrost-linux-amd64"

// BuildBifrost cross-compiles the bifrost binary for linux/amd64 and writes it to BinaryPath.
// Intended to be called once from TestMain; the result is shared across all tests in the package.
func BuildBifrost(t testing.TB) string {
	t.Helper()

	root := findModuleRoot(t)
	cmd := exec.Command("go", "build", "-o", BinaryPath, "./cmd/bifrost/")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("BuildBifrost: %v\n%s", err, out)
	}

	return BinaryPath
}

func findModuleRoot(t testing.TB) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("findModuleRoot: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("findModuleRoot: go.mod not found")
		}
		dir = parent
	}
}
