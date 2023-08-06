package cli

import (
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/fatih/color"
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
		if lwwGraph.GetVertexByFilePath(items[i], crdt.Tree) == nil {
			lwwGraph.AddVertex(models.Tree{
				FileName: items[i],
			}, crdt.Tree)
		}
	}

	if lwwGraph.GetVertexByFilePath(filepath.Base(originalPath), crdt.Blob) == nil {
		fileContent, err := os.ReadFile(originalPath)
		if err != nil {
			color.Red("Error for path %v -> ######### %v", originalPath, err)
		}
		lwwGraph.AddVertex(models.Blob{
			FileName: filepath.Base(originalPath),
			Content:  string(fileContent),
		}, crdt.Blob)
	}
}

func createEdgesInGraph(items []string) {
	for from, to := 0, 1; to < len(items); from, to = from+1, to+1 {
		color.Red("%v %v %v %v", from, to, len(items), items)

		var ToVertex *crdt.Vertex
		var FromVertex *crdt.Vertex

		if to == len(items)-1 {
			ToVertex = lwwGraph.GetVertexByFilePath(items[to], crdt.Blob)
		} else {
			ToVertex = lwwGraph.GetVertexByFilePath(items[to], crdt.Tree)
		}
		FromVertex = lwwGraph.GetVertexByFilePath(items[from], crdt.Tree)

		if !lwwGraph.EdgeExists(FromVertex, ToVertex) {
			lwwGraph.AddEdge(ToVertex, FromVertex)
		}
	}
}
