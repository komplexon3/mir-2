package actions

import (
	"errors"
	"math/rand/v2"
	"sort"
)

// Weighted random selection adapted from https://github.com/mroth/weightedrand

const (
	intSize   = 32 << (^uint(0) >> 63) // cf. strconv.IntSize
	maxInt    = 1<<(intSize-1) - 1
	maxUint64 = 1<<64 - 1
)

// Possible errors returned by NewChooser, preventing the creation of a Chooser
// with unsafe runtime states.
var (
	// If the sum of provided Action weights exceed the maximum integer value
	// for the current platform (e.g. math.MaxInt32 or math.MaxInt64), then
	// the internal running total will overflow, resulting in an imbalanced
	// distribution generating improper results.
	errWeightOverflow = errors.New("sum of Action Weights exceeds max int")
	// If there are no Choices available to the Chooser with a weight >= 1,
	// there are no valid choices and Pick would produce a runtime panic.
	errNoValidChoices = errors.New("zero Actions with Weight >= 1")
)

type RandomActions struct {
	rand    *rand.Rand
	actions []WeightedAction
	totals  []int
	max     int
}

type WeightedAction struct {
	action Action
	weight int
}

func NewWeightedAction(action Action, weight int) WeightedAction {
	return WeightedAction{
		action: action,
		weight: weight,
	}
}

func NewRandomActions(weightedActions []WeightedAction, rand *rand.Rand) (*RandomActions, error) {
	sort.Slice(weightedActions, func(i, j int) bool {
		return weightedActions[i].weight < weightedActions[j].weight
	})

	totals := make([]int, len(weightedActions))
	runningTotal := 0
	for i, c := range weightedActions {
		if c.weight < 0 {
			continue // ignore negative weights, can never be picked
		}

		// case of single ~uint64 or similar value that exceeds maxInt on its own
		if uint64(c.weight) >= maxInt {
			return nil, errWeightOverflow
		}

		weight := int(c.weight) // convert weight to int for internal counter usage
		if (maxInt - runningTotal) <= weight {
			return nil, errWeightOverflow
		}
		runningTotal += weight
		totals[i] = runningTotal
	}

	if runningTotal < 1 {
		return nil, errNoValidChoices
	}

	return &RandomActions{
		actions: weightedActions,
		totals:  totals,
		max:     runningTotal,
		rand:    rand,
	}, nil
}

func (ra *RandomActions) SelectAction() Action {
	r := ra.rand.IntN(ra.max) + 1
	i := searchInts(ra.totals, r)
	return ra.actions[i].action
}

// The standard library sort.SearchInts() just wraps the generic sort.Search()
// function, which takes a function closure to determine truthfulness. However,
// since this function is utilized within a for loop, it cannot currently be
// properly inlined by the compiler, resulting in non-trivial performance
// overhead.
//
// Thus, this is essentially manually inlined version.  In our use case here, it
// results in a significant throughput increase for Pick.
//
// See also github.com/mroth/xsort.
func searchInts(a []int, x int) int {
	// Possible further future optimization for searchInts via SIMD if we want
	// to write some Go assembly code: http://0x80.pl/articles/simd-search.html
	i, j := 0, len(a)
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		if a[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}
