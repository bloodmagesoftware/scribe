package cmd

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"scribe/internal/config"
	"scribe/internal/remote"
	"strconv"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone a repository into a new directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Invalid count of arguments for clone: %d. 2 required: clone <share> <target dir>", len(args))
		}

		shareRegexp := regexp.MustCompile(`^(.+)@([^:]+):(\d+)#(.+)$`)
		matches := shareRegexp.FindStringSubmatch(args[0])
		if len(matches) != 5 {
			return fmt.Errorf("failed to parse share string %s", args[0])
		}

		port, _ := strconv.Atoi(matches[3])
		c := &config.Config{
			Version: config.Version,
			Host:    matches[2],
			Port:    port,
			User:    matches[1],
			Path:    matches[4],
		}

		log.Println("connect to remote")
		r, err := remote.Connect(c)
		if err != nil {
			return errors.Join(errors.New("failed to connect to remote"), err)
		}
		defer r.Close()

		if err := c.SaveNew(); err != nil {
			return errors.Join(errors.New("failed to save new config file"), err)
		}

		log.Println("pull commits from remote")
		if err := r.PullCommits(); err != nil {
			return errors.Join(errors.New("failed to pull commits"), err)
		}

		log.Println("get head commit from remote")
		head, err := r.GetHeadCommit()
		if err != nil {
			return errors.Join(errors.New("failed to disconnect from remote"), err)
		}

		log.Printf("checkout commit %x\n", head.Created)
		if err := r.CheckoutCommit(head); err != nil {
			return errors.Join(errors.New("failed to disconnect from remote"), err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
