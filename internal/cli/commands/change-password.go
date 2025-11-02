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
	newProjectID       string
	newProjectPassword string
)

// changePasswordCmd represents the changePassword command
var changePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "change password of a project",
	Long: `must be logged in to use

must also be the owner of the project`,
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

		requestBody := api.ChangeProjectRequest{
			Password: newProjectPassword,
		}

		body, err := utils.PatchWithAuth(
			config.APIBaseURL+"/api/projects/"+newProjectID+"/password",
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
	projectCmd.AddCommand(changePasswordCmd)

	// Define flags
	changePasswordCmd.Flags().StringVarP(&newProjectID, "id", "i", "", "ID of the project")
	changePasswordCmd.Flags().StringVarP(&newProjectPassword, "password", "p", "", "New password for the project")

	// Mark flags as required
	changePasswordCmd.MarkFlagRequired("id")
	changePasswordCmd.MarkFlagRequired("password")
}
