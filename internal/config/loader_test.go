package config_test

import (
	"strings"
	"testing"

	"github.com/adaouat/bifrost/internal/config"
)

func TestLoad_MinimalFile(t *testing.T) {
	cfg, err := config.Load("../../testdata/bifrost-minimal.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Strategy != "atomic" {
		t.Errorf("strategy default: got %q, want %q", cfg.Strategy, "atomic")
	}
	if cfg.Paths.ReleasesRoot != "/var/www/releases" {
		t.Errorf("releases_root: got %q", cfg.Paths.ReleasesRoot)
	}
	if cfg.Settings.ReleasesToKeep != 10 {
		t.Errorf("releases_to_keep default: got %d, want 10", cfg.Settings.ReleasesToKeep)
	}
}

func TestLoad_FullFile(t *testing.T) {
	cfg, err := config.Load("../../testdata/bifrost-full.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Settings.ReleasesToKeep != 30 {
		t.Errorf("releases_to_keep: got %d, want 30", cfg.Settings.ReleasesToKeep)
	}
	prod, ok := cfg.Environments["prod"]
	if !ok {
		t.Fatal("environment prod not found")
	}
	web, ok := prod.Applications["web"]
	if !ok {
		t.Fatal("application web not found")
	}
	if len(web.Hooks.PostExtract) != 1 {
		t.Errorf("post_extract hooks: got %d, want 1", len(web.Hooks.PostExtract))
	}
	h := web.Hooks.PostExtract[0]
	if h.Cmd != "composer install --no-dev --optimize-autoloader" {
		t.Errorf("hook cmd: got %q", h.Cmd)
	}
	if h.Priority == nil || *h.Priority != 99999 {
		t.Errorf("hook priority default: want 99999, got %v", h.Priority)
	}
	if len(web.Hooks.PostEnableRelease) != 2 {
		t.Errorf("post_enable_release hooks: got %d, want 2", len(web.Hooks.PostEnableRelease))
	}
	if p := web.Hooks.PostEnableRelease[0].Priority; p == nil || *p != 10 {
		t.Errorf("first hook priority: want 10, got %v", p)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/.bifrost.yml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := config.Parse(strings.NewReader(":\tbad:\tyaml"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParse_UnknownField(t *testing.T) {
	_, err := config.Parse(strings.NewReader(`
unknown_field: oops
paths:
  releases_root: /x
  shared_root: /y
`))
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestParse_DefaultStrategy(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /x
  shared_root: /y
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Strategy != "atomic" {
		t.Errorf("strategy default: got %q, want %q", cfg.Strategy, "atomic")
	}
}

func TestParse_DefaultReleasesToKeep(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /x
  shared_root: /y
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Settings.ReleasesToKeep != 10 {
		t.Errorf("releases_to_keep default: got %d, want 10", cfg.Settings.ReleasesToKeep)
	}
}

func TestParse_HookPriorityDefault(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /x
  shared_root: /y
hooks:
  post_extract:
    - cmd: "echo hello"
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := cfg.Hooks.PostExtract[0]
	if h.Priority == nil || *h.Priority != 99999 {
		t.Errorf("hook priority default: want 99999, got %v", h.Priority)
	}
}

func TestParse_HookPriorityExplicitZero(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /x
  shared_root: /y
hooks:
  post_extract:
    - cmd: "echo first"
      priority: 0
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := cfg.Hooks.PostExtract[0]
	if h.Priority == nil || *h.Priority != 0 {
		t.Errorf("hook priority explicit 0: want 0, got %v", h.Priority)
	}
}

func TestParse_HookPriorityInNestedApp(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /x
  shared_root: /y
environments:
  prod:
    applications:
      web:
        hooks:
          pre_enable_release:
            - cmd: "echo migrate"
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := cfg.Environments["prod"].Applications["web"].Hooks.PreEnableRelease[0]
	if h.Priority == nil || *h.Priority != 99999 {
		t.Errorf("nested hook priority default: want 99999, got %v", h.Priority)
	}
}
