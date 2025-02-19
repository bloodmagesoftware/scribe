package cmd

import (
	"fmt"
	"os"
	"scribe/internal/config"
	"scribe/internal/remote"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull latest changes from remote",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.Load()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to load config:", err.Error())
			os.Exit(1)
			return
		}

		r, err := remote.Connect(c)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to connect to remote:", err.Error())
			os.Exit(1)
			return
		}

		if err := r.PullCommits(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to pull commits:", err.Error())
			_ = r.Close()
			os.Exit(1)
			return
		}

		if err := r.Close(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to disconnect from remote:", err.Error())
			os.Exit(1)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
