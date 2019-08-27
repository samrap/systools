package backups

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/samrap/systools/pkg/backups"
	"github.com/samrap/systools/pkg/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func attachBackupCommand(rootCmd *cobra.Command) {
	var flags = &backupFlags{}
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup a file or directory to a remote back end",
		Run: func(cmd *cobra.Command, args []string) {
			if err := flags.Validate(); err != nil {
				logrus.Fatal(err)
			}

			if name, err := runBackupCommand(flags); err != nil {
				logrus.Fatalf("Failed to back up %s: %v", name, err)
			} else {
				logrus.Infof("Successfully backed up %s", name)
			}
		},
	}

	backupCmd.Flags().StringVarP(&flags.File, "file", "f", "", "The file to backup. Mutually exclusive to -d")
	backupCmd.Flags().StringVarP(&flags.Directory, "directory", "d", "", "The directory to backup. Will be stored as a gzipped file. Mutually exclusive to -f")

	rootCmd.AddCommand(backupCmd)
}

type backupFlags struct {
	File      string
	Directory string
}

func (bf *backupFlags) Validate() error {
	if bf.File != "" && bf.Directory != "" {
		return errors.New("Only one of -f or -d is allowed")
	}

	if bf.File == "" && bf.Directory == "" {
		return errors.New("You must specify either a file or directory to back up")
	}

	return nil
}

func runBackupCommand(flags *backupFlags) (string, error) {
	session := session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String(os.Getenv("SYSTOOLS_BACKUPS_S3_ENDPOINT")),
		Region:   aws.String(os.Getenv("SYSTOOLS_BACKUPS_S3_REGION")),
	}))

	manager := backups.NewManager(
		backups.NewS3Backend(session, os.Getenv("SYSTOOLS_BACKUPS_S3_BUCKET")),
		backups.NewTimestampVersioner(),
	)

	if flags.File != "" {
		logrus.Infof("Backing up file %s", flags.File)

		return flags.File, backupFile(flags.File, manager)
	}

	logrus.Infof("Backing up directory: %s", flags.Directory)

	return flags.Directory, backupDirectory(flags.Directory, manager)
}

func backupFile(filename string, manager backups.Manager) error {
	reader, err := os.Open(filename)
	if err != nil {
		return err
	}

	return manager.Backup(filename, reader)
}

func backupDirectory(dirname string, manager backups.Manager) error {
	logrus.Info("Creating tarball")

	tarpath, err := filesystem.CreateTarball(dirname, os.TempDir())
	if err != nil {
		return fmt.Errorf("Unable to create tarball for directory %s: %v", dirname, err)
	}

	reader, err := os.Open(tarpath)
	if err != nil {
		return fmt.Errorf("Unable to open tarball for reading: %v", err)
	}

	logrus.Info("Tarball created. Uploading back up")

	return manager.Backup(dirname, reader)
}
