package cli

import (
	"encoding/json"
	"fmt"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/dataramol/aadvcs/network"
	"github.com/dataramol/aadvcs/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

func init() {
	rootCmd.AddCommand(startServerCmd)
}

var startServerCmd = &cobra.Command{
	Use:     "startserver",
	Short:   "This command starts the server at node",
	Example: "aadvcs startserver",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStartServerCommand()
	},
}

func runStartServerCommand() error {
	exists, err := utils.CheckPathExists(utils.AadvcsNetworkConfigFilePath)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("server not initialised")
	}

	ws := GetNetworkConfig()
	Server = network.NewServer(ws.ListAddr)
	go Server.Start()
	time.Sleep(time.Second * 1)
	noOfCommits, err := utils.GetNumberOfChildrenDir(utils.AadvcsCommitDirPath)
	if noOfCommits != 0 {
		latestCommitGraphPath := filepath.Join(utils.AadvcsCommitDirPath, fmt.Sprintf("v%v", noOfCommits), "graph.json")
		file, err := os.ReadFile(latestCommitGraphPath)
		if err != nil {
			return err
		}
		LwwGraph = &crdt.LastWriterWinsGraph{}
		err = json.Unmarshal(file, LwwGraph)
		if err != nil {
			logrus.Error("Error while reading latest commit graph ---> %+v", err)
		}
	} else {
		LwwGraph = crdt.NewLastWriterWinsGraph(ws.ListAddr)
	}
	Server.LastWriterWinsGraph = LwwGraph
	file, err := os.ReadFile(utils.AadvcsNetworkConfigFilePath)
	fp, _ := utils.CreateOrOpenFileRWMode(utils.AadvcsNetworkCommunicationFilePath)
	newWs := &models.WritableServer{}
	err = json.Unmarshal(file, newWs)
	time.Sleep(time.Second * 10)
	To := make([]string, len(newWs.Connections))
	for _, conn := range newWs.Connections {
		fmt.Printf("Dialing %v", conn)
		err := Server.Dial(conn)
		if err != nil {
			return err
		}
		To = append(To, conn)
	}

	Server.Broadcast(network.BroadcastTo{
		To:      To,
		Payload: LwwGraph,
	}, false)

	err = utils.ClearFileContent(fp)
	if err != nil {
		return err
	}
	select {}
	return nil

}
