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
	PostExtract       []HookEntry `yaml:"post_extract"        json:"post_extract,omitempty"`
	PreLink           []HookEntry `yaml:"pre_link"            json:"pre_link,omitempty"`
	PreEnableRelease  []HookEntry `yaml:"pre_enable_release"  json:"pre_enable_release,omitempty"`
	PostEnableRelease []HookEntry `yaml:"post_enable_release" json:"post_enable_release,omitempty"`
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
	Paths     Paths             `yaml:"paths"     json:"paths,omitempty"`
	Settings  Settings          `yaml:"settings"  json:"settings,omitempty"`
	Variables map[string]string `yaml:"variables" json:"variables,omitempty"`
	Hooks     Hooks             `yaml:"hooks"     json:"hooks,omitempty"`
}

// Environment groups applications that share common config overrides.
type Environment struct {
	Name         string                 `yaml:"name"         json:"name,omitempty"`
	Paths        Paths                  `yaml:"paths"        json:"paths,omitempty"`
	Settings     Settings               `yaml:"settings"     json:"settings,omitempty"`
	Variables    map[string]string      `yaml:"variables"    json:"variables,omitempty"`
	Hooks        Hooks                  `yaml:"hooks"        json:"hooks,omitempty"`
	Applications map[string]Application `yaml:"applications" json:"applications,omitempty"`
}

// Config is the top-level structure parsed from a .bifrost.yml file.
type Config struct {
	Strategy     string                 `yaml:"strategy"     json:"strategy"`
	Paths        Paths                  `yaml:"paths"        json:"paths"`
	Settings     Settings               `yaml:"settings"     json:"settings"`
	Variables    map[string]string      `yaml:"variables"    json:"variables,omitempty"`
	Hooks        Hooks                  `yaml:"hooks"        json:"hooks,omitempty"`
	Environments map[string]Environment `yaml:"environments" json:"environments,omitempty"`
}
