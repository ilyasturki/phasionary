package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"phasionary/internal/config"
	"phasionary/internal/journal"
	"phasionary/internal/syncclient"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync with a self-hosted phasionary-server",
		Long: `Sync with a self-hosted phasionary-server.

Sync is opt-in. Until this device is enrolled (sync login), phasionary is
purely local and journals nothing. Once enrolled, every change is journaled
and "sync now" exchanges it with the server; nothing happens on its own.

Stop any file syncer (Syncthing, Dropbox, …) carrying the data directory
before enrolling: two pipes moving the same files fight each other.`,
	}
	cmd.AddCommand(newSyncLoginCmd(), newSyncNowCmd(), newSyncStatusCmd(), newSyncLogoutCmd())
	return cmd
}

func newSyncLoginCmd() *cobra.Command {
	var code, name string
	cmd := &cobra.Command{
		Use:   "login <server-url>",
		Short: "Enroll this device with a server and upload its projects",
		Long: `Enroll this device with a phasionary-server.

Get a single-use code from "phasionary-server enroll" on the server, then run
this with the server's URL. The device identity and credential are stored in
` + "`~/.local/state/phasionary/`" + `, never in the data directory. Every local
project is uploaded as part of enrolling.`,
		Args: exactArgs("server-url"),
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := config.ResolveStateDir()
			if err != nil {
				return err
			}
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			if code == "" {
				if code, err = readCode(cmd.OutOrStdout(), cmd.InOrStdin()); err != nil {
					return err
				}
			}
			if name == "" {
				name, _ = os.Hostname()
			}
			dev, sum, err := syncclient.Login(cmd.Context(), stateDir, store, args[0], code, name)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Enrolled as device %s with %s.\n", dev.DeviceID, dev.ServerURL)
			printSummary(out, sum)
			return nil
		},
	}
	cmd.Flags().StringVar(&code, "code", "", "enrollment code (prompted for when omitted)")
	cmd.Flags().StringVar(&name, "name", "", "label for this device on the server (default: hostname)")
	return cmd
}

func readCode(out io.Writer, in io.Reader) (string, error) {
	fmt.Fprint(out, "Enrollment code: ")
	line, err := bufio.NewReader(in).ReadString('\n')
	code := strings.TrimSpace(line)
	if code == "" {
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return "", errors.New("no enrollment code given (pass --code)")
	}
	return code, nil
}

func newSyncNowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "now",
		Short: "Push journaled changes and pull what changed elsewhere",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := config.ResolveStateDir()
			if err != nil {
				return err
			}
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			sum, err := syncclient.Run(cmd.Context(), stateDir, store)
			if err != nil {
				return err
			}
			printSummary(cmd.OutOrStdout(), sum)
			return nil
		},
	}
}

func printSummary(out io.Writer, sum syncclient.Summary) {
	fmt.Fprintf(out, "Pushed %d change(s).\n", sum.Pushed)
	fmt.Fprintf(out, "Pulled %d project(s) written, %d unchanged, %d removed.\n", sum.Written, sum.Unchanged, sum.Removed)
	if sum.Skipped > 0 {
		fmt.Fprintf(out, "%d project(s) changed locally during the sync; run `phasionary sync now` again.\n", sum.Skipped)
	}
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
				fmt.Fprintln(out, "This device journals no changes. Enroll with: phasionary sync login <server-url>")
				return nil
			}
			fmt.Fprintln(out, "Sync: configured")
			fmt.Fprintf(out, "Device ID: %s\n", device.DeviceID)
			fmt.Fprintf(out, "Server:    %s\n", device.ServerURL)
			ops, err := journal.Open(stateDir).Entries()
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Pending:   %d journaled change(s) awaiting sync\n", len(ops))
			return nil
		},
	}
}

func newSyncLogoutCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Forget this device's enrollment and go back to local-only",
		Long: `Forget this device's enrollment.

Removes the device identity, credential and journal. Project files stay as
they are. The server keeps its record of the device; enrolling again creates
a new one.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := config.ResolveStateDir()
			if err != nil {
				return err
			}
			if _, enrolled, err := journal.LoadDevice(stateDir); err != nil && !force {
				return err
			} else if !enrolled && err == nil {
				return errors.New("this device is not enrolled")
			}
			ops, _ := journal.Open(stateDir).Entries()
			if len(ops) > 0 && !force {
				return fmt.Errorf("%d change(s) not yet synced; run `phasionary sync now` first, or pass --force to discard them", len(ops))
			}
			if err := journal.Unenroll(stateDir); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Sync: not configured (local-only). Enrollment removed.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "discard unsynced changes and a corrupt identity file")
	return cmd
}
