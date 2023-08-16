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
	"github.com/sirupsen/logrus"
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
		return runCommitCommand(utils.StagingAreaFile, msg)
	},
}

func runCommitCommand(trackedFilePath, msg string) error {
	noOfDirectory, err := utils.GetNumberOfChildrenDir(utils.AadvcsCommitDirPath)

	if err != nil {
		return err
	}

	newCommitDirName := filepath.Join(utils.AadvcsCommitDirPath, fmt.Sprintf("v%v", noOfDirectory+1))
	commitTime := time.Now()
	err = createCommitMetadataFile(newCommitDirName, msg, commitTime)
	if err != nil {
		return err
	}

	metadata, err := createMetadataMap(trackedFilePath)
	if err != nil {
		return err
	}

	for _, file := range metadata {
		destCommitFilePath := filepath.Join(newCommitDirName, file.Path)

		destFilePtr, _ := utils.CreateNestedFile(destCommitFilePath)
		originalFilePtr, _ := os.Open(file.Path)
		_, _ = io.Copy(destFilePtr, originalFilePtr)

		destFilePtr.Close()
		originalFilePtr.Close()
	}

	stagingFilePtr, _ := utils.CreateOrOpenFileRWMode(utils.StagingAreaFile)
	defer stagingFilePtr.Close()

	ws := GetNetworkConfig()

	err, LwwGraph = createLWWGraph(msg, ws)
	color.Red("Error After Graph Creation :- %v", err)
	LwwGraph.PrintGraph()
	LwwGraph.TimeStamp = commitTime
	if LwwGraph != nil {
		fp, err := utils.CreateNestedFile(filepath.Join(newCommitDirName, "graph.json"))
		color.Red("Error After Creating Graph File - %v", err)
		LwwGraph.IncrementClock()
		if err == nil {
			color.Yellow("Marshalling Graph Now")
			jsonData, err := json.MarshalIndent(LwwGraph, "", "")
			color.Red("***Error While Marshalling*** -> %v", err)
			color.Magenta("%v", string(jsonData))
			_, _ = fp.Write(jsonData)
			fp.Close()
		}
		err = SendUpdateOverNetwork(LwwGraph, ws)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"LWWGraph": LwwGraph,
				"Network":  ws,
				"Error":    err,
			}).Error("Failed to send update over network")
		}
	}

	err = utils.ClearFileContent(stagingFilePtr)

	select {}
	return nil
}

func createCommitMetadataFile(commitDirectory, commitMsg string, commitDate time.Time) error {
	commitMetadataFP, err := utils.CreateNestedFile(filepath.Join(commitDirectory, utils.AadvcsCommitMetadataFile))
	if err != nil {
		return err
	}
	defer commitMetadataFP.Close()

	_, _ = commitMetadataFP.WriteString(fmt.Sprintf("%v%v%v", commitMsg, utils.Separator, commitDate.Format(utils.AadvcsTimeFormat)))

	return nil
}

func SendUpdateOverNetwork(LwwGraph *crdt.LastWriterWinsGraph, ws *models.WritableServer) error {
	newServer := network.NewServer(ws.ListAddr)
	newServer.LastWriterWinsGraph = LwwGraph

	go newServer.Start()
	time.Sleep(time.Millisecond * 200)

	for _, connection := range ws.Connections {
		fmt.Printf("Dialing %v", connection)
		err := newServer.Dial(connection)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 2)
	}

	To := make([]string, len(ws.Connections))
	for _, peer := range ws.Connections {
		To = append(To, peer)
	}
	fmt.Printf("Now Broacasting msg")
	err := newServer.Broadcast(network.BroadcastTo{
		To:      To,
		Payload: newServer.LastWriterWinsGraph,
	}, false)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 2)

	return nil
}

func GetNetworkConfig() *models.WritableServer {
	file, err := os.ReadFile(utils.AadvcsNetworkConfigFilePath)
	ws := &models.WritableServer{}
	err = json.Unmarshal(file, ws)
	if err != nil {
		logrus.Error("Error while reading network config ---> %+v", err)
	}

	return ws
}
