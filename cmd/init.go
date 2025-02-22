package cmd

import (
	"errors"
	"fmt"
	"log"
	"scribe/internal/config"
	"scribe/internal/history"
	"scribe/internal/options"
	"scribe/internal/remote"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

const art = `
      #######                                   /                
    /       ###                          #    #/                 
   /         ##                         ###   ##                 
   ##        #                           #    ##                 
    ###                                       ##                 
   ## ###           /###   ###  /###   ###    ## /###     /##    
    ### ###        / ###  / ###/ #### / ###   ##/ ###  / / ###   
      ### ###     /   ###/   ##   ###/   ##   ##   ###/ /   ###  
        ### /##  ##          ##          ##   ##    ## ##    ### 
          #/ /## ##          ##          ##   ##    ## ########  
           #/ ## ##          ##          ##   ##    ## #######   
            # /  ##          ##          ##   ##    ## ##        
  /##        /   ###     /   ##          ##   ##    /# ####    / 
 /  ########/     ######/    ###         ### / ####/    ######/  
/     #####        #####      ###         ##/   ###      #####   
|                                                                
 \)                                                              

`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Scribe repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(art)
		c, err := config.Load()
		if err != nil {
			err = nil
			c = &config.Config{Ignore: config.DefaultIgnore}
		}
		port := "22"
		if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Host").
				Value(&c.Host),
			huh.NewInput().
				Title("Port").
				Validate(func(s string) error {
					i, err := strconv.Atoi(s)
					if err != nil {
						return err
					}
					if i < 0 || i > 65535 {
						return errors.New("out of range 0-65535")
					}
					return nil
				}).
				Value(&port),
			huh.NewInput().
				Title("User").
				Value(&c.User),
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&c.Password),
			huh.NewInput().
				Title("Path").
				Value(&c.Path),
		)).Run(); err != nil {
			return err
		}

		c.Port, err = strconv.Atoi(port)
		if err != nil {
			panic(err)
		}

		log.Println("connect to remote")
		r, err := remote.Connect(c)
		if err != nil {
			return errors.Join(errors.New("failed to connect to remote"), err)
		}

        defer r.Close()

		if !options.FlagForce {
			log.Println("check if remote repo does not exist")
			if empty, err := r.RepoIsEmpty(); err != nil {
				return errors.Join(errors.New("failed checking if remote directory is empty"), err)
			} else if !empty {
				return errors.New("Remote directory exists and is not empty. Delete it manually or choose another remote directory.")
			}
		}

		log.Println("initialize local history")
		if err := history.Init(); err != nil {
			return errors.Join(errors.New("failed to initialize history"), err)
        }



		log.Println("save local scribe config")
		if err := c.SaveNew(); err != nil {
			return errors.Join(errors.New("failed to save config"), err)
		}

		log.Println("create inital commit")
		if err := r.InitialCommit(); err != nil {
			return errors.Join(errors.New("failed to create initial commit"), err)
		}

        return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
