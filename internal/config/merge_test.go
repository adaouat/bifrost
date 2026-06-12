package config_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/config"
)

// base returns a minimal valid Config with global paths set.
func base() *config.Config {
	return &config.Config{
		Strategy: "atomic",
		Paths: config.Paths{
			ReleasesRoot: "/global/releases",
			SharedRoot:   "/global/shared",
		},
		Settings: config.Settings{ReleasesToKeep: 10},
		Environments: map[string]config.Environment{
			"prod": {
				Applications: map[string]config.Application{
					"web": {},
				},
			},
		},
	}
}

func TestMerge_PathScalarsOverride(t *testing.T) {
	cfg := base()
	cfg.Environments["prod"] = config.Environment{
		Paths: config.Paths{ReleasesRoot: "/env/releases"},
		Applications: map[string]config.Application{
			"web": {
				Paths: config.Paths{SharedRoot: "/app/shared"},
			},
		},
	}

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// env overrides global releases_root
	if m.ReleasesRoot != "/env/releases" {
		t.Errorf("releases_root: got %q, want /env/releases", m.ReleasesRoot)
	}
	// app overrides global shared_root
	if m.SharedRoot != "/app/shared" {
		t.Errorf("shared_root: got %q, want /app/shared", m.SharedRoot)
	}
}

func TestMerge_SettingsOverride(t *testing.T) {
	cfg := base()
	env := cfg.Environments["prod"]
	env.Settings.ReleasesToKeep = 5
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Settings.ReleasesToKeep != 5 {
		t.Errorf("releases_to_keep: got %d, want 5", m.Settings.ReleasesToKeep)
	}
}

func TestMerge_SettingsAppWins(t *testing.T) {
	cfg := base()
	env := cfg.Environments["prod"]
	env.Settings.ReleasesToKeep = 5
	env.Applications = map[string]config.Application{
		"web": {Settings: config.Settings{ReleasesToKeep: 3}},
	}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Settings.ReleasesToKeep != 3 {
		t.Errorf("releases_to_keep: got %d, want 3", m.Settings.ReleasesToKeep)
	}
}

func TestMerge_VariablesDeepMerge(t *testing.T) {
	cfg := base()
	cfg.Variables = map[string]string{"a": "global-a", "b": "global-b"}
	env := cfg.Environments["prod"]
	env.Variables = map[string]string{"b": "env-b", "c": "env-c"}
	env.Applications = map[string]config.Application{
		"web": {Variables: map[string]string{"c": "app-c", "d": "app-d"}},
	}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cases := map[string]string{
		"a": "global-a", // only in global
		"b": "env-b",    // env overrides global
		"c": "app-c",    // app overrides env
		"d": "app-d",    // only in app
	}
	for k, want := range cases {
		if got := m.Variables[k]; got != want {
			t.Errorf("variables[%q]: got %q, want %q", k, got, want)
		}
	}
}

func TestMerge_HooksConcatAndSort(t *testing.T) {
	p := func(n int) *int { return &n }
	cfg := base()
	cfg.Hooks.PostEnableRelease = []config.HookEntry{
		{Cmd: "global-reload", Priority: p(20)},
	}
	env := cfg.Environments["prod"]
	env.Hooks.PostEnableRelease = []config.HookEntry{
		{Cmd: "env-notify", Priority: p(30)},
	}
	env.Applications = map[string]config.Application{
		"web": {Hooks: config.Hooks{
			PostEnableRelease: []config.HookEntry{
				{Cmd: "app-cache", Priority: p(10)},
			},
		}},
	}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hooks := m.Hooks.PostEnableRelease
	if len(hooks) != 3 {
		t.Fatalf("hook count: got %d, want 3", len(hooks))
	}
	// sorted by priority: app(10) < global(20) < env(30)
	want := []string{"app-cache", "global-reload", "env-notify"}
	for i, h := range hooks {
		if h.Cmd != want[i] {
			t.Errorf("hooks[%d].cmd: got %q, want %q", i, h.Cmd, want[i])
		}
	}
}

