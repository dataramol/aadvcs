package cli

import (
	"encoding/json"
	"fmt"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
	"os"
	"path/filepath"
	"strings"
)

var lwwGraph = crdt.NewLastWriterWinsGraph("node1")

func createLWWGraph(commitMsg string) (error, *crdt.LastWriterWinsGraph) {
	noOfCommits, err := getNumberOfChildrenDir(aadvcsCommitDirPath)
	if err != nil {
		return err, nil
	}
	if noOfCommits == 1 {
		err := createGraphForFirstCommit(commitMsg)
		if err != nil {
			return err, nil
		}
	} else {
		currentDir := filepath.Join(aadvcsCommitDirPath, fmt.Sprintf("v%v", noOfCommits-1))
		color.Red("Current Dir -> ", currentDir)
		color.Red("Creating Graph For non first Commit")
		pth := filepath.Join(currentDir, "graph.json")
		color.Yellow("Path -> ", pth)
		file, err := os.ReadFile(pth)
		if err != nil {
			color.Red("Err While Reading -> %v", err)
			return err, nil
		}
		prevCommitGraph := crdt.NewLastWriterWinsGraph("node1")
		err = json.Unmarshal(file, prevCommitGraph)
		if err != nil {
			color.Red("Err While Unmarshalling -> %v", err)
			return err, nil
		}
		err = createGraphForNonFirstCommit(commitMsg, prevCommitGraph, noOfCommits)
		if err != nil {
			color.Red("Err While Graph Creation -> %v", err)
			return err, nil
		}

		color.Magenta("Before Printing Graph")

		lwwGraph.PrintGraph()

		color.Magenta("After Printing Graph")
	}
	return nil, lwwGraph
}

func createGraphForNonFirstCommit(commitMsg string, previousCommitGraph *crdt.LastWriterWinsGraph, commitVersion int) error {
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

	for _, vtx := range lwwGraph.Vertices {
		color.Yellow("Vertex :- %T", vtx.Value)
	}

	for _, edge := range previousCommitGraph.Edges {
		prevFrom := edge.From
		prevTo := edge.To
		var currFrom *crdt.Vertex
		var currTo *crdt.Vertex

		if prevFrom.ModType == crdt.Tree {
			var treeModel models.Tree
			err := mapstructure.Decode(prevFrom.Value, &treeModel)
			if err != nil {
				color.Red("Err While Map Structure Decoding -> %v", err)
				return err
			}
			color.Magenta("Previous From -> %T", prevFrom)
			if currFrom = lwwGraph.GetVertexByValue(treeModel, crdt.Tree); currFrom == nil {
				//lwwGraph.AddVertex(prevFrom.Value, crdt.Tree)
				//currFrom = lwwGraph.GetVertexByValue(prevFrom.Value.(models.Tree), crdt.Tree)
				currFrom = prevFrom
				lwwGraph.AddVtx(currFrom)
			}
		}

		if prevTo.ModType == crdt.Tree {
			var treeModel models.Tree
			err := mapstructure.Decode(prevTo.Value, &treeModel)
			if err != nil {
				color.Red("Err While Map Structure Decoding -> %v", err)
				return err
			}
			if currTo = lwwGraph.GetVertexByValue(treeModel, crdt.Tree); currTo == nil {
				//lwwGraph.AddVertex(prevTo.Value, crdt.Tree)
				//currTo = lwwGraph.GetVertexByValue(prevTo.Value.(models.Tree), crdt.Tree)
				currTo = prevTo
				lwwGraph.AddVtx(currTo)
			}
		} else if prevTo.ModType == crdt.Blob {
			var blobModel models.Blob
			err := mapstructure.Decode(prevTo.Value, &blobModel)
			if err != nil {
				color.Red("Err While Map Structure Decoding -> %v", err)
				return err
			}
			if currTo = lwwGraph.GetVertexByFilePath(blobModel.FileName, crdt.Blob); currTo == nil {
				//lwwGraph.AddVertex(prevTo.Value, crdt.Blob)
				//currTo = lwwGraph.GetVertexByValue(prevTo.Value.(models.Blob), crdt.Blob)
				currTo = prevTo
				lwwGraph.AddVtx(currTo)
			}
		}

		if currFrom != nil && currTo != nil && currFrom.ModType != crdt.Commit && !lwwGraph.EdgeExists(currFrom, currTo) {
			lwwGraph.AddEdge(currTo, currFrom)
		}

		/*prevRootVtx := previousCommitGraph.GetRootVertex()
		if prevRootVtx.ModType == crdt.Commit {
			prevCommitModel := prevRootVtx.Value.(models.CommitModel).ParentCommit
			prevCommitModel.ParentCommit = &currCommitModel
		}*/

	}

	currRootVtx := lwwGraph.GetRootVertex()
	currCommitModel := models.CommitModel{
		CommitMsg:     commitMsg,
		ParentCommit:  nil,
		CommitVersion: commitVersion,
	}

	lwwGraph.AddVertex(currCommitModel, crdt.Commit)
	lwwGraph.AddEdge(currRootVtx, lwwGraph.GetVertexByValue(currCommitModel, crdt.Commit))

	lwwGraph.Clock.Clock[lwwGraph.NodeId] = previousCommitGraph.Clock.Clock[lwwGraph.NodeId]

	return nil
}

func createGraphForFirstCommit(commitMsg string) error {
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

	rootDir := lwwGraph.GetRootVertex()

	commitVertex := models.CommitModel{
		CommitMsg:     commitMsg,
		ParentCommit:  nil,
		CommitVersion: 1,
	}

	lwwGraph.AddVertex(commitVertex, crdt.Commit)
	lwwGraph.AddEdge(rootDir, lwwGraph.GetVertexByValue(commitVertex, crdt.Commit))

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
