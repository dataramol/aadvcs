package cli

import (
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
)

var lwwGraph = crdt.NewLastWriterWinsGraph("node1")

func createLWWGraph() (error, *crdt.LastWriterWinsGraph) {
	noOfCommits, err := getNumberOfChildrenDir(aadvcsCommitDirPath)
	if err != nil {
		return err, nil
	}
	if noOfCommits == 1 {
		err := createGraphForFirstCommit()
		if err != nil {
			return err, nil
		}
	} else {

	}
	return nil, lwwGraph
}

func createGraphForFirstCommit() error {
	metadata, err := createMetadataMap(stagingAreaFile)
	if err != nil {
		return err
	}

	for _, file := range metadata {
		files := strings.Split(file.Path, "\\")
		createVerticesInGraph(files, file.Path)
	}

	for _, file := range metadata {
		files := strings.Split(file.Path, "\\")
		createEdgesInGraph(files)
	}

	lwwGraph.PrintGraph()

	return nil
}

func createVerticesInGraph(items []string, originalPath string) {
	for i := 0; i < len(items)-1; i++ {
		model := models.Tree{
			FileName: items[i],
		}
		if lwwGraph.GetVertexByValue(model) == nil {
			lwwGraph.AddVertex(model, uuid.NewString())
		}
	}

	model := models.Blob{
		FileName: filepath.Base(originalPath),
	}
	if lwwGraph.GetVertexByValue(model) == nil {
		//filePtr, err := createOrOpenFileRWMode(originalPath)
		fileContent, err := os.ReadFile(originalPath)
		if err != nil {
			color.Red("Error for path %v -> ######### %v", originalPath, err)
		}
		//fileScanner := bufio.NewScanner(filePtr)
		color.Yellow("Content --> %v", string(fileContent))
		model.Content = string(fileContent)
		lwwGraph.AddVertex(model, uuid.NewString())
	}
}

func createEdgesInGraph(items []string) {
	for from, to := 0, 1; to < len(items); from, to = from+1, to+1 {
		color.Red("%v %v %v %v", from, to, len(items), items)
		if !lwwGraph.EdgeExists(items[from], items[to]) {
			lwwGraph.AddEdge(items[to], items[from], uuid.NewString())
		}
	}
}
