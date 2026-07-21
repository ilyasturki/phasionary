package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostAllowed(t *testing.T) {
	s := &Server{cfg: Config{AllowedHosts: []string{"phasionary.tail1a2b.ts.net"}}}

	t.Run("accepts IP literals and localhost", func(t *testing.T) {
		// Reaching the server by address is how every documented deployment
		// works — loopback, SSH tunnel, LAN, Tailscale CGNAT — so none of these
		// may be broken by the check.
		for _, host := range []string{
			"127.0.0.1:7777",
			"127.0.0.1",
			"localhost:7777",
			"LocalHost",
			"[::1]:7777",
			"::1",
			"192.168.1.5:7777",
			"100.101.102.103:7777", // Tailscale CGNAT range
			"0.0.0.0:7777",
		} {
			assert.True(t, s.hostAllowed(host), host)
		}
	})

	t.Run("accepts explicitly allowed names", func(t *testing.T) {
		assert.True(t, s.hostAllowed("phasionary.tail1a2b.ts.net"))
		assert.True(t, s.hostAllowed("phasionary.tail1a2b.ts.net:7777"))
		assert.True(t, s.hostAllowed("PHASIONARY.TAIL1A2B.TS.NET"))
	})

	t.Run("refuses unrecognised names", func(t *testing.T) {
		// This is the DNS-rebinding shape: the attacker's page keeps its own
		// hostname while DNS re-points it at 127.0.0.1, so the Host header still
		// names the attacker. httptest.NewRequest's "example.com" default is
		// exactly this, which is why the API tests had to start setting a Host.
		for _, host := range []string{
			"example.com",
			"example.com:7777",
			"evil.example",
			"rebind.attacker.test:7777",
			"phasionary.tail1a2b.ts.net.evil.com",
			"",
		} {
			assert.False(t, s.hostAllowed(host), host)
		}
	})

	t.Run("refuses named hosts when none are allowed", func(t *testing.T) {
		bare := &Server{cfg: Config{}}
		assert.False(t, bare.hostAllowed("phasionary.tail1a2b.ts.net"))
		assert.True(t, bare.hostAllowed("127.0.0.1:7777"))
	})
}

func TestSafeForLog(t *testing.T) {
	// The request path is attacker-controlled and arrives percent-decoded, and
	// this middleware runs outside auth — so a rejected request still reaches
	// the operator's terminal. Anything that makes the terminal act rather than
	// display has to be neutralised on the way out.
	t.Run("escapes terminal control sequences", func(t *testing.T) {
		for _, payload := range []string{
			"/api/v1/projects/\x1b]52;c;cm0gLXJm\x07",
			"/api/v1/\x1b[2J",
			"/api/v1/K",
			"/api/v1/a\nfake log line",
			"/api/v1/\x00",
		} {
			got := safeForLog(payload)
			for _, r := range got {
				assert.False(t, r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f),
					"control character %U survived in %q", r, got)
			}
		}
	})

	t.Run("keeps ordinary paths readable", func(t *testing.T) {
		got := safeForLog("/api/v1/projects/abc-123")
		assert.Equal(t, `"/api/v1/projects/abc-123"`, got)
	})

	t.Run("prevents forging a second log line", func(t *testing.T) {
		// A newline would otherwise let a request write what looks like its own
		// complete log entry.
		got := safeForLog("/x\nGET /api/v1/projects 200 1ms")
		assert.False(t, strings.Contains(got, "\n"))
	})
}
