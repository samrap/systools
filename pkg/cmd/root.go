package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute initializes and executes a systools command.
func Execute() {
	// Create the root command.
	var rootCmd = &cobra.Command{
		Use:   "systools",
		Short: "Systools are system tools for common server tasks",
	}

	// Attach all sub commands.
	attachBackupCommand(rootCmd)
	attachRestoreCommand(rootCmd)

	// Run the command.
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
