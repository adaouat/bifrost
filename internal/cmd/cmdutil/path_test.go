package cmdutil

import (
	"errors"
	"os"
	"testing"
)

func TestResolvePath_ExplicitFlag(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) {
		t.Fatal("statFile must not be called when explicit path is given")
		return nil, nil
	}

	got := ResolvePath("/custom/bifrost.yml")
	if got != "/custom/bifrost.yml" {
		t.Errorf("got %q, want /custom/bifrost.yml", got)
	}
}

func TestResolvePath_EnvVarUsedWhenNoFlag(t *testing.T) {
	t.Setenv("BIFROST_FILE", "/env/bifrost.yml")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) {
		t.Fatal("statFile must not be called when BIFROST_FILE is set")
		return nil, nil
	}

	got := ResolvePath("")
	if got != "/env/bifrost.yml" {
		t.Errorf("got %q, want /env/bifrost.yml", got)
	}
}

func TestResolvePath_FlagWinsOverEnvVar(t *testing.T) {
	t.Setenv("BIFROST_FILE", "/env/bifrost.yml")

	got := ResolvePath("/flag/bifrost.yml")
	if got != "/flag/bifrost.yml" {
		t.Errorf("got %q, want /flag/bifrost.yml", got)
	}
}

func TestResolvePath_EnvVarWinsOverAutoDiscovery(t *testing.T) {
	t.Setenv("BIFROST_FILE", "/env/bifrost.yml")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) {
		t.Fatal("statFile must not be called when BIFROST_FILE is set")
		return nil, nil
	}

	got := ResolvePath("")
	if got != "/env/bifrost.yml" {
		t.Errorf("got %q, want /env/bifrost.yml", got)
	}
}

func TestResolvePath_EmptyEnvVarFallsThrough(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) { return nil, errors.New("not found") }

	got := ResolvePath("")
	if got != ".bifrost.yml" {
		t.Errorf("got %q, want .bifrost.yml", got)
	}
}

func TestResolvePath_WhitespaceOnlyEnvVarFallsThrough(t *testing.T) {
	t.Setenv("BIFROST_FILE", "   ")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) { return nil, errors.New("not found") }

	got := ResolvePath("")
	if got != ".bifrost.yml" {
		t.Errorf("got %q, want .bifrost.yml", got)
	}
}

func TestResolvePath_XdgConfigUsedWhenExists(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(name string) (os.FileInfo, error) {
		if name == ".config/bifrost.yml" {
			return nil, nil
		}
		return nil, errors.New("not found")
	}

	got := ResolvePath("")
	if got != ".config/bifrost.yml" {
		t.Errorf("got %q, want .config/bifrost.yml", got)
	}
}

func TestResolvePath_DotfileFallback(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) { return nil, errors.New("not found") }

	got := ResolvePath("")
	if got != ".bifrost.yml" {
		t.Errorf("got %q, want .bifrost.yml", got)
	}
}

func TestResolvePath_XdgWinsWhenBothFilesPresent(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(name string) (os.FileInfo, error) {
		// both files "exist"
		return nil, nil
	}

	got := ResolvePath("")
	if got != ".config/bifrost.yml" {
		t.Errorf("got %q, want .config/bifrost.yml", got)
	}
}

func TestInitDest_ConfigDirExists(t *testing.T) {
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(name string) (os.FileInfo, error) {
		if name == ".config" {
			return nil, nil
		}
		return nil, errors.New("not found")
	}

	got := InitDest()
	if got != ".config/bifrost.yml" {
		t.Errorf("got %q, want .config/bifrost.yml", got)
	}
}

func TestInitDest_ConfigDirMissing(t *testing.T) {
	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) { return nil, errors.New("not found") }

	got := InitDest()
	if got != ".bifrost.yml" {
		t.Errorf("got %q, want .bifrost.yml", got)
	}
}
