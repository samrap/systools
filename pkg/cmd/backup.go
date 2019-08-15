package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func attachBackupCommand(rootCmd *cobra.Command) {
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup a file or directory to a remote back end",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Not implemented")
		},
	}

	rootCmd.AddCommand(backupCmd)
}
