package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStagingPath_DefaultsToTmp(t *testing.T) {
	assert.Equal(t, "/tmp/bifrost-abc123", stagingPath("", "abc123"))
}

func TestStagingPath_UsesConfiguredBase(t *testing.T) {
	assert.Equal(t, "/srv/staging/bifrost-abc123", stagingPath("/srv/staging", "abc123"))
}

func TestStagingPath_CleansTrailingSlash(t *testing.T) {
	assert.Equal(t, "/tmp/bifrost-abc123", stagingPath("/tmp/", "abc123"))
}
