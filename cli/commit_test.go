package cli

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestCommit(t *testing.T) {
	trackedFile := "D:\\Study\\MSc Project\\Codebase\\aadvcs\\.aadvcs\\status.txt"
	//commitDir := "D:\\Study\\MSc Project\\Codebase\\aadvcs\\.aadvcs\\commit"
	//noOfDirectory, err := getNumberOfChildrenDir(commitDir)
	//newCommitDirName := filepath.Join(commitDir, fmt.Sprintf("v%v", noOfDirectory+1))
	metadata, err := createMetadataMap(trackedFile)
	assert.Nil(t, err)
	for _, file := range metadata {
		/*paths := strings.Split(file.Path, "\\")
		for _, path := range paths {
			path, _ = filepath.Dir(path)
			info, _ := os.Stat(path)
			fmt.Printf("%v is Directory ? -> %v\n", path, info.IsDir())
		}*/
		info, err := filepath.Abs(file.Path)
		fmt.Printf("%v ------ %v--------%v\n", filepath.Dir(file.Path), info, err)
	}

	/*g := crdt.NewLastWriterWinsGraph("node1")
	tm := models.Tree{FileName: "dir"}
	g.AddVertex(tm, "v1")

	/*for id, v := range g.Vertices {
		tm := v.Value.(models.Tree)
		fmt.Printf("vertex for id %v is %v", id, tm.FileName)
	}

	retval := g.GetVertexByValue(models.Tree{FileName: "dfh"})
	fmt.Printf("Vertex Value : %v", retval)
	*/
}
