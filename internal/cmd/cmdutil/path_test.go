package cmdutil

import (
	"os"
	"testing"
)

func TestResolvePath_FlagWins(t *testing.T) {
	t.Setenv("BIFROST_FILE", "/env/bifrost.yml")
	if got := ResolvePath("/flag/bifrost.yml"); got != "/flag/bifrost.yml" {
		t.Errorf("got %q, want /flag/bifrost.yml", got)
	}
}

func TestResolvePath_EnvVar(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "/env/bifrost.yml")
	if got := ResolvePath(""); got != "/env/bifrost.yml" {
		t.Errorf("got %q, want /env/bifrost.yml", got)
	}
}

func TestResolvePath_XdgDiscovered(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "")
	if err := os.MkdirAll(".config", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".config/bifrost.yml", []byte("x: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := ResolvePath(""); got != ".config/bifrost.yml" {
		t.Errorf("got %q, want .config/bifrost.yml", got)
	}
}

func TestResolvePath_DotfileFallback(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "")
	if got := ResolvePath(""); got != ".bifrost.yml" {
		t.Errorf("got %q, want .bifrost.yml", got)
	}
}

func TestResolveInitDest_EnvWins(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "/env/bifrost.yml")
	if got := ResolveInitDest(); got != "/env/bifrost.yml" {
		t.Errorf("got %q, want /env/bifrost.yml", got)
	}
}

func TestResolveInitDest_ConfigDir(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "")
	if err := os.MkdirAll(".config", 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ResolveInitDest(); got != ".config/bifrost.yml" {
		t.Errorf("got %q, want .config/bifrost.yml", got)
	}
}

func TestStatFile_Injectable(t *testing.T) {
	SetStatFile(func(string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	defer SetStatFile(nil)
	if _, err := StatFile("anything"); err == nil {
		t.Error("expected injected statFile to return its error")
	}
}
