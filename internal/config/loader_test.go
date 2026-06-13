package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/adaouat/bifrost/internal/cmderr"
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
	if len(web.Hooks.PostActivate) != 2 {
		t.Errorf("post_activate hooks: got %d, want 2", len(web.Hooks.PostActivate))
	}
	if p := web.Hooks.PostActivate[0].Priority; p == nil || *p != 10 {
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

func TestParse_RejectsOldHookKeyNames(t *testing.T) {
	_, err := config.Parse(strings.NewReader(`
strategy: atomic
paths:
  releases_root: /var/www/releases
  shared_root: /var/www/shared
hooks:
  pre_enable_release:
    - cmd: "echo old"
`))
	if err == nil {
		t.Fatal("expected error for old hook key name pre_enable_release")
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

func TestParse_ServerBlock_DefaultsApplied(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
servers:
  web-01:
    host: 192.168.1.10
    user: deploy
paths:
  releases_root: /x
  shared_root: /y
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	srv, ok := cfg.Servers["web-01"]
	if !ok {
		t.Fatal("expected web-01 in Servers map")
	}
	if srv.Host != "192.168.1.10" {
		t.Errorf("Host: got %q", srv.Host)
	}
	if srv.User != "deploy" {
		t.Errorf("User: got %q", srv.User)
	}
	if srv.Port != 22 {
		t.Errorf("Port default: got %d, want 22", srv.Port)
	}
	if srv.StagingDir != "/tmp" {
		t.Errorf("StagingDir default: got %q, want /tmp", srv.StagingDir)
	}
}

func TestParse_ServerBlock_ExplicitPort(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
servers:
  web-01:
    host: 192.168.1.10
    port: 2222
    user: deploy
paths:
  releases_root: /x
  shared_root: /y
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Servers["web-01"].Port != 2222 {
		t.Errorf("Port: got %d, want 2222", cfg.Servers["web-01"].Port)
	}
}

func TestValidateServerRefs_Valid(t *testing.T) {
	cfg := &config.Config{
		Servers: map[string]config.Server{
			"web-01": {Host: "1.2.3.4", User: "deploy"},
		},
		Environments: map[string]config.Environment{
			"prod": {
				Servers: []string{"web-01"},
				Applications: map[string]config.Application{
					"web": {},
				},
			},
		},
	}
	errs := config.ValidateServerRefs(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateServerRefs_UnknownRef(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"prod": {
				Servers: []string{"nonexistent"},
				Applications: map[string]config.Application{
					"web": {},
				},
			},
		},
	}
	errs := config.ValidateServerRefs(cfg)
	if len(errs) == 0 {
		t.Fatal("expected error for unknown server reference")
	}
}

func TestValidateServerRefs_AppUnknownRef(t *testing.T) {
	cfg := &config.Config{
		Servers: map[string]config.Server{
			"web-01": {Host: "1.2.3.4", User: "deploy"},
		},
		Environments: map[string]config.Environment{
			"prod": {
				Applications: map[string]config.Application{
					"web": {Servers: []string{"nonexistent"}},
				},
			},
		},
	}
	errs := config.ValidateServerRefs(cfg)
	if len(errs) == 0 {
		t.Fatal("expected error for unknown server reference in app")
	}
}

func TestValidateServerRefs_MissingHost(t *testing.T) {
	cfg := &config.Config{
		Servers: map[string]config.Server{
			"web-01": {User: "deploy"},
		},
	}
	errs := config.ValidateServerRefs(cfg)
	if len(errs) == 0 {
		t.Fatal("expected error for server missing host")
	}
}

func TestValidateServerRefs_MissingUser(t *testing.T) {
	cfg := &config.Config{
		Servers: map[string]config.Server{
			"web-01": {Host: "1.2.3.4"},
		},
	}
	errs := config.ValidateServerRefs(cfg)
	if len(errs) == 0 {
		t.Fatal("expected error for server missing user")
	}
}

func TestValidateServerRefs_NoServersNoEnvRefs(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"prod": {
				Applications: map[string]config.Application{"web": {}},
			},
		},
	}
	errs := config.ValidateServerRefs(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors when no servers configured, got: %v", errs)
	}
}

func TestIsFlat_WithEnvironments(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"prod": {},
		},
	}
	if config.IsFlat(cfg) {
		t.Error("expected IsFlat=false when environments are present")
	}
}

func TestIsFlat_WithoutEnvironments(t *testing.T) {
	cfg := &config.Config{}
	if !config.IsFlat(cfg) {
		t.Error("expected IsFlat=true when no environments")
	}
}

func TestParse_FlatConfig(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(`
paths:
  releases_root: /var/www/releases
  shared_root: /var/www/shared
settings:
  releases_to_keep: 5
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Paths.ReleasesRoot != "/var/www/releases" {
		t.Errorf("releases_root: got %q", cfg.Paths.ReleasesRoot)
	}
	if len(cfg.Environments) != 0 {
		t.Errorf("expected no environments, got %d", len(cfg.Environments))
	}
}

func TestLoad_InvalidServerRef_ExitCode2(t *testing.T) {
	_, err := config.Load("../../testdata/bifrost-invalid-server-ref.yml")
	if err == nil {
		t.Fatal("expected error for invalid server reference")
	}
	var exitErr *cmderr.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *cmderr.ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 2 {
		t.Errorf("exit code: got %d, want 2", exitErr.Code)
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
          pre_activate:
            - cmd: "echo migrate"
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := cfg.Environments["prod"].Applications["web"].Hooks.PreActivate[0]
	if h.Priority == nil || *h.Priority != 99999 {
		t.Errorf("nested hook priority default: want 99999, got %v", h.Priority)
	}
}
