package config

// HookEntry is a single shell command executed at a lifecycle point.
type HookEntry struct {
	Cmd         string `yaml:"cmd"          json:"cmd"`
	Priority    *int   `yaml:"priority"     json:"priority"`
	Sudo        bool   `yaml:"sudo"         json:"sudo"`
	CmdDir      string `yaml:"cmd_dir"      json:"cmd_dir,omitempty"`
	AllowFail   bool   `yaml:"allow_fail"   json:"allow_fail"`
	Interactive bool   `yaml:"interactive"  json:"interactive"`
}

// Hooks groups shell commands by deployment lifecycle point.
type Hooks struct {
	PreExtract   []HookEntry `yaml:"pre_extract"   json:"pre_extract,omitempty"`
	PostExtract  []HookEntry `yaml:"post_extract"  json:"post_extract,omitempty"`
	PreLink      []HookEntry `yaml:"pre_link"      json:"pre_link,omitempty"`
	PostLink     []HookEntry `yaml:"post_link"     json:"post_link,omitempty"`
	PreActivate  []HookEntry `yaml:"pre_activate"  json:"pre_activate,omitempty"`
	PostActivate []HookEntry `yaml:"post_activate" json:"post_activate,omitempty"`
	PrePurge     []HookEntry `yaml:"pre_purge"     json:"pre_purge,omitempty"`
	PostPurge    []HookEntry `yaml:"post_purge"    json:"post_purge,omitempty"`
}

// SharedPaths lists relative paths symlinked from shared_root into each release.
type SharedPaths struct {
	Directories []string `yaml:"directories" json:"directories,omitempty"`
	Files       []string `yaml:"files"       json:"files,omitempty"`
}

// Paths holds root directories and shared resource declarations.
type Paths struct {
	ReleasesRoot string      `yaml:"releases_root" json:"releases_root,omitempty"`
	SharedRoot   string      `yaml:"shared_root"   json:"shared_root,omitempty"`
	Shared       SharedPaths `yaml:"shared"        json:"shared,omitempty"`
}

// Settings holds deployment behaviour tunables.
type Settings struct {
	ReleasesToKeep int `yaml:"releases_to_keep" json:"releases_to_keep"`
}

// Application is the innermost config level — one deployable unit.
type Application struct {
	Name      string            `yaml:"name"      json:"name,omitempty"`
	Servers   []string          `yaml:"servers"   json:"servers,omitempty"`
	Paths     Paths             `yaml:"paths"     json:"paths,omitempty"`
	Settings  Settings          `yaml:"settings"  json:"settings,omitempty"`
	Variables map[string]string `yaml:"variables" json:"variables,omitempty"`
	Hooks     Hooks             `yaml:"hooks"     json:"hooks,omitempty"`
}

// Environment groups applications that share common config overrides.
type Environment struct {
	Name         string                 `yaml:"name"         json:"name,omitempty"`
	Servers      []string               `yaml:"servers"      json:"servers,omitempty"`
	Paths        Paths                  `yaml:"paths"        json:"paths,omitempty"`
	Settings     Settings               `yaml:"settings"     json:"settings,omitempty"`
	Variables    map[string]string      `yaml:"variables"    json:"variables,omitempty"`
	Hooks        Hooks                  `yaml:"hooks"        json:"hooks,omitempty"`
	Applications map[string]Application `yaml:"applications" json:"applications,omitempty"`
}

// Server holds SSH connection details for a named target server.
type Server struct {
	Host       string `yaml:"host"        json:"host"`
	Port       int    `yaml:"port"        json:"port,omitempty"`
	User       string `yaml:"user"        json:"user"`
	KeyFile    string `yaml:"key_file"    json:"key_file,omitempty"`
	StagingDir string `yaml:"staging_dir" json:"staging_dir,omitempty"`
}

// Config is the top-level structure parsed from a .bifrost.yml file.
type Config struct {
	Strategy     string                 `yaml:"strategy"     json:"strategy"`
	Servers      map[string]Server      `yaml:"servers,omitempty"     json:"servers,omitempty"`
	Paths        Paths                  `yaml:"paths"        json:"paths"`
	Settings     Settings               `yaml:"settings"     json:"settings"`
	Variables    map[string]string      `yaml:"variables"    json:"variables,omitempty"`
	Hooks        Hooks                  `yaml:"hooks"        json:"hooks,omitempty"`
	Environments map[string]Environment `yaml:"environments,omitempty" json:"environments,omitempty"`
}
