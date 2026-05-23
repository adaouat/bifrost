package config

import (
	"fmt"
	"sort"
)

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
			PostExtract:       sortedHooks(cfg.Hooks.PostExtract, env.Hooks.PostExtract, app.Hooks.PostExtract),
			PreLink:           sortedHooks(cfg.Hooks.PreLink, env.Hooks.PreLink, app.Hooks.PreLink),
			PreEnableRelease:  sortedHooks(cfg.Hooks.PreEnableRelease, env.Hooks.PreEnableRelease, app.Hooks.PreEnableRelease),
			PostEnableRelease: sortedHooks(cfg.Hooks.PostEnableRelease, env.Hooks.PostEnableRelease, app.Hooks.PostEnableRelease),
		},
	}

	return m, nil
}

// Validate returns a message for each required field missing from a MergedConfig.
func Validate(m *MergedConfig) []string {
	var errs []string
	if m.ReleasesRoot == "" {
		errs = append(errs, "paths.releases_root: required but not set")
	}
	if m.SharedRoot == "" {
		errs = append(errs, "paths.shared_root: required but not set")
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
