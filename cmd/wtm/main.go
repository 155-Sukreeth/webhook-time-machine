package main

import (
	"fmt"
	"os"

	"github.com/155-Sukreeth/webhook-time-machine/internal/cli"
	"github.com/155-Sukreeth/webhook-time-machine/web"
)

func main() {
	rootCmd := cli.NewRootCmd(web.Assets)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing wtm: %v\n", err)
		os.Exit(1)
	}
}
