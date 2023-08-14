package cli

import (
	"encoding/json"
	"fmt"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/dataramol/aadvcs/utils"
	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
	"os"
	"path/filepath"
	"strings"
)

//var lwwGraph = crdt.NewLastWriterWinsGraph("node1")

func createLWWGraph(commitMsg string, ws *models.WritableServer) (error, *crdt.LastWriterWinsGraph) {

	LwwGraph = crdt.NewLastWriterWinsGraph(ws.ListAddr)
	noOfCommits, err := utils.GetNumberOfChildrenDir(utils.AadvcsCommitDirPath)
	if err != nil {
		return err, nil
	}
	if noOfCommits == 1 {
		err := createGraphForFirstCommit(commitMsg)
		if err != nil {
			return err, nil
		}
	} else {
		currentDir := filepath.Join(utils.AadvcsCommitDirPath, fmt.Sprintf("v%v", noOfCommits-1))
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

		LwwGraph.PrintGraph()

		color.Magenta("After Printing Graph")
	}
	return nil, LwwGraph
}

func createGraphForNonFirstCommit(commitMsg string, previousCommitGraph *crdt.LastWriterWinsGraph, commitVersion int) error {
	metadata, err := createMetadataMap(utils.StagingAreaFile)
	if err != nil {
		return err
	}

	for _, file := range metadata {
		LwwGraph.Paths[file.Path] = ""
		files := strings.Split(file.Path, "\\")
		createVerticesInGraph(files, file.Path)
	}

	for _, file := range metadata {
		files := strings.Split(file.Path, "\\")
		createEdgesInGraph(files)
	}

	for _, vtx := range LwwGraph.Vertices {
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
			if currFrom = LwwGraph.GetVertexByValue(treeModel, crdt.Tree); currFrom == nil {
				//lwwGraph.AddVertex(prevFrom.Value, crdt.Tree)
				//currFrom = lwwGraph.GetVertexByValue(prevFrom.Value.(models.Tree), crdt.Tree)
				currFrom = prevFrom
				LwwGraph.AddVtx(currFrom)
			}
		}

		if prevTo.ModType == crdt.Tree {
			var treeModel models.Tree
			err := mapstructure.Decode(prevTo.Value, &treeModel)
			if err != nil {
				color.Red("Err While Map Structure Decoding -> %v", err)
				return err
			}
			if currTo = LwwGraph.GetVertexByValue(treeModel, crdt.Tree); currTo == nil {
				//lwwGraph.AddVertex(prevTo.Value, crdt.Tree)
				//currTo = lwwGraph.GetVertexByValue(prevTo.Value.(models.Tree), crdt.Tree)
				currTo = prevTo
				LwwGraph.AddVtx(currTo)
			}
		} else if prevTo.ModType == crdt.Blob {
			var blobModel models.Blob
			err := mapstructure.Decode(prevTo.Value, &blobModel)
			if err != nil {
				color.Red("Err While Map Structure Decoding -> %v", err)
				return err
			}
			if currTo = LwwGraph.GetVertexByFilePath(blobModel.FileName, crdt.Blob); currTo == nil {
				//lwwGraph.AddVertex(prevTo.Value, crdt.Blob)
				//currTo = lwwGraph.GetVertexByValue(prevTo.Value.(models.Blob), crdt.Blob)
				currTo = prevTo
				LwwGraph.AddVtx(currTo)
			}
		}

		if currFrom != nil && currTo != nil && currFrom.ModType != crdt.Commit && !LwwGraph.EdgeExists(currFrom, currTo) {
			LwwGraph.AddEdge(currTo, currFrom)
		}

		/*prevRootVtx := previousCommitGraph.GetRootVertex()
		if prevRootVtx.ModType == crdt.Commit {
			prevCommitModel := prevRootVtx.Value.(models.CommitModel).ParentCommit
			prevCommitModel.ParentCommit = &currCommitModel
		}*/

	}

	currRootVtx := LwwGraph.GetRootVertex()
	currCommitModel := models.CommitModel{
		CommitMsg:     commitMsg,
		ParentCommit:  nil,
		CommitVersion: commitVersion,
	}

	LwwGraph.AddVertex(currCommitModel, crdt.Commit)
	LwwGraph.AddEdge(currRootVtx, LwwGraph.GetVertexByValue(currCommitModel, crdt.Commit))

	LwwGraph.Clock.Clock[LwwGraph.NodeId] = previousCommitGraph.Clock.Clock[LwwGraph.NodeId]

	return nil
}

func createGraphForFirstCommit(commitMsg string) error {
	metadata, err := createMetadataMap(utils.StagingAreaFile)
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

	rootDir := LwwGraph.GetRootVertex()

	commitVertex := models.CommitModel{
		CommitMsg:     commitMsg,
		ParentCommit:  nil,
		CommitVersion: 1,
	}

	LwwGraph.AddVertex(commitVertex, crdt.Commit)
	LwwGraph.AddEdge(rootDir, LwwGraph.GetVertexByValue(commitVertex, crdt.Commit))

	return nil
}

func createVerticesInGraph(items []string, originalPath string) {
	for i := 0; i < len(items)-1; i++ {
		if LwwGraph.GetVertexByFilePath(items[i], crdt.Tree) == nil {
			LwwGraph.AddVertex(models.Tree{
				FileName: items[i],
			}, crdt.Tree)
		}
	}

	if LwwGraph.GetVertexByFilePath(filepath.Base(originalPath), crdt.Blob) == nil {
		fileContent, err := os.ReadFile(originalPath)
		if err != nil {
			color.Red("Error for path %v -> ######### %v", originalPath, err)
		}
		LwwGraph.AddVertex(models.Blob{
			FileName: filepath.Base(originalPath),
			Content:  string(fileContent),
		}, crdt.Blob)
		LwwGraph.Paths[originalPath] = string(fileContent)
	}
}

func createEdgesInGraph(items []string) {
	for from, to := 0, 1; to < len(items); from, to = from+1, to+1 {
		color.Red("%v %v %v %v", from, to, len(items), items)

		var ToVertex *crdt.Vertex
		var FromVertex *crdt.Vertex

		if to == len(items)-1 {
			ToVertex = LwwGraph.GetVertexByFilePath(items[to], crdt.Blob)
		} else {
			ToVertex = LwwGraph.GetVertexByFilePath(items[to], crdt.Tree)
		}
		FromVertex = LwwGraph.GetVertexByFilePath(items[from], crdt.Tree)

		if !LwwGraph.EdgeExists(FromVertex, ToVertex) {
			LwwGraph.AddEdge(ToVertex, FromVertex)
		}
	}
}
