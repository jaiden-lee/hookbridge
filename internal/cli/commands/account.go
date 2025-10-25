/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jaiden-lee/hookbridge/internal/cli/config"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "account",
	Short: "check if you are logged in and current account details",
	Long:  `check if you are logged in and current account details`,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := config.LoadUserConfig()
		if err != nil {
			return err
		}

		if config.IsLoggedIn(user) {
			fmt.Println("\nYou are logged in as:", user.Email)
			fmt.Println()
			return nil
		}

		fmt.Println("\nYou are not currently signed in")
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
