package cmd

import (
	"errors"
	"os"
	"testing"
)

func TestResolveConfigPath_ExplicitFlag(t *testing.T) {
	root := NewRootCmd()
	if err := root.PersistentFlags().Set("config", "/custom/path.yml"); err != nil {
		t.Fatalf("setting flag: %v", err)
	}

	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) {
		t.Fatal("statFile must not be called when --config is explicit")
		return nil, nil
	}

	if got := resolveConfigPath(root); got != "/custom/path.yml" {
		t.Errorf("got %q, want /custom/path.yml", got)
	}
}

func TestResolveConfigPath_XdgConfigExists(t *testing.T) {
	root := NewRootCmd()

	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(name string) (os.FileInfo, error) {
		if name == ".config/bifrost.yml" {
			return nil, nil
		}
		return nil, errors.New("not found")
	}

	if got := resolveConfigPath(root); got != ".config/bifrost.yml" {
		t.Errorf("got %q, want .config/bifrost.yml", got)
	}
}

func TestResolveConfigPath_FallbackToDotfile(t *testing.T) {
	root := NewRootCmd()

	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(string) (os.FileInfo, error) { return nil, errors.New("not found") }

	if got := resolveConfigPath(root); got != ".bifrost.yml" {
		t.Errorf("got %q, want .bifrost.yml", got)
	}
}
