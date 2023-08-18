package crdt

import (
	"fmt"
	"github.com/dataramol/aadvcs/models"
	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
	"sync"
	"time"

	"github.com/dataramol/aadvcs/clock"
)

type ModelType int

const (
	Tree ModelType = iota
	Blob
	Commit
)

type Vertex struct {
	Value            interface{}
	AdjacentVertices []*Vertex
	ModType          ModelType
	TimeStamp        time.Time
}

type Edge struct {
	From *Vertex
	To   *Vertex
}

type LastWriterWinsGraph struct {
	NodeId       string
	Vertices     []*Vertex
	Edges        []*Edge
	Clock        *clock.VectorClock
	mu           sync.Mutex
	Paths        map[string]string
	TimeStamp    time.Time
	LatestCommit *models.CommitModel
}

func NewLastWriterWinsGraph(nodeId string) *LastWriterWinsGraph {
	return &LastWriterWinsGraph{
		NodeId: nodeId,
		Clock:  clock.NewVectorClock(nodeId),
		Paths:  make(map[string]string),
	}
}

func (lwwGraph *LastWriterWinsGraph) IncrementClock() {
	lwwGraph.Clock.Increment()
}

func (lwwGraph *LastWriterWinsGraph) AddVertex(Val interface{}, ModType ModelType) {
	lwwGraph.mu.Lock()
	defer lwwGraph.mu.Unlock()

	vertex := Vertex{
		Value:     Val,
		ModType:   ModType,
		TimeStamp: time.Now(),
	}
	lwwGraph.Vertices = append(lwwGraph.Vertices, &vertex)
}

func (lwwGraph *LastWriterWinsGraph) AddVtx(vtx *Vertex) {
	if vtx.TimeStamp.IsZero() {
		vtx.TimeStamp = time.Now()
	}
	lwwGraph.Vertices = append(lwwGraph.Vertices, vtx)
}

func (lwwGraph *LastWriterWinsGraph) AddEdge(To *Vertex, From *Vertex) {
	lwwGraph.mu.Lock()
	defer lwwGraph.mu.Unlock()

	edge := Edge{
		From: From,
		To:   To,
	}

	lwwGraph.Edges = append(lwwGraph.Edges, &edge)

	From.AdjacentVertices = append(From.AdjacentVertices, To)
}

func (lwwGraph *LastWriterWinsGraph) PrintGraph() {
	fmt.Println("*****Printing Graph*****")
	for _, v := range lwwGraph.Vertices {
		color.Green("Vertex is %v :-> ", v.Value)
		for _, adjVtx := range v.AdjacentVertices {
			color.Yellow("Adjacent Vertex : %v\t", adjVtx.Value)
		}
		fmt.Println()
	}
}

func (lwwGraph *LastWriterWinsGraph) GetVertexByValue(targetValue interface{}, modType ModelType) *Vertex {
	for _, vertex := range lwwGraph.Vertices {
		if vertex.ModType == modType && vertex.Value == targetValue {
			return vertex
		}
	}

	return nil
}

func (lwwGraph *LastWriterWinsGraph) GetVertexByFilePath(filePath string, modType ModelType) *Vertex {
	var blobModel models.Blob
	var treeModel models.Tree
	for _, vertex := range lwwGraph.Vertices {
		if modType == vertex.ModType && modType == Tree {
			mapstructure.Decode(vertex.Value, &treeModel)
			if treeModel.FileName == filePath {
				return vertex
			}
		} else if modType == vertex.ModType && modType == Blob {
			mapstructure.Decode(vertex.Value, &blobModel)
			if blobModel.FileName == filePath {
				return vertex
			}
		}
	}

	return nil
}

func (lwwGraph *LastWriterWinsGraph) EdgeExists(from *Vertex, to *Vertex) bool {
	for _, edge := range lwwGraph.Edges {
		if convertValueMapToStruct(edge.From) == convertValueMapToStruct(from) && convertValueMapToStruct(edge.To) == convertValueMapToStruct(to) {
			return true
		}
	}
	return false
}

func (lwwGraph *LastWriterWinsGraph) GetRootVertex() *Vertex {
	vertexMap := make(map[*Vertex]bool)

	for _, vertex := range lwwGraph.Vertices {
		vertexMap[vertex] = false
	}

	for _, edge := range lwwGraph.Edges {
		vertexMap[edge.To] = true
	}

	for vertex, isNotRoot := range vertexMap {
		if !isNotRoot {
			return vertex
		}
	}

	return nil
}

func DeepCopy(destination *LastWriterWinsGraph, source *LastWriterWinsGraph) *LastWriterWinsGraph {
	if source == nil || destination == nil {
		return nil
	}

	var copiedVertices []*Vertex
	for _, v := range source.Vertices {
		copiedV := *v
		copiedVertices = append(copiedVertices, &copiedV)
	}

	var copiedEdges []*Edge
	for _, e := range source.Edges {
		copiedE := *e
		copiedEdges = append(copiedEdges, &copiedE)
	}

	copiedPaths := make(map[string]string)
	for k, v := range source.Paths {
		copiedPaths[k] = v
	}

	copiedCommit := &models.CommitModel{
		CommitMsg:     source.LatestCommit.CommitMsg,
		CommitVersion: source.LatestCommit.CommitVersion,
		ParentCommit:  source.LatestCommit.ParentCommit,
	}

	destination.Vertices = copiedVertices
	destination.Edges = copiedEdges
	destination.Paths = copiedPaths
	destination.LatestCommit = copiedCommit
	destination.TimeStamp = source.TimeStamp

	return destination
}

func convertValueMapToStruct(V *Vertex) any {
	var treeModel models.Tree
	var blobModel models.Blob
	var commitModel models.CommitModel

	if V.ModType == Tree {
		mapstructure.Decode(V.Value, &treeModel)
		return treeModel
	} else if V.ModType == Blob {
		mapstructure.Decode(V.Value, &blobModel)
		return blobModel
	} else if V.ModType == Commit {
		mapstructure.Decode(V.Value, &commitModel)
		return commitModel
	}

	return fmt.Errorf("unknow value type received : %v", V.ModType)
}