func TestMerge_HooksSamePriorityStableOrder(t *testing.T) {
	p := func(n int) *int { return &n }
	cfg := base()
	cfg.Hooks.PostExtract = []config.HookEntry{{Cmd: "global", Priority: p(10)}}
	env := cfg.Environments["prod"]
	env.Hooks.PostExtract = []config.HookEntry{{Cmd: "env", Priority: p(10)}}
	env.Applications = map[string]config.Application{
		"web": {Hooks: config.Hooks{
			PostExtract: []config.HookEntry{{Cmd: "app", Priority: p(10)}},
		}},
	}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// same priority: global first, then env, then app
	want := []string{"global", "env", "app"}
	for i, h := range m.Hooks.PostExtract {
		if h.Cmd != want[i] {
			t.Errorf("hooks[%d].cmd: got %q, want %q", i, h.Cmd, want[i])
		}
	}
}

func TestMerge_SharedPathsConcat(t *testing.T) {
	cfg := base()
	cfg.Paths.Shared.Directories = []string{"var/log"}
	cfg.Paths.Shared.Files = []string{".env"}
	env := cfg.Environments["prod"]
	env.Paths.Shared.Directories = []string{"var/cache"}
	env.Applications = map[string]config.Application{
		"web": {Paths: config.Paths{Shared: config.SharedPaths{
			Files: []string{"config/secrets.yml"},
		}}},
	}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantDirs := []string{"var/log", "var/cache"}
	if len(m.SharedDirs) != len(wantDirs) {
		t.Fatalf("shared dirs: got %v, want %v", m.SharedDirs, wantDirs)
	}
	for i, d := range wantDirs {
		if m.SharedDirs[i] != d {
			t.Errorf("shared dirs[%d]: got %q, want %q", i, m.SharedDirs[i], d)
		}
	}
	wantFiles := []string{".env", "config/secrets.yml"}
	if len(m.SharedFiles) != len(wantFiles) {
		t.Fatalf("shared files: got %v, want %v", m.SharedFiles, wantFiles)
	}
	for i, f := range wantFiles {
		if m.SharedFiles[i] != f {
			t.Errorf("shared files[%d]: got %q, want %q", i, m.SharedFiles[i], f)
		}
	}
}

func TestMerge_MissingEnvironment(t *testing.T) {
	_, err := config.Merge(base(), "staging", "web")
	if err == nil {
		t.Fatal("expected error for missing environment")
	}
}

func TestMerge_MissingApplication(t *testing.T) {
	_, err := config.Merge(base(), "prod", "api")
	if err == nil {
		t.Fatal("expected error for missing application")
	}
}

func TestMerge_MissingReleasesRoot_NoError(t *testing.T) {
	cfg := base()
	cfg.Paths.ReleasesRoot = ""
	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("Merge must not validate required fields: %v", err)
	}
	if m.ReleasesRoot != "" {
		t.Error("expected empty ReleasesRoot to propagate")
	}
}

func TestMerge_MissingSharedRoot_NoError(t *testing.T) {
	cfg := base()
	cfg.Paths.SharedRoot = ""
	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("Merge must not validate required fields: %v", err)
	}
	if m.SharedRoot != "" {
		t.Error("expected empty SharedRoot to propagate")
	}
}

func TestMerge_StrategyPassthrough(t *testing.T) {
	cfg := base()
	cfg.Strategy = "atomic"
	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Strategy != "atomic" {
		t.Errorf("strategy: got %q, want \"atomic\"", m.Strategy)
	}
}

func TestMerge_AllHookTypesPopulated(t *testing.T) {
	p := func(n int) *int { return &n }
	cfg := base()
	cfg.Hooks.PostExtract = []config.HookEntry{{Cmd: "post-extract", Priority: p(10)}}
	cfg.Hooks.PreLink = []config.HookEntry{{Cmd: "pre-link", Priority: p(10)}}
	cfg.Hooks.PreEnableRelease = []config.HookEntry{{Cmd: "pre-enable", Priority: p(10)}}
	cfg.Hooks.PostEnableRelease = []config.HookEntry{{Cmd: "post-enable", Priority: p(10)}}

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Hooks.PostExtract) != 1 || m.Hooks.PostExtract[0].Cmd != "post-extract" {
		t.Error("PostExtract not populated correctly")
	}
	if len(m.Hooks.PreLink) != 1 || m.Hooks.PreLink[0].Cmd != "pre-link" {
		t.Error("PreLink not populated correctly")
	}
	if len(m.Hooks.PreEnableRelease) != 1 || m.Hooks.PreEnableRelease[0].Cmd != "pre-enable" {
		t.Error("PreEnableRelease not populated correctly")
	}
	if len(m.Hooks.PostEnableRelease) != 1 || m.Hooks.PostEnableRelease[0].Cmd != "post-enable" {
		t.Error("PostEnableRelease not populated correctly")
	}
}

