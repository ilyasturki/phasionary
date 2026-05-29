package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/web"
)

const (
	envServeAddr  = "PHASIONARY_SERVE_ADDR"
	envServeToken = "PHASIONARY_SERVE_TOKEN"

	defaultServeAddr = "127.0.0.1:7777"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve a touch-friendly web UI for mobile access",
		Long: `Serve a touch-friendly htmx web UI for accessing Phasionary from a phone or tablet.

By default the server binds to ` + defaultServeAddr + ` only. Reach it from your phone
via an SSH tunnel:

    ssh -L 7777:localhost:7777 your-server

then open http://localhost:7777 in the phone's browser.

To expose on a LAN/Tailscale interface, pass --addr explicitly. A --token is
required for any non-loopback bind. The first request with ?token=... receives
a session cookie so the secret isn't carried in URLs after that.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := viper.GetString("serve.addr")
			if addr == "" {
				addr = defaultServeAddr
			}
			token := viper.GetString("serve.token")

			if !web.IsLoopbackAddr(addr) && token == "" {
				return fmt.Errorf("--token is required when binding to a non-loopback address (got %q)", addr)
			}

			store, err := storeFromViper()
			if err != nil {
				return err
			}
			if err := store.Ensure(); err != nil {
				return err
			}

			srv, err := web.New(store, web.Config{Addr: addr, Token: token})
			if err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := srv.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String("addr", defaultServeAddr, "listen address (host:port)")
	cmd.Flags().String("token", "", "auth token; required for non-loopback bind")

	_ = viper.BindPFlag("serve.addr", cmd.Flags().Lookup("addr"))
	_ = viper.BindPFlag("serve.token", cmd.Flags().Lookup("token"))
	_ = viper.BindEnv("serve.addr", envServeAddr)
	_ = viper.BindEnv("serve.token", envServeToken)

	return cmd
}
