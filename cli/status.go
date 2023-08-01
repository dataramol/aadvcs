package cli

import (
	"bufio"
	"github.com/dataramol/aadvcs/models"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"io"
	"os"
	"sort"
)

type MetadataArr []models.FileMetaData

func (arr MetadataArr) Len() int {
	return len(arr)
}

func (arr MetadataArr) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func (arr MetadataArr) Less(i, j int) bool {
	return arr[i].Path > arr[j].Path
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "This allows you to track status of all files",
	Example: "aadvcs status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatusCommand(os.Stdout)
	},
}

func runStatusCommand(writer io.Writer) error {
	metadataArr, err := readMetadataFromStagingFile()
	if err != nil {
		return err
	}

	printStatus(writer, metadataArr)
	return nil
}

func printStatus(writer io.Writer, metadataArr MetadataArr) {
	if len(metadataArr) == 0 {
		color.Green("No changes in staging area !")
		return
	}

	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"File Name", "Status", "Last Modified"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor},
	)

	for _, metadata := range metadataArr {
		row := []string{metadata.Path, string(metadata.Status), metadata.ModificationTime}
		statusColor := tablewriter.FgGreenColor
		if metadata.Status == models.StatusUpdated {
			statusColor = tablewriter.FgHiMagentaColor
		}
		table.Rich(row, []tablewriter.Colors{{}, {tablewriter.Bold, statusColor}, {}})
	}

	table.Render()
}

func readMetadataFromStagingFile() (MetadataArr, error) {
	filePtr, err := createOrOpenFileRWMode(stagingAreaFile)
	if err != nil {
		return nil, err
	}
	defer filePtr.Close()

	updatedStatusMap := make(map[string]models.FileMetaData)

	fileScanner := bufio.NewScanner(filePtr)
	for fileScanner.Scan() {
		metadata := extractFileMetadataFromLine(fileScanner.Text())
		updatedStatusMap[metadata.Path] = metadata
	}

	var metadataArr MetadataArr
	for _, val := range updatedStatusMap {
		metadataArr = append(metadataArr, val)
	}

	sort.Sort(metadataArr)

	return metadataArr, nil
}
