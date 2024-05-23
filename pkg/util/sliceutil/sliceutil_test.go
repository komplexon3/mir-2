package sliceutil_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

type containsAllArgs[T comparable] struct {
	Item1 []T
	Item2 []T
}
type containsAllTestCase[T comparable] struct {
	name string
	args containsAllArgs[T]
	want bool
}

func TestContainsAll1(t *testing.T) {
	tcString := []containsAllTestCase[string]{
		{
			name: "n1",
			args: containsAllArgs[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"a", "b"},
			},
			want: true,
		},
		{
			name: "n2",
			args: containsAllArgs[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"b"},
			},
			want: true,
		},
		{
			name: "n3",
			args: containsAllArgs[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"b", "d"},
			},
			want: false,
		},
		{
			name: "n4",
			args: containsAllArgs[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"c", "d"},
			},
			want: false,
		},
	}

	for _, tc := range tcString {
		t.Run(tc.name, func(t *testing.T) {
			actual := sliceutil.ContainsAll(tc.args.Item1, tc.args.Item2)
			assert.Equal(t, tc.want, actual)
		})
	}

	tcInt := []containsAllTestCase[int]{
		{
			name: "testInt-1",
			args: containsAllArgs[int]{
				Item1: []int{1, 2},
				Item2: []int{2},
			},
			want: true,
		},
		{
			name: "testInt-2",
			args: containsAllArgs[int]{
				Item1: []int{1, 2},
				Item2: []int{2, 1},
			},
			want: true,
		},
		{
			name: "testInt-3",
			args: containsAllArgs[int]{
				Item1: []int{1, 2},
				Item2: []int{3, 4},
			},
			want: false,
		},
		{
			name: "testInt-4",
			args: containsAllArgs[int]{
				Item1: []int{1, 2},
				Item2: []int{1, 3},
			},
			want: false,
		},
	}
	for _, tc := range tcInt {
		t.Run(tc.name, func(t *testing.T) {
			actual := sliceutil.ContainsAll(tc.args.Item1, tc.args.Item2)
			assert.Equal(t, tc.want, actual)
		})
	}
}

type vector [3]int

type topologicalSortTestCase[T comparable] struct {
	name     string
	data     []vector
	accepted [][]vector
}

func (v vector) less(other vector) bool {
	for i := range 3 {
		if v[i] > other[i] {
			return false
		}
	}

	return true
}

func TestTopologicalSort(t *testing.T) {
	tcString := []topologicalSortTestCase[string]{
		{
			name:     "empty",
			data:     []vector{},
			accepted: [][]vector{{}},
		},
		{
			name:     "one element",
			data:     []vector{{1, 2, 3}},
			accepted: [][]vector{{{1, 2, 3}}},
		},
		{
			name: "3 in a row",
			data: []vector{
				{3, 0, 0}, {1, 0, 0}, {2, 0, 0},
			},
			accepted: [][]vector{
        {{1, 0, 0}, {2, 0, 0}, {3, 0, 0}},
      },
		},
		{
			name:     "diamond",
			data:     []vector{
        {2, 1, 1}, {1, 1, 0}, {1, 0, 0}, {1, 0, 1},
      },
			accepted: [][]vector{
        {{1, 0, 0}, {1, 1, 0}, {1, 0, 1}, {2, 1, 1}},
        {{1, 0, 0}, {1, 0, 1}, {1, 1, 0}, {2, 1, 1}},
      },
		},
		{
			name: "loose start",
			data: []vector{
				{1, 0, 1}, {1, 0, 0}, {3, 1, 1}, {0, 1, 0},
			},
			accepted: [][]vector{
				{{1, 0, 0}, {1, 0, 1}, {0, 1, 0}, {3, 1, 1}},
				{{0, 1, 0}, {1, 0, 0}, {1, 0, 1}, {3, 1, 1}},
			},
		},
		{
			name: "loose ends",
			data: []vector{{1, 1, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 0}},
			accepted: [][]vector{
				{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {1, 1, 0}},
				{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {1, 1, 0}},
				{{0, 0, 0}, {0, 1, 0}, {1, 1, 0}, {0, 1, 1}},
			},
		},
	}

	for _, tc := range tcString {
		t.Run(tc.name, func(tt *testing.T) {
			result := sliceutil.TopologicalSortFunc(tc.data, func(a, b vector) bool { return a.less(b) })
			assert.True(tt, slices.ContainsFunc(tc.accepted, func(option []vector) bool {
				if len(option) != len(result) {
					return false
				}
				for i := range len(option) {
					for j := range 3 {
						if option[i][j] != result[i][j] {
							return false
						}
					}
				}

				return true
			}))
		})
	}
}
