package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

// GenerateFlatConfig merges global < env < app for the given (env, app) pair
// and writes a flat YAML config to w. The output has no environments: or
// servers: sections; it is the config the remote agent receives in SSH client mode.
func GenerateFlatConfig(cfg *Config, envName, appName string, w io.Writer) error {
	merged, err := Merge(cfg, envName, appName)
	if err != nil {
		return err
	}

	flat := &Config{
		Strategy: merged.Strategy,
		Paths: Paths{
			ReleasesRoot: merged.ReleasesRoot,
			SharedRoot:   merged.SharedRoot,
			Shared: SharedPaths{
				Directories: merged.SharedDirs,
				Files:       merged.SharedFiles,
			},
		},
		Settings:  merged.Settings,
		Variables: merged.Variables,
		Hooks:     merged.Hooks,
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(flat); err != nil {
		return err
	}
	return enc.Close()
}
