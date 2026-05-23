package config

import (
	"fmt"
	"sort"
)

// MergedConfig is the fully resolved configuration for a single env+app deployment.
type MergedConfig struct {
	Strategy     string
	ReleasesRoot string
	SharedRoot   string
	SharedDirs   []string
	SharedFiles  []string
	Settings     Settings
	Variables    map[string]string
	Hooks        Hooks
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

	if m.ReleasesRoot == "" {
		return nil, fmt.Errorf("paths.releases_root is required but not set for %q/%q", envName, appName)
	}
	if m.SharedRoot == "" {
		return nil, fmt.Errorf("paths.shared_root is required but not set for %q/%q", envName, appName)
	}

	return m, nil
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
