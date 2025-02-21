package cmd

import (
	"errors"
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
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
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
			return
		}

		c.Port, err = strconv.Atoi(port)
		if err != nil {
			panic(err)
		}

		r, err := remote.Connect(c)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to connect to remote:", err.Error())
			os.Exit(1)
			return
		}

		if !options.FlagForce {
			if empty, err := r.RepoIsEmpty(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "failed checking if remote directory is empty:", err.Error())
				os.Exit(1)
				return
			} else if !empty {
				_, _ = fmt.Fprintln(os.Stderr, "Remote directory exists and is not empty. Delete it manually or choose another remote directory.")
				os.Exit(2)
				return
			}
		}

		if err := history.Init(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to initialize history:", err.Error())
		}

		if err := c.SaveNew(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to save config:", err.Error())
			_ = r.Close()
			os.Exit(1)
			return
		}

		if err := r.InitialCommit(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to create initial commit:", err.Error())
			_ = r.Close()
			os.Exit(1)
			return
		}

		if err := r.Close(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to disconnect from remote:", err.Error())
			os.Exit(1)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
