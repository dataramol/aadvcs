package clock

import (
	"testing"
)

func TestVectorClock(t *testing.T) {
	//Created two vector clocks for two nodes
	vc1 := NewVectorClock("node1")
	vc2 := NewVectorClock("node2")

	//Initialise and setup node
	vc1.AddNode(vc2.nodeId)
	vc2.AddNode(vc1.nodeId)

	vc1.Increment()
	if !vc1.HappenedAfter(vc2) {
		t.Error("Expected VC1 To Happen After VC2")
	}
	vc2.Increment()
	if !vc1.IsConcurrent(vc2) {
		t.Error("Expected VC2 To Be Concurrent with VC1")
	}

	vc1.Merge(vc2)
	vc2.Merge(vc1)

	if vc1.Compare(vc2) != 3 {
		t.Error("Expected VC1 and VC2 to be identical")
	}

	vc2.Increment()
	if !vc2.HappenedAfter(vc1) {
		t.Error("Expected VC2 To Happen After VC1")
	}
}
