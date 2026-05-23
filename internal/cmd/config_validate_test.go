package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCmd_EnvApp_OutputsMergedJSON(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config",
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
	vars, _ := result["variables"].(map[string]any)
	assert.Equal(t, "production", vars["app_env"])
}

func TestConfigCmd_EnvApp_MissingEnv(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config",
		"--config", "../../testdata/bifrost-full.yml",
		"--env", "staging",
		"--app", "web",
	})

	err := root.Execute()
	assert.Error(t, err)
}

func TestConfigCmd_EnvApp_MissingApp(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"config",
		"--config", "../../testdata/bifrost-full.yml",
		"--env", "prod",
		"--app", "api",
	})

	err := root.Execute()
	assert.Error(t, err)
}
