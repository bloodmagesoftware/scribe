package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"scribe/internal/config"
	"scribe/internal/history"
	"scribe/internal/remote"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone a repository into a new directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Invalid count of arguments for clone: %d. 2 required: clone <share> <target dir>", len(args))
		}

		if err := os.MkdirAll(args[1], 0764); err != nil {
			return err
		}
		if dir, err := os.ReadDir(args[1]); err != nil {
			return err
		} else if len(dir) != 0 {
			return fmt.Errorf("directory %s exists and is not empty", args[1])
		}
		if err := os.Chdir(args[1]); err != nil {
			return err
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

		if pwd, err := keyring.Get(config.KeyringService, c.FullUser()); err == nil {
			c.Password = pwd
		} else if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Password for %s@%s", c.User, c.Host)).
				EchoMode(huh.EchoModePassword).
				Value(&c.Password),
		)).Run(); err != nil {
			return err
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

		log.Println("initialize local history")
		if err := history.Init(); err != nil {
			return errors.Join(errors.New("failed to initialize history"), err)
		}

		log.Println("pull commits from remote")
		if err := r.PullCommits(); err != nil {
			return errors.Join(errors.New("failed to pull commits"), err)
		}

		log.Println("get head commit from remote")
		head, err := r.GetHeadCommit()
		if err != nil {
			return errors.Join(errors.New("failed to get head commit"), err)
		}

		log.Printf("checkout commit %x\n", head.Created)
		if err := r.CloneCommit(head); err != nil {
			return errors.Join(errors.New("failed to checkout commit"), err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
