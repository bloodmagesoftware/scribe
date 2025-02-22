package cmd

import (
	"errors"
	"fmt"
	"log"
	"scribe/internal/config"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "prints the clone uri",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("load local config")
		c, err := config.Load()
		if err != nil {
			return errors.Join(errors.New("failed to load config"), err)
		}

		fmt.Printf("%s@%s:%d#%s\n", c.User, c.Host, c.Port, c.Path)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(shareCmd)
}
