package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/dataramol/aadvcs/network"
	"github.com/dataramol/aadvcs/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	LwwGraph *crdt.LastWriterWinsGraph
	Server   *network.Server
)

/*
const (
	aadvcsRootDirName           = ".aadvcs"
	aadvcsCommitDirPath         = ".aadvcs/commit"
	aadvcsCheckoutDirPath       = ".aadvcs/checkout"
	aadvcsStatusFilePath        = ".aadvcs/status.txt"
	aadvcsStagingFilePath       = ".aadvcs/staging_area.txt"
	AadvcsNetworkConfigFilePath = ".aadvcs/network.json"
	aadvcsTimeFormat            = "2006-01-02 03:04:05"
	aadvcsCommitMetadataFile    = "metadata.txt"
)*/

func init() {
	initCmd.Flags().StringP("port", "p", "", "Port for server to run")
	initCmd.Flags().BoolP("start", "s", false, "Start the server")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "This command initializes aadvcs version control system",
	Example: "aadvcs init -p <port>",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetString("port")
		startServer, _ := cmd.Flags().GetBool("start")
		return runInitCommand(port, startServer)
	},
}

func runInitCommand(port string, startServer bool) error {
	if exists, err := utils.CheckPathExists(utils.AadvcsRootDirName); err != nil {
		return err
	} else if exists {
		return errors.New("aadvcs root directory already exists")
	}

	if err := utils.CreateDirectories(utils.AadvcsRootDirName, utils.AadvcsCommitDirPath, utils.AadvcsCheckoutDirPath); err != nil {
		return err
	}

	if err := utils.CreateFile(utils.AadvcsStatusFilePath); err != nil {
		return err
	}

	if err := utils.CreateFile(utils.AadvcsStagingFilePath); err != nil {
		return err
	}

	color.Green("Repository initialised, files are within .aadvcs directory")
	fmt.Printf("Start Server -> %v", startServer)
	if port != "" {
		fp, err := utils.CreateNestedFile(utils.AadvcsNetworkConfigFilePath)
		if err != nil {
			return err
		}
		Server = network.NewServer(fmt.Sprintf(":%s", port))
		LwwGraph = crdt.NewLastWriterWinsGraph(fmt.Sprintf(":%s", port))
		Server.LastWriterWinsGraph = LwwGraph

		wrServer := &models.WritableServer{
			ListAddr: Server.ListenAddress,
		}

		data, err := json.Marshal(wrServer)
		if err != nil {
			return err
		}
		_, _ = fp.Write(data)
		fp.Close()

		if startServer {
			Server.Start()
		}
	}

	return nil
}
