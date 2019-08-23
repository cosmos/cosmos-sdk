package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	// Minimum time per block
	minTimePerBlock int64 = 10000 / 2

	// Maximum time per block
	maxTimePerBlock int64 = 10000
)

// TODO: explain transitional matrix usage
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

// AppParams defines a flat JSON of key/values for all possible configurable
// simulation parameters. It might contain: operation weights, simulation parameters
// and flattened module state parameters (i.e not stored under it's respective module name).
type AppParams map[string]json.RawMessage

// ParamSimulator creates a parameter value from a source of random number
type ParamSimulator func(r *rand.Rand)

// GetOrGenerate attempts to get a given parameter by key from the AppParams
// object. If it exists, it'll be decoded and returned. Otherwise, the provided
// ParamSimulator is used to generate a random value.
func (sp AppParams) GetOrGenerate(cdc *codec.Codec, key string, ptr interface{}, r *rand.Rand, ps ParamSimulator) {
	if v, ok := sp[key]; ok && v != nil {
		cdc.MustUnmarshalJSON(v, ptr)
		return
	}

	ps(r)
}

// Params define the parameters necessary for running the simulations
type Params struct {
	PastEvidenceFraction      float64
	NumKeys                   int
	EvidenceFraction          float64
	InitialLivenessWeightings []int
	LivenessTransitionMatrix  TransitionMatrix
	BlockSizeTransitionMatrix TransitionMatrix
}

// RandomParams for simulation
func RandomParams(r *rand.Rand) Params {
	return Params{
		PastEvidenceFraction:      r.Float64(),
		NumKeys:                   RandIntBetween(r, 2, 250),
		EvidenceFraction:          r.Float64(),
		InitialLivenessWeightings: []int{RandIntBetween(r, 1, 80), r.Intn(10), r.Intn(10)},
		LivenessTransitionMatrix:  defaultLivenessTransitionMatrix,
		BlockSizeTransitionMatrix: defaultBlockSizeTransitionMatrix,
	}
}

//-----------------------------------------------------------------------------
// Param change proposals

// SimValFn function to generate the randomized parameter change value
type SimValFn func(r *rand.Rand) string

// ParamChange defines the object used for simulating parameter change proposals
type ParamChange struct {
	Subspace string
	Key      string
	Subkey   string
	SimValue SimValFn
}

// NewSimParamChange creates a new ParamChange instance
func NewSimParamChange(subspace, key, subkey string, simVal SimValFn) ParamChange {
	return ParamChange{
		Subspace: subspace,
		Key:      key,
		Subkey:   subkey,
		SimValue: simVal,
	}
}

// ComposedKey creates a new composed key for the param change proposal
func (spc ParamChange) ComposedKey() string {
	return fmt.Sprintf("%s/%s/%s", spc.Subspace, spc.Key, spc.Subkey)
}
