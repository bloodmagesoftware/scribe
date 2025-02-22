package cmd

import (
	"errors"
	"fmt"
	"log"
	"scribe/internal/config"
	"scribe/internal/diff"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "view local changes to current commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("load local config")
		c, err := config.Load()
		if err != nil {
			return errors.Join(errors.New("failed to load config"), err)
		}

		log.Println("get current commit")
		currentCommit, err := c.CurrentCommit()
		if err != nil {
			return errors.Join(errors.New("failed to get current commit"), err)
		}

		log.Printf("diff %x to local changes\n", currentCommit.Created)
		locallyChanged, err := diff.LocalFromCommit(c, currentCommit)
		if err != nil {
			return errors.Join(errors.New("failed to diff local changes with current commit"), err)
		}

		for _, d := range locallyChanged {
			switch d.Type {
			case diff.DiffTypeCreate:
				fmt.Print("+ ")
			case diff.DiffTypeModify:
				fmt.Print("~ ")
			case diff.DiffTypeDelete:
				fmt.Print("- ")
			}
			fmt.Println(d.Path)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
