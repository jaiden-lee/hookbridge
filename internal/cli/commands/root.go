package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hookbridge",
	Short: "hookbridge is an easy way to test webhooks locally",
	Long: `hookbridge works by having a server that forwards HTTP requests 
from webhook providers (i.e. Stripe) to all connected devices (think of it 
like a tunnel with multiple endpoints).

the hookbridge cli is how you connect devices to the hookbridge server.
this connection is called a tunnel/project.

to get started, type "hookbridge login" to login or create a new account.
NOTE: you only need to create a new account if you want to create a tunnel/project.
otherwise, you can just connect directly to an existing tunnel/project`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
