/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/jaiden-lee/hookbridge/internal/cli/config"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "signs out of current account",
	Long:  `signs out of current account`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := config.DeleteUserConfig()
		if errors.Is(err, config.ErrNotSignedIn) {
			fmt.Println("\nYou are not currently signed in")
			fmt.Println()
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Println("\nYou have been successfully signed out")
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
