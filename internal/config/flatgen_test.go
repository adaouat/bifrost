package config_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/adaouat/bifrost/internal/config"
)

func flatgenBase() *config.Config {
	p := 10
	return &config.Config{
		Strategy: "atomic",
		Servers: map[string]config.ServerConfig{
			"web-01": {Host: "192.168.1.10", Port: 22, User: "deploy"},
		},
		Paths: config.Paths{
			ReleasesRoot: "/global/releases",
			SharedRoot:   "/global/shared",
			Shared: config.SharedPaths{
				Directories: []string{"var/log"},
			},
		},
		Settings:  config.Settings{ReleasesToKeep: 5},
		Variables: map[string]string{"app_env": "production"},
		Environments: map[string]config.Environment{
			"prod": {
				Servers: []string{"web-01"},
				Applications: map[string]config.Application{
					"web": {
						Paths: config.Paths{
							Shared: config.SharedPaths{
								Files: []string{".env"},
							},
						},
						Hooks: config.Hooks{
							PostEnableRelease: []config.HookEntry{
								{Cmd: "systemctl reload nginx", Sudo: true, Priority: &p},
							},
						},
					},
				},
			},
		},
	}
}

func TestGenerateFlatConfig_NoEnvironmentsOrServers(t *testing.T) {
	cfg := flatgenBase()
	var buf bytes.Buffer
	if err := config.GenerateFlatConfig(cfg, "prod", "web", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "environments:") {
		t.Errorf("output contains 'environments:' key:\n%s", out)
	}
	if strings.Contains(out, "servers:") {
		t.Errorf("output contains 'servers:' key:\n%s", out)
	}
}

