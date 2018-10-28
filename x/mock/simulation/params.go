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
	initialLivenessWeightings   = []int{40, 5, 5}
	livenessTransitionMatrix, _ = CreateTransitionMatrix([][]int{
		{90, 20, 1},
		{10, 50, 5},
		{0, 10, 1000},
	})
)

type SimulationParams struct {
	PastEvidenceFraction float64
	NumKeys              int
	EvidenceFraction     float64
}

func DefaultSimulationParams() SimulationParams {
	return SimulationParams{
		PastEvidenceFraction: 0.5,
		NumKeys:              250,
		EvidenceFraction:     0.5,
	}
}

func RandomSimulationParams(r *rand.Rand) SimulationParams {
	return SimulationParams{
		PastEvidenceFraction: r.Float64(),
		NumKeys:              r.Intn(DefaultSimulationParams().NumKeys / 2),
		EvidenceFraction:     r.Float64(),
	}
}

func (params SimulationParams) String() string {
	return fmt.Sprintf("{pastEvidenceFraction: %v, numKeys: %v, evidenceFraction: %v}",
		params.PastEvidenceFraction, params.NumKeys, params.EvidenceFraction)
}
