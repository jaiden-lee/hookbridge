package cmd

import (
	"encoding/json"

	"github.com/jaiden-lee/hookbridge/internal/cli/config"
	"github.com/jaiden-lee/hookbridge/internal/cli/utils"
	"github.com/jaiden-lee/hookbridge/pkg/api"
	"github.com/spf13/cobra"

	"fmt"
	"strings"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "displays a list of all projects",
	Long: `displays a list of all projects user owns.
must be logged in to use this method`,
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

		body, err := utils.GetWithAuth(
			config.APIBaseURL+"/api/projects",
			nil,
			user,
		)

		if err != nil {
			return err
		}

		var response api.ProjectsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return err
		}

		fmt.Printf("\n%-20s %-40s\n", "PROJECT_ID", "NAME")
		fmt.Println(strings.Repeat("-", 60))

		for _, project := range response.Projects {
			fmt.Printf("%-20s %-40s\n", project.ID, project.Name)
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
