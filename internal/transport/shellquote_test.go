package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellQuote(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"plain", "release-20260613", "'release-20260613'"},
		{"with space", "my release", "'my release'"},
		{"with semicolon", "rel; rm -rf /", "'rel; rm -rf /'"},
		{"with command substitution", "$(reboot)", "'$(reboot)'"},
		{"with single quote", "a'b", `'a'\''b'`},
		{"empty", "", "''"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ShellQuote(tc.in))
		})
	}
}
