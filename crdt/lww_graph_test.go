package crdt

import (
	"fmt"
	"testing"

	"github.com/dataramol/aadvcs/models"
)

func TestLWWGraph(t *testing.T) {
	lwwGraph := NewLastWriterWinsGraph("Node1")
	val1 := models.Tree{
		FileName: "dir",
	}
	val2 := models.Blob{
		Content:  "ndskjgjdf",
		FileName: "file1.txt",
	}
	val3 := models.Tree{
		FileName: "lib",
	}
	val4 := models.Blob{
		Content:  "123234356",
		FileName: "file2.txt",
	}
	val5 := models.Tree{
		FileName: "bin",
	}
	val6 := models.Blob{
		Content:  "abceddas",
		FileName: "file3.txt",
	}

	lwwGraph.AddVertex(val1, Tree)
	lwwGraph.AddVertex(val2, Blob)
	lwwGraph.AddVertex(val3, Tree)
	lwwGraph.AddVertex(val4, Blob)
	lwwGraph.AddVertex(val5, Tree)
	lwwGraph.AddVertex(val6, Blob)

	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("file1.txt", Blob), lwwGraph.GetVertexByFilePath("dir", Tree))
	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("lib", Tree), lwwGraph.GetVertexByFilePath("dir", Tree))
	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("file2.txt", Blob), lwwGraph.GetVertexByFilePath("lib", Tree))
	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("file3.txt", Blob), lwwGraph.GetVertexByFilePath("bin", Tree))
	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("bin", Tree), lwwGraph.GetVertexByFilePath("dir", Tree))

	rootVtx := lwwGraph.GetRootVertex()

	fmt.Printf("Root Vertex is %v", rootVtx)

	commitModel := models.CommitModel{
		CommitVersion: 1,
		ParentCommit:  nil,
		CommitMsg:     "test",
	}

	lwwGraph.AddVertex(commitModel, Commit)

	lwwGraph.AddEdge(rootVtx, lwwGraph.GetVertexByValue(commitModel, Commit))

	lwwGraph.PrintGraph()

	fmt.Printf("Root Vertex After Commit is %v", lwwGraph.GetRootVertex())
}
