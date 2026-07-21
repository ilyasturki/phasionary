package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/api"
	"phasionary/internal/config"
)

const (
	envServeHost         = "PHASIONARY_SERVE_HOST"
	envServePort         = "PHASIONARY_SERVE_PORT"
	envServeToken        = "PHASIONARY_SERVE_TOKEN"
	envServeAllowedHosts = "PHASIONARY_SERVE_ALLOWED_HOSTS"

	defaultServeHost = "127.0.0.1"
	defaultServePort = 7777
)

// resolveServeToken returns the bearer token the server will require, in order
// of precedence: --token / PHASIONARY_SERVE_TOKEN, then the stored token in
// config.json, then a freshly generated one which is persisted and announced.
//
// Generating rather than defaulting to "no auth" is the point: binding to
// loopback is not a trust boundary a browser respects. A page the user visits
// can reach 127.0.0.1 through DNS rebinding, at which point the request is
// same-origin and the response body is readable. A token the page cannot see
// closes that, along with cross-origin writes — but only if one exists by
// default, since the exposure lands hardest on people who never configured one.
func resolveServeToken(out io.Writer) (string, error) {
	if token := viper.GetString("serve.token"); token != "" {
		return token, nil
	}

	configPath, err := config.ResolveConfigPath(viper.GetString("config"))
	if err != nil {
		return "", err
	}
	cfgManager := config.NewManager(configPath)
	if err := cfgManager.Load(); err != nil {
		return "", err
	}
	if token := cfgManager.Get().ServeToken; token != "" {
		return token, nil
	}

	token, err := config.NewServeToken()
	if err != nil {
		return "", fmt.Errorf("generating serve token: %w", err)
	}
	if err := cfgManager.Update(func(c *config.Config) { c.ServeToken = token }); err != nil {
		return "", fmt.Errorf("saving serve token: %w", err)
	}

	// Printed once, on the run that creates it. An existing client will start
	// getting 401s at this point, so the message has to say what to do about it
	// rather than just announce a new secret.
	fmt.Fprintf(out, `
Generated an API token and saved it to %s:

    %s

The mobile app needs this token before it can connect again: enter it in
Settings alongside the server URL. Requests without it are rejected.

`, cfgManager.Path(), token)

	return token, nil
}

// resolveAllowedHosts returns the extra Host header values the server accepts.
//
// It exists because viper hands back a bound environment variable as one
// opaque string, and GetStringSlice then splits a string on whitespace rather
// than commas. PHASIONARY_SERVE_ALLOWED_HOSTS="a.example,b.example" therefore
// arrived as the single hostname "a.example,b.example", which matches nothing —
// so both names were refused with a 421 that named neither. The repeated
// --allowed-host flag was unaffected, pflag having already split it.
//
// Splitting on commas here makes the two paths agree. Empty entries are
// dropped so a trailing comma or an unset-but-exported variable is not turned
// into an "" host, which hostAllowed would reject anyway but only after the
// startup log had claimed to accept it.
func resolveAllowedHosts() []string {
	var hosts []string
	for _, entry := range viper.GetStringSlice("serve.allowed_hosts") {
		for _, host := range strings.Split(entry, ",") {
			if host = strings.TrimSpace(host); host != "" {
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the JSON API for the mobile app",
		Long: `Serve the JSON API (/api/v1) that the Phasionary mobile app talks to.

Every request needs an "Authorization: Bearer <token>" header. On first run a
token is generated, saved to config.json and printed once; pass --token or set
PHASIONARY_SERVE_TOKEN to use your own instead. Authentication is not optional
even on loopback, because a web page the browser loads can reach 127.0.0.1.

By default the server binds to 127.0.0.1:7777 only. Reach it from your phone
via an SSH tunnel:

    ssh -L 7777:localhost:7777 your-server

then point the app at http://localhost:7777 and enter the token.

To expose on a LAN/Tailscale interface, pass --host explicitly (e.g.
--host 0.0.0.0).

Requests are accepted when the Host header is an IP address or "localhost",
which covers reaching the server by LAN or Tailscale IP. A request naming the
server some other way — a Tailscale MagicDNS name, or a reverse proxy — is
refused until that name is passed with --allowed-host. This is what stops a web
page from reaching the server by re-pointing its own domain at it.

--allowed-host is repeatable; PHASIONARY_SERVE_ALLOWED_HOSTS takes the same
names as a comma-separated list.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			host := viper.GetString("serve.host")
			if host == "" {
				host = defaultServeHost
			}
			port := viper.GetInt("serve.port")
			if port == 0 {
				port = defaultServePort
			}
			addr := net.JoinHostPort(host, strconv.Itoa(port))
			// Always non-empty: resolveServeToken generates and persists one
			// rather than returning "". The old "--token is required for a
			// non-loopback bind" guard is gone because every bind now requires
			// one — loopback included.
			token, err := resolveServeToken(cmd.OutOrStdout())
			if err != nil {
				return err
			}

			store, err := storeFromViper()
			if err != nil {
				return err
			}
			if err := store.Ensure(); err != nil {
				return err
			}

			// Same state.json the TUI uses: folds set from the phone show up
			// in the TUI the next time it opens the project, and vice versa.
			state, err := stateManagerFromViper()
			if err != nil {
				return err
			}

			srv := api.New(store, state, api.Config{
				Addr:         addr,
				Token:        token,
				AllowedHosts: resolveAllowedHosts(),
			})

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := srv.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String("host", defaultServeHost, "listen host/IP")
	cmd.Flags().Int("port", defaultServePort, "listen port")
	cmd.Flags().String("token", "", "auth token; generated and saved on first run if unset")
	cmd.Flags().StringSlice("allowed-host", nil,
		"extra Host header value to accept, e.g. a Tailscale MagicDNS name (repeatable)")

	_ = viper.BindPFlag("serve.host", cmd.Flags().Lookup("host"))
	_ = viper.BindPFlag("serve.port", cmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("serve.token", cmd.Flags().Lookup("token"))
	_ = viper.BindPFlag("serve.allowed_hosts", cmd.Flags().Lookup("allowed-host"))
	_ = viper.BindEnv("serve.host", envServeHost)
	_ = viper.BindEnv("serve.port", envServePort)
	_ = viper.BindEnv("serve.token", envServeToken)
	_ = viper.BindEnv("serve.allowed_hosts", envServeAllowedHosts)

	return cmd
}
