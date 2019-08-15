package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func attachRestoreCommand(rootCmd *cobra.Command) {
	var restoreCommand = &cobra.Command{
		Use:   "restore",
		Short: "Restore a file or directory from a remote back end",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Not implemented")
		},
	}

	rootCmd.AddCommand(restoreCommand)
}
