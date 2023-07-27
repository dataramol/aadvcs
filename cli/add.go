package cli

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/dataramol/aadvcs/models"
)

type metadataMap map[string]models.FileMetaData

func init() {
	rootCmd.AddCommand(addCommand)
}

var addCommand = &cobra.Command{
	Use:     "add",
	Short:   "For tracking file status : Created or Modified",
	Example: "aadvcs add test.txt",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAddCommand(stagingAreaFile, args)
	},
}

func runAddCommand(stagingAreaFilePath string, filePaths []string) error {

	metadata, err := createMetadataMap(statusFile)

	if err != nil {
		return err
	}

	prepareStagingArea(filePaths, metadata)

	statusFilePtr, _ := createOrOpenFileRWMode(stagingAreaFilePath)
	defer statusFilePtr.Close()

	stagingFilePtr, _ := CreateOrOpenFileAppendMode(stagingAreaFilePath)
	defer stagingFilePtr.Close()

	if err := clearFileContent(statusFilePtr); err != nil {
		return err
	}

	for _, mtdata := range metadata {
		line := formatMetaData(&mtdata)

		_, _ = statusFilePtr.WriteString(line)
		if mtdata.GoToStaging {
			_, _ = stagingFilePtr.WriteString(line)
		}
	}

	return nil
}

func formatMetaData(mtdata *models.FileMetaData) string {
	return fmt.Sprintf("%v%v%v%v%v\n", mtdata.Path, separator, mtdata.ModificationTime, separator, string(mtdata.Status))
}

func extractMetadata(metadata metadataMap, filePath string, fileCurrModTime time.Time) {
	fileCurrModTimeStr := fileCurrModTime.Format(aadvcsTimeFormat)
	mtdata, ok := metadata[filePath]

	metadataStruct := models.FileMetaData{
		Path:             filePath,
		ModificationTime: fileCurrModTimeStr,
		GoToStaging:      true,
		Status:           models.StatusCreated,
	}

	if ok && fileCurrModTimeStr != mtdata.ModificationTime {
		metadataStruct.Status = models.StatusUpdated
	}

	metadata[filePath] = metadataStruct
}

func createMetadataMap(filePath string) (metadataMap, error) {
	trackedFile, err := createOrOpenFileRWMode(filePath)
	if err != nil {
		return nil, err
	}
	defer trackedFile.Close()

	fileUpdateMap := make(metadataMap)
	fileScanner := bufio.NewScanner(trackedFile)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		fmd := extractFileMetadataFromLine(line)
		fileUpdateMap[fmd.Path] = fmd
	}

	if err := fileScanner.Err(); err != nil {
		return nil, err
	}

	return fileUpdateMap, nil
}

func walkDirectory(root string, metadata metadataMap) {
	_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			extractMetadata(metadata, path, info.ModTime())
		}
		return nil
	})
}

func prepareStagingArea(filePaths []string, metadata metadataMap) {
	for _, filePath := range filePaths {
		fileInfo, _ := os.Stat(filePath)

		if fileInfo.IsDir() {
			walkDirectory(filePath, metadata)
		} else {
			extractMetadata(metadata, filePath, fileInfo.ModTime())
		}
	}
}
