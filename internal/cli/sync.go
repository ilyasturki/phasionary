package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"phasionary/internal/config"
	"phasionary/internal/journal"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "App-level sync with a self-hosted phasionary server",
		Long: "App-level sync with a self-hosted phasionary server.\n\n" +
			"Sync is opt-in: until this device is enrolled with a server, phasionary\n" +
			"is purely local and journals nothing. The server and the enrollment\n" +
			"command ship in a future release; see docs/sync-design.md.",
	}
	cmd.AddCommand(newSyncStatusCmd())
	return cmd
}

func newSyncStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show this device's sync state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := config.ResolveStateDir()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			device, enrolled, err := journal.LoadDevice(stateDir)
			if err != nil {
				return err
			}
			if !enrolled {
				fmt.Fprintln(out, "Sync: not configured (local-only)")
				fmt.Fprintln(out, "This device journals no changes. Server support is planned; see docs/sync-design.md.")
				return nil
			}
			fmt.Fprintln(out, "Sync: configured")
			fmt.Fprintf(out, "Device ID: %s\n", device.DeviceID)
			if device.ServerURL != "" {
				fmt.Fprintf(out, "Server:    %s\n", device.ServerURL)
			}
			ops, err := journal.Open(stateDir).Entries()
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Pending:   %d journaled change(s) awaiting sync\n", len(ops))
			return nil
		},
	}
}
