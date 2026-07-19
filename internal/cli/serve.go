package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/api"
)

const (
	envServeHost  = "PHASIONARY_SERVE_HOST"
	envServePort  = "PHASIONARY_SERVE_PORT"
	envServeToken = "PHASIONARY_SERVE_TOKEN"

	defaultServeHost = "127.0.0.1"
	defaultServePort = 7777
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the JSON API for the mobile app",
		Long: `Serve the JSON API (/api/v1) that the Phasionary mobile app talks to.

By default the server binds to 127.0.0.1:7777 only. Reach it from your phone
via an SSH tunnel:

    ssh -L 7777:localhost:7777 your-server

then point the app at http://localhost:7777.

To expose on a LAN/Tailscale interface, pass --host explicitly (e.g.
--host 0.0.0.0). A --token is required for any non-loopback bind; clients send
it as an "Authorization: Bearer <token>" header.`,
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
			token := viper.GetString("serve.token")

			if !api.IsLoopbackAddr(addr) && token == "" {
				return fmt.Errorf("--token is required when binding to a non-loopback address (got %q)", addr)
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

			srv := api.New(store, state, api.Config{Addr: addr, Token: token})

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
	cmd.Flags().String("token", "", "auth token; required for non-loopback bind")

	_ = viper.BindPFlag("serve.host", cmd.Flags().Lookup("host"))
	_ = viper.BindPFlag("serve.port", cmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("serve.token", cmd.Flags().Lookup("token"))
	_ = viper.BindEnv("serve.host", envServeHost)
	_ = viper.BindEnv("serve.port", envServePort)
	_ = viper.BindEnv("serve.token", envServeToken)

	return cmd
}
