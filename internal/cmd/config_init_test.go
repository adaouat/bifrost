package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/adaouat/bifrost/internal/cmd/cmdutil"
	configcmd "github.com/adaouat/bifrost/internal/cmd/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigInitCmd_WritesDefaultConfig(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	root := cmd.NewRootCmd("dev")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init", "--config", "/tmp/test-bifrost-init.yml"})

	var written []byte
	configcmd.SetInitWrite(func(_ string, data []byte, _ uint32) error {
		written = data
		return nil
	})
	cmdutil.SetStatFile(func(string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	defer func() {
		configcmd.SetInitWrite(nil)
		cmdutil.SetStatFile(nil)
	}()

	err := root.Execute()
	require.NoError(t, err)

	assert.NotEmpty(t, written, "should have written config content")
	assert.Contains(t, string(written), "releases_root")
	assert.Contains(t, string(written), "shared_root")
}

func TestConfigInitCmd_RefusesOverwriteWithoutForce(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init", "--config", "/tmp/test-bifrost-init.yml"})

	configcmd.SetInitWrite(func(_ string, _ []byte, _ uint32) error { return nil })
	cmdutil.SetStatFile(func(string) (os.FileInfo, error) { return nil, nil }) // file exists
	defer func() {
		configcmd.SetInitWrite(nil)
		cmdutil.SetStatFile(nil)
	}()

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConfigInitCmd_EnvVarDeterminesWriteDest(t *testing.T) {
	t.Setenv("BIFROST_FILE", "/env/dest.yml")
	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init"})

	var writtenPath string
	configcmd.SetInitWrite(func(path string, _ []byte, _ uint32) error {
		writtenPath = path
		return nil
	})
	cmdutil.SetStatFile(func(string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	defer func() {
		configcmd.SetInitWrite(nil)
		cmdutil.SetStatFile(nil)
	}()

	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, "/env/dest.yml", writtenPath)
}

func TestConfigInitCmd_FlagWinsOverEnvVar(t *testing.T) {
	t.Setenv("BIFROST_FILE", "/env/dest.yml")
	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init", "--config", "/flag/dest.yml"})

	var writtenPath string
	configcmd.SetInitWrite(func(path string, _ []byte, _ uint32) error {
		writtenPath = path
		return nil
	})
	cmdutil.SetStatFile(func(string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	defer func() {
		configcmd.SetInitWrite(nil)
		cmdutil.SetStatFile(nil)
	}()

	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, "/flag/dest.yml", writtenPath)
}

func TestConfigInitCmd_AutoDestUsesConfigDir(t *testing.T) {
	// Auto-dest discovery now uses forge's resolver (real os.Stat), so create a
	// real .config/ dir in a temp cwd rather than injecting statFile.
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "")
	require.NoError(t, os.MkdirAll(".config", 0o755))

	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init"})

	var writtenPath string
	configcmd.SetInitWrite(func(path string, _ []byte, _ uint32) error {
		writtenPath = path
		return nil
	})
	defer configcmd.SetInitWrite(nil)

	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, ".config/bifrost.yml", writtenPath)
}

func TestConfigInitCmd_OverwritesWithForce(t *testing.T) {
	t.Setenv("BIFROST_FILE", "")
	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init", "--config", "/tmp/test-bifrost-init.yml", "--force"})

	var written []byte
	configcmd.SetInitWrite(func(_ string, data []byte, _ uint32) error {
		written = data
		return nil
	})
	cmdutil.SetStatFile(func(string) (os.FileInfo, error) { return nil, nil }) // file exists, --force overrides
	defer func() {
		configcmd.SetInitWrite(nil)
		cmdutil.SetStatFile(nil)
	}()

	err := root.Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, written)
}
