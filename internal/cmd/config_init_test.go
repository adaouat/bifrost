package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/adaouat/bifrost/internal/cmd"
	configcmd "github.com/adaouat/bifrost/internal/cmd/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigInitCmd_WritesDefaultConfig(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init", "--config", "/tmp/test-bifrost-init.yml"})

	var written []byte
	configcmd.SetInitWrite(func(_ string, data []byte, _ uint32) error {
		written = data
		return nil
	})
	configcmd.SetInitStat(func(_ string) error { return os.ErrNotExist })
	defer func() {
		configcmd.SetInitWrite(nil)
		configcmd.SetInitStat(nil)
	}()

	err := root.Execute()
	require.NoError(t, err)

	assert.NotEmpty(t, written, "should have written config content")
	assert.Contains(t, string(written), "releases_root")
	assert.Contains(t, string(written), "shared_root")
}

func TestConfigInitCmd_RefusesOverwriteWithoutForce(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init", "--config", "/tmp/test-bifrost-init.yml"})

	configcmd.SetInitWrite(func(_ string, _ []byte, _ uint32) error { return nil })
	configcmd.SetInitStat(func(_ string) error { return nil }) // file exists
	defer func() {
		configcmd.SetInitWrite(nil)
		configcmd.SetInitStat(nil)
	}()

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConfigInitCmd_OverwritesWithForce(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "init", "--config", "/tmp/test-bifrost-init.yml", "--force"})

	var written []byte
	configcmd.SetInitWrite(func(_ string, data []byte, _ uint32) error {
		written = data
		return nil
	})
	configcmd.SetInitStat(func(_ string) error { return nil }) // file exists, but --force is set
	defer func() {
		configcmd.SetInitWrite(nil)
		configcmd.SetInitStat(nil)
	}()

	err := root.Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, written)
}
