package cli

import (
	"encoding/json"
	"github.com/dataramol/aadvcs/crdt"
	"os"
	"testing"
)

func TestCommit(t *testing.T) {
	graphFile1 := "D:\\Study\\MSc Project\\Codebase\\aadvcs\\.aadvcs\\commit\\v1\\graph.json"
	graphFile2 := "D:\\Study\\MSc Project\\Codebase\\aadvcs\\.aadvcs\\commit\\v2\\graph.json"
	file1, err := os.ReadFile(graphFile1)
	if err != nil {
		return
	}
	file2, err := os.ReadFile(graphFile2)
	if err != nil {
		return
	}
	var graph1 crdt.LastWriterWinsGraph
	var graph2 crdt.LastWriterWinsGraph
	err = json.Unmarshal(file1, &graph1)
	if err != nil {
		return
	}
	err = json.Unmarshal(file2, &graph2)
	if err != nil {
		return
	}

	graph1.PrintGraph()
	graph2.PrintGraph()
}