func TestMerge_GlobalFallback(t *testing.T) {
	// env and app don't set releases_to_keep — should fall back to global
	m, err := config.Merge(base(), "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Settings.ReleasesToKeep != 10 {
		t.Errorf("releases_to_keep fallback: got %d, want 10", m.Settings.ReleasesToKeep)
	}
}

func TestMerge_ServersNilBothLevels(t *testing.T) {
	m, err := config.Merge(base(), "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Servers) != 0 {
		t.Errorf("expected empty Servers, got %v", m.Servers)
	}
}

func TestMerge_ServersEnvLevel(t *testing.T) {
	cfg := base()
	cfg.Servers = map[string]config.Server{
		"web-01": {Host: "1.2.3.4", Port: 22, User: "deploy"},
	}
	env := cfg.Environments["prod"]
	env.Servers = []string{"web-01"}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(m.Servers))
	}
	if m.Servers[0].Name != "web-01" {
		t.Errorf("server name: got %q, want web-01", m.Servers[0].Name)
	}
	if m.Servers[0].Host != "1.2.3.4" {
		t.Errorf("server host: got %q, want 1.2.3.4", m.Servers[0].Host)
	}
}

func TestMerge_ServersAppOverridesEnv(t *testing.T) {
	cfg := base()
	cfg.Servers = map[string]config.Server{
		"web-01": {Host: "1.2.3.4", Port: 22, User: "deploy"},
		"web-02": {Host: "1.2.3.5", Port: 22, User: "deploy"},
	}
	env := cfg.Environments["prod"]
	env.Servers = []string{"web-01", "web-02"}
	env.Applications = map[string]config.Application{
		"web": {Servers: []string{"web-01"}},
	}
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Servers) != 1 {
		t.Fatalf("expected 1 server (app overrides env), got %d", len(m.Servers))
	}
	if m.Servers[0].Name != "web-01" {
		t.Errorf("server name: got %q, want web-01", m.Servers[0].Name)
	}
}

func TestMerge_ServersNilAppInheritsEnv(t *testing.T) {
	cfg := base()
	cfg.Servers = map[string]config.Server{
		"web-01": {Host: "1.2.3.4", Port: 22, User: "deploy"},
		"web-02": {Host: "1.2.3.5", Port: 22, User: "deploy"},
	}
	env := cfg.Environments["prod"]
	env.Servers = []string{"web-01", "web-02"}
	// app has no Servers set → nil → inherits env
	cfg.Environments["prod"] = env

	m, err := config.Merge(cfg, "prod", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Servers) != 2 {
		t.Fatalf("expected 2 servers (inherited from env), got %d", len(m.Servers))
	}
}

func TestMergeFlat_TopLevelFields(t *testing.T) {
	cfg := &config.Config{
		Strategy: "atomic",
		Paths: config.Paths{
			ReleasesRoot: "/flat/releases",
			SharedRoot:   "/flat/shared",
			Shared: config.SharedPaths{
				Directories: []string{"var/log"},
				Files:       []string{".env"},
			},
		},
		Settings:  config.Settings{ReleasesToKeep: 5},
		Variables: map[string]string{"key": "val"},
	}

	m := config.MergeFlat(cfg)
	if m.ReleasesRoot != "/flat/releases" {
		t.Errorf("ReleasesRoot: got %q", m.ReleasesRoot)
	}
	if m.SharedRoot != "/flat/shared" {
		t.Errorf("SharedRoot: got %q", m.SharedRoot)
	}
	if len(m.SharedDirs) != 1 || m.SharedDirs[0] != "var/log" {
		t.Errorf("SharedDirs: got %v", m.SharedDirs)
	}
	if len(m.SharedFiles) != 1 || m.SharedFiles[0] != ".env" {
		t.Errorf("SharedFiles: got %v", m.SharedFiles)
	}
	if m.Settings.ReleasesToKeep != 5 {
		t.Errorf("ReleasesToKeep: got %d, want 5", m.Settings.ReleasesToKeep)
	}
	if m.Variables["key"] != "val" {
		t.Errorf("Variables: got %v", m.Variables)
	}
}
