package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"phasionary/internal/config"
	"phasionary/internal/version"
)

const (
	EnvHost         = "PHASIONARY_SERVER_HOST"
	EnvPort         = "PHASIONARY_SERVER_PORT"
	EnvAllowedHosts = "PHASIONARY_SERVER_ALLOWED_HOSTS"

	defaultHost = "127.0.0.1"
	defaultPort = "7777"
)

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func envHosts(name string) []string {
	return strings.Fields(strings.ReplaceAll(os.Getenv(name), ",", " "))
}

func NewRootCmd() *cobra.Command {
	var (
		dataDir      string
		host         string
		port         string
		allowedHosts []string
	)
	cmd := &cobra.Command{
		Use:   "phasionary-server",
		Short: "Self-hosted sync server for phasionary",
		Long: `Self-hosted sync server for phasionary.

Devices enroll once with a single-use code (see "phasionary-server enroll"),
then exchange their change journals with this server over HTTP. The server
merges every device's changes and keeps the result in a SQLite database.

It binds to 127.0.0.1:7777 by default. Put it on a private network (Tailscale,
WireGuard, an SSH tunnel) or behind a TLS reverse proxy rather than exposing
the plain-HTTP port to the internet: the per-device bearer tokens travel in
every request.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := config.ResolveServerDataDir(dataDir)
			if err != nil {
				return err
			}
			db, err := OpenDB(dir)
			if err != nil {
				return err
			}
			defer db.Close()
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return New(db, net.JoinHostPort(host, port), allowedHosts...).Run(ctx)
		},
	}
	cmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version.Version, version.Commit, version.BuildDate)

	cmd.PersistentFlags().StringVar(&dataDir, "data", "", "database directory (default: $"+config.EnvServerDataPath+" or ~/.local/share/phasionary-server)")
	cmd.Flags().StringVar(&host, "host", envOr(EnvHost, defaultHost), "listen host/IP ($"+EnvHost+")")
	cmd.Flags().StringVar(&port, "port", envOr(EnvPort, defaultPort), "listen port ($"+EnvPort+")")
	cmd.Flags().StringArrayVar(&allowedHosts, "allowed-host", envHosts(EnvAllowedHosts),
		"hostname browsers may reach the app by; repeatable, IPs and localhost always pass ($"+EnvAllowedHosts+", comma-separated)")

	cmd.AddCommand(newEnrollCmd(&dataDir, &host, &port))
	return cmd
}

func newEnrollCmd(dataDir, host, port *string) *cobra.Command {
	return &cobra.Command{
		Use:   "enroll",
		Short: "Print a single-use code that lets one device join",
		Long: `Print a single-use enrollment code.

The code is valid for ten minutes and admits exactly one device. It is stored
in the server's database, so run this against the same --data directory (and
as the same user) as the running server.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := config.ResolveServerDataDir(*dataDir)
			if err != nil {
				return err
			}
			db, err := OpenDB(dir)
			if err != nil {
				return err
			}
			defer db.Close()
			code, err := db.NewEnrollCode(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Enrollment code: %s (single use, valid %s)\n\n"+
					"On the device to enroll, run\n\n    phasionary sync login http://%s\n\n"+
					"and enter the code when prompted. Replace the address with the one that\ndevice reaches this server at.\n",
				code, CodeTTL, net.JoinHostPort(*host, *port))
			return nil
		},
	}
}
