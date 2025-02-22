package cmd

import (
	"os"
	"scribe/internal/options"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "scb",
	Short: "SFTP based VCS",
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVar(&options.FlagForce, "force", false, "enforce an illegal action, which could lead to unintentional data loss")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
