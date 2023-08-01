package cli

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"time"
)

func init() {
	commitCmd.Flags().StringP("message", "m", "", "commit message")
	_ = commitCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(commitCmd)
}

var commitCmd = &cobra.Command{
	Use:     "commit",
	Short:   "This allows you to save all file changes",
	Example: `aadvcs commit -m "<your message>"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		if msg == "" {
			return errors.New("commit message cannot be empty")
		}
		return runCommitCommand(stagingAreaFile, msg)
	},
}

func runCommitCommand(trackedFilePath, msg string) error {
	noOfDirectory, err := getNumberOfChildrenDir(aadvcsCommitDirPath)

	if err != nil {
		return err
	}

	newCommitDirName := filepath.Join(aadvcsCommitDirPath, fmt.Sprintf("v%v", noOfDirectory+1))
	err = createCommitMetadataFile(newCommitDirName, msg, time.Now())
	if err != nil {
		return err
	}

	metadata, err := createMetadataMap(trackedFilePath)
	if err != nil {
		return err
	}

	for _, file := range metadata {
		destCommitFilePath := filepath.Join(newCommitDirName, file.Path)

		destFilePtr, _ := createNestedFile(destCommitFilePath)
		originalFilePtr, _ := os.Open(file.Path)
		_, _ = io.Copy(destFilePtr, originalFilePtr)

		destFilePtr.Close()
		originalFilePtr.Close()
	}

	stagingFilePtr, _ := createOrOpenFileRWMode(stagingAreaFile)
	defer stagingFilePtr.Close()

	err = clearFileContent(stagingFilePtr)

	return nil
}

func createCommitMetadataFile(commitDirectory, commitMsg string, commitDate time.Time) error {
	commitMetadataFP, err := createNestedFile(filepath.Join(commitDirectory, aadvcsCommitMetadataFile))
	if err != nil {
		return err
	}
	defer commitMetadataFP.Close()

	_, _ = commitMetadataFP.WriteString(fmt.Sprintf("%v%v%v", commitMsg, separator, commitDate.Format(aadvcsTimeFormat)))

	return nil
}
