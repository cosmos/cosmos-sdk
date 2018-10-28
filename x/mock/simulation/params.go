package simulation

import (
	"fmt"
	"math/rand"
)

const (
	// Minimum time per block
	minTimePerBlock int64 = 1000 / 2

	// Maximum time per block
	maxTimePerBlock int64 = 1000

	// TODO Remove in favor of binary search for invariant violation
	onOperation bool = false
)

var (
	// Currently there are 3 different liveness types, fully online, spotty connection, offline.
	livenessTransitionMatrix, _ = CreateTransitionMatrix([][]int{
		{90, 20, 1},
		{10, 50, 5},
		{0, 10, 1000},
	})
)

// Simulation parameters
type Params struct {
	PastEvidenceFraction      float64
	NumKeys                   int
	EvidenceFraction          float64
	InitialLivenessWeightings []int
}

// Return default simulation parameters
func DefaultParams() Params {
	return Params{
		PastEvidenceFraction:      0.5,
		NumKeys:                   250,
		EvidenceFraction:          0.5,
		InitialLivenessWeightings: []int{40, 5, 5},
	}
}

// Return random simulation parameters
func RandomParams(r *rand.Rand) Params {
	return Params{
		PastEvidenceFraction:      r.Float64(),
		NumKeys:                   r.Intn(250),
		EvidenceFraction:          r.Float64(),
		InitialLivenessWeightings: []int{r.Intn(80), r.Intn(10), r.Intn(10)},
	}
}

func (params Params) String() string {
	return fmt.Sprintf("{pastEvidenceFraction: %v, numKeys: %v, evidenceFraction: %v, initialLivenessWeightings: %v}",
		params.PastEvidenceFraction, params.NumKeys, params.EvidenceFraction, params.InitialLivenessWeightings)
}
