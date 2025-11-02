/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jaiden-lee/hookbridge/internal/cli/config"
	"github.com/jaiden-lee/hookbridge/internal/cli/utils"
	"github.com/jaiden-lee/hookbridge/pkg/api"
	"github.com/spf13/cobra"
)

var projectID string

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "deletes a specified project",
	Long: `deletes a specified project
	
you must be logged in to use this method.
you must also be the owner/original creator of the specified project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := config.LoadUserConfig()
		if err != nil {
			return err
		}

		if !config.IsLoggedIn(user) {
			fmt.Println("\nYou aren't logged in, please login or create an account before creating a project")
			fmt.Println()
			return nil
		}

		body, err := utils.DeleteWithAuth(
			config.APIBaseURL+"/api/projects/"+projectID,
			user,
		)

		if err != nil {
			return err
		}

		var response api.MessageResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return err
		}

		fmt.Println("\n" + response.Message)
		fmt.Println()
		return nil
	},
}

func init() {
	projectCmd.AddCommand(deleteCmd)

	// Define flags
	deleteCmd.Flags().StringVarP(&projectID, "id", "i", "", "ID of the project")

	// Mark flags as required
	deleteCmd.MarkFlagRequired("id")
}
