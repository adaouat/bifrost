package cmd

import (
	"os"
	"testing"
)

func TestResolveConfigPath_ExplicitFlag(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	root := NewRootCmd()
	if err := root.PersistentFlags().Set("config", "/custom/path.yml"); err != nil {
		t.Fatalf("setting flag: %v", err)
	}
	if got := resolveConfigPath(root); got != "/custom/path.yml" {
		t.Errorf("got %q, want /custom/path.yml", got)
	}
}

func TestResolveConfigPath_XdgConfigExists(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "")
	if err := os.MkdirAll(".config", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".config/bifrost.yml", []byte("x: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := resolveConfigPath(NewRootCmd()); got != ".config/bifrost.yml" {
		t.Errorf("got %q, want .config/bifrost.yml", got)
	}
}

func TestResolveConfigPath_FallbackToDotfile(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "")
	if got := resolveConfigPath(NewRootCmd()); got != ".bifrost.yml" {
		t.Errorf("got %q, want .bifrost.yml", got)
	}
}
