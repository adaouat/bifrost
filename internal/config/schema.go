package config

// HookEntry is a single shell command executed at a lifecycle point.
type HookEntry struct {
	Cmd         string `yaml:"cmd"`
	Priority    *int   `yaml:"priority"`
	Sudo        bool   `yaml:"sudo"`
	CmdDir      string `yaml:"cmd_dir"`
	AllowFail   bool   `yaml:"allow_fail"`
	Interactive bool   `yaml:"interactive"`
}

// Hooks groups shell commands by deployment lifecycle point.
type Hooks struct {
	PostExtract       []HookEntry `yaml:"post_extract"`
	PreLink           []HookEntry `yaml:"pre_link"`
	PreEnableRelease  []HookEntry `yaml:"pre_enable_release"`
	PostEnableRelease []HookEntry `yaml:"post_enable_release"`
}

// SharedPaths lists relative paths symlinked from shared_root into each release.
type SharedPaths struct {
	Directories []string `yaml:"directories"`
	Files       []string `yaml:"files"`
}

// Paths holds root directories and shared resource declarations.
type Paths struct {
	ReleasesRoot string      `yaml:"releases_root"`
	SharedRoot   string      `yaml:"shared_root"`
	Shared       SharedPaths `yaml:"shared"`
}

// Settings holds deployment behaviour tunables.
type Settings struct {
	ReleasesToKeep int `yaml:"releases_to_keep"`
}

// Application is the innermost config level — one deployable unit.
type Application struct {
	Name      string            `yaml:"name"`
	Paths     Paths             `yaml:"paths"`
	Settings  Settings          `yaml:"settings"`
	Variables map[string]string `yaml:"variables"`
	Hooks     Hooks             `yaml:"hooks"`
}

// Environment groups applications that share common config overrides.
type Environment struct {
	Name         string                 `yaml:"name"`
	Paths        Paths                  `yaml:"paths"`
	Settings     Settings               `yaml:"settings"`
	Variables    map[string]string      `yaml:"variables"`
	Hooks        Hooks                  `yaml:"hooks"`
	Applications map[string]Application `yaml:"applications"`
}

// Config is the top-level structure parsed from a .bifrost.yml file.
type Config struct {
	Strategy     string                 `yaml:"strategy"`
	Paths        Paths                  `yaml:"paths"`
	Settings     Settings               `yaml:"settings"`
	Variables    map[string]string      `yaml:"variables"`
	Hooks        Hooks                  `yaml:"hooks"`
	Environments map[string]Environment `yaml:"environments"`
}
