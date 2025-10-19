package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "sign in or create a new account",
	Long:  `sign in or create a new account`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("signing user in...")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
