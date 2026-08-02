package cli

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/155-Sukreeth/webhook-time-machine/internal/app"
	"github.com/155-Sukreeth/webhook-time-machine/internal/config"
)

var (
	cfgFile         string
	flagPort        int
	flagUIPort      int
	flagForwardURL  string
	flagDBPath      string
	flagStripSigs   bool
)

func NewRootCmd(webFS embed.FS) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "wtm",
		Short: "Local Webhook Time Machine — capture, inspect, edit, and replay webhooks locally",
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".wtm.yaml", "config file path")
	rootCmd.PersistentFlags().IntVarP(&flagPort, "port", "p", 8080, "proxy listening port")
	rootCmd.PersistentFlags().IntVar(&flagUIPort, "ui-port", 8081, "dashboard UI port")
	rootCmd.PersistentFlags().StringVarP(&flagForwardURL, "forward-to", "f", "http://localhost:3000", "local application forward target URL")
	rootCmd.PersistentFlags().StringVar(&flagDBPath, "db", "./wtm.db", "SQLite database file path")
	rootCmd.PersistentFlags().BoolVar(&flagStripSigs, "strip-signatures", true, "automatically strip webhook signature headers on replay")

	// wtm init
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize default .wtm.yaml configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ".wtm.yaml"
			if cfgFile != "" {
				target = cfgFile
			}
			if err := app.InitConfig(target); err != nil {
				return err
			}
			fmt.Printf("Successfully created configuration file: %s\n", target)
			return nil
		},
	}

	// wtm start
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start proxy server and dashboard UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				return err
			}

			// CLI Flag overrides
			if cmd.Flags().Changed("port") {
				cfg.Port = flagPort
			}
			if cmd.Flags().Changed("ui-port") {
				cfg.UIPort = flagUIPort
			}
			if cmd.Flags().Changed("forward-to") {
				cfg.ForwardURL = flagForwardURL
			}
			if cmd.Flags().Changed("db") {
				cfg.DBPath = flagDBPath
			}
			if cmd.Flags().Changed("strip-signatures") {
				cfg.StripSignatures = flagStripSigs
			}

			application := app.New(cfg, webFS)
			return application.Run(cmd.Context())
		},
	}

	// wtm version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print wtm version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("wtm v1.0.0 (Local Webhook Time Machine)")
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}
