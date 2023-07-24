package clock

import (
	"fmt"
	"sync"
)

// Vector clock represents the vector clock for a node in system.
type VectorClock struct {
	nodeId string
	clock  map[string]int64
	mu     sync.Mutex
}

type Ordering int

const (
	CONCURRENT Ordering = iota
	HAPPENS_BEFORE
	HAPPENS_AFTER
	IDENTICAL
	NOT_COMPARABLE
)

// Creates new instance of vector clock
func NewVectorClock(id string) *VectorClock {
	clock := make(map[string]int64)
	clock[id] = 0
	return &VectorClock{
		nodeId: id,
		clock:  clock,
	}
}

func (v *VectorClock) AddNode(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, exists := v.clock[id]; !exists {
		v.clock[id] = 0
	}
}

// Increments the local counter for given node
func (v *VectorClock) Increment() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.clock[v.nodeId]++
}

// merges the current vector clock with another vector clock
func (v *VectorClock) Merge(other *VectorClock) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for nodeId, counter := range other.clock {
		if counter > v.clock[nodeId] {
			v.clock[nodeId] = counter
		}
	}
}

// compares the two vector clocks and returns the ordering relationship
// return relation of vector clock in argument with respect to current vector clock
// If method returns "HAPPENS_AFTER" that means , other happened after v
// If method returns "HAPPENS_BEFORE" that means, other happened before v
// other __________ v
func (v *VectorClock) Compare(other *VectorClock) Ordering {
	v.mu.Lock()
	defer v.mu.Unlock()
	var ordering Ordering
	if len(v.clock) != len(other.clock) {
		return NOT_COMPARABLE
	}

	for node := range v.clock {
		if _, exists := other.clock[node]; !exists {
			return NOT_COMPARABLE
		}
	}

	vAfterOther := false
	vBeforeOther := false
	concurrent := false

	for node := range v.clock {
		clock1 := v.clock[node]
		clock2 := other.clock[node]

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
		ordering = HAPPENS_AFTER
	} else if vAfterOther && !vBeforeOther {
		ordering = HAPPENS_BEFORE
	} else if !vAfterOther && !vBeforeOther && !concurrent {
		ordering = IDENTICAL
	}

	return ordering
}

// v.HappenedBefore(other) --> true means v happened before other
func (v *VectorClock) HappenedBefore(other *VectorClock) bool {
	return other.Compare(v) == HAPPENS_BEFORE
}

// v.HappenedAfter(other) --> true means v happened after other
func (v *VectorClock) HappenedAfter(other *VectorClock) bool {
	return other.Compare(v) == HAPPENS_AFTER
}

func (v *VectorClock) IsConcurrent(other *VectorClock) bool {
	return other.Compare(v) == CONCURRENT
}

func (v *VectorClock) Print() {
	fmt.Printf("Vector Clock at Node Id : %v\n", v.nodeId)
	for nodeId, clock := range v.clock {
		fmt.Printf("Clock for Node %v is %v\n", nodeId, clock)
	}
}
