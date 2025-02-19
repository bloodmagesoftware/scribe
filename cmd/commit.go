package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:     "commit",
	Aliases: []string{"push"},
	Short:   "commit changes to remote",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("commit not implemented")
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
