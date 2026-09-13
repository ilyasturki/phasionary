package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"phasionary/internal/config"
	"phasionary/internal/data"
	"phasionary/internal/journal"
	"phasionary/internal/syncclient"
)

const autoSyncTimeout = 3 * time.Second

func syncAtLaunch(store *data.Store) string {
	if err := autoSyncRound(store); err != nil {
		return "Sync failed: " + err.Error()
	}
	return ""
}

func syncAtExit(store *data.Store) {
	if err := autoSyncRound(store); err != nil {
		fmt.Fprintf(os.Stderr, "Sync failed: %v\nRun `phasionary sync now` when the server is back.\n", err)
	}
}

func autoSyncRound(store *data.Store) error {
	stateDir, err := config.ResolveStateDir()
	if err != nil {
		return nil
	}
	if d, ok, err := journal.LoadDevice(stateDir); err != nil || !ok || d.Token == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), autoSyncTimeout)
	defer cancel()
	_, err = syncclient.Run(ctx, stateDir, store)
	return err
}
