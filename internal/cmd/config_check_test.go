package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCheckCmd_ValidConfig_OutputsMergedJSON(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config", "check",
		"--config", "../../testdata/bifrost-full.yml",
		"--env", "prod",
		"--app", "web",
	})

	err := root.Execute()
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result), "output must be valid JSON:\n%s", buf.String())

	assert.Equal(t, "/var/www/webroot/ROOT", result["releases_root"])
	assert.Equal(t, "/var/nas/shared", result["shared_root"])
}

func TestConfigCheckCmd_MissingEnv_Errors(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config", "check",
		"--config", "../../testdata/bifrost-full.yml",
		"--env", "staging",
		"--app", "web",
	})

	err := root.Execute()
	assert.Error(t, err)
}

func TestConfigCheckCmd_MissingApp_Errors(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config", "check",
		"--config", "../../testdata/bifrost-full.yml",
		"--env", "prod",
		"--app", "api",
	})

	err := root.Execute()
	assert.Error(t, err)
}

func TestConfigCheckCmd_MissingPaths_ExitCode2(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config", "check",
		"--config", "../../testdata/bifrost-missing-paths.yml",
		"--env", "prod",
		"--app", "web",
	})

	err := root.Execute()
	require.Error(t, err)
	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 2, exitErr.Code)
}

func TestConfigCheckCmd_RequiresEnvAndApp(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config", "check",
		"--config", "../../testdata/bifrost-full.yml",
	})

	err := root.Execute()
	assert.Error(t, err)
}
