package cmd

import (
	"errors"
	"log"
	"scribe/internal/config"
	"scribe/internal/options"
	"scribe/internal/remote"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:     "commit",
	Aliases: []string{"push"},
	Short:   "commit changes to remote",
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

		msg := strings.Join(options.FlagMessage, "\n")
		if len(options.FlagMessage) == 0 {
			if err := huh.NewForm(huh.NewGroup(
				huh.NewText().
					Title("commit message").
					ShowLineNumbers(true).
					Validate(func(s string) error {
						if len(s) != 0 {
							return nil
						} else {
							return errors.New("message must not be empty")
						}
					}).
					Value(&msg),
			)).Run(); err != nil {
				return err
			}
		}

		log.Println("creating commit")
		if err := r.Commit(msg); err != nil {
			return errors.Join(errors.New("failed to create initial commit"), err)
		}

		return nil
	},
}

func init() {
	commitCmd.Flags().StringArrayVarP(&options.FlagMessage, "message", "m", options.FlagMessage, "Use the given value as the commit message. If multiple -m options are given, their values are concatenated as separate paragraphs.")
	rootCmd.AddCommand(commitCmd)
}
