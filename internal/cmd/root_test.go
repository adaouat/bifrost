package cmd_test

import (
	"testing"

	"github.com/bchatard/bifrost/internal/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootFlagsDefaults(t *testing.T) {
	root := cmd.NewRootCmd()
	flags := root.PersistentFlags()

	config, err := flags.GetString("config")
	require.NoError(t, err)
	assert.Equal(t, ".bifrost.yml", config)

	output, err := flags.GetString("output")
	require.NoError(t, err)
	assert.Equal(t, "human", output)

	dryRun, err := flags.GetBool("dry-run")
	require.NoError(t, err)
	assert.False(t, dryRun)

	verbose, err := flags.GetBool("verbose")
	require.NoError(t, err)
	assert.False(t, verbose)
}

func TestValidateOutputMode(t *testing.T) {
	tests := []struct {
		mode    string
		wantErr bool
	}{
		{"human", false},
		{"json", false},
		{"plain", false},
		{"xml", true},
		{"", true},
		{"Human", true},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			err := cmd.ValidateOutputMode(tt.mode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
