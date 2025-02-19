package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest changes from remote",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pull called")
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
