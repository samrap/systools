package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/samrap/systools/pkg/backups"
	"github.com/samrap/systools/pkg/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func attachRestoreCommand(rootCmd *cobra.Command) {
	var flags = &restoreFlags{}
	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore a file or directory from a remote back end",
		Run: func(cmd *cobra.Command, args []string) {
			if name, err := runRestoreCommand(flags); err != nil {
				logrus.Fatalf("Failed to restore %s: %v", name, err)
			} else {
				logrus.Infof("Successfully restored %s", name)
			}
		},
	}

	restoreCmd.Flags().StringVarP(&flags.File, "file", "f", "", "The file to backup. Mutually exclusive to -d")
	restoreCmd.Flags().StringVarP(&flags.Directory, "directory", "d", "", "The directory to backup. Will be stored as a gzipped file. Mutually exclusive to -f")

	rootCmd.AddCommand(restoreCmd)
}

type restoreFlags struct {
	File      string
	Directory string
}

func (rf *restoreFlags) Validate() error {
	if rf.File != "" && rf.Directory != "" {
		return errors.New("Only one of -f or -d is allowed")
	}

	if rf.File == "" && rf.Directory == "" {
		return errors.New("You must specify either a file or directory to back up")
	}

	return nil
}

func runRestoreCommand(flags *restoreFlags) (string, error) {
	session := session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String(os.Getenv("SYSTOOLS_BACKUPS_S3_ENDPOINT")),
		Region:   aws.String(os.Getenv("SYSTOOLS_BACKUPS_S3_REGION")),
	}))

	manager := backups.NewManager(
		backups.NewS3Backend(session, os.Getenv("SYSTOOLS_BACKUPS_S3_BUCKET")),
		backups.NewTimestampVersioner(),
	)

	if flags.File != "" {
		logrus.Info("Restoring file %s", flags.File)

		return flags.File, restoreFile(flags.File, manager)
	}

	logrus.Infof("Restoring directory: %s", flags.Directory)

	return flags.Directory, restoreDirectory(flags.Directory, manager)
}

func restoreFile(filename string, manager backups.Manager) error {
	reader, err := manager.Restore(filename)
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, bytes, os.FileMode(0666))
}

func restoreDirectory(dirname string, manager backups.Manager) error {
	reader, err := manager.Restore(dirname)
	if err != nil {
		return err
	}

	if err = filesystem.ExtractTarball(reader, path.Dir(dirname)); err != nil {
		return fmt.Errorf("Could not restore from tarball: %v", err)
	}
	return nil
}
