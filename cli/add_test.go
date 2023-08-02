package cli

import (
	"fmt"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"strings"
	"testing"
)

func TestCreateMetadataMap(t *testing.T) {
	//rootDir := "D:\\Study\\MSc Project\\Codebase\\aadvcs\\"
	mp, err := createMetadataMap("D:\\Study\\MSc Project\\Codebase\\aadvcs\\.aadvcs\\status.txt")

	//lwwGraph := crdt.NewLastWriterWinsGraph("node1")
	//cntr := 0
	/*for path, _ := range mp {
		if fileInfo, _ := os.Stat(rootDir + path); fileInfo.IsDir() {
			fmt.Printf("%v is directory", path)
			//lwwGraph.AddVertex(createVertex(createTreeModel()), string(rune(cntr)))
		} else {
			fmt.Printf("%v is not directory", path)
			//lwwGraph.AddVertex(createVertex(createBlobModel()), string(rune(cntr)))
		}
		cntr++
		fmt.Println()
	}
	*/

	for path, _ := range mp {
		components := strings.Split(path, "\\")
		currDir := ""
		for _, component := range components[:len(components)-1] {
			currDir = joinPath(currDir, component)
			fmt.Printf("%v\n", currDir)
		}
	}

	// Print the map

	if err != nil {
		t.Fatalf("Error : %v", err)
	}
	fmt.Printf("Map -> %v", mp)

}

func joinPath(dir, component string) string {
	if dir == "" {
		return component
	}
	return fmt.Sprintf("%s\\%s", dir, component)
}

func createVertex(val interface{}) *crdt.Vertex {
	vtx := crdt.Vertex{
		Value: val,
	}
	return &vtx
}

func createBlobModel() *models.Blob {
	blob := models.Blob{}
	return &blob
}

func createTreeModel() *models.Tree {
	tree := models.Tree{}
	return &tree
}
