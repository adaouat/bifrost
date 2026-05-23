package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCmd_NoFlags_OutputsJSON(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "--config", "../../testdata/bifrost-full.yml"})

	err := root.Execute()
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result), "output must be valid JSON:\n%s", buf.String())

	assert.Equal(t, "atomic", result["strategy"])

	paths, ok := result["paths"].(map[string]any)
	require.True(t, ok, "paths must be an object")
	assert.Equal(t, "/var/www/webroot/ROOT", paths["releases_root"])
	assert.Equal(t, "/var/nas/shared", paths["shared_root"])
}

func TestConfigCmd_MissingConfigFile(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "--config", "/nonexistent/.bifrost.yml"})

	err := root.Execute()
	assert.Error(t, err)
}
