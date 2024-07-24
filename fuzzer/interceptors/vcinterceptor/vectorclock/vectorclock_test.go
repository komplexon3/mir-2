package vectorclock_test

import (
	"reflect"
	"testing"

	"github.com/filecoin-project/mir/fuzzer/interceptors/vcinterceptor/vectorclock"
	"github.com/filecoin-project/mir/stdtypes"
)

func TestVectorClockIncrement(t *testing.T) {
	vc := vectorclock.NewVectorClock()
	nodeID := stdtypes.NodeID("1")

	vc.Increment(nodeID)

	expected := vectorclock.VectorClockFromMap(map[stdtypes.NodeID]uint32{nodeID: 1})
	if !reflect.DeepEqual(vc, expected) {
		t.Errorf("Incrementing node %s: expected %v, got %v", nodeID, expected, vc)
	}
}

func TestVectorClockString(t *testing.T) {
	vc := vectorclock.NewVectorClock()
	vc.Increment("1")
	vc.Increment("2")

	expected := "[1: 1, 2: 1]"
	result := vc.String()

	if result != expected {
		t.Errorf("Expected string representation %s, got %s", expected, result)
	}
}

func TestVectorClockCombine(t *testing.T) {
	vcA := vectorclock.NewVectorClock()
	vcA.Increment("1")
	vcA.Increment("2")
	vcA.Increment("4")

	vcB := vectorclock.NewVectorClock()
	vcB.Increment("1")
	vcB.Increment("1")
	vcB.Increment("2")
	vcB.Increment("3")

	vcA.Combine(vcB)

	expected := vectorclock.VectorClockFromMap(map[stdtypes.NodeID]uint32{"1": 2, "2": 1, "3": 1, "4": 1})
	if !reflect.DeepEqual(vcA, expected) {
		t.Errorf("Combining vector clocks: expected %v, got %v", expected, vcA)
	}
}

func TestVectorClockCompare(t *testing.T) {
	vcA := vectorclock.NewVectorClock()
	vcA.Increment("1")

	vcB := vectorclock.NewVectorClock()
	vcB.Increment("1")

	result := vectorclock.Compare(vcA, vcB)

	if result != 0 {
		t.Errorf("Comparing equal vector clocks: expected 0, got %d", result)
	}

	vcA = vectorclock.NewVectorClock()
	vcA.Increment("1")

	vcB = vectorclock.NewVectorClock()
	vcB.Increment("2")

	result = vectorclock.Compare(vcA, vcB)

	vcA = vectorclock.NewVectorClock()
	vcA.Increment("1")

	vcB = vectorclock.NewVectorClock()

	result = vectorclock.Compare(vcA, vcB)

	if result != 1 {
		t.Errorf("Comparing different vector clocks: expected 1, got %d", result)
	}

	vcA = vectorclock.NewVectorClock()

	vcB = vectorclock.NewVectorClock()
	vcB.Increment("1")

	result = vectorclock.Compare(vcA, vcB)

	if result != -1 {
		t.Errorf("Comparing different vector clocks: expected -1, got %d", result)
	}

	vcA = vectorclock.NewVectorClock()
	vcA.Increment("1")
	vcA.Increment("1")

	vcB = vectorclock.NewVectorClock()
	vcB.Increment("1")
	vcB.Increment("1")
	vcB.Increment("2")

	result = vectorclock.Compare(vcA, vcB)

	if result != -1 {
		t.Errorf("Comparing different vector clocks: expected -1, got %d", result)
	}

	vcA = vectorclock.NewVectorClock()
	vcA.Increment("1")
	vcA.Increment("1")
	vcA.Increment("2")

	vcB = vectorclock.NewVectorClock()
	vcB.Increment("1")
	vcB.Increment("1")

	result = vectorclock.Compare(vcA, vcB)

	if result != 1 {
		t.Errorf("Comparing different vector clocks: expected 1, got %d", result)
	}

	vcA = vectorclock.NewVectorClock()
	vcA.Increment("1")
	vcA.Increment("1")
	vcA.Increment("2")

	vcB = vectorclock.NewVectorClock()
	vcB.Increment("1")
	vcB.Increment("1")
	vcB.Increment("1")

	result = vectorclock.Compare(vcA, vcB)

	if result != 0 {
		t.Errorf("Comparing different vector clocks: expected 0, got %d", result)
	}

	vcA = vectorclock.VectorClockFromMap(map[stdtypes.NodeID]uint32{"0": 340, "1": 340, "2": 340, "3": 340, "4": 327})
	vcB = vectorclock.VectorClockFromMap(map[stdtypes.NodeID]uint32{"0": 352, "1": 348, "2": 340, "3": 348, "4": 345})

	result = vectorclock.Compare(vcA, vcB)

	if result != -1 {
		t.Errorf("Comparing different vector clocks: expected -1, got %d", result)
	}

	result = vectorclock.Compare(vcB, vcA)

	if result != 1 {
		t.Errorf("Comparing different vector clocks: expected 1, got %d", result)
	}
}
