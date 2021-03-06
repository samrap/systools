package cmd

import (
	"github.com/samrap/systools/pkg/cmd/backups"
	"github.com/sirupsen/logrus"
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
	backups.AttachBackupsCommand(rootCmd)

	// Run the command.
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
