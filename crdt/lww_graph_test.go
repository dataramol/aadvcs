package crdt

import (
	"testing"

	"github.com/dataramol/aadvcs/models"
)

func TestLWWGraph(t *testing.T) {
	lwwGraph := NewLastWriterWinsGraph("Node1")
	val1 := models.Tree{
		FileMode: "R",
		FileName: "dir",
	}
	val2 := models.Blob{
		Content:  "ndskjgjdf",
		FileName: "file1.txt",
	}
	val3 := models.Tree{
		FileMode: "R",
		FileName: "lib",
	}
	val4 := models.Blob{
		Content:  "123234356",
		FileName: "file2.txt",
	}

	lwwGraph.AddVertex(val1, "val1")
	lwwGraph.AddVertex(val2, "val2")
	lwwGraph.AddVertex(val3, "val3")
	lwwGraph.AddVertex(val4, "val4")

	lwwGraph.AddEdge("val2", "val1", "edge1")
	lwwGraph.AddEdge("val3", "val1", "edge2")
	lwwGraph.AddEdge("val4", "val3", "edge3")

	lwwGraph.PrintGraph()

}
