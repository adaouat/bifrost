package config

import (
	"fmt"
	"io"
	"strings"

	"github.com/adaouat/bifrost/internal/cmderr"
	forgeconfig "github.com/adaouat/forge/config"
)

// Load reads and strictly parses a .bifrost.yml file at the given path.
// Returns *cmderr.ExitError with code 2 if server references are invalid.
func Load(path string) (*Config, error) {
	var cfg Config
	if err := forgeconfig.Load(path, &cfg); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)
	var errs forgeconfig.ValidationErrors
	errs = append(errs, ValidateServerRefs(&cfg)...)
	errs = append(errs, ValidateHookCmds(&cfg)...)
	if len(errs) > 0 {
		return nil, &cmderr.ExitError{Code: cmderr.Config, Message: errs.Error()}
	}
	return &cfg, nil
}

// IsFlat reports whether cfg has no environments defined.
// A flat config is the merged result used by the remote agent — it has no environments or server refs.
func IsFlat(cfg *Config) bool {
	return len(cfg.Environments) == 0
}

// ValidateServerRefs checks that:
//   - every Server in cfg.Servers has a non-empty Host and User
//   - every name in env/app Servers lists exists in cfg.Servers
//
// Returns a slice of human-readable error messages; empty means valid.
func ValidateServerRefs(cfg *Config) forgeconfig.ValidationErrors {
	var errs forgeconfig.ValidationErrors

	for name, srv := range cfg.Servers {
		if srv.Host == "" {
			errs = append(errs, forgeconfig.ValidationError{Path: "servers." + name, Message: "host is required"})
		}
		if srv.User == "" {
			errs = append(errs, forgeconfig.ValidationError{Path: "servers." + name, Message: "user is required"})
		}
	}

	for envName, env := range cfg.Environments {
		for _, ref := range env.Servers {
			if _, ok := cfg.Servers[ref]; !ok {
				errs = append(errs, forgeconfig.ValidationError{
					Path:    fmt.Sprintf("environments.%s.servers", envName),
					Message: fmt.Sprintf("unknown server %q", ref),
				})
			}
		}
		for appName, app := range env.Applications {
			for _, ref := range app.Servers {
				if _, ok := cfg.Servers[ref]; !ok {
					errs = append(errs, forgeconfig.ValidationError{
						Path:    fmt.Sprintf("environments.%s.applications.%s.servers", envName, appName),
						Message: fmt.Sprintf("unknown server %q", ref),
					})
				}
			}
		}
	}

	return errs
}

// ValidateHookCmds reports an error for every hook entry with an empty cmd, at
// any config level (global, environment, application).
func ValidateHookCmds(cfg *Config) forgeconfig.ValidationErrors {
	var errs forgeconfig.ValidationErrors
	check := func(path string, h Hooks) {
		for _, st := range hookStages(h) {
			for i, e := range st.hooks {
				if strings.TrimSpace(e.Cmd) == "" {
					errs = append(errs, forgeconfig.ValidationError{
						Path:    fmt.Sprintf("%s.%s[%d]", path, st.name, i),
						Message: "cmd is required",
					})
				}
			}
		}
	}
	check("hooks", cfg.Hooks)
	for name, env := range cfg.Environments {
		check(fmt.Sprintf("environments.%s.hooks", name), env.Hooks)
		for appName, app := range env.Applications {
			check(fmt.Sprintf("environments.%s.applications.%s.hooks", name, appName), app.Hooks)
		}
	}
	return errs
}

type namedHookList struct {
	name  string
	hooks []HookEntry
}

// hookStages pairs each hook list in h with its lifecycle name.
func hookStages(h Hooks) []namedHookList {
	return []namedHookList{
		{"pre_extract", h.PreExtract},
		{"post_extract", h.PostExtract},
		{"pre_link", h.PreLink},
		{"post_link", h.PostLink},
		{"pre_activate", h.PreActivate},
		{"post_activate", h.PostActivate},
		{"pre_purge", h.PrePurge},
		{"post_purge", h.PostPurge},
	}
}

// Parse strictly parses .bifrost.yml content from r and applies defaults.
func Parse(r io.Reader) (*Config, error) {
	var cfg Config
	if err := forgeconfig.Decode(r, &cfg); err != nil {
		return nil, err
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
	applyServerDefaults(cfg)
	applyHookDefaults(&cfg.Hooks)
	for envName, env := range cfg.Environments {
		applyHookDefaults(&env.Hooks)
		for appName, app := range env.Applications {
			applyHookDefaults(&app.Hooks)
			env.Applications[appName] = app
		}
		cfg.Environments[envName] = env
	}
}

func applyServerDefaults(cfg *Config) {
	for name, srv := range cfg.Servers {
		if srv.Port == 0 {
			srv.Port = 22
		}
		if srv.StagingDir == "" {
			srv.StagingDir = "/tmp"
		}
		cfg.Servers[name] = srv
	}
}

func applyHookDefaults(h *Hooks) {
	setDefaultPriority(h.PreExtract)
	setDefaultPriority(h.PostExtract)
	setDefaultPriority(h.PreLink)
	setDefaultPriority(h.PostLink)
	setDefaultPriority(h.PreActivate)
	setDefaultPriority(h.PostActivate)
	setDefaultPriority(h.PrePurge)
	setDefaultPriority(h.PostPurge)
}

func setDefaultPriority(hooks []HookEntry) {
	for i := range hooks {
		if hooks[i].Priority == nil {
			p := 99999
			hooks[i].Priority = &p
		}
	}
}
