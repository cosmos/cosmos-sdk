package simulation

import (
	"math/rand"
)

const (
	// Minimum time per block
	minTimePerBlock int64 = 10000 / 2

	// Maximum time per block
	maxTimePerBlock int64 = 10000

	// TODO Remove in favor of binary search for invariant violation
	onOperation bool = false
)

// TODO explain transitional matrix usage
var (
	// Currently there are 3 different liveness types,
	// fully online, spotty connection, offline.
	defaultLivenessTransitionMatrix, _ = CreateTransitionMatrix([][]int{
		{90, 20, 1},
		{10, 50, 5},
		{0, 10, 1000},
	})

	// 3 states: rand in range [0, 4*provided blocksize],
	// rand in range [0, 2 * provided blocksize], 0
	defaultBlockSizeTransitionMatrix, _ = CreateTransitionMatrix([][]int{
		{85, 5, 0},
		{15, 92, 1},
		{0, 3, 99},
	})
)

// Simulation parameters
type Params struct {
	PastEvidenceFraction      float64
	NumKeys                   int
	EvidenceFraction          float64
	InitialLivenessWeightings []int
	LivenessTransitionMatrix  TransitionMatrix
	BlockSizeTransitionMatrix TransitionMatrix
}

// getBlockSize returns a block size as determined from the transition matrix.
// It targets making average block size the provided parameter. The three
// states it moves between are:
//  - "over stuffed" blocks with average size of 2 * avgblocksize,
//  - normal sized blocks, hitting avgBlocksize on average,
//  - and empty blocks, with no txs / only txs scheduled from the past.
func getBlockSize(r *rand.Rand, params Params,
	lastBlockSizeState, avgBlockSize int) (state, blocksize int) {
	// TODO: Make default blocksize transition matrix actually make the average
	// blocksize equal to avgBlockSize.
	state = params.BlockSizeTransitionMatrix.NextState(r, lastBlockSizeState)
	if state == 0 {
		blocksize = r.Intn(avgBlockSize * 4)
	} else if state == 1 {
		blocksize = r.Intn(avgBlockSize * 2)
	} else {
		blocksize = 0
	}
	return state, blocksize
}

// Return default simulation parameters
func DefaultParams() Params {
	return Params{
		PastEvidenceFraction:      0.5,
		NumKeys:                   250,
		EvidenceFraction:          0.5,
		InitialLivenessWeightings: []int{40, 5, 5},
		LivenessTransitionMatrix:  defaultLivenessTransitionMatrix,
		BlockSizeTransitionMatrix: defaultBlockSizeTransitionMatrix,
	}
}

// Return random simulation parameters
func RandomParams(r *rand.Rand) Params {
	return Params{
		PastEvidenceFraction:      r.Float64(),
		NumKeys:                   r.Intn(250),
		EvidenceFraction:          r.Float64(),
		InitialLivenessWeightings: []int{r.Intn(80), r.Intn(10), r.Intn(10)},
		LivenessTransitionMatrix:  defaultLivenessTransitionMatrix,
		BlockSizeTransitionMatrix: defaultBlockSizeTransitionMatrix,
	}
}
