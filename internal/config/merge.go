package config

import (
	"fmt"
	"sort"

	forgeconfig "github.com/adaouat/forge/config"
)

// ResolvedServer is a named server entry resolved from the top-level servers map.
type ResolvedServer struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	KeyFile    string `json:"key_file,omitempty"`
	StagingDir string `json:"staging_dir,omitempty"`
}

// MergedConfig is the fully resolved configuration for a single env+app deployment.
type MergedConfig struct {
	Strategy     string            `json:"strategy"`
	ReleasesRoot string            `json:"releases_root"`
	SharedRoot   string            `json:"shared_root"`
	SharedDirs   []string          `json:"shared_dirs,omitempty"`
	SharedFiles  []string          `json:"shared_files,omitempty"`
	Settings     Settings          `json:"settings"`
	Variables    map[string]string `json:"variables,omitempty"`
	Hooks        Hooks             `json:"hooks,omitempty"`
	Servers      []ResolvedServer  `json:"servers,omitempty"`
}

// Merge resolves the three-level hierarchy (global < env < app) for the given
// environment and application names.
func Merge(cfg *Config, envName, appName string) (*MergedConfig, error) {
	env, ok := cfg.Environments[envName]
	if !ok {
		return nil, fmt.Errorf("environment %q not found", envName)
	}
	app, ok := env.Applications[appName]
	if !ok {
		return nil, fmt.Errorf("application %q not found in environment %q", appName, envName)
	}

	// Server resolution: app-level overrides env-level; nil inherits from parent.
	serverNames := app.Servers
	if serverNames == nil {
		serverNames = env.Servers
	}

	m := &MergedConfig{
		Strategy:     cfg.Strategy,
		ReleasesRoot: firstNonEmpty(app.Paths.ReleasesRoot, env.Paths.ReleasesRoot, cfg.Paths.ReleasesRoot),
		SharedRoot:   firstNonEmpty(app.Paths.SharedRoot, env.Paths.SharedRoot, cfg.Paths.SharedRoot),
		SharedDirs:   concat(cfg.Paths.Shared.Directories, env.Paths.Shared.Directories, app.Paths.Shared.Directories),
		SharedFiles:  concat(cfg.Paths.Shared.Files, env.Paths.Shared.Files, app.Paths.Shared.Files),
		Settings: Settings{
			ReleasesToKeep: firstNonZeroInt(app.Settings.ReleasesToKeep, env.Settings.ReleasesToKeep, cfg.Settings.ReleasesToKeep),
		},
		Variables: mergeMaps(cfg.Variables, env.Variables, app.Variables),
		Hooks: Hooks{
			PreExtract:   sortedHooks(cfg.Hooks.PreExtract, env.Hooks.PreExtract, app.Hooks.PreExtract),
			PostExtract:  sortedHooks(cfg.Hooks.PostExtract, env.Hooks.PostExtract, app.Hooks.PostExtract),
			PreLink:      sortedHooks(cfg.Hooks.PreLink, env.Hooks.PreLink, app.Hooks.PreLink),
			PostLink:     sortedHooks(cfg.Hooks.PostLink, env.Hooks.PostLink, app.Hooks.PostLink),
			PreActivate:  sortedHooks(cfg.Hooks.PreActivate, env.Hooks.PreActivate, app.Hooks.PreActivate),
			PostActivate: sortedHooks(cfg.Hooks.PostActivate, env.Hooks.PostActivate, app.Hooks.PostActivate),
			PrePurge:     sortedHooks(cfg.Hooks.PrePurge, env.Hooks.PrePurge, app.Hooks.PrePurge),
			PostPurge:    sortedHooks(cfg.Hooks.PostPurge, env.Hooks.PostPurge, app.Hooks.PostPurge),
		},
		Servers: resolveServers(cfg.Servers, serverNames),
	}

	return m, nil
}

// MergeFlat converts a flat Config (no environments) directly to a MergedConfig.
// Used when the agent receives a pre-merged config without an environments block.
func MergeFlat(cfg *Config) *MergedConfig {
	return &MergedConfig{
		Strategy:     cfg.Strategy,
		ReleasesRoot: cfg.Paths.ReleasesRoot,
		SharedRoot:   cfg.Paths.SharedRoot,
		SharedDirs:   cfg.Paths.Shared.Directories,
		SharedFiles:  cfg.Paths.Shared.Files,
		Settings:     cfg.Settings,
		Variables:    cfg.Variables,
		Hooks:        cfg.Hooks,
	}
}

// resolveServers returns a ResolvedServer slice for the given server names.
func resolveServers(serverMap map[string]Server, names []string) []ResolvedServer {
	if len(names) == 0 {
		return nil
	}
	resolved := make([]ResolvedServer, 0, len(names))
	for _, name := range names {
		if srv, ok := serverMap[name]; ok {
			resolved = append(resolved, ResolvedServer{
				Name:       name,
				Host:       srv.Host,
				Port:       srv.Port,
				User:       srv.User,
				KeyFile:    srv.KeyFile,
				StagingDir: srv.StagingDir,
			})
		}
	}
	return resolved
}

// Validate returns a ValidationError for each required field missing from a MergedConfig.
func Validate(m *MergedConfig) forgeconfig.ValidationErrors {
	var errs forgeconfig.ValidationErrors
	if m.ReleasesRoot == "" {
		errs = append(errs, forgeconfig.ValidationError{Path: "paths.releases_root", Message: "required but not set"})
	}
	if m.SharedRoot == "" {
		errs = append(errs, forgeconfig.ValidationError{Path: "paths.shared_root", Message: "required but not set"})
	}
	// Empty means unset (defaults to atomic); only an explicit non-atomic value is rejected.
	if m.Strategy != "" && m.Strategy != "atomic" {
		errs = append(errs, forgeconfig.ValidationError{
			Path:    "strategy",
			Message: fmt.Sprintf("unsupported strategy %q (only \"atomic\" is valid in v0)", m.Strategy),
		})
	}
	return errs
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstNonZeroInt(vals ...int) int {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}

func mergeMaps(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func concat(slices ...[]string) []string {
	var out []string
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

func sortedHooks(slices ...[]HookEntry) []HookEntry {
	var out []HookEntry
	for _, s := range slices {
		out = append(out, s...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return *out[i].Priority < *out[j].Priority
	})
	return out
}
