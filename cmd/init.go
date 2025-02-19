package cmd

import (
	"errors"
	"fmt"
	"os"
	"scribe/internal/config"
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
		c := &config.Config{Ignore: config.DefaultIgnore}
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

		var err error
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

		if err := c.Save(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to save config:", err.Error())
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
