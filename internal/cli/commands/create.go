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

var (
	projectName     string
	projectPassword string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new project",
	Long: `create a new project. 

a project is an instance of a tunnel that anyone can connect to
(provided they have the right password)
	
you must be signed in to use this command. `,
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

		requestBody := api.CreateProjectRequest{
			Name:     projectName,
			Password: projectPassword,
		}

		body, err := utils.PostWithAuth(
			config.APIBaseURL+"/api/projects",
			requestBody,
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
	rootCmd.AddCommand(createCmd)

	// Define flags
	createCmd.Flags().StringVarP(&projectName, "name", "n", "", "Name of the project")
	createCmd.Flags().StringVarP(&projectPassword, "password", "p", "", "Password for the project")

	// Mark flags as required
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("password")
}
