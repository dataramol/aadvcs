package cli

import (
	"errors"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	statusFile      = filepath.Join(aadvcsRootDirName, "status.txt")
	stagingAreaFile = filepath.Join(aadvcsRootDirName, "staging_area.txt")
)

const (
	aadvcsRootDirName        = ".aadvcs"
	aadvcsCommitDirPath      = ".aadvcs/commit"
	aadvcsCheckoutDirPath    = ".aadvcs/checkout"
	aadvcsStatusFilePath     = ".aadvcs/status.txt"
	aadvcsStagingFilePath    = ".aadvcs/staging_area.txt"
	aadvcsTimeFormat         = "2006-01-02 03:04:05"
	aadvcsCommitMetadataFile = "metadata.txt"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "This command initializes aadvcs version control system",
	Example: "aadvcs init",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInitCommand()
	},
}

func runInitCommand() error {
	if exists, err := checkPathExists(aadvcsRootDirName); err != nil {
		return err
	} else if exists {
		return errors.New("aadvcs root directory already exists")
	}

	if err := createDirectories(aadvcsRootDirName, aadvcsCommitDirPath, aadvcsCheckoutDirPath); err != nil {
		return err
	}

	if err := createFile(aadvcsStatusFilePath); err != nil {
		return err
	}

	if err := createFile(aadvcsStagingFilePath); err != nil {
		return err
	}

	color.Green("Repository initialised, files are within .aadvcs directory")

	return nil
}
