package simulation

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

// ParamSimulator creates a parameter value from a source of random number
type ParamSimulator func(r *rand.Rand)

// ContentSimulatorFn defines a function type alias for generating random proposal
// content.
type ContentSimulatorFn func(r *rand.Rand, ctx sdk.Context, accs []module.Account) govtypes.Content

// Params define the parameters necessary for running the simulations
type Params struct {
	PastEvidenceFraction      float64
	NumKeys                   int
	EvidenceFraction          float64
	InitialLivenessWeightings []int
	LivenessTransitionMatrix  TransitionMatrix
	BlockSizeTransitionMatrix TransitionMatrix
}

// RandomParams returns random simulation parameters
func RandomParams(r *rand.Rand) Params {
	return Params{
		PastEvidenceFraction:      r.Float64(),
		NumKeys:                   module.RandIntBetween(r, 2, 2500), // number of accounts created for the simulation
		EvidenceFraction:          r.Float64(),
		InitialLivenessWeightings: []int{module.RandIntBetween(r, 1, 80), r.Intn(10), r.Intn(10)},
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
	subspace string
	key      string
	simValue module.SimValFn
}

func (spc ParamChange) Subspace() string {
	return spc.subspace
}

func (spc ParamChange) Key() string {
	return spc.key
}

func (spc ParamChange) SimValue() module.SimValFn {
	return spc.simValue
}

// NewSimParamChange creates a new ParamChange instance
func NewSimParamChange(subspace, key string, simVal module.SimValFn) module.ParamChange {
	return ParamChange{
		subspace: subspace,
		key:      key,
		simValue: simVal,
	}
}

// ComposedKey creates a new composed key for the param change proposal
func (spc ParamChange) ComposedKey() string {
	return fmt.Sprintf("%s/%s", spc.Subspace, spc.Key)
}

//-----------------------------------------------------------------------------
// Proposal Contents

// WeightedProposalContent defines a common struct for proposal contents defined by
// external modules (i.e outside gov)
type WeightedProposalContent struct {
	appParamsKey       string                    // key used to retrieve the value of the weight from the simulation application params
	defaultWeight      int                       // default weight
	contentSimulatorFn module.ContentSimulatorFn // content simulator function
}

func NewWeightedProposalContent(appParamsKey string, defaultWeight int, contentSimulatorFn module.ContentSimulatorFn) *WeightedProposalContent {
	return &WeightedProposalContent{appParamsKey: appParamsKey, defaultWeight: defaultWeight, contentSimulatorFn: contentSimulatorFn}
}

func (w WeightedProposalContent) AppParamsKey() string {
	return w.appParamsKey
}

func (w WeightedProposalContent) DefaultWeight() int {
	return w.defaultWeight
}

func (w WeightedProposalContent) ContentSimulatorFn() module.ContentSimulatorFn {
	return w.contentSimulatorFn
}
