package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock a file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lock not implemented")
	},
}

func init() {
	rootCmd.AddCommand(lockCmd)
}
