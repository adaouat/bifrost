package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and strictly parses a .bifrost.yml file at the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

// Parse strictly parses .bifrost.yml content from r and applies defaults.
func Parse(r io.Reader) (*Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Strategy == "" {
		cfg.Strategy = "atomic"
	}
	if cfg.Settings.ReleasesToKeep == 0 {
		cfg.Settings.ReleasesToKeep = 10
	}
	applyHookDefaults(&cfg.Hooks)
	for envName, env := range cfg.Environments {
		if env.Settings.ReleasesToKeep == 0 {
			env.Settings.ReleasesToKeep = 10
		}
		applyHookDefaults(&env.Hooks)
		for appName, app := range env.Applications {
			if app.Settings.ReleasesToKeep == 0 {
				app.Settings.ReleasesToKeep = 10
			}
			applyHookDefaults(&app.Hooks)
			env.Applications[appName] = app
		}
		cfg.Environments[envName] = env
	}
}

func applyHookDefaults(h *Hooks) {
	setDefaultPriority(h.PostExtract)
	setDefaultPriority(h.PreLink)
	setDefaultPriority(h.PreEnableRelease)
	setDefaultPriority(h.PostEnableRelease)
}

func setDefaultPriority(hooks []HookEntry) {
	for i := range hooks {
		if hooks[i].Priority == nil {
			p := 99999
			hooks[i].Priority = &p
		}
	}
}
