package cmd

import (
	"errors"
	"log"
	"scribe/internal/config"
	"scribe/internal/remote"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull latest changes from remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("load local config")
		c, err := config.Load()
		if err != nil {
			return errors.Join(errors.New("failed to load config"), err)
		}

		log.Println("connect to remote")
		r, err := remote.Connect(c)
		if err != nil {
			return errors.Join(errors.New("failed to connect to remote"), err)
		}

		defer r.Close()

		log.Println("pull commits from remote")
		if err := r.PullCommits(); err != nil {
			return errors.Join(errors.New("failed to pull commits"), err)
		}

		log.Println("get head commit from remote")
		head, err := r.GetHeadCommit()
		if err != nil {
			return errors.Join(errors.New("failed to get head commit from remote"), err)
		}

		log.Printf("checkout commit %x\n", head.Created)
		if err := r.CheckoutCommit(head); err != nil {
			return errors.Join(errors.New("failed to checkout commit"), err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
