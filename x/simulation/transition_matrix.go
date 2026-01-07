package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// TransitionMatrix is _almost_ a left stochastic matrix.  It is technically
// not one due to not normalizing the column values.  In the future, if we want
// to find the steady state distribution, it will be quite easy to normalize
// these values to get a stochastic matrix.  Floats aren't currently used as
// the default due to non-determinism across architectures
type TransitionMatrix struct {
	weights [][]int
	// total in each column
	totals []int
	n      int
}

// CreateTransitionMatrix creates a transition matrix from the provided weights.
//
// Example:
//
//	weights := [][]int{
//	  // From state 0 to states 0,1,2
//	  {90, 10, 0},
//	  // From state 1 to states 0,1,2
//	  {20, 70, 10},
//	  // From state 2 to states 0,1,2
//	  {5,  15, 80},
//	}
//	tm, err := CreateTransitionMatrix(weights)
//	if err != nil {
//	  // handle error
//	}
//
//	// NextState picks the next state from current state i.
//	// r should be a deterministic *rand.Rand when used in simulations.
//	// next := tm.NextState(r, i)
func CreateTransitionMatrix(weights [][]int) (simulation.TransitionMatrix, error) {
	n := len(weights)
	for i := range n {
		if len(weights[i]) != n {
			return TransitionMatrix{},
				fmt.Errorf("transition matrix: non-square matrix provided, error on row %d", i)
		}
	}

	totals := make([]int, n)

	for row := range n {
		for col := range n {
			totals[col] += weights[row][col]
		}
	}

	return TransitionMatrix{weights, totals, n}, nil
}

// NextState returns the next state randomly chosen using r, and the weightings
// provided in the transition matrix.
func (t TransitionMatrix) NextState(r *rand.Rand, i int) int {
	randNum := r.Intn(t.totals[i])
	for row := range t.n {
		if randNum < t.weights[row][i] {
			return row
		}

		randNum -= t.weights[row][i]
	}
	// This line should never get executed
	return -1
}

// GetMemberOfInitialState takes an initial array of weights, of size n.
// It returns a weighted random number in [0,n).
func GetMemberOfInitialState(r *rand.Rand, weights []int) int {
	n := len(weights)
	total := 0

	for i := range n {
		total += weights[i]
	}

	randNum := r.Intn(total)

	for state := range n {
		if randNum < weights[state] {
			return state
		}

		randNum -= weights[state]
	}
	// This line should never get executed
	return -1
}
