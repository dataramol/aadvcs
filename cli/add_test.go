package cli

import (
	"fmt"
	"testing"
)

func TestCreateMetadataMap(t *testing.T) {
	mp, err := createMetadataMap("D:/Study/MSc Project/Codebase/aadvcs/.aadvcs/status.txt")
	if err != nil {
		t.Fatalf("Error : %v", err)
	}
	fmt.Printf("Map -> %v", mp)

}
