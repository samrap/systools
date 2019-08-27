package backups

import "github.com/spf13/cobra"

// AttachBackupsCommand attaches the backups command and all of its subcommands.
func AttachBackupsCommand(rootCmd *cobra.Command) {
	var backupsCmd = &cobra.Command{
		Use:   "backups",
		Short: "Manage backups of the filesystem",
	}

	attachBackupCommand(backupsCmd)
	attachRestoreCommand(backupsCmd)

	rootCmd.AddCommand(backupsCmd)
}
