package cli

import (
	"testing"

	"github.com/spf13/viper"
)

// Tests must not run in parallel: resolveAllowedHosts reads the global viper,
// so concurrent runs would race on it.
func TestResolveAllowedHosts(t *testing.T) {
	// The env var is the case that regressed: viper returns a bound variable as
	// one opaque string, and GetStringSlice splits a string on whitespace, so a
	// comma-separated list used to arrive as a single unmatched hostname.
	t.Run("comma-separated env var", func(t *testing.T) {
		viper.Reset()
		t.Setenv(envServeAllowedHosts, "a.example,b.example")
		if err := viper.BindEnv("serve.allowed_hosts", envServeAllowedHosts); err != nil {
			t.Fatalf("bind env: %v", err)
		}

		got := resolveAllowedHosts()
		want := []string{"a.example", "b.example"}
		assertHosts(t, got, want)
	})

	t.Run("repeated flag", func(t *testing.T) {
		viper.Reset()
		cmd := newServeCmd()
		if err := cmd.Flags().Parse([]string{
			"--allowed-host", "a.example",
			"--allowed-host", "b.example",
		}); err != nil {
			t.Fatalf("parse flags: %v", err)
		}

		assertHosts(t, resolveAllowedHosts(), []string{"a.example", "b.example"})
	})

	t.Run("whitespace and empty entries dropped", func(t *testing.T) {
		viper.Reset()
		t.Setenv(envServeAllowedHosts, " a.example , ,b.example,")
		if err := viper.BindEnv("serve.allowed_hosts", envServeAllowedHosts); err != nil {
			t.Fatalf("bind env: %v", err)
		}

		assertHosts(t, resolveAllowedHosts(), []string{"a.example", "b.example"})
	})

	t.Run("unset yields no hosts", func(t *testing.T) {
		viper.Reset()
		if got := resolveAllowedHosts(); len(got) != 0 {
			t.Fatalf("resolveAllowedHosts() = %q, want empty", got)
		}
	})
}

func assertHosts(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("resolveAllowedHosts() = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("resolveAllowedHosts() = %q, want %q", got, want)
		}
	}
}
