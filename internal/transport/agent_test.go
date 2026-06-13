package transport

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
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
	assert.Equal(t, "https://github.com/adaouat/bifrost/releases/download/v0.9.0/bifrost_0.9.0_linux_amd64", url)
}

func TestDownloadURL_Darwin(t *testing.T) {
	url := downloadURL("1.2.3", Platform{OS: "darwin", Arch: "arm64"})
	assert.Equal(t, "https://github.com/adaouat/bifrost/releases/download/v1.2.3/bifrost_1.2.3_darwin_arm64", url)
}

func TestDownloadURL_NormalizesLeadingV(t *testing.T) {
	url := downloadURL("v1.2.3", Platform{OS: "linux", Arch: "amd64"})
	assert.Equal(t, "https://github.com/adaouat/bifrost/releases/download/v1.2.3/bifrost_1.2.3_linux_amd64", url)
}

func TestDownloadAgent_ReturnsRawBinary(t *testing.T) {
	want := []byte("\x7fELF not-an-archive")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1.2.3/bifrost_1.2.3_linux_amd64", r.URL.Path)
		_, _ = w.Write(want)
	}))
	defer srv.Close()

	orig := downloadBaseURL
	downloadBaseURL = srv.URL
	defer func() { downloadBaseURL = orig }()

	got, err := downloadAgent("1.2.3", Platform{OS: "linux", Arch: "amd64"})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestDownloadAgent_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	orig := downloadBaseURL
	downloadBaseURL = srv.URL
	defer func() { downloadBaseURL = orig }()

	_, err := downloadAgent("9.9.9", Platform{OS: "linux", Arch: "amd64"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--agent-binary")
}

func TestVerifyChecksum_Match(t *testing.T) {
	data := []byte("agent-binary-bytes")
	sum := sha256.Sum256(data)
	checksums := hex.EncodeToString(sum[:]) + "  bifrost_1.2.3_linux_amd64\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1.2.3/checksums.txt", r.URL.Path)
		_, _ = w.Write([]byte(checksums))
	}))
	defer srv.Close()

	orig := downloadBaseURL
	downloadBaseURL = srv.URL
	defer func() { downloadBaseURL = orig }()

	err := verifyChecksum("1.2.3", Platform{OS: "linux", Arch: "amd64"}, data)
	require.NoError(t, err)
}

func TestVerifyChecksum_Tampered(t *testing.T) {
	good := sha256.Sum256([]byte("the-real-agent"))
	checksums := hex.EncodeToString(good[:]) + "  bifrost_1.2.3_linux_amd64\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(checksums))
	}))
	defer srv.Close()

	orig := downloadBaseURL
	downloadBaseURL = srv.URL
	defer func() { downloadBaseURL = orig }()

	err := verifyChecksum("1.2.3", Platform{OS: "linux", Arch: "amd64"}, []byte("a-tampered-agent"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestVerifyChecksum_NoEntryForPlatform(t *testing.T) {
	sum := sha256.Sum256([]byte("data"))
	checksums := hex.EncodeToString(sum[:]) + "  bifrost_1.2.3_linux_amd64\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(checksums))
	}))
	defer srv.Close()

	orig := downloadBaseURL
	downloadBaseURL = srv.URL
	defer func() { downloadBaseURL = orig }()

	err := verifyChecksum("1.2.3", Platform{OS: "darwin", Arch: "arm64"}, []byte("data"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no checksum")
}
