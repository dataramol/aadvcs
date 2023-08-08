package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
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

	err, lwwGraph = createLWWGraph(msg)
	color.Red("Error After Graph Creation :- %v", err)
	lwwGraph.PrintGraph()
	if lwwGraph != nil {
		fp, err := createNestedFile(filepath.Join(newCommitDirName, "graph.json"))
		color.Red("Error After Creating Graph File - %v", err)
		lwwGraph.IncrementClock()
		if err == nil {
			color.Yellow("Marshalling Graph Now")
			jsonData, err := json.MarshalIndent(lwwGraph, "", "")
			color.Red("***Error While Marshalling*** -> %v", err)
			color.Magenta("%v", string(jsonData))
			_, _ = fp.Write(jsonData)
			fp.Close()
		}
	}

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
