package config_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/config"
)

func TestValidate_OK(t *testing.T) {
	m := &config.MergedConfig{
		ReleasesRoot: "/var/www/releases",
		SharedRoot:   "/var/www/shared",
	}
	errs := config.Validate(m)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_MissingReleasesRoot(t *testing.T) {
	m := &config.MergedConfig{SharedRoot: "/var/www/shared"}
	errs := config.Validate(m)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidate_MissingSharedRoot(t *testing.T) {
	m := &config.MergedConfig{ReleasesRoot: "/var/www/releases"}
	errs := config.Validate(m)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidate_MissingBoth(t *testing.T) {
	m := &config.MergedConfig{}
	errs := config.Validate(m)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidate_UnsupportedStrategy(t *testing.T) {
	m := &config.MergedConfig{
		ReleasesRoot: "/var/www/releases",
		SharedRoot:   "/var/www/shared",
		Strategy:     "docker",
	}
	errs := config.Validate(m)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for unsupported strategy, got %d: %v", len(errs), errs)
	}
}

func TestValidate_AtomicStrategyOK(t *testing.T) {
	m := &config.MergedConfig{
		ReleasesRoot: "/var/www/releases",
		SharedRoot:   "/var/www/shared",
		Strategy:     "atomic",
	}
	if errs := config.Validate(m); len(errs) != 0 {
		t.Errorf("expected no errors for atomic, got: %v", errs)
	}
}

func TestValidateHookCmds_RejectsEmptyCmd(t *testing.T) {
	cfg := &config.Config{
		Hooks: config.Hooks{
			PostExtract: []config.HookEntry{{Cmd: ""}},
		},
	}
	errs := config.ValidateHookCmds(cfg)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for empty hook cmd, got %d: %v", len(errs), errs)
	}
}

func TestValidateHookCmds_RejectsEmptyCmdInApplication(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"prod": {
				Applications: map[string]config.Application{
					"web": {Hooks: config.Hooks{PostActivate: []config.HookEntry{{Cmd: "  "}}}},
				},
			},
		},
	}
	errs := config.ValidateHookCmds(cfg)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for whitespace hook cmd, got %d: %v", len(errs), errs)
	}
}

func TestValidateHookCmds_AllowsNonEmpty(t *testing.T) {
	cfg := &config.Config{
		Hooks: config.Hooks{PostExtract: []config.HookEntry{{Cmd: "echo hi"}}},
		Environments: map[string]config.Environment{
			"prod": {
				Hooks: config.Hooks{PreActivate: []config.HookEntry{{Cmd: "migrate"}}},
				Applications: map[string]config.Application{
					"web": {Hooks: config.Hooks{PostActivate: []config.HookEntry{{Cmd: "reload"}}}},
				},
			},
		},
	}
	if errs := config.ValidateHookCmds(cfg); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}
