package config_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/config"
)

func TestHookEntry_Fields(t *testing.T) {
	p := 10
	h := config.HookEntry{
		Cmd:         "echo hello",
		Priority:    &p,
		Sudo:        true,
		CmdDir:      "/tmp",
		AllowFail:   true,
		Interactive: true,
	}
	if h.Cmd != "echo hello" {
		t.Errorf("cmd: got %q", h.Cmd)
	}
	if h.Priority == nil || *h.Priority != 10 {
		t.Errorf("priority: got %v, want 10", h.Priority)
	}
	if !h.Sudo {
		t.Error("sudo: got false, want true")
	}
	if h.CmdDir != "/tmp" {
		t.Errorf("cmd_dir: got %q", h.CmdDir)
	}
	if !h.AllowFail {
		t.Error("allow_fail: got false, want true")
	}
	if !h.Interactive {
		t.Error("interactive: got false, want true")
	}
}

func TestHooks_AllLifecyclePoints(t *testing.T) {
	h := config.Hooks{}
	_ = h.PostExtract
	_ = h.PreLink
	_ = h.PreEnableRelease
	_ = h.PostEnableRelease
}

func TestPaths_FlatRoots(t *testing.T) {
	p := config.Paths{
		ReleasesRoot: "/var/www/releases",
		SharedRoot:   "/var/www/shared",
	}
	if p.ReleasesRoot == "" {
		t.Error("ReleasesRoot must be settable")
	}
	if p.SharedRoot == "" {
		t.Error("SharedRoot must be settable")
	}
	_ = p.Shared.Directories
	_ = p.Shared.Files
}

func TestConfig_TopLevel(t *testing.T) {
	c := config.Config{
		Strategy: "atomic",
		Paths: config.Paths{
			ReleasesRoot: "/var/www/releases",
			SharedRoot:   "/var/www/shared",
		},
		Settings: config.Settings{
			ReleasesToKeep: 10,
		},
		Variables: map[string]string{"key": "value"},
		Environments: map[string]config.Environment{
			"prod": {
				Name: "Production",
				Applications: map[string]config.Application{
					"web": {Name: "Web"},
				},
			},
		},
	}
	if c.Strategy != "atomic" {
		t.Errorf("strategy: got %q, want %q", c.Strategy, "atomic")
	}
}

func TestEnvironment_InheritsCommonFields(t *testing.T) {
	e := config.Environment{}
	_ = e.Name
	_ = e.Paths
	_ = e.Settings
	_ = e.Variables
	_ = e.Hooks
	_ = e.Applications
}

func TestApplication_InheritsCommonFields(t *testing.T) {
	a := config.Application{}
	_ = a.Name
	_ = a.Paths
	_ = a.Settings
	_ = a.Variables
	_ = a.Hooks
}
