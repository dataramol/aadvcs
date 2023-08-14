package crdt

import (
	"fmt"
	"github.com/dataramol/aadvcs/models"
	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
	"sync"

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
}

type Edge struct {
	From *Vertex
	To   *Vertex
}

type LastWriterWinsGraph struct {
	NodeId   string
	Vertices []*Vertex
	Edges    []*Edge
	Clock    *clock.VectorClock
	mu       sync.Mutex
	Paths    map[string]string
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
		Value:   Val,
		ModType: ModType,
	}
	lwwGraph.Vertices = append(lwwGraph.Vertices, &vertex)
}

func (lwwGraph *LastWriterWinsGraph) AddVtx(vtx *Vertex) {
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
	for _, vertex := range lwwGraph.Vertices {
		if modType == vertex.ModType && modType == Tree {
			if vertex.Value.(models.Tree).FileName == filePath {
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
		if edge.From == from && edge.To == to {
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

	for _, vertex := range lwwGraph.Vertices {
		for _, adjVertex := range vertex.AdjacentVertices {
			vertexMap[adjVertex] = true
		}
	}

	for vertex, isNotRoot := range vertexMap {
		if !isNotRoot {
			return vertex
		}
	}

	return nil
}