func TestGenerateFlatConfig_MergedValuesAtTopLevel(t *testing.T) {
	cfg := flatgenBase()
	var buf bytes.Buffer
	if err := config.GenerateFlatConfig(cfg, "prod", "web", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := config.Parse(&buf)
	if err != nil {
		t.Fatalf("parsing flat config: %v", err)
	}

	if parsed.Paths.ReleasesRoot != "/global/releases" {
		t.Errorf("releases_root: got %q, want /global/releases", parsed.Paths.ReleasesRoot)
	}
	if parsed.Paths.SharedRoot != "/global/shared" {
		t.Errorf("shared_root: got %q, want /global/shared", parsed.Paths.SharedRoot)
	}
	if parsed.Settings.ReleasesToKeep != 5 {
		t.Errorf("releases_to_keep: got %d, want 5", parsed.Settings.ReleasesToKeep)
	}
	if parsed.Variables["app_env"] != "production" {
		t.Errorf("variables.app_env: got %q, want production", parsed.Variables["app_env"])
	}
}

func TestGenerateFlatConfig_SharedPathsMerged(t *testing.T) {
	cfg := flatgenBase()
	var buf bytes.Buffer
	if err := config.GenerateFlatConfig(cfg, "prod", "web", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := config.Parse(&buf)
	if err != nil {
		t.Fatalf("parsing flat config: %v", err)
	}

	if len(parsed.Paths.Shared.Directories) != 1 || parsed.Paths.Shared.Directories[0] != "var/log" {
		t.Errorf("shared.directories: got %v, want [var/log]", parsed.Paths.Shared.Directories)
	}
	if len(parsed.Paths.Shared.Files) != 1 || parsed.Paths.Shared.Files[0] != ".env" {
		t.Errorf("shared.files: got %v, want [.env]", parsed.Paths.Shared.Files)
	}
}

func TestGenerateFlatConfig_HooksMerged(t *testing.T) {
	cfg := flatgenBase()
	var buf bytes.Buffer
	if err := config.GenerateFlatConfig(cfg, "prod", "web", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := config.Parse(&buf)
	if err != nil {
		t.Fatalf("parsing flat config: %v", err)
	}

	if len(parsed.Hooks.PostEnableRelease) != 1 {
		t.Fatalf("post_enable_release: got %d hooks, want 1", len(parsed.Hooks.PostEnableRelease))
	}
	if parsed.Hooks.PostEnableRelease[0].Cmd != "systemctl reload nginx" {
		t.Errorf("hook cmd: got %q, want systemctl reload nginx", parsed.Hooks.PostEnableRelease[0].Cmd)
	}
	if !parsed.Hooks.PostEnableRelease[0].Sudo {
		t.Error("hook sudo: got false, want true")
	}
}

func TestGenerateFlatConfig_IsFlat(t *testing.T) {
	cfg := flatgenBase()
	var buf bytes.Buffer
	if err := config.GenerateFlatConfig(cfg, "prod", "web", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := config.Parse(&buf)
	if err != nil {
		t.Fatalf("parsing flat config: %v", err)
	}

	if !config.IsFlat(parsed) {
		t.Error("IsFlat: got false, want true — flat config must have no environments")
	}
}

func TestGenerateFlatConfig_UnknownEnv(t *testing.T) {
	cfg := flatgenBase()
	var buf bytes.Buffer
	err := config.GenerateFlatConfig(cfg, "staging", "web", &buf)
	if err == nil {
		t.Error("expected error for unknown environment, got nil")
	}
}

// flatgenLayered populates every mergeable field at all three levels (global,
// env, app) so precedence can be verified end-to-end through the generator.
func flatgenLayered() *config.Config {
	p := func(n int) *int { return &n }
	return &config.Config{
		Strategy: "atomic",
		Paths: config.Paths{
			ReleasesRoot: "/global/releases",
			SharedRoot:   "/global/shared",
			Shared: config.SharedPaths{
				Directories: []string{"var/log"},
				Files:       []string{".env"},
			},
		},
		Settings:  config.Settings{ReleasesToKeep: 10},
		Variables: map[string]string{"a": "global-a", "b": "global-b"},
		Hooks: config.Hooks{
			PostEnableRelease: []config.HookEntry{
				{Cmd: "global-reload", Priority: p(20)},
			},
		},
		Environments: map[string]config.Environment{
			"prod": {
				Paths: config.Paths{
					ReleasesRoot: "/env/releases",
					Shared:       config.SharedPaths{Directories: []string{"var/cache"}},
				},
				Settings:  config.Settings{ReleasesToKeep: 7},
				Variables: map[string]string{"b": "env-b", "c": "env-c"},
				Hooks: config.Hooks{
					PostEnableRelease: []config.HookEntry{
						{Cmd: "env-notify", Priority: p(30)},
					},
				},
				Applications: map[string]config.Application{
					"web": {
						Paths: config.Paths{
							SharedRoot: "/app/shared",
							Shared:     config.SharedPaths{Files: []string{"config/secrets.yml"}},
						},
						Variables: map[string]string{"c": "app-c", "d": "app-d"},
						Hooks: config.Hooks{
							PostEnableRelease: []config.HookEntry{
								{Cmd: "app-cache", Priority: p(10)},
							},
						},
					},
				},
			},
		},
	}
}

func flatten(t *testing.T, cfg *config.Config) *config.Config {
	t.Helper()
	var buf bytes.Buffer
	if err := config.GenerateFlatConfig(cfg, "prod", "web", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := config.Parse(&buf)
	if err != nil {
		t.Fatalf("parsing flat config: %v", err)
	}
	return parsed
}

func TestGenerateFlatConfig_ScalarOverridePrecedence(t *testing.T) {
	flat := flatten(t, flatgenLayered())

	if flat.Paths.ReleasesRoot != "/env/releases" {
		t.Errorf("releases_root: got %q, want /env/releases (env overrides global)", flat.Paths.ReleasesRoot)
	}
	if flat.Paths.SharedRoot != "/app/shared" {
		t.Errorf("shared_root: got %q, want /app/shared (app overrides global)", flat.Paths.SharedRoot)
	}
	if flat.Settings.ReleasesToKeep != 7 {
		t.Errorf("releases_to_keep: got %d, want 7 (env overrides global)", flat.Settings.ReleasesToKeep)
	}
}

func TestGenerateFlatConfig_VariablesPrecedence(t *testing.T) {
	flat := flatten(t, flatgenLayered())

	cases := map[string]string{
		"a": "global-a", // only in global
		"b": "env-b",    // env overrides global
		"c": "app-c",    // app overrides env
		"d": "app-d",    // only in app
	}
	for k, want := range cases {
		if got := flat.Variables[k]; got != want {
			t.Errorf("variables[%q]: got %q, want %q", k, got, want)
		}
	}
}

func TestGenerateFlatConfig_SharedPathsConcatAllLevels(t *testing.T) {
	flat := flatten(t, flatgenLayered())

	wantDirs := []string{"var/log", "var/cache"}
	if len(flat.Paths.Shared.Directories) != len(wantDirs) {
		t.Fatalf("shared.directories: got %v, want %v", flat.Paths.Shared.Directories, wantDirs)
	}
	for i, d := range wantDirs {
		if flat.Paths.Shared.Directories[i] != d {
			t.Errorf("shared.directories[%d]: got %q, want %q", i, flat.Paths.Shared.Directories[i], d)
		}
	}

	wantFiles := []string{".env", "config/secrets.yml"}
	if len(flat.Paths.Shared.Files) != len(wantFiles) {
		t.Fatalf("shared.files: got %v, want %v", flat.Paths.Shared.Files, wantFiles)
	}
	for i, f := range wantFiles {
		if flat.Paths.Shared.Files[i] != f {
			t.Errorf("shared.files[%d]: got %q, want %q", i, flat.Paths.Shared.Files[i], f)
		}
	}
}

func TestGenerateFlatConfig_HooksSortedAcrossLevels(t *testing.T) {
	flat := flatten(t, flatgenLayered())

	hooks := flat.Hooks.PostEnableRelease
	if len(hooks) != 3 {
		t.Fatalf("post_enable_release: got %d hooks, want 3", len(hooks))
	}
	// sorted by priority: app(10) < global(20) < env(30)
	want := []string{"app-cache", "global-reload", "env-notify"}
	for i, h := range hooks {
		if h.Cmd != want[i] {
			t.Errorf("hooks[%d].cmd: got %q, want %q", i, h.Cmd, want[i])
		}
	}
}
