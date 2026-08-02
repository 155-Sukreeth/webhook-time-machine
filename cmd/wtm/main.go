package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/155-Sukreeth/webhook-time-machine/internal/cli"
)

//go:embed web/*
var webFS embed.FS

func main() {
	rootCmd := cli.NewRootCmd(webFS)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing wtm: %v\n", err)
		os.Exit(1)
	}
}
