/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "manage hookbridge projects",
	Long:  `create, list, delete, update password, etc...`,
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
