package vectorclock

import (
	"fmt"
	"maps"

	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/stdtypes"
)

// TODO: docs!

type VectorClock struct {
	V map[stdtypes.NodeID]uint32
}

func NewVectorClock() *VectorClock {
	return &VectorClock{
		V: make(map[stdtypes.NodeID]uint32),
	}
}

func VectorClockFromMap(values map[stdtypes.NodeID]uint32) *VectorClock {
	vc := VectorClock{
		V: values,
	}
	return &vc
}

func (vc *VectorClock) Increment(nodeID stdtypes.NodeID) {
	vc.V[nodeID]++
}

func (vc *VectorClock) String() string {
	if len(vc.V) == 0 {
		return "[]"
	}

	str := "["
	// TODO: probably really inefficient as strings are immutable (I think) - problem for later...
	// But - do we really care for a 'to string'?
	sortedNodeIds := maputil.GetSortedKeys(vc.V)
	for _, nodeId := range sortedNodeIds {
		str = str + fmt.Sprintf("%s: %d, ", nodeId, vc.V[nodeId])
	}

	str = str[:len(str)-2]
	str = str + "]"

	return str
}

func (vcA *VectorClock) Combine(vcB *VectorClock) {
	for keyB, valB := range vcB.V {
		if valA, ok := vcA.V[keyB]; !ok || valB > valA {
			vcA.V[keyB] = valB
		}
	}
}

func (vcA *VectorClock) CombineAndIncrement(vcB *VectorClock, nodeId stdtypes.NodeID) {
	vcA.Combine(vcB)
	vcA.Increment(nodeId)
}

func (vc *VectorClock) Clone() *VectorClock {
	return &VectorClock{
		V: maps.Clone(vc.V),
	}
}

func Compare(vcA, vcB *VectorClock) int {
	dir := 0

	for keyA, valA := range vcA.V {
		if valB, ok := vcB.V[keyA]; !ok || valA > valB {
			dir = 1
		}
	}

	for keyB, valB := range vcB.V {
		if valA, ok := vcA.V[keyB]; !ok || valB > valA {
			if dir == 1 {
				return 0
			}
			return -1
		}
	}

	return dir
}

func Less(vcA, vcB *VectorClock) bool {
	for keyA, valA := range vcA.V {
		if valB, ok := vcB.V[keyA]; !ok || valA > valB {
			return false
		}
	}

	return true
}
