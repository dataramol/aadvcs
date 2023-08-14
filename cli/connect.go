package cli

import (
	"encoding/json"
	"fmt"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/dataramol/aadvcs/network"
	"github.com/dataramol/aadvcs/utils"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func init() {
	connectCmd.Flags().StringP("node", "n", "", "Node for server to connect")
	rootCmd.AddCommand(connectCmd)
}

var connectCmd = &cobra.Command{
	Use:     "connect",
	Short:   "This command connects to another node forming a p2p network",
	Example: "aadvcs connect -n <node>",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("node")
		return runConnectCommand(host)
	},
}

func runConnectCommand(host string) error {
	file, err := os.ReadFile(utils.AadvcsNetworkConfigFilePath)
	ws := &models.WritableServer{}
	err = json.Unmarshal(file, ws)
	if err != nil {
		return err
	}

	newServer := network.NewServer(ws.ListAddr)
	newServer.LastWriterWinsGraph = crdt.NewLastWriterWinsGraph(ws.ListAddr)
	go newServer.Start()
	time.Sleep(time.Second * 1)
	err = newServer.Dial(fmt.Sprintf(":%s", host))
	time.Sleep(time.Second * 1)
	if err != nil {
		return err
	}
	fp, err := utils.CreateOrOpenFileRWMode(utils.AadvcsNetworkConfigFilePath)
	err = utils.ClearFileContent(fp)
	if err != nil {
		return err
	}

	ws.Connections = append(ws.Connections, fmt.Sprintf(":%s", host))

	data, err := json.Marshal(ws)
	fp.Write(data)
	fp.Close()
	time.Sleep(time.Second * 2)
	return nil
}
