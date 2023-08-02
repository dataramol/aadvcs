package cli

import (
	"bufio"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/google/uuid"
	"path/filepath"
	"strings"
)

var lwwGraph *crdt.LastWriterWinsGraph

func createLWWGraph() (error, *crdt.LastWriterWinsGraph) {
	noOfCommits, err := getNumberOfChildrenDir(aadvcsCommitDirPath)
	if err != nil {
		return err, nil
	}

	if noOfCommits == 0 {
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
		createEdgesinGraph(files)
	}

	return nil
}

func createVerticesInGraph(items []string, originalPath string) {
	for i := 0; i < len(items)-1; i++ {
		model := models.Tree{
			FileName: items[i],
		}
		if lwwGraph.GetVertexByValue(model) != nil {
			lwwGraph.AddVertex(model, uuid.NewString())
		}
	}

	model := models.Blob{
		FileName: filepath.Base(originalPath),
	}
	if lwwGraph.GetVertexByValue(model) != nil {
		filePtr, _ := createOrOpenFileRWMode(originalPath)
		fileScanner := bufio.NewScanner(filePtr)
		model.Content = string(fileScanner.Bytes())
		lwwGraph.AddVertex(model, uuid.NewString())
		filePtr.Close()
	}
}

func createEdgesinGraph(items []string) {
	for from, to := 0, 1; to < len(items); from, to = from+1, to+1 {
		if !lwwGraph.EdgeExists(items[from], items[to]) {
			lwwGraph.AddEdge(items[to], items[from], uuid.NewString())
		}
	}
}
