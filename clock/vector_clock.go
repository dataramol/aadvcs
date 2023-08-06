package clock

import (
	"fmt"
	"sync"
)

// VectorClock represents the vector Clock for a node in system.
type VectorClock struct {
	NodeId string
	Clock  map[string]int64
	mu     sync.Mutex
}

type Ordering int

const (
	CONCURRENT Ordering = iota
	HappensBefore
	HappensAfter
	IDENTICAL
	NotComparable
)

// NewVectorClock Creates new instance of vector Clock
func NewVectorClock(id string) *VectorClock {
	clock := make(map[string]int64)
	clock[id] = 0
	return &VectorClock{
		NodeId: id,
		Clock:  clock,
	}
}

func (v *VectorClock) AddNode(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, exists := v.Clock[id]; !exists {
		v.Clock[id] = 0
	}
}

// Increment Increments the local counter for given node
func (v *VectorClock) Increment() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Clock[v.NodeId]++
}

// Merge merges the current vector Clock with another vector Clock
func (v *VectorClock) Merge(other *VectorClock) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for nodeId, counter := range other.Clock {
		if counter > v.Clock[nodeId] {
			v.Clock[nodeId] = counter
		}
	}
}

// Compare compares the two vector clocks and returns the ordering relationship
// return relation of vector Clock in argument with respect to current vector Clock
// If method returns "HAPPENS_AFTER" that means , other happened after v
// If method returns "HAPPENS_BEFORE" that means, other happened before v
// other __________ v
func (v *VectorClock) Compare(other *VectorClock) Ordering {
	v.mu.Lock()
	defer v.mu.Unlock()
	var ordering Ordering
	if len(v.Clock) != len(other.Clock) {
		return NotComparable
	}

	for node := range v.Clock {
		if _, exists := other.Clock[node]; !exists {
			return NotComparable
		}
	}

	vAfterOther := false
	vBeforeOther := false
	concurrent := false

	for node := range v.Clock {
		clock1 := v.Clock[node]
		clock2 := other.Clock[node]

		if clock1 > clock2 {
			vAfterOther = true
		}
		if clock2 > clock1 {
			vBeforeOther = true
		}
		if vBeforeOther && vAfterOther {
			ordering = CONCURRENT
			concurrent = true
			break
		}
	}
	if vBeforeOther && !vAfterOther {
		ordering = HappensAfter
	} else if vAfterOther && !vBeforeOther {
		ordering = HappensBefore
	} else if !vAfterOther && !vBeforeOther && !concurrent {
		ordering = IDENTICAL
	}

	return ordering
}

// HappenedBefore v.HappenedBefore(other) --> true means v happened before other
func (v *VectorClock) HappenedBefore(other *VectorClock) bool {
	return other.Compare(v) == HappensBefore
}

// HappenedAfter v.HappenedAfter(other) --> true means v happened after other
func (v *VectorClock) HappenedAfter(other *VectorClock) bool {
	return other.Compare(v) == HappensAfter
}

func (v *VectorClock) IsConcurrent(other *VectorClock) bool {
	return other.Compare(v) == CONCURRENT
}

func (v *VectorClock) Print() {
	fmt.Printf("Vector Clock at Node Id : %v\n", v.NodeId)
	for nodeId, clock := range v.Clock {
		fmt.Printf("Clock for Node %v is %v\n", nodeId, clock)
	}
}
