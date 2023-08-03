package crdt

import (
	"fmt"
	"github.com/dataramol/aadvcs/models"
	"github.com/fatih/color"
	"sync"

	"github.com/dataramol/aadvcs/clock"
)

type Vertex struct {
	Id               string
	Value            interface{}
	AdjacentVertices []string
}

type Edge struct {
	Id   string
	From string
	To   string
}

type LastWriterWinsGraph struct {
	NodeId   string
	Vertices map[string]*Vertex
	Edges    map[string]*Edge
	Clock    *clock.VectorClock
	mu       sync.Mutex
}

func NewLastWriterWinsGraph(nodeId string) *LastWriterWinsGraph {
	return &LastWriterWinsGraph{
		NodeId:   nodeId,
		Vertices: make(map[string]*Vertex),
		Edges:    make(map[string]*Edge),
		Clock:    clock.NewVectorClock(nodeId),
	}
}

func (lwwgraph *LastWriterWinsGraph) IncrementClock() {
	lwwgraph.Clock.Increment()
}

func (lwwgraph *LastWriterWinsGraph) AddVertex(Val interface{}, Id string) {
	lwwgraph.mu.Lock()
	defer lwwgraph.mu.Unlock()

	vertex := Vertex{
		Id:    Id,
		Value: Val,
	}
	lwwgraph.Vertices[vertex.Id] = &vertex
}

func (lwwgraph *LastWriterWinsGraph) AddEdge(To string, From string, Id string) {
	lwwgraph.mu.Lock()
	defer lwwgraph.mu.Unlock()

	edge := Edge{
		Id:   Id,
		From: From,
		To:   To,
	}

	lwwgraph.Edges[edge.Id] = &edge
	srcVtx := lwwgraph.GetVertexByValue(models.Tree{
		FileName: From,
	})

	srcVtx.AdjacentVertices = append(srcVtx.AdjacentVertices, To)
}

func (lwwgraph *LastWriterWinsGraph) PrintGraph() {
	fmt.Println("*****Printing Graph*****")
	for _, v := range lwwgraph.Vertices {
		color.Green("Vertex is %v :-> ", v)
		for _, adjVtx := range v.AdjacentVertices {
			color.Yellow("Adjacent Vertex : %v\t", adjVtx)
		}
		fmt.Println()
	}
}

func (lwwgraph *LastWriterWinsGraph) GetVertexByValue(targetValue interface{}) *Vertex {
	for _, vertex := range lwwgraph.Vertices {
		if vertex.Value == targetValue {
			return vertex
		}
	}

	return nil
}

func (lwwgraph *LastWriterWinsGraph) EdgeExists(from string, to string) bool {
	for _, edge := range lwwgraph.Edges {
		if edge.From == from && edge.To == to {
			return true
		}
	}
	return false
}
