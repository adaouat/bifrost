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
	_ = h.PreExtract
	_ = h.PostExtract
	_ = h.PreLink
	_ = h.PostLink
	_ = h.PreActivate
	_ = h.PostActivate
	_ = h.PrePurge
	_ = h.PostPurge
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

func TestServer_Fields(t *testing.T) {
	s := config.Server{
		Host:       "192.168.1.1",
		Port:       22,
		User:       "deploy",
		KeyFile:    "~/.ssh/id_rsa",
		StagingDir: "/tmp",
	}
	if s.Host != "192.168.1.1" {
		t.Errorf("Host: got %q", s.Host)
	}
	if s.Port != 22 {
		t.Errorf("Port: got %d, want 22", s.Port)
	}
	if s.User != "deploy" {
		t.Errorf("User: got %q", s.User)
	}
	if s.KeyFile != "~/.ssh/id_rsa" {
		t.Errorf("KeyFile: got %q", s.KeyFile)
	}
	if s.StagingDir != "/tmp" {
		t.Errorf("StagingDir: got %q", s.StagingDir)
	}
}

func TestConfig_ServersMap(t *testing.T) {
	c := config.Config{
		Servers: map[string]config.Server{
			"web-01": {Host: "1.2.3.4", Port: 22, User: "deploy"},
		},
	}
	srv, ok := c.Servers["web-01"]
	if !ok {
		t.Fatal("expected web-01 in Servers map")
	}
	if srv.Host != "1.2.3.4" {
		t.Errorf("Host: got %q", srv.Host)
	}
}

func TestEnvironment_HasServers(t *testing.T) {
	e := config.Environment{Servers: []string{"web-01", "web-02"}}
	if len(e.Servers) != 2 {
		t.Errorf("Servers length: got %d, want 2", len(e.Servers))
	}
	if e.Servers[0] != "web-01" {
		t.Errorf("Servers[0]: got %q, want web-01", e.Servers[0])
	}
}

func TestApplication_HasServers(t *testing.T) {
	a := config.Application{Servers: []string{"web-01"}}
	if len(a.Servers) != 1 || a.Servers[0] != "web-01" {
		t.Errorf("Servers: got %v, want [web-01]", a.Servers)
	}
}
