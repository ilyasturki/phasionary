package main

import (
	"os"

	"phasionary/internal/server"
)

func main() {
	if err := server.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
