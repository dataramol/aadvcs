package network

import (
	"encoding/json"
	"fmt"
	"github.com/dataramol/aadvcs/crdt"
	"os"
	"testing"
)

func TestHandleMerge(t *testing.T) {
	data1, _ := os.ReadFile("D:\\Study\\MSc Project\\test1\\.aadvcs\\commit\\v1\\graph.json")
	data2, _ := os.ReadFile("D:\\Study\\MSc Project\\test2\\.aadvcs\\commit\\v1\\graph.json")

	current := &crdt.LastWriterWinsGraph{}
	incoming := &crdt.LastWriterWinsGraph{}
	_ = json.Unmarshal(data1, current)
	_ = json.Unmarshal(data2, incoming)

	fmt.Printf("Hihello")
	HandleMerge(incoming, current, nil)
}
