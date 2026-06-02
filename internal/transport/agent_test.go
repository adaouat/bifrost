package transport

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapPlatform_SupportedTargets(t *testing.T) {
	cases := []struct {
		unameS, unameM string
		want           Platform
	}{
		{"Linux", "x86_64", Platform{OS: "linux", Arch: "amd64"}},
		{"Linux", "aarch64", Platform{OS: "linux", Arch: "arm64"}},
		{"Darwin", "x86_64", Platform{OS: "darwin", Arch: "amd64"}},
		{"Darwin", "arm64", Platform{OS: "darwin", Arch: "arm64"}},
	}
	for _, tc := range cases {
		got, err := mapPlatform(tc.unameS, tc.unameM)
		require.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}
}

func TestMapPlatform_Unsupported(t *testing.T) {
	_, err := mapPlatform("Linux", "i386")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Linux/i386")
	assert.Contains(t, err.Error(), "--agent-binary")
}

func TestCacheKey(t *testing.T) {
	key := cacheKey("0.9.0", Platform{OS: "linux", Arch: "amd64"})
	assert.Equal(t, "bifrost/agents/0.9.0/linux_amd64/bifrost", filepath.ToSlash(key))
}

func TestCacheKey_Darwin(t *testing.T) {
	key := cacheKey("1.2.3", Platform{OS: "darwin", Arch: "arm64"})
	assert.Equal(t, "bifrost/agents/1.2.3/darwin_arm64/bifrost", filepath.ToSlash(key))
}

func TestDownloadURL(t *testing.T) {
	url := downloadURL("0.9.0", Platform{OS: "linux", Arch: "amd64"})
	assert.Equal(t, "https://github.com/adaouat/bifrost/releases/download/v0.9.0/bifrost_0.9.0_linux_amd64.tar.gz", url)
}

func TestDownloadURL_Darwin(t *testing.T) {
	url := downloadURL("1.2.3", Platform{OS: "darwin", Arch: "arm64"})
	assert.Equal(t, "https://github.com/adaouat/bifrost/releases/download/v1.2.3/bifrost_1.2.3_darwin_arm64.tar.gz", url)
}
